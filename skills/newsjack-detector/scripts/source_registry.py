"""Source registry for the v0 newsjack monitoring engine."""

from __future__ import annotations

from typing import Any

import news_search
from lib import hackernews, reddit_public, xurl_x


DEFAULT_SOURCES = ["news_search", "x_news", "x"]
OPTIONAL_SOURCES = ["x_trends", "reddit", "hackernews"]
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
        if key in {"twitter", "x_posts"}:
            key = "x"
        if key not in ALL_SOURCES:
            raise ValueError(f"Unsupported source for v0: {source}")
        if key not in sources:
            sources.append(key)
    return sources


def available_sources(config: dict[str, Any], requested: list[str]) -> list[str]:
    available = []
    xurl_available = xurl_x.is_available()
    bearer_token = _bearer_token(config)
    for source in requested:
        if source == "news_search" and news_search.is_available(config):
            available.append(source)
        elif source == "x" and xurl_available:
            available.append(source)
        elif source == "x_news" and (xurl_available or bearer_token):
            available.append(source)
        elif source == "x_trends" and (xurl_available or bearer_token):
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
            if response.get("error"):
                return [], response["error"]
            counts = xurl_x.recent_count_summary(query, bearer_token=_bearer_token(config))
            items = xurl_x.parse_x_response(response, topic=query, counts_summary=counts)
            mapped = [_map_x(item) for item in items if xurl_x.keep_x_item(item)]
            return mapped, None
        if source == "x_news":
            response = xurl_x.search_x_news(
                query,
                depth=depth,
                max_age_hours=_lookback_hours(from_date, to_date),
                bearer_token=_bearer_token(config),
            )
            if response.get("error"):
                return [], response["error"]
            items = xurl_x.parse_x_news_response(response, topic=query)
            return [_map_x_news(item) for item in items], None
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


def collect_x_trends(
    trends_config: dict[str, Any],
    *,
    depth: str,
    config: dict[str, Any],
) -> tuple[list[dict[str, Any]], str | None]:
    items, error = xurl_x.collect_x_trends(
        trends_config,
        depth=depth,
        bearer_token=_bearer_token(config),
    )
    return [_map_x_trend(item) for item in items], error


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
        "metadata": item.get("metadata") or {},
    }


def _map_x_news(item: dict[str, Any]) -> dict[str, Any]:
    return {
        "id": item.get("id"),
        "source": "x_news",
        "title": item.get("title") or item.get("text", "")[:120],
        "url": item.get("url", ""),
        "author": item.get("author_handle"),
        "container": "x.com/news",
        "published_at": item.get("date"),
        "excerpt": item.get("text", ""),
        "engagement": item.get("engagement") or {},
        "metadata": item.get("metadata") or {},
    }


def _map_x_trend(item: dict[str, Any]) -> dict[str, Any]:
    return {
        "id": item.get("id"),
        "source": "x_trends",
        "title": item.get("title") or item.get("text", "")[:120],
        "url": item.get("url", ""),
        "author": item.get("author_handle"),
        "container": "x.com/trends",
        "published_at": item.get("date"),
        "excerpt": item.get("text", ""),
        "engagement": item.get("engagement") or {},
        "metadata": item.get("metadata") or {},
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


def _bearer_token(config: dict[str, Any]) -> str | None:
    for key in ("TWITTER_BEARER_TOKEN", "X_BEARER_TOKEN", "X_API_BEARER_TOKEN", "TWITTER_API_BEARER_TOKEN"):
        value = config.get(key)
        if value:
            return str(value)
    return None


def _lookback_hours(from_date: str, to_date: str) -> int:
    try:
        start = from_date[:10]
        end = to_date[:10]
        from datetime import date

        days = (date.fromisoformat(end) - date.fromisoformat(start)).days + 1
        return max(24, days * 24)
    except Exception:
        return 168


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
