"""Source registry for the v0 newsjack monitoring engine."""

from __future__ import annotations

from typing import Any

import news_search
from lib import hackernews, reddit_public, xurl_x


DEFAULT_SOURCES = ["news_search", "x"]
OPTIONAL_SOURCES = ["reddit", "hackernews"]
ALL_SOURCES = [*DEFAULT_SOURCES, *OPTIONAL_SOURCES]


def parse_sources(raw: str | None) -> list[str]:
    if not raw:
        return list(DEFAULT_SOURCES)
    sources = []
    for source in raw.split(","):
        key = source.strip().lower()
        if not key:
            continue
        if key == "hn":
            key = "hackernews"
        if key == "news":
            key = "news_search"
        if key not in ALL_SOURCES:
            raise ValueError(f"Unsupported source for v0: {source}")
        if key not in sources:
            sources.append(key)
    return sources


def available_sources(config: dict[str, Any], requested: list[str]) -> list[str]:
    available = []
    for source in requested:
        if source == "news_search" and news_search.is_available(config):
            available.append(source)
        elif source == "x" and xurl_x.is_available():
            available.append(source)
        elif source in {"reddit", "hackernews"}:
            available.append(source)
    return available


def collect_source(
    source: str,
    query: str,
    *,
    from_date: str,
    to_date: str,
    depth: str,
    config: dict[str, Any],
) -> tuple[list[dict[str, Any]], str | None]:
    try:
        if source == "news_search":
            return news_search.search_news(
                query,
                from_date=from_date,
                to_date=to_date,
                limit=_limit(depth),
                config=config,
            ), None
        if source == "x":
            response = xurl_x.search_x(query, depth=depth)
            items = xurl_x.parse_x_response(response, topic=query)
            return [_map_x(item) for item in items], None
        if source == "reddit":
            items = reddit_public.search_reddit_public(query, from_date, to_date, depth=depth)
            return [_map_reddit(item) for item in items], None
        if source == "hackernews":
            response = hackernews.search_hackernews(query, from_date, to_date, depth=depth)
            items = hackernews.parse_hackernews_response(response, query=query)
            return [_map_hackernews(item) for item in items], None
    except Exception as exc:
        return [], f"{type(exc).__name__}: {exc}"
    return [], f"Unsupported source: {source}"


def _limit(depth: str) -> int:
    return {"quick": 10, "default": 25, "deep": 50}.get(depth, 25)


def _map_x(item: dict[str, Any]) -> dict[str, Any]:
    return {
        "id": item.get("id"),
        "source": "x",
        "title": item.get("text", "")[:120],
        "url": item.get("url", ""),
        "author": item.get("author_handle"),
        "container": "x.com",
        "published_at": item.get("date"),
        "excerpt": item.get("text", ""),
        "engagement": item.get("engagement") or {},
        "metadata": {},
    }


def _map_reddit(item: dict[str, Any]) -> dict[str, Any]:
    return {
        "id": item.get("id"),
        "source": "reddit",
        "title": item.get("title", ""),
        "url": item.get("url", ""),
        "author": item.get("author"),
        "container": item.get("subreddit"),
        "published_at": item.get("date"),
        "excerpt": item.get("selftext") or item.get("body") or item.get("snippet") or "",
        "engagement": item.get("engagement") or {},
        "metadata": {"top_comments": item.get("top_comments") or []},
    }


def _map_hackernews(item: dict[str, Any]) -> dict[str, Any]:
    return {
        "id": item.get("id"),
        "source": "hackernews",
        "title": item.get("title", ""),
        "url": item.get("url", ""),
        "author": item.get("author"),
        "container": "news.ycombinator.com",
        "published_at": item.get("date"),
        "excerpt": item.get("text") or item.get("snippet") or "",
        "engagement": item.get("engagement") or {},
        "metadata": {},
    }
