#!/usr/bin/env python3
"""Compile final_report.md for a detector run dir — the 3-bucket scan.

Two axes drive the shape: STANDING (can the client act?) and MAGNITUDE (how big
is the public story?). They are kept separate so a big story is never buried for
lack of standing and a small story with standing is never lost.

Sections:
  ✅ Pitch-Ready          — fresh + real standing + angles (act on this)
  🔥 Big Stories Worth a Look — fresh + high/major story_size, suggestions only,
                            sorted by coverage spread, NEVER dropped (your call)
  👀 Watch / Context      — stale / freshness-unverified / small-off-beat, disclosed

Every link carries a date and a provenance tag, so worker-introduced links are
never laundered into the report as established coverage. Usage: build_report.py RUN_DIR
"""
import json, os, re, sys
from collections import defaultdict

CLIENTS = {"bluebottle": "Blue Bottle Coffee", "clearnym": "Clearnym", "localfalcon": "Local Falcon",
           "nofar-method": "Nofar Method", "property-saviour": "Property Saviour", "simular": "Simular", "slite": "Slite"}
# Domains that are aggregators / low-authority republishers, not a source of record.
AGGREGATORS = ("kucoin.com", "startupfortune.com", "vocal.media", "x.com", "twitter.com",
               "bitget.com", "stocktitan", "markets.businessinsider", "einpresswire", "manilatimes")
GATE_LABEL = {
    "stale": "Stale — older than 24h, or a newer pickup of an older original",
    "unverified_no_corroboration": "Unverified — fewer than 2 independent corroborating sources",
    "unverified_no_timestamp": "Unverified — no first-public timestamp recovered",
    "unverified_boundary": "Unverified — date-only clock straddling the cutoff",
}
# Map the coarse worker's weakness reason to a human confidence flag on a big-story suggestion.
WEAKNESS_FLAG = {
    "keyword_collision": "⚠ possible keyword match",
    "seo_landing_page": "⚠ looks like an SEO / landing page",
    "owned_docs_or_product_page": "⚠ product / docs page",
    "not_news": "⚠ may not be news reporting",
    "off_beat": "⚠ off-beat for this client",
    "no_profile_bridge": "⚠ no obvious client bridge",
    "major_news_no_bridge": "⚠ big news, no obvious client bridge",
}
BAND_LABEL = {"major": "major story", "high": "large story", "moderate": "moderate story", "low": "small story"}


def load(run, name):
    p = os.path.join(run, name)
    return json.load(open(p)) if os.path.exists(p) else None


def norm_url(u):
    u = (u or "").strip().lower()
    u = re.sub(r"^https?://", "", u)
    u = re.sub(r"[?#].*$", "", u)
    return u.rstrip("/")


def domain(u):
    m = re.sub(r"^https?://", "", (u or "").strip().lower())
    return m.split("/")[0].removeprefix("www.")


def mdlink(t, u):
    t = (t or "").strip().replace("[", "(").replace("]", ")") or domain(u)
    return f"[{t}]({u})" if u else (t or "")


def evidence_index(candidates):
    """signal_id -> list of surfaced evidence dicts (real provenance)."""
    idx = {}
    for s in (candidates.get("signals") or []):
        idx[s["id"]] = s.get("evidence") or []
    return idx


def main_and_related(rep, dups, ev_idx, story_origin):
    """Return (main_link, related_links, surfaced source-domain count)."""
    rep_ev = ev_idx.get(rep["id"], rep.get("evidence") or [])
    surfaced = []
    for e in rep_ev:
        surfaced.append({"title": e.get("title") or e.get("container"), "url": e.get("url"),
                         "src": e.get("container") or e.get("source"), "date": e.get("published_at"),
                         "prov": "surfaced"})
    related = []
    for d in dups:
        for u in (d.get("evidence_urls") or []):
            de = next((e for e in ev_idx.get(d.get("signal_id"), []) if e.get("url") == u), {})
            related.append({"title": d.get("signal_title"), "url": u,
                            "src": de.get("container") or (d.get("sources") or ["?"])[0],
                            "date": de.get("published_at"), "prov": "surfaced duplicate"})
    surfaced_urls = {norm_url(x["url"]) for x in surfaced + related if x["url"]}
    main = None
    for x in surfaced:
        if x["url"] and not any(a in domain(x["url"]) for a in AGGREGATORS):
            main = x
            break
    if main is None and surfaced:
        main = surfaced[0]
    for x in surfaced:
        if x is not main:
            related.append(x)
    so = story_origin or {}
    for url, datekey, label in [
        (so.get("canonical_coverage_url"), so.get("canonical_coverage_published_at"), so.get("canonical_coverage_source") or "canonical"),
        (so.get("original_url"), so.get("first_public_at"), so.get("original_source") or "original"),
    ]:
        if not url:
            continue
        if norm_url(url) in surfaced_urls:
            continue
        if any(norm_url(url) == norm_url(r["url"]) for r in related):
            continue
        related.append({"title": label, "url": url, "src": domain(url), "date": datekey,
                        "prov": "proposed by research — UNVERIFIED"})
    src_domains = {domain(x["url"]) for x in surfaced if x["url"]} | {domain(r["url"]) for r in related if r["url"] and r["prov"].startswith("surfaced")}
    return main, related, len(src_domains)


def fmt_link(x):
    date = x.get("date") or "no date"
    tag = x["prov"]
    if tag.startswith("proposed"):
        date = f"{date} — date unverified"
    dom = domain(x["url"]) if x.get("url") else x.get("src")
    return f"{mdlink(x.get('title'), x.get('url'))} — {dom}, {date} · `{tag}`"


def freshness_line(s, so):
    gate = (s.get("freshness_gate") or {}).get("computed_status")
    fp = so.get("first_public_at")
    if gate == "fresh_new_development":
        return f"- **Freshness:** `{gate}` — first public **{fp}**; new development **{so.get('new_development_at')}**: {so.get('new_development')}"
    return f"- **Freshness:** `{gate}` — first public **{fp}**"


def source_lines(s, dups, ev_idx, so):
    main, related, nsrc = main_and_related(s, dups, ev_idx, so)
    flag = " ⚠ **single source**" if nsrc <= 1 else ""
    if main and any(a in domain(main["url"]) for a in AGGREGATORS):
        flag += " ⚠ **source of record is an aggregator**"
    out = [f"- **Main source:** {fmt_link(main) if main else '_none surfaced_'}",
           f"- **Provenance:** {nsrc} surfaced source domain(s){flag}"]
    if related:
        out.append("- **Related coverage:**")
        out += [f"  - {fmt_link(r)}" for r in related]
    return out, nsrc


def angle_lines(run, sid, label="angle-generator angles"):
    ang = load(run, f"angles.{sid[:8]}.json")
    if not ang or not ang.get("angles"):
        return []
    out = [f"- **{label}:**"]
    for a in ang["angles"]:
        js = a.get("journalist_shape") or {}
        sug = " *(SUGGESTION)*" if a.get("suggestion") else ""
        out.append(f"  - *{a.get('headline_frame')}*{sug} — {(js.get('beat_description') or '')[:110]} _(decay {(a.get('decay') or {}).get('stage')})_")
    return out


def build(run):
    name = os.path.basename(run).split("_", 1)[1]
    client = CLIENTS.get(name, name)
    candidates = load(run, "candidates.json") or {}
    clustered = load(run, "clustered_candidates.json") or {}
    targeted = load(run, "targeted_candidates.json") or {}
    triaged = (load(run, "triaged_candidates.json") or {}).get("triaged", [])
    ev_idx = evidence_index(candidates)
    clm = clustered.get("clustering", {})
    sel = {s["id"]: s for s in (targeted.get("signals") or [])}
    dups_by_rep = {}
    for d in (clustered.get("clustered_duplicates") or []):
        dups_by_rep.setdefault(d.get("representative_id"), []).append(d)
    rej = (targeted.get("freshness_gate") or {}).get("rejected_signals") or []

    # Tier routing. Back-compat: fall back to the old gate field if tier is absent.
    def tier_of(t):
        if t.get("tier"):
            return t["tier"]
        return {"advance": "pitch_ready", "drop": "watch"}.get(t.get("gate"), "watch")
    pitch = [t for t in triaged if tier_of(t) == "pitch_ready"]
    big = [t for t in triaged if tier_of(t) == "big_story"]
    watch_tri = [t for t in triaged if tier_of(t) == "watch"]

    # Sort big stories by coverage spread (distinct surfaced outlet count) desc.
    def outlet_count(sid):
        s = sel.get(sid, {})
        return main_and_related(s, dups_by_rep.get(sid, []), ev_idx, s.get("story_origin") or {})[2]
    big.sort(key=lambda t: outlet_count(t["signal_id"]), reverse=True)

    P = [f"# {client} — Newsjack Scan", ""]
    big_word = "big story" if len(big) == 1 else "big stories"
    P.append(f"**Today's read:** {len(pitch)} pitch-ready · {len(big)} {big_word} worth a look · "
             f"{len(watch_tri) + len(rej)} watched.")
    P.append(f"**Funnel:** {len(candidates.get('signals') or [])} candidates → "
             f"{clm.get('representative_count','?')} stories ({clm.get('duplicate_count',0)} dups collapsed) → "
             f"{len(sel)} fresh → **{len(pitch)} pitch-ready + {len(big)} big-story suggestions**. "
             f"Nothing pitchable or big was dropped off-screen.")
    P.append("")

    # ── ✅ Pitch-Ready ──────────────────────────────────────────────
    P.append("## ✅ Pitch-Ready")
    P.append("_Fresh, real standing, and a journalist-shaped angle. Act on these._")
    P.append("")
    if not pitch:
        P.append("**Nothing cleared the standing gate this window.**")
        P.append("")
    for i, t in enumerate(pitch, 1):
        sid = t["signal_id"]; s = sel.get(sid, {}); so = s.get("story_origin") or {}
        dups = dups_by_rep.get(sid, [])
        cf = t.get("consolidated_from") or []
        note = f" _(consolidates {len(cf)+1} same-event pickups)_" if cf else (f" _(collapses {len(dups)+1} pickups)_" if dups else "")
        P.append(f"### {i}. {t.get('signal_title') or s.get('title','')}{note}")
        P.append(freshness_line(s, so))
        P.append(f"- **Standing:** `{t.get('standing')}` — {t.get('standing_rationale','')}"
                 + ("  _(proof-gated — lead with the human ask)_" if t.get("proof_gated") else ""))
        lines, _ = source_lines(s, dups, ev_idx, so)
        P += lines
        P += angle_lines(run, sid)
        P.append("")

    # ── 🔥 Big Stories Worth a Look ─────────────────────────────────
    P.append("## 🔥 Big Stories Worth a Look")
    P.append("_Your call — relevance unverified, suggestions only. We are **not** asserting you have standing; "
             "a big story is always worth a look and the drop decision is yours. Sorted by how widely it is covered._")
    P.append("")
    if not big:
        P.append("_No fresh big stories outside the pitch-ready set this window._")
        P.append("")
    for i, t in enumerate(big, 1):
        sid = t["signal_id"]; s = sel.get(sid, {}); so = s.get("story_origin") or {}
        dups = dups_by_rep.get(sid, [])
        band = (s.get("story_size") or {}).get("band", "")
        lines, nsrc = source_lines(s, dups, ev_idx, so)
        wflag = WEAKNESS_FLAG.get((s.get("coarse_relevance") or {}).get("weakness_flag"))
        P.append(f"### {i}. {t.get('signal_title') or s.get('title','')} — {BAND_LABEL.get(band, band or 'story')} · {nsrc} outlet(s)")
        P.append(freshness_line(s, so))
        conf = t.get("relevance_confidence")
        bridge = t.get("bridge_note") or "No clear bridge — awareness only."
        P.append(f"- **Why surfaced:** big public story; relevance is your call"
                 + (f" (confidence: `{conf}`)" if conf else "") + (f" · {wflag}" if wflag else ""))
        P.append(f"- **Possible way in (SUGGESTION):** {bridge}")
        P += lines
        al = angle_lines(run, sid, label="suggested angle")
        P += al if al else ["- _No clean angle — listed for awareness only._"]
        P.append("")

    # ── 👀 Watch / Context ──────────────────────────────────────────
    P.append("## 👀 Watch / Context")
    P.append("_Seen, not pitchable now. Disclosed so nothing is hidden._")
    P.append("")
    any_w = False
    if watch_tri:
        any_w = True
        P.append(f"**Fresh but dropped by standing triage** ({len(watch_tri)})")
        for t in watch_tri:
            P.append(f"- {t.get('signal_title','')[:90]} — `{t.get('watch_reason') or t.get('drop_reason')}`: {t.get('standing_rationale','')[:130]}")
        P.append("")
    g = defaultdict(list)
    for r in rej:
        g[(r.get("freshness_gate") or {}).get("computed_status")].append(r)
    for st in ("stale", "unverified_no_corroboration", "unverified_boundary", "unverified_no_timestamp"):
        items = g.get(st) or []
        if not items:
            continue
        any_w = True
        P.append(f"**{GATE_LABEL.get(st, st)}** ({len(items)})")
        for r in items:
            so = r.get("story_origin") or {}
            link = so.get("canonical_coverage_url") or (so.get("evidence_urls") or r.get("evidence_urls") or [None])[0]
            band = (r.get("story_size") or {}).get("band", "")
            big_marker = " _(big story — stale)_" if band in ("high", "major") else ""
            P.append(f"- {mdlink(r.get('signal_title'), link)} — first public {so.get('first_public_at') or 'unverified'}{big_marker}")
        P.append("")
    if not any_w:
        P.append("_Nothing gated out this window._")
        P.append("")

    open(os.path.join(run, "final_report.md"), "w").write("\n".join(P))
    return len(pitch), len(big), len(watch_tri), len(rej)


if __name__ == "__main__":
    run = sys.argv[1]
    p, b, w, r = build(run)
    print(f"{os.path.basename(run)}: pitch_ready={p} big_story={b} watch_triage={w} freshness_gated={r}")
