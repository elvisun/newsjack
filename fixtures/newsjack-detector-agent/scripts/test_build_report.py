import importlib.util
import json
import tempfile
import unittest
from pathlib import Path
from urllib.parse import parse_qs, urlparse


MODULE_PATH = Path(__file__).with_name("build_report.py")
SPEC = importlib.util.spec_from_file_location("build_report", MODULE_PATH)
build_report = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(build_report)


class MediaListHandoffTest(unittest.TestCase):
    def test_handoff_url_round_trips_public_campaign_prompt(self):
        with tempfile.TemporaryDirectory(prefix="newsjack-report_") as run:
            Path(run, "angles.signal-1.json").write_text(json.dumps({
                "angles": [{
                    "headline_frame": "AI & PR: what changes now?",
                    "journalist_shape": {
                        "beat_description": "PR technology reporters",
                        "why_they_care_now": "A new agent workflow is live",
                        "do_not_target": "consumer AI roundups",
                    },
                }]
            }))
            prompt, url = build_report.media_list_handoff(
                run,
                "Newsjack",
                {
                    "signal_id": "signal-12345678",
                    "signal_title": "Newsjack ships agent-driven media lists",
                    "standing_rationale": "Newsjack maintains the open-source workflow",
                },
                {},
                {"url": "https://example.com/story?a=1&b=2"},
            )

        self.assertLessEqual(len(prompt), 2000)
        parsed = urlparse(url)
        self.assertEqual(parsed.scheme, "https")
        self.assertEqual(parsed.netloc, "medialyst.ai")
        self.assertEqual(parsed.path, "/app/_/workflow/campaign")
        self.assertEqual(parse_qs(parsed.query)["prompt"], [prompt])
        self.assertIn("AI & PR: what changes now?", prompt)
        self.assertIn("https://example.com/story?a=1&b=2", prompt)

    def test_handoff_shortens_optional_prose_without_cutting_safety_or_source(self):
        exclusion = "consumer AI roundups and general startup lists"
        source = "https://example.com/source-of-record?story=ai&region=ca"
        with tempfile.TemporaryDirectory(prefix="newsjack-report_") as run:
            Path(run, "angles.signal-1.json").write_text(json.dumps({
                "angles": [{
                    "headline_frame": "A" * 1800,
                    "journalist_shape": {
                        "beat_description": "enterprise AI reporters",
                        "do_not_target": exclusion,
                    },
                }]
            }))
            prompt, url = build_report.media_list_handoff(
                run,
                "Newsjack",
                {
                    "signal_id": "signal-12345678",
                    "signal_title": "Long campaign brief",
                    "standing_rationale": "S" * 1800,
                },
                {},
                {"url": source},
            )

        self.assertLessEqual(len(prompt), 2000)
        self.assertIn(f"Exclude: {exclusion}.", prompt)
        self.assertIn(f"Source of record: {source}", prompt)
        self.assertTrue(prompt.endswith(source))
        self.assertEqual(parse_qs(urlparse(url).query)["prompt"], [prompt])

    def test_handoff_requires_a_viable_angle(self):
        with tempfile.TemporaryDirectory(prefix="newsjack-report_") as run:
            prompt, url = build_report.media_list_handoff(
                run,
                "Newsjack",
                {"signal_id": "signal-12345678", "signal_title": "No angle"},
                {},
                {"url": "https://example.com/story"},
            )
        self.assertIsNone(prompt)
        self.assertIsNone(url)

    def test_report_links_only_pitch_ready_opportunities(self):
        with tempfile.TemporaryDirectory(prefix="newsjack-report_") as parent:
            run = Path(parent, "run_newsjack")
            run.mkdir()

            signals = []
            for signal_id, title in (
                ("pitch0001", "Pitch-ready story"),
                ("noangle01", "Angleless pitch candidate"),
                ("big000001", "Big story"),
                ("watch0001", "Watch story"),
            ):
                signals.append({
                    "id": signal_id,
                    "title": title,
                    "evidence": [{
                        "title": title,
                        "url": f"https://example.com/{signal_id}",
                        "container": "Example News",
                        "published_at": "2026-09-14T12:00:00Z",
                    }],
                    "freshness_gate": {"computed_status": "fresh"},
                    "story_origin": {"first_public_at": "2026-09-14T12:00:00Z"},
                })

            Path(run, "candidates.json").write_text(json.dumps({"signals": signals}))
            Path(run, "clustered_candidates.json").write_text(json.dumps({
                "clustering": {"representative_count": 4, "duplicate_count": 0},
                "clustered_duplicates": [],
            }))
            Path(run, "targeted_candidates.json").write_text(json.dumps({
                "signals": signals,
                "freshness_gate": {"rejected_signals": []},
            }))
            Path(run, "triaged_candidates.json").write_text(json.dumps({
                "triaged": [
                    {
                        "signal_id": "pitch0001",
                        "signal_title": "Pitch-ready story",
                        "tier": "pitch_ready",
                        "standing": "strong",
                        "standing_rationale": "Newsjack maintains the workflow",
                    },
                    {
                        "signal_id": "big000001",
                        "signal_title": "Big story",
                        "tier": "big_story",
                    },
                    {
                        "signal_id": "noangle01",
                        "signal_title": "Angleless pitch candidate",
                        "tier": "pitch_ready",
                        "standing": "strong",
                        "standing_rationale": "No angle cleared the angle gate",
                    },
                    {
                        "signal_id": "watch0001",
                        "signal_title": "Watch story",
                        "tier": "watch",
                        "watch_reason": "no standing",
                    },
                ],
            }))
            Path(run, "angles.pitch000.json").write_text(json.dumps({
                "angles": [{
                    "headline_frame": "What agent-driven PR changes",
                    "journalist_shape": {"beat_description": "PR technology reporters"},
                }]
            }))

            build_report.build(str(run))
            report = Path(run, "final_report.md").read_text()

        self.assertEqual(report.count("/app/_/workflow/campaign?prompt="), 1)
        link_position = report.index("/app/_/workflow/campaign?prompt=")
        self.assertLess(link_position, report.index("## 🔥 Big Stories Worth a Look"))
        self.assertIn("credits start only after you approve it in Medialyst", report)
        pitch_section, watch_section = report.split("## 👀 Watch / Context", 1)
        self.assertNotIn("Angleless pitch candidate", pitch_section)
        self.assertIn("Angleless pitch candidate", watch_section)
        self.assertIn("no_viable_angle", watch_section)


if __name__ == "__main__":
    unittest.main()
