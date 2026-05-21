#!/usr/bin/env python3
"""v0 newsjack monitoring engine.

This command ingests public signals, clusters related evidence, and computes
mechanical monitor scores. It does not decide whether to pitch. The skill owns
that PR judgment.
"""

from __future__ import annotations

import argparse
import hashlib
import json
import math
import os
import re
import sys
from concurrent.futures import ThreadPoolExecutor, as_completed
from dataclasses import dataclass, field
from datetime import datetime, timezone
from pathlib import Path
from typing import Any

SCRIPT_DIR = Path(__file__).parent.resolve()
sys.path.insert(0, str(SCRIPT_DIR))

import brand_safety_flags
import newsjack_store
import source_registry
from lib import dates, rss_feed
from monitor_profile import MonitorProfile


STOP_WORDS = {
    "about", "after", "again", "against", "also", "and", "are", "because",
    "been", "being", "but", "can", "could", "did", "does", "doing", "for",
    "from", "had", "has", "have", "her", "here", "him", "his", "how",
    "into", "its", "just", "more", "new", "not", "now", "our", "out",
    "over", "per", "said", "say", "says", "she", "should", "than", "that",
    "the", "their", "them", "then", "there", "these", "they", "this",
    "those", "through", "too", "under", "was", "what", "when", "where",
    "which", "while", "who", "why", "will", "with", "would", "you", "your",
}

SOURCE_QUALITY = {
    "major_feed": 0.88,
    "news_search": 0.95,
    "x": 0.70,
    "reddit": 0.62,
    "hackernews": 0.72,
}

ENGAGEMENT_FIELDS = (
    "score", "num_comments", "comments", "likes", "reposts", "replies",
    "quotes", "bookmarks", "views", "points",
)

DEFAULT_MAJOR_FEEDS = (
    "https://www.techmeme.com/feed.xml",
)

MAJOR_NEWS_TERMS = {
    "acquire", "acquired", "acquisition", "antitrust", "ban", "billion",
    "breach", "contract", "deal", "funding", "hack", "investigation",
    "ipo", "lawsuit", "launch", "launched", "layoffs", "merger", "outage",
    "probe", "regulation", "regulator", "ruling", "sec", "settlement",
    "shutdown", "sues", "valuation",
}

MAJOR_ENTITY_TERMS = {
    "amazon", "anthropic", "apple", "doj", "ftc", "google", "meta",
    "microsoft", "nvidia", "openai", "pentagon", "salesforce", "sec",
    "spacex", "tesla", "white house", "xai",
}


@dataclass
class EvidenceItem:
    source: str
    title: str
    url: str
    excerpt: str = ""
    author: str | None = None
    container: str | None = None
    published_at: str | None = None
    engagement: dict[str, Any] = field(default_factory=dict)
    metadata: dict[str, Any] = field(default_factory=dict)

    @classmethod
    def from_dict(cls, payload: dict[str, Any]) -> "EvidenceItem":
        return cls(
            source=str(payload.get("source") or "unknown"),
            title=str(payload.get("title") or payload.get("excerpt") or "").strip(),
            url=str(payload.get("url") or "").strip(),
            excerpt=str(payload.get("excerpt") or "").strip(),
            author=payload.get("author"),
            container=payload.get("container"),
            published_at=payload.get("published_at"),
            engagement=dict(payload.get("engagement") or {}),
            metadata=dict(payload.get("metadata") or {}),
        )

    def text(self) -> str:
        return " ".join(part for part in [self.title, self.excerpt, self.author or "", self.container or ""] if part)

    def public_dict(self) -> dict[str, Any]:
        payload = {
            "source": self.source,
            "title": self.title,
            "url": self.url,
            "author": self.author,
            "container": self.container,
            "published_at": self.published_at,
            "excerpt": self.excerpt[:500],
            "engagement": self.engagement,
        }
        if self.metadata:
            payload["metadata"] = self.metadata
        return payload


@dataclass
class SignalCluster:
    evidence: list[EvidenceItem] = field(default_factory=list)

    def add(self, item: EvidenceItem) -> None:
        self.evidence.append(item)

    @property
    def title(self) -> str:
        for item in self.evidence:
            if item.source == "news_search" and item.title:
                return item.title
        return self.evidence[0].title if self.evidence else ""

    def text(self) -> str:
        return " ".join(item.text() for item in self.evidence)

    def sources(self) -> list[str]:
        return sorted({item.source for item in self.evidence})

    def urls(self) -> list[str]:
        return [item.url for item in self.evidence if item.url]


def _tokens(text: str) -> set[str]:
    return {
        token
        for token in re.findall(r"[a-z0-9][a-z0-9+._-]{2,}", text.lower())
        if token not in STOP_WORDS
    }


def _jaccard(left: str, right: str) -> float:
    left_tokens = _tokens(left)
    right_tokens = _tokens(right)
    if not left_tokens or not right_tokens:
        return 0.0
    return len(left_tokens & right_tokens) / len(left_tokens | right_tokens)


def _profile_matches(profile: MonitorProfile, text: str) -> list[str]:
    lower = text.lower()
    terms = [
        *(profile.topics or []),
        *(profile.competitors or []),
        *(profile.standing or []),
        *(profile.proof_assets or []),
    ]
    matches = []
    for term in terms:
        term = term.strip()
        if term and term.lower() in lower and term not in matches:
            matches.append(term)
    return matches[:12]


def _profile_match_score(profile: MonitorProfile, text: str) -> float:
    context = profile.match_text()
    if not context:
        return 0.4
    overlap = _jaccard(context, text)
    phrase_bonus = min(0.4, 0.08 * len(_profile_matches(profile, text)))
    return min(1.0, overlap + phrase_bonus)


def _parse_time(value: str | None) -> datetime | None:
    if not value:
        return None
    raw = value.strip()
    if not raw:
        return None
    if raw.endswith("Z"):
        raw = f"{raw[:-1]}+00:00"
    try:
        parsed = datetime.fromisoformat(raw)
    except ValueError:
        try:
            parsed = datetime.fromisoformat(f"{raw[:10]}T00:00:00+00:00")
        except ValueError:
            return None
    if parsed.tzinfo is None:
        parsed = parsed.replace(tzinfo=timezone.utc)
    return parsed.astimezone(timezone.utc)


def _min_age_hours(cluster: SignalCluster, now: datetime) -> float | None:
    ages = []
    for item in cluster.evidence:
        parsed = _parse_time(item.published_at)
        if parsed:
            ages.append(max(0.0, (now - parsed).total_seconds() / 3600.0))
    return min(ages) if ages else None


def _freshness_score(age_hours: float | None, lookback_days: int) -> float:
    if age_hours is None:
        return 0.35
    if age_hours <= 4:
        return 1.0
    if age_hours <= 24:
        return 0.86
    window_hours = max(24, lookback_days * 24)
    return max(0.05, 1.0 - (age_hours / window_hours))


def _decay_bucket(age_hours: float | None) -> str:
    if age_hours is None:
        return "unknown"
    if age_hours <= 1:
        return "30min"
    if age_hours <= 4:
        return "4hr"
    if age_hours <= 24:
        return "24hr"
    if age_hours <= 168:
        return "week"
    return "month"


def _filter_items_by_age(
    items: list[EvidenceItem],
    *,
    now: datetime,
    max_age_hours: float | None,
) -> list[EvidenceItem]:
    if max_age_hours is None or max_age_hours <= 0:
        return items
    output = []
    for item in items:
        parsed = _parse_time(item.published_at)
        if parsed is None:
            output.append(item)
            continue
        age_hours = max(0.0, (now - parsed).total_seconds() / 3600.0)
        if age_hours <= max_age_hours:
            output.append(item)
    return output


def _source_agreement_score(sources: list[str]) -> float:
    if len(sources) >= 3:
        return 1.0
    if len(sources) == 2:
        return 0.78
    if sources and sources[0] == "major_feed":
        return 0.62
    if sources and sources[0] == "news_search":
        return 0.55
    return 0.32


def _source_quality_score(cluster: SignalCluster) -> float:
    if not cluster.evidence:
        return 0.0
    return sum(SOURCE_QUALITY.get(item.source, 0.5) for item in cluster.evidence) / len(cluster.evidence)


def _engagement_score(cluster: SignalCluster) -> float:
    total = 0.0
    for item in cluster.evidence:
        for field in ENGAGEMENT_FIELDS:
            try:
                value = float(item.engagement.get(field) or 0)
            except (TypeError, ValueError):
                value = 0.0
            if value > 0:
                total += math.log1p(value)
    return min(1.0, total / 24.0)


def _novelty_score(urls: list[str], seen: dict[str, dict[str, Any]]) -> float:
    if not urls:
        return 0.50
    unseen = [url for url in urls if url not in seen]
    if not unseen:
        return 0.10
    return len(unseen) / len(urls)


def _major_news_score(cluster: SignalCluster, age_hours: float | None) -> float:
    """Estimate broad news importance for curated feed items.

    This is intentionally mechanical. The skill still decides whether a major
    story is newsjackable for the client.
    """
    feed_items = [item for item in cluster.evidence if item.source == "major_feed"]
    if not feed_items:
        return 0.0

    positions = []
    for item in feed_items:
        try:
            positions.append(int(item.metadata.get("feed_position") or 999))
        except (TypeError, ValueError):
            positions.append(999)
    best_position = min(positions or [999])
    if best_position <= 3:
        position_score = 1.0
    elif best_position <= 10:
        position_score = 0.82
    elif best_position <= 25:
        position_score = 0.64
    else:
        position_score = 0.42

    text = cluster.text().lower()
    tokens = _tokens(text)
    stake_hits = sum(1 for term in MAJOR_NEWS_TERMS if _term_matches(term, text, tokens))
    entity_hits = sum(1 for term in MAJOR_ENTITY_TERMS if _term_matches(term, text, tokens))
    stake_score = min(1.0, 0.22 * stake_hits)
    entity_score = min(1.0, 0.18 * entity_hits)
    freshness = _freshness_score(age_hours, 7)

    return min(
        1.0,
        0.44 * position_score
        + 0.24 * freshness
        + 0.20 * stake_score
        + 0.12 * entity_score,
    )


def _term_matches(term: str, text: str, tokens: set[str]) -> bool:
    if " " in term:
        return term in text
    return term in tokens


def _rank_signal(
    cluster: SignalCluster,
    *,
    profile: MonitorProfile,
    seen: dict[str, dict[str, Any]],
    now: datetime,
    lookback_days: int,
) -> dict[str, Any]:
    text = cluster.text()
    sources = cluster.sources()
    urls = cluster.urls()
    age_hours = _min_age_hours(cluster, now)
    freshness = _freshness_score(age_hours, lookback_days)
    novelty = _novelty_score(urls, seen)
    source_agreement = _source_agreement_score(sources)
    source_quality = _source_quality_score(cluster)
    engagement = _engagement_score(cluster)
    profile_match = _profile_match_score(profile, text)
    major_news = _major_news_score(cluster, age_hours)
    lane = "major_news" if major_news > 0 else "profile_relevance"
    if lane == "major_news":
        rank = round(
            100
            * (
                0.30 * major_news
                + 0.22 * freshness
                + 0.14 * novelty
                + 0.12 * profile_match
                + 0.12 * source_quality
                + 0.10 * source_agreement
            ),
            1,
        )
    else:
        rank = round(
            100
            * (
                0.24 * freshness
                + 0.20 * source_agreement
                + 0.18 * novelty
                + 0.16 * profile_match
                + 0.12 * source_quality
                + 0.10 * engagement
            ),
            1,
        )
    evidence = [item.public_dict() for item in cluster.evidence[:8]]
    signal_id = _signal_id(cluster.title, urls, text)
    return {
        "id": signal_id,
        "title": cluster.title,
        "sources": sources,
        "evidence": evidence,
        "features": {
            "age_hours": None if age_hours is None else round(age_hours, 2),
            "decay_bucket": _decay_bucket(age_hours),
            "source_count": len(sources),
            "evidence_count": len(cluster.evidence),
            "lane": lane,
            "seen_before": bool(urls and all(url in seen for url in urls)),
            "seen_urls": {url: seen[url] for url in urls if url in seen},
            "profile_matches": _profile_matches(profile, text),
            "safety_flags": brand_safety_flags.flag_text(text, exclusions=profile.exclusions),
        },
        "scores": {
            "rank": rank,
            "freshness": round(freshness, 3),
            "source_agreement": round(source_agreement, 3),
            "novelty": round(novelty, 3),
            "profile_match": round(profile_match, 3),
            "source_quality": round(source_quality, 3),
            "momentum": round(engagement, 3),
            "major_news": round(major_news, 3),
        },
    }


def _signal_id(title: str, urls: list[str], text: str) -> str:
    basis = "|".join([title, *urls[:5]]) or text[:400]
    return hashlib.sha1(basis.encode("utf-8")).hexdigest()[:16]


def _signal_is_seen(signal: dict[str, Any]) -> bool:
    return bool((signal.get("features") or {}).get("seen_before"))


def _cluster_items(items: list[EvidenceItem]) -> list[SignalCluster]:
    clusters: list[SignalCluster] = []
    sorted_items = sorted(
        items,
        key=lambda item: (
            0 if item.source == "news_search" else 1,
            item.published_at or "",
        ),
    )
    for item in sorted_items:
        if not item.title and not item.excerpt:
            continue
        placed = False
        for cluster in clusters:
            if item.url and item.url in cluster.urls():
                cluster.add(item)
                placed = True
                break
            if _jaccard(item.text(), cluster.text()) >= 0.32:
                cluster.add(item)
                placed = True
                break
        if not placed:
            clusters.append(SignalCluster([item]))
    return clusters


def _build_queries(args: argparse.Namespace, profile: MonitorProfile) -> list[str]:
    queries: list[str] = []
    queries.extend(args.topic or [])
    if args.query:
        queries.append(" ".join(args.query).strip())
    queries.extend(profile.query_terms())
    return _dedupe([query for query in queries if query])


def _feed_urls(args: argparse.Namespace, profile: MonitorProfile) -> list[str]:
    feeds: list[str] = []
    if not args.no_profile_feeds:
        feeds.extend(profile.feed_urls)
    if args.major_feeds:
        feeds.extend(_env_major_feeds() or ([] if profile.feed_urls else DEFAULT_MAJOR_FEEDS))
    feeds.extend(args.feed_url or [])
    for feed_file in args.feed_file or []:
        feeds.extend(_read_feed_urls(Path(feed_file).expanduser()))
    return _dedupe(feeds)


def _env_major_feeds() -> list[str]:
    raw = os.environ.get("NEWSJACK_MAJOR_FEEDS", "")
    if not raw.strip():
        return []
    return [part.strip() for part in re.split(r"[\n,]", raw) if part.strip()]


def _read_feed_urls(path: Path) -> list[str]:
    urls = []
    for raw_line in path.read_text(encoding="utf-8").splitlines():
        line = raw_line.strip()
        if not line or line.startswith("#"):
            continue
        urls.append(line)
    return urls


def _dedupe(values: list[str]) -> list[str]:
    seen = set()
    output = []
    for value in values:
        key = value.lower().strip()
        if not key or key in seen:
            continue
        seen.add(key)
        output.append(value.strip())
    return output


def _load_env_file(path: Path) -> dict[str, str]:
    values: dict[str, str] = {}
    if not path.exists():
        return values
    for raw_line in path.read_text(encoding="utf-8").splitlines():
        line = raw_line.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue
        key, _, value = line.partition("=")
        key = key.strip()
        value = value.strip()
        if not key:
            continue
        if value and value[0] in {"'", '"'} and value[-1:] == value[0]:
            value = value[1:-1]
        values[key] = value
    return values


def _env_file_values() -> dict[str, str]:
    repo_root = SCRIPT_DIR.parents[2]
    paths = [
        repo_root / ".env",
        Path.cwd() / ".env",
    ]
    merged: dict[str, str] = {}
    for path in paths:
        merged.update(_load_env_file(path))
    return merged


def _config_from_env() -> dict[str, str | None]:
    file_env = _env_file_values()

    def get(key: str) -> str | None:
        return os.environ.get(key) or file_env.get(key)

    return {
        "MEDIALYST_API_KEY": get("MEDIALYST_API_KEY"),
        "MEDIALYST_API_BASE": get("MEDIALYST_API_BASE"),
        "MEDIALYST_NEWS_PATH": get("MEDIALYST_NEWS_PATH"),
    }


def _mock_items(query: str, now: datetime) -> list[EvidenceItem]:
    today = now.date().isoformat()
    return [
        EvidenceItem(
            source="news_search",
            title=f"Regulators open inquiry tied to {query}",
            url=f"https://example.com/news/{hashlib.sha1(query.encode()).hexdigest()[:8]}",
            container="Example News",
            published_at=today,
            excerpt=f"Officials are examining claims and compliance practices around {query}.",
        ),
        EvidenceItem(
            source="x",
            title=f"Experts are reacting to {query}",
            url=f"https://x.com/example/status/{hashlib.sha1((query + 'x').encode()).hexdigest()[:8]}",
            author="example",
            container="x.com",
            published_at=today,
            excerpt=f"Thread: the {query} inquiry is moving faster than vendors expected.",
            engagement={"likes": 120, "reposts": 22, "replies": 9},
        ),
    ]


def _mock_feed_items(now: datetime) -> list[EvidenceItem]:
    published_at = now.isoformat()
    return [
        EvidenceItem(
            source="major_feed",
            title="Salesforce launches free AI customer service agents for startups",
            url="https://example.com/major/salesforce-ai-agents",
            container="Example Major Feed",
            published_at=published_at,
            excerpt=(
                "A major CRM vendor is targeting startup and SMB customer-support "
                "workflows with free AI agents."
            ),
            metadata={
                "feed_title": "Example Major Feed",
                "feed_url": "mock://major-feed",
                "feed_position": 1,
            },
        ),
        EvidenceItem(
            source="major_feed",
            title="Pentagon launches task force for safe deployment of AI tools",
            url="https://example.com/major/pentagon-ai-task-force",
            container="Example Major Feed",
            published_at=published_at,
            excerpt=(
                "The Pentagon is studying how to deploy leading AI tools across "
                "sensitive government workflows."
            ),
            metadata={
                "feed_title": "Example Major Feed",
                "feed_url": "mock://major-feed",
                "feed_position": 2,
            },
        ),
    ]


def collect_query(
    query: str,
    *,
    sources: list[str],
    config: dict[str, Any],
    depth: str,
    lookback_days: int,
    mock: bool,
    now: datetime,
) -> tuple[list[EvidenceItem], dict[str, str]]:
    if mock:
        return _mock_items(query, now), {}
    from_date, to_date = dates.get_date_range(lookback_days)
    errors: dict[str, str] = {}
    items: list[EvidenceItem] = []
    with ThreadPoolExecutor(max_workers=min(4, max(1, len(sources)))) as executor:
        futures = {
            executor.submit(
                source_registry.collect_source,
                source,
                query,
                from_date=from_date,
                to_date=to_date,
                depth=depth,
                config=config,
            ): source
            for source in sources
        }
        for future in as_completed(futures):
            source = futures[future]
            raw_items, error = future.result()
            if error:
                errors[source] = error
            for raw in raw_items:
                item = EvidenceItem.from_dict(raw)
                if item.source != "news_search" and _jaccard(query, item.text()) < 0.08:
                    continue
                if item.title or item.excerpt:
                    items.append(item)
    return items, errors


def collect_feeds(
    feed_urls: list[str],
    *,
    depth: str,
    mock: bool,
    now: datetime,
) -> tuple[list[EvidenceItem], dict[str, str]]:
    if mock:
        return _mock_feed_items(now), {}

    limit = {"quick": 15, "default": 30, "deep": 60}.get(depth, 30)
    items: list[EvidenceItem] = []
    errors: dict[str, str] = {}
    with ThreadPoolExecutor(max_workers=min(4, max(1, len(feed_urls)))) as executor:
        futures = {
            executor.submit(rss_feed.collect_feed, feed_url, limit=limit): feed_url
            for feed_url in feed_urls
        }
        for future in as_completed(futures):
            feed_url = futures[future]
            raw_items, error = future.result()
            if error:
                errors[feed_url] = error
            for raw in raw_items:
                item = EvidenceItem.from_dict(raw)
                if item.title or item.excerpt:
                    items.append(item)
    return items, errors


def run(args: argparse.Namespace) -> int:
    profile = MonitorProfile.from_file(args.profile) if args.profile else MonitorProfile()
    queries = [] if args.feed_only else _build_queries(args, profile)
    feed_urls = _feed_urls(args, profile)
    if not queries and not feed_urls:
        raise SystemExit("Provide a query, --topic, --major-feeds, --feed-url, or --profile with topics/competitors.")

    config = _config_from_env()
    requested_sources = source_registry.parse_sources(args.sources)
    sources = (
        (requested_sources if args.mock else source_registry.available_sources(config, requested_sources))
        if queries
        else []
    )
    if queries and not sources:
        raise SystemExit(
            "No requested sources are available. Configure MEDIALYST_API_KEY and xurl auth, "
            "or rerun with --mock."
        )

    now = datetime.now(timezone.utc)
    db_path = Path(args.store).expanduser() if args.store else None
    all_signals: list[dict[str, Any]] = []
    seen_urls_to_mark: list[str] = []
    source_errors: dict[str, dict[str, str]] = {}

    for query in queries:
        items, errors = collect_query(
            query,
            sources=sources,
            config=config,
            depth=args.depth,
            lookback_days=args.lookback_days,
            mock=args.mock,
            now=now,
        )
        items = _filter_items_by_age(items, now=now, max_age_hours=args.max_age_hours)
        if errors:
            source_errors[query] = errors
        clusters = _cluster_items(items)
        urls = [url for cluster in clusters for url in cluster.urls()]
        seen_urls_to_mark.extend(urls)
        seen = newsjack_store.seen_status(urls, db_path=db_path)
        for cluster in clusters:
            signal = _rank_signal(
                cluster,
                profile=profile,
                seen=seen,
                now=now,
                lookback_days=args.lookback_days,
            )
            signal["query"] = query
            if not (args.new_only and _signal_is_seen(signal)):
                all_signals.append(signal)

    if feed_urls:
        items, errors = collect_feeds(
            feed_urls,
            depth=args.depth,
            mock=args.mock,
            now=now,
        )
        items = _filter_items_by_age(items, now=now, max_age_hours=args.max_age_hours)
        if errors:
            source_errors["major_feeds"] = errors
        clusters = _cluster_items(items)
        urls = [url for cluster in clusters for url in cluster.urls()]
        seen_urls_to_mark.extend(urls)
        seen = newsjack_store.seen_status(urls, db_path=db_path)
        for cluster in clusters:
            signal = _rank_signal(
                cluster,
                profile=profile,
                seen=seen,
                now=now,
                lookback_days=args.lookback_days,
            )
            signal["query"] = "major_news_feed"
            if not (args.new_only and _signal_is_seen(signal)):
                all_signals.append(signal)

    all_signals.sort(key=lambda signal: signal["scores"]["rank"], reverse=True)
    signals = all_signals[: args.limit]
    run_id = None
    if args.save:
        run_id = newsjack_store.record_run(
            monitor_name=args.monitor_name,
            profile=profile.public_dict(),
            queries=queries,
            signals=signals,
            seen_urls=seen_urls_to_mark,
            db_path=db_path,
        )

    payload = {
        "monitor": {
            "name": args.monitor_name,
            "generated_at": now.isoformat(),
            "profile": profile.public_dict(),
            "queries": queries,
            "feed_urls": feed_urls,
            "sources_requested": requested_sources,
            "sources_used": [*sources, *(["major_feed"] if feed_urls else [])],
            "lookback_days": args.lookback_days,
            "max_age_hours": args.max_age_hours,
            "new_only": args.new_only,
            "depth": args.depth,
            "mock": args.mock,
        },
        "signals": signals,
        "source_errors": source_errors,
        "store": {
            "saved": args.save,
            "run_id": run_id,
            "path": str(db_path or newsjack_store.db_path_from_env()) if args.save else None,
        },
    }
    if args.emit == "brief":
        print(_brief(payload))
    else:
        print(json.dumps(payload, indent=2, sort_keys=True))
    return 0


def diagnose(args: argparse.Namespace) -> int:
    config = _config_from_env()
    requested = source_registry.parse_sources(args.sources)
    payload = {
        "sources_requested": requested,
        "sources_available": source_registry.available_sources(config, requested),
        "news_search_configured": bool(config.get("MEDIALYST_API_KEY")),
        "xurl_available": "x" in source_registry.available_sources(config, ["x"]),
        "store_path": str(Path(args.store).expanduser()) if args.store else str(newsjack_store.db_path_from_env()),
    }
    print(json.dumps(payload, indent=2, sort_keys=True))
    return 0


def recent(args: argparse.Namespace) -> int:
    db_path = Path(args.store).expanduser() if args.store else None
    print(json.dumps(newsjack_store.recent_runs(args.limit, db_path=db_path), indent=2))
    return 0


def _brief_evidence_links(signal: dict[str, Any], limit: int = 3) -> list[str]:
    links = []
    seen_urls = set()
    for evidence in signal.get("evidence") or []:
        url = evidence.get("url")
        if not url or url in seen_urls:
            continue
        seen_urls.add(url)
        source = evidence.get("source") or "source"
        title = evidence.get("title") or evidence.get("container") or ""
        if title:
            links.append(f"{source}: {title[:110]} - {url}")
        else:
            links.append(f"{source}: {url}")
        if len(links) >= limit:
            break
    return links


def _brief(payload: dict[str, Any]) -> str:
    lines = ["newsjack monitor", ""]
    for index, signal in enumerate(payload.get("signals") or [], start=1):
        features = signal["features"]
        lines.append(
            f"{index}. {signal['title']} "
            f"({signal['scores']['rank']}, {features['lane']}, {features['decay_bucket']}, "
            f"{', '.join(signal['sources'])})"
        )
        for link in _brief_evidence_links(signal):
            lines.append(f"   evidence: {link}")
        if features.get("safety_flags"):
            lines.append(f"   safety_flags={len(features['safety_flags'])}")
    if not payload.get("signals"):
        lines.append("No signals returned.")
    return "\n".join(lines)


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(description="Collect and rank newsjack monitoring evidence.")
    sub = parser.add_subparsers(dest="command")

    run_parser = sub.add_parser("run", help="Run a monitor pass.")
    run_parser.add_argument("query", nargs="*", help="Ad hoc monitor query.")
    run_parser.add_argument("--topic", action="append", help="Topic to monitor. Repeatable.")
    run_parser.add_argument("--profile", help="Path to monitor profile JSON.")
    run_parser.add_argument("--sources", help="Comma-separated sources. Default: news_search,x.")
    run_parser.add_argument("--major-feeds", action="store_true", help="Include default curated major-news RSS feeds when the profile has no feed_urls.")
    run_parser.add_argument("--feed-url", action="append", help="RSS/Atom feed URL or local XML path. Repeatable.")
    run_parser.add_argument("--feed-file", action="append", help="File containing RSS/Atom feed URLs, one per line. Repeatable.")
    run_parser.add_argument("--feed-only", action="store_true", help="Skip query/profile searches and only ingest RSS/Atom feeds.")
    run_parser.add_argument("--no-profile-feeds", action="store_true", help="Do not include feed_urls from the monitor profile.")
    run_parser.add_argument("--depth", choices=["quick", "default", "deep"], default="quick")
    run_parser.add_argument("--lookback-days", type=int, default=7)
    run_parser.add_argument("--max-age-hours", type=float, default=168.0, help="Drop items older than this many hours when a published time is available. Use 0 to disable.")
    run_parser.add_argument("--new-only", action="store_true", help="Suppress signals whose evidence URLs were already seen in the monitor store.")
    run_parser.add_argument("--limit", type=int, default=20)
    run_parser.add_argument("--mock", action="store_true")
    run_parser.add_argument("--save", action="store_true")
    run_parser.add_argument("--store", help="Override SQLite store path.")
    run_parser.add_argument("--monitor-name")
    run_parser.add_argument("--emit", choices=["json", "brief"], default="json")
    run_parser.set_defaults(func=run)

    diag_parser = sub.add_parser("diagnose", help="Show source availability.")
    diag_parser.add_argument("--sources")
    diag_parser.add_argument("--store")
    diag_parser.set_defaults(func=diagnose)

    recent_parser = sub.add_parser("recent", help="List recent stored monitor runs.")
    recent_parser.add_argument("--limit", type=int, default=10)
    recent_parser.add_argument("--store")
    recent_parser.set_defaults(func=recent)

    return parser


def main() -> int:
    parser = build_parser()
    commands = {"run", "diagnose", "recent"}
    if len(sys.argv) > 1 and sys.argv[1] not in commands and sys.argv[1] not in {"-h", "--help"}:
        argv = ["run", *sys.argv[1:]]
    else:
        argv = sys.argv[1:]
    args = parser.parse_args(argv)
    if not args.command:
        parser.print_help()
        return 1
    return args.func(args)


if __name__ == "__main__":
    raise SystemExit(main())
