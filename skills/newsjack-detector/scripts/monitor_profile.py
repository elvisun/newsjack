"""Monitor profile parsing for newsjack evidence collection."""

from __future__ import annotations

import json
import re
from dataclasses import dataclass, field
from pathlib import Path
from typing import Any


@dataclass(frozen=True)
class MonitorProfile:
    company: str | None = None
    description: str | None = None
    topics: list[str] = field(default_factory=list)
    competitors: list[str] = field(default_factory=list)
    search_terms: list[str] = field(default_factory=list)
    feed_urls: list[str] = field(default_factory=list)
    x_news: dict[str, Any] = field(default_factory=dict)
    x_trends: dict[str, Any] = field(default_factory=dict)
    spokespeople: list[str] = field(default_factory=list)
    proof_assets: list[str] = field(default_factory=list)
    standing: list[str] = field(default_factory=list)
    exclusions: list[str] = field(default_factory=list)
    raw: dict[str, Any] = field(default_factory=dict)

    @classmethod
    def from_file(cls, path: str | Path) -> "MonitorProfile":
        with Path(path).expanduser().open("r", encoding="utf-8") as handle:
            payload = json.load(handle)
        if not isinstance(payload, dict):
            raise ValueError("monitor profile must be a JSON object")
        return cls.from_dict(payload)

    @classmethod
    def from_dict(cls, payload: dict[str, Any]) -> "MonitorProfile":
        return cls(
            company=_string_or_none(payload.get("company") or payload.get("client")),
            description=_string_or_none(payload.get("description")),
            topics=_string_list(payload.get("topics") or payload.get("beats")),
            competitors=_string_list(payload.get("competitors")),
            search_terms=_string_list(payload.get("search_terms") or payload.get("queries")),
            feed_urls=_string_list(payload.get("feed_urls") or payload.get("feeds") or payload.get("rss_feeds")),
            x_news=_dict_value(payload.get("x_news"), default={"enabled": True}),
            x_trends=_dict_value(payload.get("x_trends"), default={"mode": "none", "woeids": [], "locations": []}),
            spokespeople=_string_list(payload.get("spokespeople") or payload.get("experts")),
            proof_assets=_string_list(payload.get("proof_assets") or payload.get("proof")),
            standing=_string_list(payload.get("standing") or payload.get("expertise")),
            exclusions=_string_list(payload.get("exclusions") or payload.get("do_not_newsjack")),
            raw=payload,
        )

    def query_terms(self) -> list[str]:
        if self.search_terms:
            return _dedupe(self.search_terms)
        return _dedupe([*self.topics, *[_competitor_query(term) for term in self.competitors]])

    def match_text(self) -> str:
        parts = [
            self.company or "",
            self.description or "",
            *self.topics,
            *self.competitors,
            *self.feed_urls,
            *(_string_list(self.x_trends.get("locations")) if self.x_trends else []),
            *self.spokespeople,
            *self.proof_assets,
            *self.standing,
        ]
        return " ".join(part for part in parts if part).strip()

    def public_dict(self) -> dict[str, Any]:
        return {
            "company": self.company,
            "topics": self.topics,
            "competitors": self.competitors,
            "search_terms": self.search_terms,
            "feed_urls": self.feed_urls,
            "x_news": self.x_news,
            "x_trends": self.x_trends,
            "spokespeople": self.spokespeople,
            "proof_assets": self.proof_assets,
            "standing": self.standing,
            "exclusions": self.exclusions,
        }


def _string_or_none(value: Any) -> str | None:
    if value is None:
        return None
    text = str(value).strip()
    return text or None


def _dict_value(value: Any, *, default: dict[str, Any]) -> dict[str, Any]:
    if isinstance(value, dict):
        return dict(value)
    return dict(default)


def _string_list(value: Any) -> list[str]:
    if not value:
        return []
    if isinstance(value, str):
        return [item.strip() for item in re.split(r"[,;\n]", value) if item.strip()]
    if isinstance(value, list):
        output = []
        for item in value:
            if isinstance(item, dict):
                output.extend(str(v).strip() for v in item.values() if str(v).strip())
            elif str(item).strip():
                output.append(str(item).strip())
        return output
    return [str(value).strip()]


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


def _competitor_query(value: str) -> str:
    term = value.strip()
    if not term or term.startswith('"') or " " not in term:
        return term
    escaped = term.replace('"', r'\"')
    return f'"{escaped}"'
