#!/usr/bin/env python3
"""Novelty store for the newsjack monitoring engine."""

from __future__ import annotations

import json
import os
import sqlite3
from datetime import datetime, timezone
from pathlib import Path
from typing import Any


DB_PATH = Path.home() / ".local" / "share" / "newsjack" / "monitor.db"

SCHEMA = """
PRAGMA journal_mode=WAL;
PRAGMA synchronous=NORMAL;

CREATE TABLE IF NOT EXISTS seen_urls (
    url TEXT PRIMARY KEY,
    first_seen TEXT NOT NULL,
    last_seen TEXT NOT NULL,
    sighting_count INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS monitor_runs (
    id INTEGER PRIMARY KEY,
    monitor_name TEXT,
    profile_json TEXT,
    query_json TEXT NOT NULL,
    generated_at TEXT NOT NULL,
    signal_count INTEGER NOT NULL DEFAULT 0,
    created_at TEXT DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS signal_snapshots (
    id INTEGER PRIMARY KEY,
    run_id INTEGER REFERENCES monitor_runs(id) ON DELETE CASCADE,
    signal_id TEXT NOT NULL,
    title TEXT NOT NULL,
    rank_score REAL NOT NULL,
    payload_json TEXT NOT NULL,
    created_at TEXT DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_signal_snapshots_run ON signal_snapshots(run_id);
CREATE INDEX IF NOT EXISTS idx_signal_snapshots_rank ON signal_snapshots(rank_score DESC);
"""


def db_path_from_env() -> Path:
    return Path(os.environ.get("NEWSJACK_STORE", str(DB_PATH))).expanduser()


def _connect(db_path: Path | None = None) -> sqlite3.Connection:
    path = db_path or db_path_from_env()
    path.parent.mkdir(parents=True, exist_ok=True)
    conn = sqlite3.connect(str(path))
    conn.row_factory = sqlite3.Row
    return conn


def init_db(db_path: Path | None = None) -> Path:
    path = db_path or db_path_from_env()
    with _connect(path) as conn:
        conn.executescript(SCHEMA)
        conn.commit()
    return path


def seen_status(urls: list[str], db_path: Path | None = None) -> dict[str, dict[str, Any]]:
    init_db(db_path)
    urls = [url for url in urls if url]
    if not urls:
        return {}
    placeholders = ",".join("?" for _ in urls)
    with _connect(db_path) as conn:
        rows = conn.execute(
            f"SELECT url, first_seen, last_seen, sighting_count FROM seen_urls WHERE url IN ({placeholders})",
            urls,
        ).fetchall()
    return {
        row["url"]: {
            "first_seen": row["first_seen"],
            "last_seen": row["last_seen"],
            "sighting_count": row["sighting_count"],
        }
        for row in rows
    }


def mark_seen(urls: list[str], db_path: Path | None = None) -> None:
    init_db(db_path)
    now = datetime.now(timezone.utc).isoformat()
    with _connect(db_path) as conn:
        for url in [u for u in urls if u]:
            conn.execute(
                """
                INSERT INTO seen_urls (url, first_seen, last_seen, sighting_count)
                VALUES (?, ?, ?, 1)
                ON CONFLICT(url) DO UPDATE SET
                    last_seen = excluded.last_seen,
                    sighting_count = seen_urls.sighting_count + 1
                """,
                (url, now, now),
            )
        conn.commit()


def record_run(
    *,
    monitor_name: str | None,
    profile: dict[str, Any],
    queries: list[str],
    signals: list[dict[str, Any]],
    seen_urls: list[str] | None = None,
    db_path: Path | None = None,
) -> int:
    init_db(db_path)
    generated_at = datetime.now(timezone.utc).isoformat()
    with _connect(db_path) as conn:
        cursor = conn.execute(
            """
            INSERT INTO monitor_runs (monitor_name, profile_json, query_json, generated_at, signal_count)
            VALUES (?, ?, ?, ?, ?)
            """,
            (
                monitor_name,
                json.dumps(profile, sort_keys=True),
                json.dumps(queries),
                generated_at,
                len(signals),
            ),
        )
        run_id = int(cursor.lastrowid)
        for signal in signals:
            conn.execute(
                """
                INSERT INTO signal_snapshots (run_id, signal_id, title, rank_score, payload_json)
                VALUES (?, ?, ?, ?, ?)
                """,
                (
                    run_id,
                    signal.get("id"),
                    signal.get("title"),
                    float((signal.get("scores") or {}).get("rank", 0.0)),
                    json.dumps(signal, sort_keys=True),
                ),
            )
        mark_urls = seen_urls
        if mark_urls is None:
            mark_urls = [
                evidence.get("url")
                for signal in signals
                for evidence in signal.get("evidence", [])
                if evidence.get("url")
            ]
        for url in dict.fromkeys(url for url in mark_urls if url):
            conn.execute(
                """
                INSERT INTO seen_urls (url, first_seen, last_seen, sighting_count)
                VALUES (?, ?, ?, 1)
                ON CONFLICT(url) DO UPDATE SET
                    last_seen = excluded.last_seen,
                    sighting_count = seen_urls.sighting_count + 1
                """,
                (url, generated_at, generated_at),
            )
        conn.commit()
    return run_id


def recent_runs(limit: int = 10, db_path: Path | None = None) -> list[dict[str, Any]]:
    init_db(db_path)
    with _connect(db_path) as conn:
        rows = conn.execute(
            """
            SELECT id, monitor_name, query_json, generated_at, signal_count
            FROM monitor_runs
            ORDER BY id DESC
            LIMIT ?
            """,
            (limit,),
        ).fetchall()
    return [
        {
            "id": row["id"],
            "monitor_name": row["monitor_name"],
            "queries": json.loads(row["query_json"]),
            "generated_at": row["generated_at"],
            "signal_count": row["signal_count"],
        }
        for row in rows
    ]
