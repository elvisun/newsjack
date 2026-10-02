# Preregistered hypotheses — protocol v1

Status: frozen before paid collection. The eight hypotheses below are primary;
all other patterns are exploratory.

| ID | Preregistered thesis | Primary outcome | Falsifier or limit | Planned lever |
|---|---|---|---|---|
| H0 | Retrieval eligibility, organic position, publisher/domain effects, and topic fit explain most citation variance; writing features add little. | Per-platform source selection | Writing-feature estimates lose direction or practical size after matched-query, rank, domain, page-type, age, topic, and length controls. | Null/control |
| H1 | A short, self-contained, scoped answer near a relevant descriptive heading improves citation selection and accurate answer use versus a diffuse answer. | Citation selection; absorption | No positive paired signal after relevance/length matching, or a single-platform-only effect. | L1 direct scoped answer |
| H2 | Explicit, high-quality evidence and attribution improve selection and accurate use; citation quantity alone does not. | Selection; accurate use | Signal disappears after evidence quality and publisher controls, or low-quality citation volume performs equally. | L2 evidence context |
| H3 | Original data, firsthand analysis, or unique information gain outperforms commodity summary when authority and rank are held comparable. | Selection; accurate use | Signal disappears after controls or cannot be distinguished from publisher effects. | L3 unique information |
| H5 | Descriptive headings and coherent single-purpose chunks improve extractability without requiring question-shaped headings. | Selection; human quality | No cross-topic signal, or benefit is explained by publisher/length and harms prose. | L4 coherent structure |
| H8 | Specific, neutral, qualified prose performs better than promotional or absolute prose when facts stay constant. | Selection; human quality | No paired benefit or neutralization lowers reader quality/voice. | L5 neutral precision |
| H11 | Google AI Overviews and ChatGPT web search differ materially in sources and response stability, requiring caveats rather than separate hack lists. | URL/domain overlap; repeats | Source sets and factor directions are substantially stable across platforms and repetitions. | Platform interaction |
| H15 | Keyword repetition and checklist saturation reduce human quality without a reliable citation-selection benefit. | Human quality; selection | Saturation improves both selection and human quality across topics and executors. | L6 remove repetition |

## Multiple testing and dispositions

- Primary family size: 8.
- Two-sided alpha: 0.05.
- False-discovery control: Benjamini-Hochberg at q = 0.10 across primary
  writing-feature tests; H0 and H11 descriptive/control estimates are reported
  with intervals and are not used to dilute the family.
- `supported`: adjusted direction survives controls and FDR, appears in at
  least two topic families or document types, and shows no material human-
  quality harm.
- `promising`: directional evidence exists but an interval, interaction,
  platform difference, or confounder prevents `supported`.
- `insufficient`: data or identification is too weak.
- `rejected`: result is null/contradictory, unsafe, or quality-degrading.

No observational association will be described as the causal effect of an edit.
