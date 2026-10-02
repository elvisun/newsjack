# Lever 1–10 final scorecard

Decision: **INCONCLUSIVE**. There is **no supported efficiency winner**.
L4, descriptive coherent structure, is the atomic post-retrieval simulation
leader. That label is not a live ranking, retrieval, mention, or citation effect.

| Efficiency rank | Lever | Atomic citation / accurate-use lift | Human quality / fidelity delta | Median text changed | Google development → fresh association | Cross-provider result | Disposition |
|---:|---|---:|---:|---:|---|---|---|
| 1 | L4 Descriptive coherent structure | +25.0 pp / +25.0 pp | +0.125 / 0 | 15.6% | −7.3 → −21.9 pp/SD | Codex +50 pp; Claude 0 | Atomic leader; not supported overall |
| 2 | L5 Neutral qualified specificity | +12.5 / +12.5 pp | +0.625 / 0 | 0.0% median | −1.7 → −0.7 pp/SD | Codex 0; Claude +25 pp | Atomic promising; not supported overall |
| 3 | L6 Remove repetition and padding | 0 / +12.5 pp | +0.250 / 0 | 9.9% | +2.8 → +4.9 pp/SD | Citation direction disagreed | Insufficient |
| 4 | L1 Direct scoped answer | 0 / 0 pp | +0.500 / 0 | 19.3% | −1.4 → −0.8 pp/SD | Both flat | Insufficient |
| 5 | L2 Evidence context and attribution | 0 / 0 pp | 0 / −0.125 | 10.3% | +4.6 → −1.6 pp/SD | Both flat | Insufficient; fidelity cost |
| 6 | L3 Unique verifiable information | 0 / 0 pp | 0 / 0 | 0.0% | +1.4 → +0.7 pp/SD | Both flat | Insufficient; fixtures could not add real evidence |
| 7 | L7 Clarify ambiguous entities | Not tested | — | — | −0.5 → −1.6 pp/SD | ChatGPT unestimable | Not atomically tested |
| 8 | L8 Intent-fit table, list, or steps | Not tested | — | — | +16.5 → −11.7 pp/SD | Reversed on fresh data | Exploratory and unstable |
| 9 | L9 Expose a legitimate date | Not tested | — | — | −2.9 → −6.9 pp/SD | ChatGPT unestimable | Not atomically tested |
| 10 | L10 Preserve precision while clarifying syntax | Not tested | — | — | −0.2 → +3.8 pp/SD | ChatGPT unestimable | Human-quality guard only |

Atomic L1–L6 each used eight constructed post-retrieval pairs, including one
null case. L7–L10 were observational or guard-only. The efficiency index was
`mean(citation lift, accurate-use lift) × 0.10 / max(median changed fraction,
0.10)`: L4 scored 15.98, L5 12.50, and L6 6.25.

L4 cannot be called the winner because the overall composite was 59.79, audit
factor F1 was 0.521, combined citation win rate was 0.438, four of 12 combined
rewrites failed the deterministic fact gate, judged fidelity failed, and no
preregistered Google writing factor passed FDR on development or fresh data.
Its fresh Google proxy was negative (−21.9 pp/SD; 95% CI −40.3 to −3.5;
q=0.120), and its simulated lift was provider-heterogeneous.

All Google values are observational associations among measurable returned
organic pages after query, rank, domain, and page controls. ChatGPT does not
expose retrieved-but-rejected pages, so adjusted ChatGPT factor effects are not
estimable. None of these numbers measures a published edit's live impact.
