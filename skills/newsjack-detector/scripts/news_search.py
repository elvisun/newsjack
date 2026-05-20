"""News search source.

This is the monitor's primary "what happened?" surface. The default backend is
the configured news API endpoint, which preserves Serper News request/response
shape.
"""

from __future__ import annotations

import os
import re
from datetime import datetime, timedelta, timezone
from email.utils import parsedate_to_datetime
from typing import Any

from lib import http


DEFAULT_BASE_URL = "https://medialyst.ai/api"
DEFAULT_NEWS_PATH = "/v1/news/search"


def is_available(config: dict[str, Any]) -> bool:
    return bool(config.get("MEDIALYST_API_KEY") or os.environ.get("MEDIALYST_API_KEY"))


def search_news(
    query: str,
    *,
    from_date: str,
    to_date: str,
    limit: int,
    config: dict[str, Any],
) -> list[dict[str, Any]]:
    api_key = config.get("MEDIALYST_API_KEY") or os.environ.get("MEDIALYST_API_KEY")
    if not api_key:
        return []
    base_url = (
        config.get("MEDIALYST_API_BASE")
        or os.environ.get("MEDIALYST_API_BASE")
        or DEFAULT_BASE_URL
    ).rstrip("/")
    path = config.get("MEDIALYST_NEWS_PATH") or os.environ.get("MEDIALYST_NEWS_PATH") or DEFAULT_NEWS_PATH
    url = f"{base_url}{path}"
    payload = http.request(
        "POST",
        url,
        headers={"Authorization": f"Bearer {api_key}"},
        json_data={
            "q": query,
            "num": limit,
            "gl": "us",
            "hl": "en",
            "tbs": _tbs_for_range(from_date, to_date),
        },
        timeout=30,
        retries=2,
    )
    return parse_news_response(payload)


def parse_news_response(payload: dict[str, Any]) -> list[dict[str, Any]]:
    raw_items = (
        payload.get("items")
        or payload.get("results")
        or payload.get("news")
        or payload.get("organic")
        or []
    )
    items: list[dict[str, Any]] = []
    for index, item in enumerate(raw_items):
        if not isinstance(item, dict):
            continue
        title = _first(item, "title", "headline", "name")
        url = _first(item, "url", "link")
        snippet = _first(item, "snippet", "summary", "description", "content")
        source = _first(item, "source", "publication", "publisher", "site")
        if isinstance(source, dict):
            source = _first(source, "name", "domain")
        published_at = _first(
            item,
            "published_at",
            "publishedAt",
            "published",
            "date",
            "created_at",
        )
        if not title and not snippet:
            continue
        items.append(
            {
                "id": str(_first(item, "id", "uuid") or f"ML{index + 1}"),
                "source": "news_search",
                "title": str(title or snippet or "").strip(),
                "url": str(url or "").strip(),
                "author": _first(item, "author", "byline"),
                "container": str(source or "").strip() or None,
                "published_at": _normalize_date(str(published_at)) if published_at else None,
                "excerpt": str(snippet or "").strip(),
                "engagement": {},
                "metadata": {"raw_source": source},
            }
        )
    return items


def _first(payload: dict[str, Any], *keys: str) -> Any:
    for key in keys:
        value = payload.get(key)
        if value not in (None, ""):
            return value
    return None


def _normalize_date(value: str) -> str:
    value = value.strip()
    if not value:
        return value
    if len(value) >= 10 and value[4] == "-" and value[7] == "-":
        return value
    relative = _relative_to_iso(value)
    if relative:
        return relative
    try:
        return parsedate_to_datetime(value).astimezone(timezone.utc).isoformat()
    except (TypeError, ValueError, OverflowError):
        pass
    return value


def _relative_to_iso(value: str) -> str | None:
    match = re.match(r"^\s*(\d+)\s+(minute|minutes|hour|hours|day|days|week|weeks|month|months)\s+ago\s*$", value, re.I)
    if not match:
        return None
    amount = int(match.group(1))
    unit = match.group(2).lower()
    if unit.startswith("minute"):
        delta = timedelta(minutes=amount)
    elif unit.startswith("hour"):
        delta = timedelta(hours=amount)
    elif unit.startswith("day"):
        delta = timedelta(days=amount)
    elif unit.startswith("week"):
        delta = timedelta(weeks=amount)
    else:
        delta = timedelta(days=amount * 30)
    return (datetime.now(timezone.utc) - delta).isoformat()


def _tbs_for_range(from_date: str, to_date: str) -> str:
    try:
        start = datetime.fromisoformat(from_date[:10])
        end = datetime.fromisoformat(to_date[:10])
    except ValueError:
        return "qdr:w"
    days = max(1, (end - start).days)
    if days <= 1:
        return "qdr:d"
    if days <= 7:
        return "qdr:w"
    return "qdr:m"
