#!/usr/bin/env python3
"""Fetch robots-permitted public pages and derive inspectable prose features."""

from __future__ import annotations

import argparse
import hashlib
import json
import math
import re
import time
import urllib.error
import urllib.request
import urllib.robotparser
from collections import Counter
from concurrent.futures import ThreadPoolExecutor, as_completed
from datetime import date, datetime, timezone
from html.parser import HTMLParser
from pathlib import Path
from urllib.parse import urlsplit

from pilotlib import canonical_json, canonicalize_url, read_jsonl, utc_now


EVAL = Path(__file__).resolve().parents[1]
USER_AGENT = "NewsjackAIVisibilityPilot/1.0"
MAX_BYTES = 2_000_000
ROBOTS_TIMEOUT = 5
PAGE_TIMEOUT = 10
FETCH_WORKERS = 12
STOP = {"about", "after", "again", "also", "been", "being", "could", "does", "from", "have", "into", "more", "most", "only", "other", "should", "some", "than", "that", "their", "them", "then", "there", "these", "they", "this", "those", "through", "under", "very", "what", "when", "where", "which", "while", "with", "would", "your"}
PROMO = {"best", "breakthrough", "game-changing", "groundbreaking", "leading", "revolutionary", "transformative", "unmatched", "unprecedented", "world-class"}
GENERIC_HEADINGS = {"about", "conclusion", "details", "introduction", "more", "overview", "summary"}


class VisibleHTML(HTMLParser):
    def __init__(self) -> None:
        super().__init__(convert_charrefs=True)
        self.skip = 0
        self.capture: str | None = None
        self.buffers: list[str] = []
        self.text: list[str] = []
        self.headings: list[str] = []
        self.links = 0
        self.lists = 0
        self.tables = 0
        self.meta: dict[str, str] = {}

    def handle_starttag(self, tag: str, attrs: list[tuple[str, str | None]]) -> None:
        values = {key.lower(): value or "" for key, value in attrs}
        if tag in {"script", "style", "svg", "noscript", "nav", "footer"}:
            self.skip += 1
            return
        if self.skip:
            return
        if tag in {"p", "li", "h1", "h2", "h3", "h4", "h5", "h6", "td", "th", "title"}:
            self.capture = tag
            self.buffers = []
        if tag == "a" and values.get("href", "").startswith(("http://", "https://")):
            self.links += 1
        if tag in {"ul", "ol"}:
            self.lists += 1
        if tag == "table":
            self.tables += 1
        if tag == "meta":
            key = values.get("property") or values.get("name") or values.get("itemprop")
            if key and values.get("content"):
                self.meta[key.lower()] = values["content"].strip()

    def handle_endtag(self, tag: str) -> None:
        if tag in {"script", "style", "svg", "noscript", "nav", "footer"} and self.skip:
            self.skip -= 1
            return
        if self.skip or tag != self.capture:
            return
        value = re.sub(r"\s+", " ", " ".join(self.buffers)).strip()
        if value:
            self.text.append(value)
            if tag and tag.startswith("h"):
                self.headings.append(value)
        self.capture = None
        self.buffers = []

    def handle_data(self, data: str) -> None:
        if not self.skip and self.capture:
            self.buffers.append(data)


def parse_date(parser: VisibleHTML, text: str) -> date | None:
    keys = ("article:published_time", "article:modified_time", "datepublished", "datemodified", "date")
    candidates = [parser.meta[key] for key in keys if key in parser.meta]
    candidates += re.findall(r"\b(?:19|20)\d{2}-\d{2}-\d{2}\b", text[:3000])
    for raw in candidates:
        try:
            return datetime.fromisoformat(raw.replace("Z", "+00:00")).date()
        except ValueError:
            try:
                return date.fromisoformat(raw[:10])
            except ValueError:
                continue
    return None


def document_type(url: str, parser: VisibleHTML) -> str:
    haystack = (url + " " + " ".join(parser.meta.values())).lower()
    if any(value in haystack for value in ("press-release", "press_release", "news-release", "press release")):
        return "press_release"
    if any(value in haystack for value in ("opinion", "contributed", "commentary", "author")):
        return "contributed_article"
    if any(value in haystack for value in ("/blog/", "blogposting", "article")):
        return "blog"
    return "unknown"


def feature_vector(*, html: str, query: str, intent: str, url: str, observed: date) -> dict:
    parser = VisibleHTML()
    parser.feed(html)
    text = "\n".join(parser.text)
    tokens = re.findall(r"[A-Za-z][A-Za-z'-]*|\d[\d,.%]*", text)
    lower = [token.lower() for token in tokens]
    words = [token for token in lower if re.match(r"[a-z]", token)]
    content = max(len(words), 1)
    query_terms = {token for token in re.findall(r"[a-z]{4,}", query.lower()) if token not in STOP}
    opening = set(words[:120])
    sentences = [part for part in re.split(r"[.!?]+", text) if part.strip()]
    mean_sentence = content / max(len(sentences), 1)
    method_hits = len(re.findall(r"\b(?:analysis|data|experiment|firsthand|measured|method|pilot|sample|study|survey|tested)\b", text, re.I))
    numeric_hits = len(re.findall(r"\b\d[\d,.]*%?\b", text))
    evidence_hits = len(re.findall(r"\b(?:according to|analysis|data|method|reported|research|source|study|survey)\b", text, re.I))
    acronym_hits = len(re.findall(r"\b[A-Z][A-Za-z -]{2,50}\s+\([A-Z]{2,8}\)", text))
    generic = sum(heading.strip().lower() in GENERIC_HEADINGS for heading in parser.headings)
    promo_hits = sum(token in PROMO for token in lower)
    counts = Counter(token for token in words if len(token) >= 4 and token not in STOP)
    repetition = max(counts.values(), default=0) / content
    published = parse_date(parser, text)
    age = (observed - published).days if published else None
    return {
        "content_length": len(words),
        "publication_date": published.isoformat() if published else None,
        "publication_age_days": age,
        "document_type": document_type(url, parser),
        "content_sha256": hashlib.sha256(text.encode()).hexdigest(),
        "allowed_excerpt": text[:300],
        "html_signals": {"headings": len(parser.headings), "outbound_links": parser.links, "lists": parser.lists, "tables": parser.tables},
        "factor_features": {
            "F1": min(1.0, (method_hits + numeric_hits * 0.5) / 8),
            "F2": len(query_terms & opening) / max(len(query_terms), 1),
            "F3": min(1.0, (evidence_hits + min(parser.links, 3) + (1 if published else 0)) / 8),
            "F4": min(1.0, acronym_hits / 2 + (0.25 if re.search(r"\b[A-Z][a-z]+\s+[A-Z][a-z]+\b", text) else 0)),
            "F5": min(1.0, max(0.0, (len(parser.headings) - generic) / max(len(parser.headings), 1))),
            "F6": 1.0 if intent in {"comparative", "procedural"} and (parser.lists or parser.tables) else 0.0,
            "F7": 1.0 if intent == "freshness" and published else 0.0,
            "F8": max(0.0, 1 - 50 * promo_hits / content),
            "F9": max(0.0, 1 - abs(mean_sentence - 20) / 30),
            "F10": max(0.0, 1 - 8 * repetition),
        },
    }


def robots_for_root(root: str) -> urllib.robotparser.RobotFileParser | None:
    parser = urllib.robotparser.RobotFileParser(root + "/robots.txt")
    request = urllib.request.Request(parser.url, headers={"User-Agent": USER_AGENT})
    try:
        with urllib.request.urlopen(request, timeout=ROBOTS_TIMEOUT) as response:
            parser.parse(response.read(500_000).decode(response.headers.get_content_charset() or "utf-8", errors="replace").splitlines())
    except Exception:
        return None
    return parser


def fetch(url: str, parser: urllib.robotparser.RobotFileParser | None) -> tuple[str | None, str | None]:
    if parser is None:
        return None, "robots_unavailable"
    if not parser.can_fetch(USER_AGENT, url):
        return None, "robots_disallowed"
    request = urllib.request.Request(url, headers={"User-Agent": USER_AGENT, "Accept": "text/html"})
    try:
        with urllib.request.urlopen(request, timeout=PAGE_TIMEOUT) as response:
            content_type = response.headers.get_content_type()
            if content_type not in {"text/html", "application/xhtml+xml"}:
                return None, "non_html"
            payload = response.read(MAX_BYTES + 1)
            if len(payload) > MAX_BYTES:
                return None, "too_large"
            return payload.decode(response.headers.get_content_charset() or "utf-8", errors="replace"), None
    except (urllib.error.HTTPError, urllib.error.URLError, TimeoutError):
        return None, "fetch_failed"


def fetch_domain(urls: list[str]) -> dict[str, tuple[str | None, str | None]]:
    parts = urlsplit(urls[0])
    parser = robots_for_root(f"{parts.scheme}://{parts.netloc}")
    results = {}
    for url in urls:
        results[url] = fetch(url, parser)
        time.sleep(0.25)
    return results


def fetch_all(urls: list[str]) -> dict[str, tuple[str | None, str | None]]:
    by_root: dict[str, list[str]] = {}
    for url in urls:
        parts = urlsplit(url)
        by_root.setdefault(f"{parts.scheme}://{parts.netloc}", []).append(url)
    results = {}
    with ThreadPoolExecutor(max_workers=FETCH_WORKERS) as executor:
        futures = {executor.submit(fetch_domain, values): root for root, values in by_root.items()}
        for future in as_completed(futures):
            results.update(future.result())
    return results


def event_rows(observations: list[dict]) -> list[dict]:
    rows: dict[tuple[str, str], dict] = {}
    for observation in observations:
        base = {key: observation.get(key) for key in ("request_tag", "paired_unit_id", "platform", "query", "topic_family", "intent", "observed_at")}
        for record in observation.get("organic_results", []):
            url = record.get("canonical_url") or canonicalize_url(record.get("url", ""))
            if url:
                rows[(observation["request_tag"], url)] = {
                    **base, **record, "canonical_url": url,
                    "cited": record.get("label") != "organic_not_ai_cited",
                    "risk_set": "google_organic",
                }
        for record in observation.get("citations", []):
            url = record.get("canonical_url") or canonicalize_url(record.get("url", ""))
            if url:
                key = (observation["request_tag"], url)
                if key in rows:
                    organic = rows[key]
                    rows[key] = {
                        **organic, **record, "canonical_url": url, "cited": True,
                        "organic_position": organic.get("organic_position"),
                        "risk_set": "google_organic",
                    }
                else:
                    rows[key] = {
                        **base, **record, "canonical_url": url, "cited": True,
                        "risk_set": "chatgpt_cited" if observation.get("platform") == "chatgpt" else "google_citation_only",
                    }
    return list(rows.values())


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--observations", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--html-fixture", type=Path, help="Use one constructed HTML fixture for deterministic smoke tests only.")
    args = parser.parse_args()
    rows = event_rows(read_jsonl(args.observations))
    urls = list(dict.fromkeys(row["canonical_url"] for row in rows))
    if args.html_fixture:
        html = args.html_fixture.read_text(encoding="utf-8")
        fetched = {url: (html, None) for url in urls}
    else:
        fetched = fetch_all(urls)
    output_rows = []
    for row in rows:
        url = row["canonical_url"]
        html, exclusion = fetched[url]
        if html is None:
            features = None
        else:
            observed = datetime.fromisoformat((row.get("observed_at") or utc_now()).replace("Z", "+00:00")).date()
            features = feature_vector(html=html, query=row.get("query") or "", intent=row.get("intent") or "", url=url, observed=observed)
        output_rows.append({**row, "publisher_domain": urlsplit(url).hostname, "features": features, "exclusion_reason": exclusion})
    args.output.parent.mkdir(parents=True, exist_ok=True)
    with args.output.open("w", encoding="utf-8") as handle:
        for row in output_rows:
            handle.write(canonical_json(row) + "\n")
    print(json.dumps({"events": len(output_rows), "pages": len(fetched), "excluded": dict(Counter(row["exclusion_reason"] for row in output_rows if row["exclusion_reason"]))}, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
