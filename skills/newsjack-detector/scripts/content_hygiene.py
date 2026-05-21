"""Deterministic hygiene filters for obvious non-news retrieval results."""

from __future__ import annotations

import re
from urllib.parse import urlparse


SOCIAL_SOURCES = {"x", "x_news", "x_trends", "reddit", "hackernews"}

DOCS_HOST_PREFIXES = (
    "docs.",
    "doc.",
    "help.",
    "support.",
    "developer.",
    "developers.",
)

DOCS_PATH_PARTS = {
    "docs",
    "documentation",
    "help",
    "support",
    "kb",
    "knowledge-base",
    "api",
    "api-reference",
    "reference",
    "manual",
    "guide",
    "guides",
    "tutorial",
    "tutorials",
    "connector",
    "connectors",
    "integration",
    "integrations",
}

PRODUCT_PATH_PARTS = {
    "product",
    "products",
    "shop",
    "store",
    "cart",
    "checkout",
    "pricing",
    "plans",
    "marketplace",
    "app-store",
    "apps",
}

SEO_PATH_PATTERNS = (
    r"\bsell[-_]house[-_]fast\b",
    r"\bsell[-_]my[-_]house[-_]fast\b",
    r"\bcash[-_]house[-_]buyers?\b",
    r"\bwe[-_]buy[-_]houses?\b",
    r"\bquick[-_]house[-_]sale\b",
    r"\bbest[-_][a-z0-9-]+",
)

SEO_TITLE_PATTERNS = (
    r"\bbest\s+\d+\b",
    r"\bbest\s+[a-z0-9\s-]+\s+(tools|services|companies|platforms)\b",
    r"\bhow to sell (your )?house fast\b",
    r"\bcash house buyers?\b",
)


def rejection_reason(
    *,
    source: str,
    title: str,
    url: str,
    container: str | None = None,
    excerpt: str = "",
) -> str | None:
    """Return a deterministic rejection reason for obvious non-news pages."""
    if source in SOCIAL_SOURCES:
        return None

    parsed = urlparse(url or "")
    host = (parsed.hostname or "").lower()
    path_parts = [part.strip() for part in parsed.path.lower().split("/") if part.strip()]
    path_text = parsed.path.lower()
    title_text = (title or "").lower()
    combined_text = " ".join([title_text, (container or "").lower(), (excerpt or "").lower()])

    if host.startswith(DOCS_HOST_PREFIXES) or "readthedocs.io" in host:
        return "owned_docs_or_help"
    if any(part in DOCS_PATH_PARTS for part in path_parts):
        return "owned_docs_or_help"
    if re.search(r"\b(documentation|api reference|developer docs|help center|support article)\b", combined_text):
        return "owned_docs_or_help"

    if any(part in PRODUCT_PATH_PARTS for part in path_parts):
        return "product_or_ecommerce_page"
    if re.search(r"\b(add to cart|buy now|pricing plans|product page|shop now)\b", combined_text):
        return "product_or_ecommerce_page"

    if any(re.search(pattern, path_text) for pattern in SEO_PATH_PATTERNS):
        return "seo_landing_page"
    if any(re.search(pattern, title_text) for pattern in SEO_TITLE_PATTERNS):
        return "seo_landing_page"

    return None
