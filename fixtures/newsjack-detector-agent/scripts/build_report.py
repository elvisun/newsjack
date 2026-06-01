#!/usr/bin/env python3
"""Compile final_report.md for a detector run dir.

Story-first report. Every story shows ONE main link (the detector-surfaced
source of record) plus related links (clustered duplicate pickups + any
canonical the research worker proposed). Every link carries a date and a
provenance tag, so worker-introduced links are never laundered into the
report as established coverage. Usage: build_report.py RUN_DIR
"""
import json, os, re, sys

CLIENTS = {"bluebottle": "Blue Bottle Coffee", "clearnym": "Clearnym", "localfalcon": "Local Falcon",
           "nofar-method": "Nofar Method", "property-saviour": "Property Saviour", "simular": "Simular", "slite": "Slite"}
# Domains that are aggregators / low-authority republishers, not a source of record.
AGGREGATORS = ("kucoin.com", "startupfortune.com", "vocal.media", "x.com", "twitter.com",
               "bitget.com", "stocktitan", "markets.businessinsider", "einpresswire", "manilatimes")
GATE_LABEL = {
    "stale": "Stale — older than 24h, or a newer pickup of an older original",
    "unverified_no_corroboration": "Possibly fresh (single source) — a recent date was claimed but only one source backs it; confirm before using",
    "unverified_no_timestamp": "Undated — couldn't establish age; could be fresh or old",
    "unverified_boundary": "Possibly fresh (exact date unclear) — date-only clock straddling the 24h cutoff",
}


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
    """Return (main_link, related_links, surfaced_url_set, source_count)."""
    rep_ev = ev_idx.get(rep["id"], rep.get("evidence") or [])
    surfaced = []
    for e in rep_ev:
        surfaced.append({"title": e.get("title") or e.get("container"), "url": e.get("url"),
                         "src": e.get("container") or e.get("source"), "date": e.get("published_at"),
                         "prov": "surfaced"})
    # cluster duplicates (same story, other outlets) — surfaced, pull dates from candidates
    related = []
    for d in dups:
        for u in (d.get("evidence_urls") or []):
            de = next((e for e in ev_idx.get(d.get("signal_id"), []) if e.get("url") == u), {})
            related.append({"title": d.get("signal_title"), "url": u,
                            "src": de.get("container") or (d.get("sources") or ["?"])[0],
                            "date": de.get("published_at"), "prov": "surfaced duplicate"})
    surfaced_urls = {norm_url(x["url"]) for x in surfaced + related if x["url"]}
    # main link = source of record: prefer a non-aggregator surfaced source, else first surfaced
    main = None
    for x in surfaced:
        if x["url"] and not any(a in domain(x["url"]) for a in AGGREGATORS):
            main = x
            break
    if main is None and surfaced:
        main = surfaced[0]
    # any extra surfaced beyond main -> related
    for x in surfaced:
        if x is not main:
            related.append(x)
    # worker-proposed canonical / original -> related, flagged unverified if not surfaced
    so = story_origin or {}
    for url, datekey, label in [
        (so.get("canonical_coverage_url"), so.get("canonical_coverage_published_at"), so.get("canonical_coverage_source") or "canonical"),
        (so.get("original_url"), so.get("first_public_at"), so.get("original_source") or "original"),
    ]:
        if not url:
            continue
        if norm_url(url) in surfaced_urls:
            continue  # already shown as a real surfaced link
        if any(norm_url(url) == norm_url(r["url"]) for r in related):
            continue
        related.append({"title": label, "url": url, "src": domain(url), "date": datekey,
                        "prov": "proposed by research — UNVERIFIED"})
    # distinct surfaced source domains
    src_domains = {domain(x["url"]) for x in surfaced if x["url"]} | {domain(r["url"]) for r in related if r["url"] and r["prov"].startswith("surfaced")}
    return main, related, len(src_domains)


def fmt_link(x):
    date = x.get("date") or "no date"
    tag = x["prov"]
    if tag.startswith("proposed"):
        date = f"{date} — date unverified"
    dom = domain(x["url"]) if x.get("url") else x.get("src")
    return f"{mdlink(x.get('title'), x.get('url'))} — {dom}, {date} · `{tag}`"


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
    adv = [t for t in triaged if t.get("gate") == "advance"]
    drop = [t for t in triaged if t.get("gate") == "drop"]

    P = [f"# {client} — Newsjack Detector Report", ""]
    P.append(f"**Funnel:** {len(candidates.get('signals') or [])} candidates → "
             f"{clm.get('representative_count','?')} story representatives "
             f"({clm.get('duplicate_count',0)} duplicates collapsed) → {len(sel)} fresh → "
             f"**{len(adv)} advanced by standing triage**.")
    P.append("")
    P.append("## Top News Today")
    P.append("")
    if not adv:
        P.append(f"**Nothing cleared the standing gate this window.** See Watch / Not A Fit." if sel
                 else "**Nothing cleared the freshness gate this window.**")
    for i, t in enumerate(adv, 1):
        sid = t["signal_id"]; s = sel.get(sid, {}); so = s.get("story_origin") or {}
        dups = dups_by_rep.get(sid, [])
        main, related, nsrc = main_and_related(s, dups, ev_idx, so)
        cf = t.get("consolidated_from") or []
        note = f" _(consolidates {len(cf)+1} same-event pickups)_" if cf else (f" _(collapses {len(dups)+1} pickups)_" if dups else "")
        P.append(f"### {i}. {t.get('signal_title') or s.get('title','')}{note}")
        # standing + freshness with BOTH dates for new-development
        gate = (s.get("freshness_gate") or {}).get("computed_status")
        fp = so.get("first_public_at")
        if gate == "fresh_new_development":
            P.append(f"- **Freshness:** `{gate}` — first public **{fp}**; new development **{so.get('new_development_at')}**: {so.get('new_development')}")
        else:
            P.append(f"- **Freshness:** `{gate}` — first public **{fp}**")
        P.append(f"- **Standing:** `{t.get('standing')}` — {t.get('standing_rationale','')}"
                 + ("  _(proof-gated)_" if t.get("proof_gated") else ""))
        flag = " ⚠ **single source**" if nsrc <= 1 else ""
        if main and any(a in domain(main['url']) for a in AGGREGATORS):
            flag += " ⚠ **source of record is an aggregator**"
        P.append(f"- **Main source:** {fmt_link(main) if main else '_none surfaced_'}")
        P.append(f"- **Provenance:** {nsrc} surfaced source domain(s){flag}")
        if related:
            P.append("- **Related coverage:**")
            for r in related:
                P.append(f"  - {fmt_link(r)}")
        # angles
        ang = load(run, f"angles.{sid[:8]}.json")
        if ang and ang.get("angles"):
            P.append("- **angle-generator angles:**")
            for a in ang["angles"]:
                js = a.get("journalist_shape") or {}
                P.append(f"  - *{a.get('headline_frame')}* — {(js.get('beat_description') or '')[:110]} _(decay {(a.get('decay') or {}).get('stage')})_")
        P.append("")
    if drop:
        P.append("**Also fresh — dropped by triage:**")
        for t in drop:
            P.append(f"- **[NOT A FIT]** {t.get('signal_title','')[:80]} — `{t.get('drop_reason')}`: {t.get('standing_rationale','')[:130]}")
        P.append("")

    P.append("## Watch / Not A Fit")
    P.append("")
    from collections import defaultdict
    g = defaultdict(list)
    for r in rej:
        g[(r.get("freshness_gate") or {}).get("computed_status")].append(r)
    any_w = False
    for st in ("unverified_no_corroboration", "unverified_boundary", "unverified_no_timestamp", "stale"):
        items = g.get(st) or []
        if not items:
            continue
        any_w = True
        P.append(f"**{GATE_LABEL.get(st, st)}** ({len(items)})")
        for r in items:
            so = r.get("story_origin") or {}
            link = so.get("canonical_coverage_url") or (so.get("evidence_urls") or r.get("evidence_urls") or [None])[0]
            P.append(f"- {mdlink(r.get('signal_title'), link)} — first public {so.get('first_public_at') or 'unverified'}")
        P.append("")
    if not any_w:
        P.append("_Nothing gated out this window._")
    open(os.path.join(run, "final_report.md"), "w").write("\n".join(P))
    return len(adv), len(drop), len(rej)


if __name__ == "__main__":
    run = sys.argv[1]
    a, d, r = build(run)
    print(f"{os.path.basename(run)}: advanced={a} dropped={d} gated={r}")
