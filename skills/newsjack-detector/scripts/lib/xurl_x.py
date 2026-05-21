"""X (Twitter) search via xurl CLI — official X API v2 with OAuth2.

xurl is an open-source CLI for the X API (https://github.com/openclaw/xurl).
It uses OAuth2 with PKCE and automatic token refresh, requiring only a free
X Developer App. No xAI subscription or browser cookies needed.

Install: npm install -g xurl
Auth:    xurl auth oauth2 login

This is the default X source for newsjack v0.
"""

import json
import os
import re
import subprocess
import sys
import urllib.parse
import urllib.error
import urllib.request
from datetime import datetime, timedelta, timezone
from typing import Any, Dict, List, Optional

from .relevance import token_overlap_relevance as _compute_relevance


def _log(msg: str) -> None:
    sys.stderr.write(f"[xurl] {msg}\n")
    sys.stderr.flush()


# Depth configurations: number of results to request
DEPTH_CONFIG = {
    "quick": 10,
    "default": 30,
    "deep": 60,
}

TWEET_FIELDS = "created_at,public_metrics,author_id,conversation_id,referenced_tweets,lang,possibly_sensitive"
USER_FIELDS = "username,name,verified,is_identity_verified,public_metrics"
SEARCH_EXPANSIONS = "author_id"

DEFAULT_MIN_ENGAGEMENT = 3
DEFAULT_MIN_AUTHOR_FOLLOWERS = 2000
DEFAULT_MIN_VIEWS = 1000
DEFAULT_TREND_MIN_24H = 25
DEFAULT_TREND_MIN_6H = 8
DEFAULT_TREND_MIN_VELOCITY = 2.0
NEWS_FIELDS = "id,name,summary,hook,contexts,cluster_posts_results,updated_at,keywords,category"
PERSONALIZED_TREND_FIELDS = "trend_name,post_count,category,trending_since"


def is_available() -> bool:
    """Check if xurl is installed and has valid authentication.

    Returns True only if xurl binary is found AND the user is authenticated
    (i.e. ``xurl whoami`` exits 0 and returns a username field).
    """
    try:
        result = subprocess.run(
            ["xurl", "whoami"],
            capture_output=True,
            text=True,
            timeout=10,
        )
        return result.returncode == 0 and '"username"' in result.stdout
    except (OSError, subprocess.TimeoutExpired):
        # OSError covers FileNotFoundError (no xurl on PATH) and
        # PermissionError (a non-executable match on PATH, e.g. WSL's
        # /mnt/c/.../WindowsApps shim returning EACCES on exec).
        return False


def search_x(
    query: str,
    depth: str = "default",
) -> Dict[str, Any]:
    """Search X via xurl CLI using X API v2 search/recent.

    Args:
        query: Search query string
        depth: "quick", "default", or "deep"

    Returns:
        Raw JSON response from X API v2 tweets/search/recent, or a dict
        with an "error" key on failure.
    """
    max_results = DEPTH_CONFIG.get(depth, DEPTH_CONFIG["default"])
    # X API v2 search/recent requires max_results in 10–100 range
    max_results = max(10, min(100, max_results))

    normalized_query = _normalize_query(query)
    params = {
        "query": normalized_query,
        "max_results": str(max_results),
        "sort_order": "relevancy",
        "tweet.fields": TWEET_FIELDS,
        "expansions": SEARCH_EXPANSIONS,
        "user.fields": USER_FIELDS,
    }
    response = _xurl_get(f"/2/tweets/search/recent?{urllib.parse.urlencode(params)}")
    if "error" not in response:
        response["_newsjack_query"] = normalized_query
        return response

    # Older xurl builds may not handle raw URL query strings the same way as
    # shortcut commands. Fall back to the shortcut, then filter locally.
    fallback = _search_x_shortcut(query, max_results=max_results)
    if "error" in fallback:
        return response
    fallback["_newsjack_query"] = query
    return fallback


def recent_count_summary(query: str, *, hours: int = 24, bearer_token: str | None = None) -> Dict[str, Any] | None:
    """Return query-volume summary from X recent counts.

    Counts are advisory. If the endpoint is unavailable, search still works and
    individual posts are filtered by reach/engagement.
    """
    now = datetime.now(timezone.utc).replace(microsecond=0)
    start = now - timedelta(hours=hours)
    params = {
        "query": _normalize_query(query),
        "granularity": "hour",
        "start_time": _iso_z(start),
        "end_time": _iso_z(now),
    }
    path = f"/2/tweets/counts/recent?{urllib.parse.urlencode(params)}"
    response = _api_get(path, bearer_token=bearer_token, timeout=30) if bearer_token else _xurl_get(path, timeout=30, auth="app")
    if "error" in response:
        if os.environ.get("NEWSJACK_DEBUG"):
            _log(f"Counts unavailable: {response['error']}")
        return None
    return _summarize_counts(response)


def search_x_news(
    query: str,
    *,
    depth: str = "default",
    max_age_hours: int = 168,
    bearer_token: str | None = None,
) -> Dict[str, Any]:
    """Search X News story clusters.

    X News is a better discovery shape than raw recent posts: it returns
    clustered stories with summaries, hooks, entities, and representative post
    IDs. Treat summaries as discovery context, not final fact-checking.
    """
    max_results = {"quick": 5, "default": 10, "deep": 20}.get(depth, 10)
    params = {
        "query": query.strip(),
        "max_results": str(max_results),
        "max_age_hours": str(max_age_hours),
        "news.fields": NEWS_FIELDS,
    }
    path = f"/2/news/search?{urllib.parse.urlencode(params)}"
    if bearer_token:
        return _api_get(path, bearer_token=bearer_token)
    return _xurl_get(path, headers={"Accept-Language": "en-US,en;q=0.9"})


def parse_x_news_response(response: Dict[str, Any], *, topic: str = "") -> List[Dict[str, Any]]:
    items: List[Dict[str, Any]] = []
    if "error" in response:
        if os.environ.get("NEWSJACK_DEBUG"):
            _log(f"X News unavailable: {response['error']}")
        return items

    for index, story in enumerate(response.get("data") or []):
        name = str(story.get("name") or "").strip()
        hook = str(story.get("hook") or "").strip()
        summary = str(story.get("summary") or "").strip()
        if not name and not hook and not summary:
            continue

        post_ids = _story_post_ids(story)
        contexts = story.get("contexts") or {}
        text = " ".join(part for part in [name, hook, summary] if part)
        url = f"https://x.com/search?q={urllib.parse.quote(name or topic)}&f=live"
        engagement = {"score": min(500, max(1, len(post_ids)) * 10)}
        if post_ids:
            engagement["comments"] = len(post_ids)
        items.append({
            "id": f"XNEWS{index + 1}",
            "title": name or hook[:120],
            "text": text[:1200],
            "url": url,
            "author_handle": "x-news",
            "date": story.get("updated_at"),
            "engagement": engagement,
            "metadata": {
                "x_signal_type": "story_cluster",
                "x_news_id": story.get("id") or story.get("rest_id"),
                "x_news_category": story.get("category"),
                "x_news_keywords": story.get("keywords") or [],
                "x_news_contexts": contexts,
                "x_news_cluster_post_ids": post_ids,
                "x_news_cluster_post_count": len(post_ids),
                "x_news_disclaimer": story.get("disclaimer"),
            },
            "why_relevant": "X News story cluster",
            "relevance": _compute_relevance(topic, text) if topic else 0.5,
        })
    return items


def collect_x_trends(
    trends_config: Dict[str, Any] | None,
    *,
    depth: str = "default",
    bearer_token: str | None = None,
) -> tuple[List[Dict[str, Any]], str | None]:
    config = trends_config or {}
    mode = str(config.get("mode") or "none").strip().lower()
    if mode in {"", "none", "off", "false"}:
        return [], None

    try:
        if mode == "personalized":
            response = personalized_trends(depth=depth)
            return parse_trends_response(response, mode=mode), _response_error(response)
        if mode == "location":
            if not bearer_token:
                return [], "x_trends location mode requires TWITTER_BEARER_TOKEN or X_BEARER_TOKEN"
            items: List[Dict[str, Any]] = []
            errors = []
            locations = [str(value) for value in config.get("locations") or []]
            for index, woeid in enumerate(config.get("woeids") or []):
                response = woeid_trends(str(woeid), depth=depth, bearer_token=bearer_token)
                error = _response_error(response)
                if error:
                    errors.append(f"{woeid}: {error}")
                    continue
                location = locations[index] if index < len(locations) else str(woeid)
                items.extend(parse_trends_response(response, mode=mode, woeid=str(woeid), location=location))
            return items, "; ".join(errors) if errors else None
        return [], f"Unsupported x_trends mode: {mode}"
    except Exception as exc:
        return [], f"{type(exc).__name__}: {exc}"


def personalized_trends(*, depth: str = "default") -> Dict[str, Any]:
    params = {
        "personalized_trend.fields": PERSONALIZED_TREND_FIELDS,
    }
    return _xurl_get(f"/2/users/personalized_trends?{urllib.parse.urlencode(params)}")


def woeid_trends(woeid: str, *, depth: str = "default", bearer_token: str) -> Dict[str, Any]:
    max_results = {"quick": 10, "default": 20, "deep": 50}.get(depth, 20)
    params = {"max_trends": str(max_results)}
    return _api_get(f"/2/trends/by/woeid/{urllib.parse.quote(str(woeid))}?{urllib.parse.urlencode(params)}", bearer_token=bearer_token)


def parse_trends_response(
    response: Dict[str, Any],
    *,
    mode: str,
    woeid: str | None = None,
    location: str | None = None,
) -> List[Dict[str, Any]]:
    items: List[Dict[str, Any]] = []
    if "error" in response:
        return items
    for index, trend in enumerate(response.get("data") or []):
        name = str(trend.get("trend_name") or "").strip()
        if not name:
            continue
        count_text = str(trend.get("post_count") or trend.get("tweet_count") or "").strip()
        count = _parse_count_text(count_text)
        context = ", ".join(part for part in [trend.get("category"), count_text, trend.get("trending_since"), location] if part)
        items.append({
            "id": f"XTREND{index + 1}",
            "title": name,
            "text": f"{name}. {context}".strip(),
            "url": f"https://x.com/search?q={urllib.parse.quote(name)}&f=live",
            "author_handle": "x-trends",
            "date": datetime.now(timezone.utc).replace(microsecond=0).isoformat(),
            "engagement": {"score": min(500, count)} if count else {},
            "metadata": {
                "x_signal_type": "trend",
                "x_trend_mode": mode,
                "x_trend_woeid": woeid,
                "x_trend_location": location,
                "x_trend_category": trend.get("category"),
                "x_trend_post_count": count_text,
                "x_trend_since": trend.get("trending_since"),
            },
            "why_relevant": "X trend",
            "relevance": 0.5,
        })
    return items


def parse_x_response(
    response: Dict[str, Any],
    topic: str = "",
    counts_summary: Dict[str, Any] | None = None,
) -> List[Dict[str, Any]]:
    """Parse xurl search response into normalized item dicts.

    Output format matches the monitor engine's generic X item shape:
    id, text, url, author_handle, date, engagement, why_relevant, relevance.

    Args:
        response: Raw X API v2 response dict from search_x()
        topic: Original search topic (used for relevance scoring)

    Returns:
        List of item dicts.  Empty list on error or no results.
    """
    items: List[Dict[str, Any]] = []

    if "error" in response:
        _log(f"Error in response: {response['error']}")
        return items

    data = response.get("data") or []
    if counts_summary and counts_summary.get("is_trending"):
        items.append(_trend_item(topic, counts_summary))
    if not data:
        return items

    # Build author lookup from includes.users
    authors: Dict[str, Dict[str, Any]] = {}
    for user in (response.get("includes") or {}).get("users") or []:
        authors[user["id"]] = user

    for i, tweet in enumerate(data):
        author_id = tweet.get("author_id", "")
        author = authors.get(author_id, {})
        username = author.get("username", "")
        author_metrics = author.get("public_metrics") or {}

        tweet_id = tweet.get("id", "")
        url = f"https://x.com/{username}/status/{tweet_id}" if username else ""

        # Parse public_metrics
        engagement: Dict[str, Any] = {}
        metrics = tweet.get("public_metrics") or {}
        if metrics:
            engagement = {
                "likes": metrics.get("like_count", 0),
                "reposts": metrics.get("retweet_count", 0),
                "replies": metrics.get("reply_count", 0),
                "quotes": metrics.get("quote_count", 0),
                "bookmarks": metrics.get("bookmark_count", 0),
                "views": metrics.get("impression_count", 0),
            }

        verified = bool(author.get("verified") or author.get("is_identity_verified"))
        social_proof = _social_proof(metrics, author_metrics, verified=verified)

        # Preserve timestamp precision when X returns ISO 8601.
        date: Optional[str] = None
        created = tweet.get("created_at", "")
        if created:
            date = created
            if not re.match(r"\d{4}-\d{2}-\d{2}T", created):
                m = re.match(r"(\d{4}-\d{2}-\d{2})", created)
                if m:
                    date = m.group(1)

        text = tweet.get("text", "").strip()

        # Relevance score via shared token-overlap function
        relevance = _compute_relevance(topic, text) if topic else 0.5

        items.append({
            "id": f"XURL{i + 1}",
            "text": text[:500],
            "url": url,
            "author_handle": username,
            "date": date,
            "engagement": engagement,
            "metadata": {
                "x_signal_type": "post",
                "x_author_followers": _int_metric(author_metrics, "followers_count"),
                "x_author_listed": _int_metric(author_metrics, "listed_count"),
                "x_author_verified": verified,
                "x_low_reach": not social_proof,
                "x_social_proof": social_proof,
                "x_query_counts": counts_summary,
            },
            "why_relevant": "",
            "relevance": relevance,
        })

    return items


def keep_x_item(item: Dict[str, Any]) -> bool:
    metadata = item.get("metadata") or {}
    if metadata.get("x_signal_type") == "query_trend":
        return True
    return not metadata.get("x_low_reach")


def _xurl_get(
    path: str,
    timeout: int = 30,
    auth: str | None = None,
    headers: Dict[str, str] | None = None,
) -> Dict[str, Any]:
    command = ["xurl"]
    if auth:
        command.extend(["--auth", auth])
    for key, value in (headers or {}).items():
        command.extend(["-H", f"{key}: {value}"])
    command.append(path)
    try:
        result = subprocess.run(
            command,
            capture_output=True,
            text=True,
            timeout=timeout,
        )
        if result.returncode != 0:
            error_text = _clean_error(result.stderr.strip() or result.stdout.strip())
            return {"error": f"xurl request failed: {error_text}"}
        return json.loads(result.stdout)
    except FileNotFoundError:
        return {"error": "xurl not found in PATH"}
    except subprocess.TimeoutExpired:
        return {"error": f"xurl request timed out ({timeout}s)"}
    except json.JSONDecodeError as exc:
        return {"error": f"Invalid JSON from xurl: {exc}"}
    except Exception as exc:
        return {"error": f"{type(exc).__name__}: {exc}"}


def _api_get(path: str, *, bearer_token: str, timeout: int = 30) -> Dict[str, Any]:
    request = urllib.request.Request(
        f"https://api.x.com{path}",
        headers={
            "Authorization": f"Bearer {bearer_token}",
            "Accept": "application/json",
            "Accept-Language": "en-US,en;q=0.9",
        },
    )
    try:
        with urllib.request.urlopen(request, timeout=timeout) as response:
            return json.loads(response.read().decode("utf-8"))
    except urllib.error.HTTPError as exc:
        body = exc.read().decode("utf-8", errors="replace")
        try:
            payload = json.loads(body)
            detail = payload.get("detail") or payload.get("title") or body[:300]
            reason = payload.get("reason")
            if reason:
                detail = f"{detail} ({reason})"
        except json.JSONDecodeError:
            detail = body[:300]
        return {"error": f"X API HTTP {exc.code}: {detail}"}
    except json.JSONDecodeError as exc:
        return {"error": f"Invalid JSON from X API: {exc}"}
    except Exception as exc:
        return {"error": f"{type(exc).__name__}: {exc}"}


def _response_error(response: Dict[str, Any]) -> str | None:
    return response.get("error") if isinstance(response, dict) else "invalid response"


def _search_x_shortcut(query: str, *, max_results: int) -> Dict[str, Any]:
    try:
        result = subprocess.run(
            ["xurl", "search", query, "-n", str(max_results)],
            capture_output=True,
            text=True,
            timeout=30,
        )
        if result.returncode != 0:
            error_text = _clean_error(result.stderr.strip() or result.stdout.strip())
            return {"error": f"xurl search failed: {error_text}"}
        return json.loads(result.stdout)
    except FileNotFoundError:
        return {"error": "xurl not found in PATH"}
    except subprocess.TimeoutExpired:
        return {"error": "xurl search timed out (30s)"}
    except json.JSONDecodeError as exc:
        return {"error": f"Invalid JSON from xurl: {exc}"}
    except Exception as exc:
        return {"error": f"{type(exc).__name__}: {exc}"}


def _normalize_query(query: str) -> str:
    output = query.strip()
    additions = []
    if "lang:" not in output:
        additions.append("lang:en")
    for operator in ("-is:retweet", "-is:reply", "-is:nullcast"):
        if operator not in output and operator[1:] not in output:
            additions.append(operator)
    if additions:
        output = f"{output} {' '.join(additions)}"
    return output[:512]


def _clean_error(text: str) -> str:
    return re.sub(r"\x1b\[[0-?]*[ -/]*[@-~]", "", text).strip()


def _story_post_ids(story: Dict[str, Any]) -> List[str]:
    seen = set()
    output = []
    for raw in story.get("cluster_posts_results") or []:
        post_id = str(raw.get("post_id") or raw.get("id") or "").strip()
        if not post_id or post_id in seen:
            continue
        seen.add(post_id)
        output.append(post_id)
    return output


def _parse_count_text(value: str) -> int:
    text = str(value or "").strip().lower().replace(",", "")
    if not text:
        return 0
    match = re.search(r"(\d+(?:\.\d+)?)\s*([km]?)", text)
    if not match:
        return 0
    number = float(match.group(1))
    suffix = match.group(2)
    if suffix == "k":
        number *= 1_000
    elif suffix == "m":
        number *= 1_000_000
    return int(number)


def _social_proof(
    metrics: Dict[str, Any],
    author_metrics: Dict[str, Any],
    *,
    verified: bool,
) -> List[str]:
    proof = []
    likes = _int_metric(metrics, "like_count")
    reposts = _int_metric(metrics, "retweet_count")
    replies = _int_metric(metrics, "reply_count")
    quotes = _int_metric(metrics, "quote_count")
    views = _int_metric(metrics, "impression_count")
    followers = _int_metric(author_metrics, "followers_count")
    listed = _int_metric(author_metrics, "listed_count")
    engagement_total = likes + reposts + replies + quotes
    has_view_metric = "impression_count" in metrics

    min_engagement = _env_int("NEWSJACK_X_MIN_ENGAGEMENT", DEFAULT_MIN_ENGAGEMENT)
    min_followers = _env_int("NEWSJACK_X_MIN_AUTHOR_FOLLOWERS", DEFAULT_MIN_AUTHOR_FOLLOWERS)
    min_views = _env_int("NEWSJACK_X_MIN_VIEWS", DEFAULT_MIN_VIEWS)

    if engagement_total >= min_engagement:
        proof.append("post_engagement")
    if reposts or quotes:
        proof.append("reshared")
    if views >= min_views:
        proof.append("views")
    if followers >= min_followers and not has_view_metric:
        proof.append("author_followers")
    if listed >= 25 and not has_view_metric:
        proof.append("author_listed")
    if verified and followers >= min_followers and not has_view_metric:
        proof.append("verified_author")
    return proof


def _summarize_counts(response: Dict[str, Any]) -> Dict[str, Any]:
    buckets = response.get("data") or []
    counts = [_int_value(bucket.get("tweet_count")) for bucket in buckets]
    total_24h = sum(counts)
    recent_6h = sum(counts[-6:])
    previous = counts[:-6]
    recent_per_hour = recent_6h / 6.0 if counts else 0.0
    previous_per_hour = (sum(previous) / len(previous)) if previous else 0.0
    velocity = (recent_per_hour + 1.0) / (previous_per_hour + 1.0)

    min_24h = _env_int("NEWSJACK_X_TREND_MIN_24H", DEFAULT_TREND_MIN_24H)
    min_6h = _env_int("NEWSJACK_X_TREND_MIN_6H", DEFAULT_TREND_MIN_6H)
    min_velocity = _env_float("NEWSJACK_X_TREND_MIN_VELOCITY", DEFAULT_TREND_MIN_VELOCITY)
    is_trending = total_24h >= min_24h and (recent_6h >= min_6h or velocity >= min_velocity)

    return {
        "total_24h": total_24h,
        "recent_6h": recent_6h,
        "previous_hourly_avg": round(previous_per_hour, 2),
        "velocity": round(velocity, 2),
        "is_trending": is_trending,
        "bucket_count": len(counts),
    }


def _trend_item(topic: str, counts: Dict[str, Any]) -> Dict[str, Any]:
    query = topic.strip()
    text = (
        f'X conversation volume for "{query}" is elevated: '
        f'{counts.get("recent_6h", 0)} posts in the last 6h, '
        f'{counts.get("total_24h", 0)} in the last 24h '
        f'(velocity {counts.get("velocity", 0)}x).'
    )
    return {
        "id": "X-TREND",
        "text": text,
        "url": f"https://x.com/search?q={urllib.parse.quote(query)}&f=live",
        "author_handle": "x-search",
        "date": datetime.now(timezone.utc).replace(microsecond=0).isoformat(),
        "engagement": {
            "score": min(500, _int_value(counts.get("total_24h"))),
            "comments": _int_value(counts.get("recent_6h")),
        },
        "metadata": {
            "x_signal_type": "query_trend",
            "x_social_proof": ["query_volume"],
            "x_query_counts": counts,
        },
        "why_relevant": "X query-volume trend",
        "relevance": 0.7,
    }


def _iso_z(value: datetime) -> str:
    return value.astimezone(timezone.utc).isoformat().replace("+00:00", "Z")


def _int_metric(metrics: Dict[str, Any], key: str) -> int:
    return _int_value(metrics.get(key))


def _int_value(value: Any) -> int:
    try:
        return int(value or 0)
    except (TypeError, ValueError):
        return 0


def _env_int(key: str, default: int) -> int:
    try:
        return int(os.environ.get(key, default))
    except (TypeError, ValueError):
        return default


def _env_float(key: str, default: float) -> float:
    try:
        return float(os.environ.get(key, default))
    except (TypeError, ValueError):
        return default
