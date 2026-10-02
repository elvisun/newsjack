from __future__ import annotations

import base64
import json
import os
import subprocess
import sys
import tempfile
import unittest
from collections import Counter
from pathlib import Path
from unittest import mock


EVAL = Path(__file__).resolve().parents[1]
SCRIPTS = EVAL / "scripts"
FIXTURES = EVAL / "fixtures"
sys.path.insert(0, str(SCRIPTS))

import pilotlib  # noqa: E402
import model_runner  # noqa: E402
import collect as collector  # noqa: E402
import extract_features  # noqa: E402
import model_features  # noqa: E402
import generate as generation  # noqa: E402
import analyze as analyze_script  # noqa: E402


def request(platform: str = "google") -> dict:
    return {
        "request_tag": f"calibration-0001-{platform}",
        "paired_unit_id": "calibration-0001",
        "phase": "calibration",
        "platform": platform,
        "query": "mock query",
        "topic_family": "workplace_technology",
        "intent": "explanatory",
    }


class ParsingTests(unittest.TestCase):
    def load(self, name: str) -> dict:
        return json.loads((FIXTURES / "api" / name).read_text(encoding="utf-8"))

    def test_google_overview_and_organic(self):
        record = pilotlib.parse_google_response(self.load("google.json"), request())
        self.assertEqual(record["outcome"], "has_citations")
        self.assertEqual(record["citations"][0]["canonical_url"], "https://evidence.example/report")
        labels = {row["canonical_url"]: row["label"] for row in record["organic_results"]}
        self.assertEqual(labels["https://evidence.example/report"], "inline_cited_not_named")
        self.assertEqual(labels["https://control.example/page"], "organic_not_ai_cited")

    def test_google_handed_and_queued_states_are_retryable(self):
        for code in (40601, 40602):
            raw = {"status_code": 20000, "tasks": [{"status_code": code}]}
            with self.subTest(code=code), self.assertRaises(pilotlib.RetryableResponse):
                pilotlib.parse_google_response(raw, request())

    def test_google_task_id_is_saved_before_polling_and_reused(self):
        posted = {"status_code": 20000, "tasks": [{"status_code": 20100, "id": "task-1"}]}
        handed = {"status_code": 20000, "tasks": [{"status_code": 40601}]}
        completed = self.load("google.json")
        with tempfile.TemporaryDirectory() as directory:
            pending_path = Path(directory) / "pending.json"
            with mock.patch.object(collector, "http_json", side_effect=[posted, handed, completed]) as http, \
                    mock.patch.object(collector.time, "sleep"):
                raw = collector.google_live(request(), ("x", "y"), pending_path)
            self.assertEqual(raw, completed)
            self.assertEqual(json.loads(pending_path.read_text())["task_id"], "task-1")
            self.assertEqual([call.args[0] for call in http.call_args_list], ["POST", "GET", "GET"])

            with mock.patch.object(collector, "http_json", return_value=completed) as http:
                raw = collector.google_live(request(), ("x", "y"), pending_path)
            self.assertEqual(raw, completed)
            self.assertEqual(http.call_count, 1)
            self.assertEqual(http.call_args.args[0], "GET")

    def test_google_post_error_retains_raw_cost(self):
        failed = {"status_code": 20000, "tasks": [{"status_code": 40501, "cost": 0.0012}]}
        with tempfile.TemporaryDirectory() as directory, \
                mock.patch.object(collector, "http_json", return_value=failed):
            with self.assertRaises(collector.GooglePostError) as caught:
                collector.ensure_google_task(request(), ("x", "y"), Path(directory) / "pending.json")
        self.assertEqual(collector.response_cost(caught.exception.raw), 0.0012)

    def test_google_no_overview_is_retained(self):
        record = pilotlib.parse_google_response(self.load("google-no-overview.json"), request())
        self.assertEqual(record["outcome"], "no_ai_answer")
        self.assertEqual(len(record["organic_results"]), 1)

    def test_google_overview_without_citation_is_retained(self):
        record = pilotlib.parse_google_response(self.load("google-no-citations.json"), request())
        self.assertEqual(record["outcome"], "ai_answer_no_citations")

    def test_chatgpt_citation_and_no_citation(self):
        cited = pilotlib.parse_chatgpt_response(self.load("chatgpt.json"), request("chatgpt"))
        self.assertEqual(cited["outcome"], "has_citations")
        self.assertEqual(cited["citations"][0]["label"], "answer_mentioned_and_cited")
        empty = pilotlib.parse_chatgpt_response(self.load("chatgpt-no-citations.json"), request("chatgpt"))
        self.assertEqual(empty["outcome"], "ai_answer_no_citations")

    def test_retry_and_error_are_distinct(self):
        with self.assertRaises(pilotlib.RetryableResponse):
            pilotlib.parse_google_response(self.load("google-retry.json"), request())
        with self.assertRaises(pilotlib.PilotError):
            pilotlib.parse_google_response(self.load("google-error.json"), request())
        with self.assertRaises(pilotlib.RetryableResponse):
            pilotlib.parse_chatgpt_response(self.load("chatgpt-retry.json"), request("chatgpt"))
        with self.assertRaises(pilotlib.PilotError):
            pilotlib.parse_chatgpt_response(self.load("chatgpt-error.json"), request("chatgpt"))


class CanonicalizationTests(unittest.TestCase):
    def test_tracking_fragment_case_and_ports(self):
        left = pilotlib.canonicalize_url("HTTPS://Example.COM:443/a//b/?utm_source=x&b=2&a=1#frag")
        right = pilotlib.canonicalize_url("https://example.com/a/b?a=1&b=2")
        self.assertEqual(left, right)

    def test_dedupe_keeps_best_position(self):
        rows = [
            {"canonical_url": "https://example.com/a", "organic_position": 5},
            {"canonical_url": "https://example.com/a", "organic_position": 2},
        ]
        self.assertEqual(pilotlib.dedupe_records(rows, keep_position=True)[0]["organic_position"], 2)

    def test_cross_query_dedupe_preserves_events(self):
        observations = [
            {"request_tag":"q1","paired_unit_id":"p1","platform":"google","citations":[{"url":"https://Example.com/a?utm_source=x","label":"inline_cited_not_named"}],"organic_results":[]},
            {"request_tag":"q2","paired_unit_id":"p2","platform":"chatgpt","citations":[{"url":"https://example.com/a#part","label":"answer_mentioned_and_cited"}],"organic_results":[]},
        ]
        pages = pilotlib.build_page_index(observations)
        self.assertEqual(len(pages), 1)
        self.assertEqual(len(pages[0]["events"]), 2)


class BudgetTests(unittest.TestCase):
    def event(self, tag: str, phase: str, platform: str, cost: float) -> dict:
        return {"request_tag": tag, "phase": phase, "platform": platform, "terminal": True, "status": "success", "cost_usd": cost}

    def test_ledger_chain_arithmetic_and_tamper(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "ledger.jsonl"
            pilotlib.append_ledger(path, self.event("c-g", "calibration", "google", 0.0012))
            pilotlib.append_ledger(path, self.event("c-c", "calibration", "chatgpt", 0.012))
            total, phases = pilotlib.verify_ledger(path)
            self.assertEqual(total, 0.0132)
            self.assertEqual(phases["calibration"], 0.0132)
            text = path.read_text(encoding="utf-8").replace('"cost_usd":0.012', '"cost_usd":0.011', 1)
            path.write_text(text, encoding="utf-8")
            with self.assertRaises(pilotlib.PilotError):
                pilotlib.verify_ledger(path)

    def test_prebatch_rejects_phase_and_combined_batch(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "ledger.jsonl"
            with self.assertRaises(pilotlib.BudgetError):
                pilotlib.check_budget(path, "main", "chatgpt", 121)
            with self.assertRaises(pilotlib.BudgetError):
                pilotlib.check_batch_budget(path, "calibration", {"google": 100, "chatgpt": 25})

    def test_every_phase_rejects_an_overallocation_before_send(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "ledger.jsonl"
            for phase, limit in pilotlib.PHASE_LIMITS.items():
                count = int(limit / pilotlib.PRECALIBRATION_CEILINGS["google"]) + 1
                with self.assertRaises(pilotlib.BudgetError, msg=phase):
                    pilotlib.check_batch_budget(path, phase, {"google": count})
            self.assertFalse(path.exists())

    def test_partially_spent_fresh_phase_can_resume_within_remaining_caps(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "ledger.jsonl"
            pilotlib.append_ledger(path, self.event("c-g", "calibration", "google", 0.0012))
            pilotlib.append_ledger(path, self.event("f-c", "fresh", "chatgpt", 1.5))
            projection = pilotlib.check_budget(path, "fresh", "google", 49)
            self.assertEqual(projection.phase_actual, 1.5)
            self.assertLess(projection.phase_actual + projection.projected, pilotlib.PHASE_LIMITS["fresh"])

    def test_calibrated_p95_allows_affordable_balanced_main(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "ledger.jsonl"
            pilotlib.append_ledger(path, self.event("c-g", "calibration", "google", 0.0012))
            pilotlib.append_ledger(path, self.event("c-c", "calibration", "chatgpt", 0.012))
            projections = pilotlib.check_batch_budget(path, "main", {"google": 180, "chatgpt": 180})
            self.assertLess(sum(item.projected for item in projections.values()), 4.0)

    def test_unknown_cost_stops_next_batch(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "ledger.jsonl"
            pilotlib.append_ledger(path, {"request_tag":"x","phase":"calibration","platform":"chatgpt","terminal":True,"status":"unknown_cost","cost_usd":None})
            with self.assertRaises(pilotlib.BudgetError):
                pilotlib.check_budget(path, "calibration", "google", 1)

    def test_unknown_cost_can_be_reconciled_append_only(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "ledger.jsonl"
            pilotlib.append_ledger(path, {
                "request_tag":"lost", "phase":"calibration", "platform":"google",
                "terminal":True, "status":"unknown_cost", "cost_usd":None,
            })
            pilotlib.append_ledger(path, {
                "reconciles_request_tag":"lost", "phase":"calibration", "platform":"google",
                "terminal":False, "status":"error", "cost_usd":0.0012,
                "cost_source":"appendix_user_data_daily_total",
            })
            total, phases = pilotlib.verify_ledger(path)
            self.assertEqual(total, 0.0012)
            self.assertEqual(phases["calibration"], 0.0012)
            self.assertEqual(pilotlib.terminal_events(path)["lost"]["status"], "error")
            self.assertEqual(pilotlib.unresolved_unknown_tags(path), set())
            pilotlib.check_budget(path, "calibration", "google", 1)
            with self.assertRaises(pilotlib.PilotError):
                pilotlib.append_ledger(path, {
                    "reconciles_request_tag":"lost", "phase":"calibration", "platform":"google",
                    "terminal":False, "status":"error", "cost_usd":0.0012,
                })

    def test_total_cap_is_independent(self):
        with tempfile.TemporaryDirectory() as directory, mock.patch.object(pilotlib, "TOTAL_LIMIT", 0.01):
            path = Path(directory) / "ledger.jsonl"
            pilotlib.append_ledger(path, self.event("c-g", "calibration", "google", 0.006))
            with self.assertRaises(pilotlib.PilotError):
                pilotlib.append_ledger(path, self.event("c-c", "calibration", "chatgpt", 0.006))
            self.assertEqual(len(pilotlib.read_jsonl(path)), 1)

    def test_error_response_cost_is_recoverable(self):
        raw = json.loads((FIXTURES / "api" / "google-error.json").read_text(encoding="utf-8"))
        self.assertIsInstance(collector.response_cost(raw), float)

    def test_model_resolution_requires_web_search_and_non_reasoning(self):
        raw = {"tasks":[{"result":[
            {"model_name":"gpt-4.1", "reasoning":False, "web_search_supported":True},
            {"model_name":"o4-mini", "reasoning":True, "web_search_supported":True},
            {"model_name":"gpt-5.2", "reasoning":False, "web_search_supported":True},
            {"model_name":"gpt-9", "reasoning":False, "web_search_supported":False}
        ]}]}
        with mock.patch.object(collector, "http_json", return_value=raw):
            self.assertEqual(collector.choose_chatgpt_model(("x", "y"), None), "gpt-5.2")

    def test_credentials_load_base64_api_key_from_env_file(self):
        token = base64.b64encode(b"api-login:api-password").decode()
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / ".env.local"
            path.write_text(f'DATAFORSEO_API_KEY="{token}"\n', encoding="utf-8")
            with mock.patch.dict(os.environ, {}, clear=True):
                self.assertEqual(collector.credentials(path), ("api-login", "api-password"))

    def test_credentials_reject_invalid_api_key_without_echoing_it(self):
        secret = "not-a-basic-credential"
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / ".env.local"
            path.write_text(f"DATAFORSEO_API_KEY={secret}\n", encoding="utf-8")
            with mock.patch.dict(os.environ, {}, clear=True), self.assertRaises(pilotlib.PilotError) as caught:
                collector.credentials(path)
        self.assertNotIn(secret, str(caught.exception))

    def test_provider_cap_is_verified_from_free_user_data(self):
        raw = {
            "status_code": 20000, "cost": 0,
            "tasks": [{"status_code": 20000, "result": [{"money": {"limits": {"day": {"total": 10}}}}]}],
        }
        with mock.patch.object(collector, "http_json", return_value=raw):
            self.assertEqual(collector.verify_provider_cap(("x", "y")), 10.0)

    def test_provider_cap_rejects_high_missing_and_ambiguous_limits(self):
        for limit in (1000, 0, None):
            raw = {
                "status_code": 20000, "cost": 0,
                "tasks": [{"status_code": 20000, "result": [{"money": {"limits": {"day": {"total": limit}}}}]}],
            }
            with self.subTest(limit=limit), mock.patch.object(collector, "http_json", return_value=raw), self.assertRaises(pilotlib.BudgetError):
                collector.verify_provider_cap(("x", "y"))

    def test_free_credit_only_mode_retains_reported_limit(self):
        raw = {
            "status_code": 20000, "cost": 0,
            "tasks": [{"status_code": 20000, "result": [{"money": {"limits": {"day": {"total": 1000}}}}]}],
        }
        with mock.patch.object(collector, "http_json", return_value=raw):
            mode, limit = collector.provider_safety(("x", "y"), capped=False, free_credit_only=True)
        self.assertEqual(mode, "user_confirmed_free_credit_only_no_payment_method")
        self.assertEqual(limit, 1000.0)

    def test_provider_safety_requires_exactly_one_mode(self):
        for value in (False, True):
            with self.subTest(value=value), self.assertRaises(pilotlib.PilotError):
                collector.provider_safety(("x", "y"), capped=value, free_credit_only=value)


class FactTests(unittest.TestCase):
    def test_exact_numbers_dates_quotes_and_urls_survive(self):
        original = 'On May 14, 2026, Dr. Mina Rao said, “The 6.2% result applies only to 840 adults.” See https://example.com/report.'
        same = 'Evidence note: On May 14, 2026, Dr. Mina Rao said, “The 6.2% result applies only to 840 adults.” See https://example.com/report.'
        changed = 'On May 15, 2026, Dr. Mina Rao said the result applies to 900 adults.'
        self.assertTrue(pilotlib.compare_fact_ledgers(original, same)["pass"])
        self.assertFalse(pilotlib.compare_fact_ledgers(original, changed)["pass"])

    def test_audit_wrapper_checks_only_revision_section(self):
        case = {
            "draft": "The fund's fee is 0.25% for 840 accounts.",
            "protected_items": ["0.25%", "840 accounts"],
        }
        wrapped = """## Verdict

1. Keep the fund's scope clear.
2. Do not invent a quote like \"best in class\".

## Revision

# A factual headline

The fund's fee is 0.25% for 840 accounts.

## Fact-preservation note

Checked.
"""
        result = generation.fact_guard_result(case, wrapped)
        self.assertTrue(result["pass"])
        self.assertEqual(result["compared_scope"], "revision_section")

    def test_revision_scope_rejects_changed_number(self):
        case = {"draft":"The sample included 840 adults.", "protected_items":["840 adults"]}
        result = generation.fact_guard_result(case, "## Revision\n\nThe sample included 900 adults.\n")
        self.assertFalse(result["pass"])

    def test_protected_match_is_case_insensitive_and_allows_two_intervening_words(self):
        self.assertTrue(generation.protected_item_preserved(
            "not the fund's total trading cost",
            "The ratio does not represent the fund's total trading cost.",
        ))
        self.assertTrue(generation.protected_item_preserved(
            "future returns are not guaranteed",
            "Future returns are not guaranteed.",
        ))
        self.assertFalse(generation.protected_item_preserved(
            "future returns are not guaranteed",
            "Future returns are guaranteed.",
        ))
        self.assertFalse(generation.protected_item_preserved(
            "not the fund's total trading cost",
            "The expense ratio describes one annual fee.",
        ))

    def test_protected_match_allows_ing_inflection_without_losing_polarity(self):
        self.assertTrue(generation.protected_item_preserved(
            "transfer heat rather than generate it",
            "Heat pumps work by transferring heat rather than generating it.",
        ))
        self.assertFalse(generation.protected_item_preserved(
            "transfer heat rather than generate it",
            "Heat pumps work by generating heat.",
        ))

    def test_protected_match_normalizes_other_language_scope_not_polarity(self):
        protected = "did not evaluate children or calls in other languages"
        self.assertTrue(generation.protected_item_preserved(
            protected,
            "The research did not evaluate children or calls in languages other than English.",
        ))
        self.assertFalse(generation.protected_item_preserved(
            protected,
            "The research did evaluate children and calls in languages other than English.",
        ))

    def test_possessive_apostrophes_are_not_quoted_spans(self):
        ledger = pilotlib.fact_ledger("The fund's fee differs from another fund's fee.")
        self.assertEqual(ledger["quotes"], [])

    def test_quote_typography_and_trailing_date_comma_are_normalized(self):
        original = 'On May 14, 2026, Mina said, “The result is limited.”'
        rewrite = 'On May 14, 2026 Mina said, "The result is limited."'
        self.assertTrue(pilotlib.compare_fact_ledgers(original, rewrite)["pass"])

    def test_quote_attribution_punctuation_is_normalized_but_words_are_not(self):
        original = 'Mina said, “The result applies only to adults.”'
        reordered = '“The result applies only to adults,” Mina said.'
        changed = '“The result applies to everyone,” Mina said.'
        self.assertTrue(pilotlib.compare_fact_ledgers(original, reordered)["pass"])
        self.assertFalse(pilotlib.compare_fact_ledgers(original, changed)["pass"])

    def test_url_path_numbers_are_not_prose_facts_and_only_supplied_url_is_allowed(self):
        url = "https://example.test/releases/2026/05/result"
        ledger = pilotlib.fact_ledger(f"Source: {url}")
        self.assertEqual(ledger["numbers"], [])
        self.assertEqual(ledger["dates"], [])
        self.assertEqual(ledger["urls"], [url])
        case = {
            "draft": "The result is limited.",
            "constructed": False,
            "source_url": url,
            "protected_items": ["result is limited"],
        }
        allowed = generation.fact_guard_result(case, f"The result is limited. Source: {url}")
        rejected = generation.fact_guard_result(case, "The result is limited. Source: https://other.test/result")
        self.assertTrue(allowed["pass"])
        self.assertEqual(allowed["allowed_provenance_urls"], [url])
        self.assertFalse(rejected["pass"])

    def test_list_markers_and_one_word_scare_quotes_are_not_facts(self):
        ledger = pilotlib.fact_ledger('1. Check the “flexible” term.\n2. Read the policy.')
        self.assertEqual(ledger["numbers"], [])
        self.assertEqual(ledger["quotes"], [])
        self.assertEqual(
            pilotlib.fact_ledger('Mina said, “The result is limited.”')["quotes"],
            ["The result is limited"],
        )

    def test_numeric_citation_markers_are_not_facts_but_prose_numbers_are(self):
        ledger = pilotlib.fact_ledger("Three sources [1][2][3] report a sample of 840 adults.")
        self.assertEqual(ledger["numbers"], ["840"])

    def test_baseline_fact_failure_is_retained_but_candidate_failure_is_hard(self):
        result = {"pass":False, "ledger_diff":{"pass":False}, "protected_missing":["6.2%"]}
        with tempfile.TemporaryDirectory() as directory:
            output = Path(directory) / "answer.json"
            generation.enforce_fact_guard("bare", result, output)
            self.assertTrue(output.with_suffix(".validation.json").exists())
            with self.assertRaises(pilotlib.PilotError):
                generation.enforce_fact_guard("skill", result, output)


class FeatureModelTests(unittest.TestCase):
    def test_visible_html_features_exclude_script_and_keep_method_signals(self):
        html = (FIXTURES / "page.html").read_text(encoding="utf-8")
        values = extract_features.feature_vector(
            html=html, query="What did Northline measure?", intent="freshness",
            url="https://example.test/press-release/pilot", observed=extract_features.date(2026, 7, 21),
        )
        self.assertEqual(values["document_type"], "press_release")
        self.assertNotIn("999999", values["allowed_excerpt"])
        self.assertGreater(values["factor_features"]["F1"], 0)
        self.assertGreater(values["factor_features"]["F3"], 0)
        self.assertEqual(values["factor_features"]["F7"], 1.0)

    def test_query_fixed_effect_model_recovers_direction_and_fdr(self):
        rows = []
        for query in range(40):
            for page in range(4):
                cited = page in ({0, 1} if query % 5 else {0, 2})
                signal = 1.0 if page in {0, 1} else 0.0
                features = {f"F{i}": ((query + page + i) % 7) / 6 for i in range(1, 11)}
                features["F2"] = signal
                rows.append({
                    "platform":"google", "request_tag":f"q-{query}", "cited":cited,
                    "publisher_domain":f"d{page % 3}.test", "organic_position":page + 1,
                    "risk_set":"google_organic",
                    "topic_family":f"topic-{query % 2}", "intent":f"intent-{query % 2}",
                    "features":{"factor_features":features, "publication_age_days":30 + page,
                                "content_length":500 + 20 * page, "document_type":["blog", "press_release", "contributed_article"][page % 3]},
                    "exclusion_reason":None,
                })
        result = model_features.analyze(rows)
        self.assertEqual(result["google_adjusted"]["status"], "estimated")
        self.assertGreater(result["google_adjusted"]["effects"]["F2"]["estimate_pp_per_sd"], 0)
        self.assertIn("fdr_q_value", result["google_adjusted"]["effects"]["F2"])
        self.assertEqual(result["chatgpt_adjusted"]["status"], "unestimable")

    def test_repeat_stability_joins_identical_queries_not_generated_ids(self):
        rows = [
            {"request_tag":"main-0001-chatgpt", "paired_unit_id":"main-0001", "platform":"chatgpt", "query":"same query", "citations":[{"canonical_url":"https://a.test/one"}]},
            {"request_tag":"repeats-0001-chatgpt", "paired_unit_id":"repeats-0001", "platform":"chatgpt", "query":"same query", "citations":[{"canonical_url":"https://a.test/one"}, {"canonical_url":"https://b.test/two"}]},
        ]
        result = analyze_script.summarize(rows, [])
        self.assertEqual(result["repeat_stability"]["n"], 1)
        self.assertEqual(result["repeat_stability"]["mean_url_jaccard"], 0.5)

    def test_google_citation_merge_preserves_organic_rank_and_risk_set(self):
        observations = [{
            "request_tag":"q-1", "paired_unit_id":"p-1", "platform":"google",
            "query":"query", "topic_family":"family", "intent":"explanatory",
            "citations":[{"canonical_url":"https://example.test/page", "label":"inline_cited_not_named"}],
            "organic_results":[{"canonical_url":"https://example.test/page", "label":"inline_cited_not_named", "organic_position":3}],
        }]
        rows = extract_features.event_rows(observations)
        self.assertEqual(len(rows), 1)
        self.assertEqual(rows[0]["organic_position"], 3)
        self.assertEqual(rows[0]["risk_set"], "google_organic")
        self.assertTrue(rows[0]["cited"])


class HarnessTests(unittest.TestCase):
    def run_score(self, name: str) -> subprocess.CompletedProcess:
        return subprocess.run(
            [sys.executable, str(EVAL / "harness" / "score.py"), "--input", str(FIXTURES / name)],
            text=True, capture_output=True,
        )

    def test_scorer_separates_good_and_bad(self):
        good = self.run_score("scoring-good.json")
        bad = self.run_score("scoring-bad.json")
        self.assertEqual(good.returncode, 0)
        self.assertEqual(bad.returncode, 0)
        self.assertGreater(float(good.stdout.strip()), float(bad.stdout.strip()) + 40)

    def test_hard_gate_voids_without_details(self):
        result = self.run_score("scoring-void.json")
        self.assertEqual(result.returncode, 3)
        self.assertEqual(result.stdout, "VOID: constraint violation\n")
        self.assertEqual(result.stderr, "")

    def test_each_hard_gate_individually_voids_without_details(self):
        good = json.loads((FIXTURES / "scoring-good.json").read_text(encoding="utf-8"))
        with tempfile.TemporaryDirectory() as directory:
            for gate in list(good["hard_gates"]):
                planted = json.loads(json.dumps(good))
                planted["hard_gates"][gate] = False
                path = Path(directory) / f"{gate}.json"
                path.write_text(json.dumps(planted), encoding="utf-8")
                result = subprocess.run([sys.executable, str(EVAL / "harness" / "score.py"), "--input", str(path)], text=True, capture_output=True)
                self.assertEqual((result.returncode, result.stdout, result.stderr), (3, "VOID: constraint violation\n", ""), gate)

    def test_every_constraint_plant_voids_full_score_without_details(self):
        for plant in ("capacity", "claim", "credential", "ledger", "checksum", "corpus"):
            result = subprocess.run([str(EVAL / "harness" / "score.sh"), "--input", str(FIXTURES / "scoring-good.json"), "--plant", plant], text=True, capture_output=True)
            self.assertEqual((result.returncode, result.stdout, result.stderr), (3, "VOID: constraint violation\n", ""), plant)

    def test_manifest_balance_and_case_balance(self):
        with tempfile.TemporaryDirectory() as directory:
            output = Path(directory) / "main.json"
            subprocess.run([sys.executable, str(SCRIPTS / "generate_manifest.py"), "--phase", "main", "--output", str(output)], check=True)
            manifest = json.loads(output.read_text(encoding="utf-8"))
            self.assertEqual(manifest["pair_count"], 120)
            family = {}
            intent = {}
            for row in manifest["units"]:
                family[row["topic_family"]] = family.get(row["topic_family"], 0) + 1
                intent[row["intent"]] = intent.get(row["intent"], 0) + 1
            self.assertEqual(sorted(family.values()), [30, 30, 30, 30])
            self.assertEqual(sorted(intent.values()), [24, 24, 24, 24, 24])
        cases = json.loads((FIXTURES / "cases.json").read_text(encoding="utf-8"))["cases"]
        self.assertEqual(len(cases), 24)
        for field, expected in (("document_type", 8), ("behavior", 6)):
            counts = {}
            for case in cases:
                counts[case[field]] = counts.get(case[field], 0) + 1
            self.assertTrue(all(value == expected for value in counts.values()))

    def test_repeat_manifest_reuses_main_queries_and_ablation_floors(self):
        with tempfile.TemporaryDirectory() as directory:
            temp = Path(directory)
            main_path = temp / "main.json"
            repeat_path = temp / "repeats.json"
            calibration_path = temp / "calibration.json"
            subprocess.run([sys.executable, str(SCRIPTS / "generate_manifest.py"), "--phase", "main", "--output", str(main_path)], check=True)
            subprocess.run([sys.executable, str(SCRIPTS / "generate_manifest.py"), "--phase", "repeats", "--output", str(repeat_path)], check=True)
            subprocess.run([sys.executable, str(SCRIPTS / "generate_manifest.py"), "--phase", "calibration", "--output", str(calibration_path)], check=True)
            main = json.loads(main_path.read_text(encoding="utf-8"))
            repeats = json.loads(repeat_path.read_text(encoding="utf-8"))
            main_by_id = {row["paired_unit_id"]: row for row in main["units"]}
            self.assertEqual(repeats["pair_count"], 49)
            repeat_family = Counter(row["topic_family"] for row in repeats["units"])
            repeat_intent = Counter(row["intent"] for row in repeats["units"])
            self.assertLessEqual(max(repeat_family.values()) - min(repeat_family.values()), 1)
            self.assertLessEqual(max(repeat_intent.values()) - min(repeat_intent.values()), 1)
            for row in repeats["units"]:
                self.assertEqual(row["query"], main_by_id[row["repeat_of"]]["query"])
            calibration = json.loads(calibration_path.read_text(encoding="utf-8"))
            self.assertEqual(calibration["pair_count"], 6)
            self.assertEqual(len({row["topic_family"] for row in calibration["units"]}), 6)
        cases_by_id = {row["id"]: row for row in json.loads((FIXTURES / "cases.json").read_text(encoding="utf-8"))["cases"]}
        ablations = json.loads((FIXTURES / "ablation-manifest.json").read_text(encoding="utf-8"))["levers"]
        self.assertEqual(set(ablations), {f"L{i}" for i in range(1, 7)})
        for lever, ids in ablations.items():
            self.assertEqual(len(ids), 8, lever)
            self.assertEqual(len(set(ids)), 8, lever)
            self.assertGreaterEqual(len({cases_by_id[item]["document_type"] for item in ids}), 2, lever)
            self.assertGreaterEqual(len({cases_by_id[item]["topic_family"] for item in ids}), 5, lever)
        full_ablation = json.loads((FIXTURES / "ablation-manifest.json").read_text(encoding="utf-8"))
        combined = [cases_by_id[item] for item in full_ablation["combined_skill_cases"]]
        self.assertEqual(len(combined), 12)
        self.assertEqual({kind: sum(row["document_type"] == kind for row in combined) for kind in ("press_release", "blog", "contributed_article")}, {"press_release":4, "blog":4, "contributed_article":4})

    def test_recursive_schema_validation_and_simulation_dry_runs(self):
        schema = json.loads((EVAL / "harness" / "citation-schema.json").read_text(encoding="utf-8"))
        valid = {"answer":"x", "cited_candidate_ids":["SOURCE-1"], "accurately_used_candidate_ids":[], "unsupported_claims":[]}
        model_runner._validate(valid, schema)
        with self.assertRaises(pilotlib.PilotError):
            model_runner._validate({**valid, "cited_candidate_ids":["SOURCE-1", "SOURCE-1"]}, schema)
        codex_schema = model_runner._codex_schema(schema)
        self.assertNotIn("uniqueItems", codex_schema["properties"]["cited_candidate_ids"])
        self.assertTrue(schema["properties"]["cited_candidate_ids"]["uniqueItems"])
        with tempfile.TemporaryDirectory() as directory:
            temp = Path(directory)
            generate = subprocess.run([
                sys.executable, str(SCRIPTS / "generate.py"), "--case-id", "workplace-01",
                "--executor", "codex", "--condition", "lever", "--lever-id", "L1",
                "--output", str(temp / "rewrite.json"), "--dry-run",
            ], text=True, capture_output=True)
            self.assertEqual(generate.returncode, 0, generate.stderr)
            citation = subprocess.run([
                sys.executable, str(SCRIPTS / "citation_simulate.py"), "--case-id", "workplace-01",
                "--lever-id", "L1", "--variant", "original", "--executor", "claude",
                "--generator", "codex", "--output", str(temp / "citation.json"),
                "--metadata", str(temp / "citation-meta.json"), "--dry-run",
            ], text=True, capture_output=True)
            self.assertEqual(citation.returncode, 0, citation.stderr)
        plan = subprocess.run([sys.executable, str(SCRIPTS / "run_simulations.py"), "--dry-run"], text=True, capture_output=True)
        self.assertEqual(plan.returncode, 0, plan.stderr)
        values = json.loads(plan.stdout)
        self.assertEqual(values["planned_invocations"], 436)
        self.assertGreater(values["fresh_plan_buffer"], 0)

    def test_clean_context_runner_resolves_relative_output_before_chdir(self):
        valid = {"markdown":"ok", "behavior":"audit", "fact_ledger":[], "protected_items_preserved":True, "blocking":False}
        with tempfile.TemporaryDirectory() as directory:
            temp = Path(directory)
            output = Path(os.path.relpath(temp / "nested" / "answer.json", Path.cwd()))
            invocation_path = temp / "invocations.jsonl"
            def fake_run(command, **kwargs):
                target = Path(command[command.index("--output-last-message") + 1])
                self.assertTrue(target.is_absolute())
                target.parent.mkdir(parents=True, exist_ok=True)
                target.write_text(json.dumps(valid), encoding="utf-8")
                return subprocess.CompletedProcess(command, 0, stdout='{"type":"turn.completed","model":"gpt-test"}\n', stderr="")
            with mock.patch.object(model_runner, "INVOCATIONS", invocation_path), mock.patch.object(model_runner.subprocess, "run", side_effect=fake_run):
                value = model_runner.run_structured(
                    executor="codex", prompt="test", schema_path=EVAL / "harness" / "generation-schema.json",
                    output=output, kind="test", case_id="relative", condition="bare", model="gpt-test",
                )
            self.assertEqual(value, valid)
            self.assertTrue((temp / "nested" / "answer.json").exists())

    def test_claude_usage_is_not_double_counted(self):
        metadata = {
            "usage":{"input_tokens":4, "cache_creation_input_tokens":5057, "cache_read_input_tokens":2659, "output_tokens":4159,
                     "iterations":[{"input_tokens":2, "output_tokens":1801}]},
            "modelUsage":{"claude-haiku-4-5":{"inputTokens":823}, "claude-opus-4-8":{"inputTokens":4}},
        }
        usage = model_runner._usage(metadata)
        self.assertEqual((usage["input_tokens"], usage["cache_creation_input_tokens"], usage["cache_read_input_tokens"], usage["output_tokens"]), (4, 5057, 2659, 4159))
        self.assertEqual(model_runner._find_model(metadata, "claude-opus-4-8"), "claude-opus-4-8")

    def test_mock_collection_is_idempotent_and_retains_negatives(self):
        with tempfile.TemporaryDirectory() as directory:
            temp = Path(directory)
            manifest = {
                "phase": "calibration",
                "units": [{
                    "paired_unit_id":"calibration-0001", "phase":"calibration",
                    "topic_family":"workplace_technology", "intent":"explanatory",
                    "query":"mock query", "requests":[
                        {"platform":"google","request_tag":"calibration-0001-google"},
                        {"platform":"chatgpt","request_tag":"calibration-0001-chatgpt"}
                    ]
                }]
            }
            manifest_path = temp / "manifest.json"
            manifest_path.write_text(json.dumps(manifest), encoding="utf-8")
            command = [sys.executable, str(SCRIPTS / "collect.py"), "--manifest", str(manifest_path), "--ledger", str(temp / "ledger.jsonl"), "--run-dir", str(temp / "run"), "--mock-dir", str(FIXTURES / "api")]
            subprocess.run(command, check=True, capture_output=True)
            subprocess.run(command, check=True, capture_output=True)
            observations = pilotlib.read_jsonl(temp / "run" / "observations.jsonl")
            self.assertEqual(len(observations), 2)
            self.assertEqual(len(pilotlib.read_jsonl(temp / "ledger.jsonl")), 2)
        no_overview = pilotlib.parse_google_response(json.loads((FIXTURES / "api" / "google-no-overview.json").read_text()), request())
        no_citation = pilotlib.parse_chatgpt_response(json.loads((FIXTURES / "api" / "chatgpt-no-citations.json").read_text()), request("chatgpt"))
        self.assertEqual({no_overview["outcome"], no_citation["outcome"]}, {"no_ai_answer", "ai_answer_no_citations"})


if __name__ == "__main__":
    unittest.main()
