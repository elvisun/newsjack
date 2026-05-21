"""RSS/Atom feed ingestion for broad major-news monitoring.

This module intentionally uses only stdlib XML parsing. It normalizes feed
items into the same evidence shape as the rest of the monitor engine.
"""

from __future__ import annotations

import html
import re
import xml.etree.ElementTree as ET
from datetime import datetime, timezone
from email.utils import parsedate_to_datetime
from pathlib import Path
from typing import Any

from . import http


def collect_feed(url_or_path: str, *, limit: int = 25) -> tuple[list[dict[str, Any]], str | None]:
    try:
        xml_text = _read_feed(url_or_path)
        return parse_feed(xml_text, feed_url=url_or_path, limit=limit), None
    except Exception as exc:
        return [], f"{type(exc).__name__}: {exc}"


def parse_feed(xml_text: str, *, feed_url: str, limit: int = 25) -> list[dict[str, Any]]:
    root = ET.fromstring(xml_text)
    if _local_name(root.tag) == "rss":
        return _parse_rss(root, feed_url=feed_url, limit=limit)
    if _local_name(root.tag) == "feed":
        return _parse_atom(root, feed_url=feed_url, limit=limit)
    return []


def _read_feed(url_or_path: str) -> str:
    if re.match(r"^https?://", url_or_path, re.I):
        return str(
            http.request(
                "GET",
                url_or_path,
                headers={"Accept": "application/rss+xml, application/atom+xml, text/xml, */*"},
                timeout=20,
                retries=2,
                raw=True,
            )
        )
    return Path(url_or_path).expanduser().read_text(encoding="utf-8")


def _parse_rss(root: ET.Element, *, feed_url: str, limit: int) -> list[dict[str, Any]]:
    channel = _first_child(root, "channel")
    if channel is None:
        return []
    feed_title = _child_text(channel, "title") or feed_url
    items = []
    for index, item in enumerate(_children(channel, "item")[:limit], start=1):
        title = _child_text(item, "title")
        link = _child_text(item, "link") or _child_text(item, "guid")
        description = _child_text(item, "description")
        published_at = _normalize_date(_child_text(item, "pubDate") or _child_text(item, "published"))
        source = _child_text(item, "source")
        normalized = _item_dict(
            title=title,
            url=link,
            excerpt=description,
            container=source or feed_title,
            published_at=published_at,
            feed_title=feed_title,
            feed_url=feed_url,
            position=index,
            guid=_child_text(item, "guid"),
        )
        if normalized:
            items.append(normalized)
    return items


def _parse_atom(root: ET.Element, *, feed_url: str, limit: int) -> list[dict[str, Any]]:
    feed_title = _child_text(root, "title") or feed_url
    items = []
    for index, entry in enumerate(_children(root, "entry")[:limit], start=1):
        title = _child_text(entry, "title")
        link = _atom_link(entry) or _child_text(entry, "id")
        excerpt = _child_text(entry, "summary") or _child_text(entry, "content")
        published_at = _normalize_date(_child_text(entry, "published") or _child_text(entry, "updated"))
        normalized = _item_dict(
            title=title,
            url=link,
            excerpt=excerpt,
            container=feed_title,
            published_at=published_at,
            feed_title=feed_title,
            feed_url=feed_url,
            position=index,
            guid=_child_text(entry, "id"),
        )
        if normalized:
            items.append(normalized)
    return items


def _item_dict(
    *,
    title: str | None,
    url: str | None,
    excerpt: str | None,
    container: str | None,
    published_at: str | None,
    feed_title: str,
    feed_url: str,
    position: int,
    guid: str | None,
) -> dict[str, Any] | None:
    clean_title = _clean_text(title)
    clean_excerpt = _clean_text(excerpt)
    if not clean_title and not clean_excerpt:
        return None
    item_url = (url or guid or "").strip()
    return {
        "id": guid or item_url or f"{feed_url}#{position}",
        "source": "major_feed",
        "title": clean_title or clean_excerpt[:120],
        "url": item_url,
        "author": None,
        "container": _clean_text(container) or feed_title,
        "published_at": published_at,
        "excerpt": clean_excerpt,
        "engagement": {},
        "metadata": {
            "feed_title": feed_title,
            "feed_url": feed_url,
            "feed_position": position,
        },
    }


def _atom_link(entry: ET.Element) -> str | None:
    for child in list(entry):
        if _local_name(child.tag) != "link":
            continue
        rel = child.attrib.get("rel")
        href = child.attrib.get("href")
        if href and rel in (None, "", "alternate"):
            return href.strip()
    return None


def _children(node: ET.Element, name: str) -> list[ET.Element]:
    target = name.lower()
    return [child for child in list(node) if _local_name(child.tag) == target]


def _first_child(node: ET.Element, name: str) -> ET.Element | None:
    target = name.lower()
    for child in list(node):
        if _local_name(child.tag) == target:
            return child
    return None


def _child_text(node: ET.Element, name: str) -> str | None:
    child = _first_child(node, name)
    if child is None or child.text is None:
        return None
    return child.text.strip()


def _local_name(tag: str) -> str:
    return tag.rsplit("}", 1)[-1].lower()


def _clean_text(value: str | None) -> str:
    if not value:
        return ""
    text = html.unescape(value)
    text = re.sub(r"(?is)<(script|style).*?</\1>", " ", text)
    text = re.sub(r"(?s)<[^>]+>", " ", text)
    text = re.sub(r"\s+", " ", text)
    return text.strip()


def _normalize_date(value: str | None) -> str | None:
    if not value:
        return None
    raw = value.strip()
    if not raw:
        return None
    if raw.endswith("Z"):
        raw = f"{raw[:-1]}+00:00"
    try:
        parsed = datetime.fromisoformat(raw)
        if parsed.tzinfo is None:
            parsed = parsed.replace(tzinfo=timezone.utc)
        return parsed.astimezone(timezone.utc).isoformat()
    except ValueError:
        pass
    try:
        parsed = parsedate_to_datetime(raw)
    except (TypeError, ValueError, IndexError, OverflowError):
        try:
            parsed = parsedate_to_datetime(raw.replace(" ET", " -0500").replace(" EDT", " -0400"))
        except (TypeError, ValueError, IndexError, OverflowError):
            return raw
    if parsed.tzinfo is None:
        parsed = parsed.replace(tzinfo=timezone.utc)
    return parsed.astimezone(timezone.utc).isoformat()
