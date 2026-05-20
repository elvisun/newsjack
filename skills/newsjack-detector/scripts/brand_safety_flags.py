"""Deterministic safety flags for newsjack evidence packets.

This module does not decide whether a signal is pitchable. It only marks text
that the skill must review against the doctrine files.
"""

from __future__ import annotations


HARD_SAFETY_TERMS = {
    "abuse",
    "assault",
    "bombing",
    "child abuse",
    "earthquake",
    "genocide",
    "hate crime",
    "hostage",
    "humanitarian crisis",
    "mass shooting",
    "missing child",
    "missing person",
    "murder",
    "rape",
    "sexual violence",
    "terror attack",
    "war crime",
}


def flag_text(text: str, exclusions: list[str] | None = None) -> list[dict[str, str]]:
    lower = text.lower()
    flags: list[dict[str, str]] = []
    for term in sorted(HARD_SAFETY_TERMS, key=len, reverse=True):
        if term in lower:
            flags.append(
                {
                    "type": "hard_safety_term",
                    "term": term,
                    "note": "Review against tragedy and human-suffering newsjacking rules.",
                }
            )
    for term in exclusions or []:
        if term and term.lower() in lower:
            flags.append(
                {
                    "type": "profile_exclusion",
                    "term": term,
                    "note": "Matched a monitor-profile exclusion.",
                }
            )
    return flags
