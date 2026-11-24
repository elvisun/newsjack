#!/usr/bin/env node
/**
 * Deterministic scoring for ai-answer-audit graded JSON.
 * Node >= 18, standard library only.
 */

import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { resolve } from 'node:path';

const VERDICTS = new Set([
  'correct',
  'wrong',
  'outdated_or_partial',
  'silent',
  'off_brand',
  'not_asked_or_engine_failed',
]);

const SUBSTANTIVE_VERDICTS = new Set(['correct', 'wrong', 'outdated_or_partial']);
const VERDICT_PRIORITY = ['wrong', 'outdated_or_partial', 'correct', 'off_brand', 'silent', 'not_asked_or_engine_failed'];

const VERDICT_LABELS = {
  correct: 'correct',
  wrong: 'wrong',
  outdated_or_partial: 'outdated/partial',
  silent: 'silent',
  off_brand: 'off-brand',
  not_asked_or_engine_failed: 'not asked/failed',
};

function requiredArray(value, label) {
  if (!Array.isArray(value)) throw new Error(`${label} must be an array`);
  return value;
}

function engineRecord(value) {
  if (typeof value === 'string') return { id: value, label: value, models: [] };
  if (!value || typeof value !== 'object' || !String(value.id || '').trim()) {
    throw new Error('each engine must be a string or an object with an id');
  }
  return {
    ...value,
    id: String(value.id).trim(),
    label: String(value.label || value.id).trim(),
    models: Array.isArray(value.models) ? value.models.filter(Boolean).map(String) : [],
  };
}

function percent(numerator, denominator) {
  if (!denominator) return 'n/a';
  return `${(100 * numerator / denominator).toFixed(1)}%`;
}

function emptyCounts() {
  return { correct: 0, wrong: 0, outdated_or_partial: 0, silent: 0, off_brand: 0, not_asked_or_engine_failed: 0 };
}

function scoreCounts(counts) {
  const graded = counts.correct + counts.wrong + counts.outdated_or_partial;
  const asked = graded + counts.silent + counts.off_brand;
  return {
    ...counts,
    graded,
    asked,
    accuracy: graded ? counts.correct / graded : null,
    coverage: asked ? graded / asked : null,
  };
}

function normalizedAnswerVerdicts(value, label) {
  if (value === undefined) return {};
  if (!value || typeof value !== 'object' || Array.isArray(value)) {
    throw new Error(`${label} answer_verdicts must be an object`);
  }
  const out = {};
  for (const [questionId, verdict] of Object.entries(value)) {
    if (!questionId.trim()) throw new Error(`${label} answer_verdicts has an empty question id`);
    if (!VERDICTS.has(verdict)) throw new Error(`${label} has unknown answer verdict ${verdict}`);
    out[questionId] = verdict;
  }
  return out;
}

function substantiveAnswerVerdicts(grade) {
  return Object.values(grade.answer_verdicts).filter((verdict) => SUBSTANTIVE_VERDICTS.has(verdict));
}

function isInconsistent(grade) {
  return new Set(substantiveAnswerVerdicts(grade)).size > 1;
}

function consistentOutcome(grade) {
  const verdicts = substantiveAnswerVerdicts(grade);
  if (verdicts.includes('correct')) return 'correct';
  if (verdicts.length && verdicts.every((verdict) => verdict === 'wrong' || verdict === 'outdated_or_partial')) return 'error';
  if (grade.verdict === 'correct') return 'correct';
  if (grade.verdict === 'wrong' || grade.verdict === 'outdated_or_partial') return 'error';
  return null;
}

function addConsistency(score, grades) {
  const outcomes = grades.map(consistentOutcome).filter(Boolean);
  const correct = outcomes.filter((outcome) => outcome === 'correct').length;
  return {
    ...score,
    consistent_correct: correct,
    consistent_graded: outcomes.length,
    consistent_accuracy: outcomes.length ? correct / outcomes.length : null,
    inconsistent_cells: grades.filter(isInconsistent).length,
  };
}

function modelAndRunMetadata(audit, engineId, grades, engine) {
  const models = new Set(engine.models || []);
  const runIds = new Set(grades.map((grade) => grade.run_id));
  for (const run of audit.runs || []) {
    if (!runIds.has(String(run.id))) continue;
    for (const answer of run.answers || []) {
      if (answer.engine === engineId && answer.model) models.add(String(answer.model));
    }
  }
  return { models: [...models], run_count: runIds.size };
}

export function summarizeAudit(audit) {
  if (!audit || typeof audit !== 'object') throw new Error('input must be a JSON object');
  const facts = requiredArray(audit.facts, 'facts').map((fact) => {
    if (!fact || !String(fact.id || '').trim() || !String(fact.text || '').trim()) {
      throw new Error('each fact must have id and text');
    }
    return { ...fact, id: String(fact.id), text: String(fact.text) };
  });
  const engines = requiredArray(audit.engines, 'engines').map(engineRecord);
  const grades = requiredArray(audit.grades, 'grades').map((grade) => ({
    ...grade,
    fact_id: String(grade?.fact_id || ''),
    engine: String(grade?.engine || ''),
    run_id: String(grade?.run_id || ''),
    verdict: String(grade?.verdict || ''),
    answer_verdicts: normalizedAnswerVerdicts(
      grade?.answer_verdicts,
      `grade for ${grade?.fact_id || '?'}/${grade?.engine || '?'}/${grade?.run_id || '?'}`,
    ),
  }));

  const factIds = new Set(facts.map((fact) => fact.id));
  const engineIds = new Set(engines.map((engine) => engine.id));
  const keys = new Set();
  for (const grade of grades) {
    if (!factIds.has(grade.fact_id)) throw new Error(`grade references unknown fact ${grade.fact_id}`);
    if (!engineIds.has(grade.engine)) throw new Error(`grade references unknown engine ${grade.engine}`);
    if (!grade.run_id) throw new Error(`grade for ${grade.fact_id}/${grade.engine} has no run_id`);
    if (!VERDICTS.has(grade.verdict)) throw new Error(`unknown verdict ${grade.verdict}`);
    const observed = Object.values(grade.answer_verdicts);
    if (observed.length) {
      const expected = VERDICT_PRIORITY.find((verdict) => observed.includes(verdict));
      if (grade.verdict !== expected) {
        throw new Error(`grade for ${grade.fact_id}/${grade.engine}/${grade.run_id} must use most-serious verdict ${expected}`);
      }
    }
    const key = `${grade.fact_id}\u0000${grade.engine}\u0000${grade.run_id}`;
    if (keys.has(key)) throw new Error(`duplicate grade for ${grade.fact_id}/${grade.engine}/${grade.run_id}`);
    keys.add(key);
  }

  const byEngine = [];
  const overallCounts = emptyCounts();
  for (const engine of engines) {
    const engineGrades = grades.filter((grade) => grade.engine === engine.id);
    const counts = emptyCounts();
    for (const grade of engineGrades) {
      counts[grade.verdict] += 1;
      overallCounts[grade.verdict] += 1;
    }
    byEngine.push({
      ...engine,
      ...addConsistency(scoreCounts(counts), engineGrades),
      ...modelAndRunMetadata(audit, engine.id, engineGrades, engine),
    });
  }

  return {
    audit,
    facts,
    engines,
    grades,
    by_engine: byEngine,
    overall: addConsistency(scoreCounts(overallCounts), grades),
  };
}

function answerFlags(audit, grade) {
  const run = (audit.runs || []).find((item) => String(item.id) === grade.run_id);
  const questionIds = new Set(Object.keys(grade.answer_verdicts));
  const answers = (run?.answers || []).filter((answer) => (
    answer.engine === grade.engine && questionIds.has(String(answer.question_id))
  ));
  return {
    off_brand: answers.filter((answer) => answer.off_brand).length,
    lookalike_brand_citation: answers.filter((answer) => answer.lookalike_brand_citation).length,
  };
}

function mixedGradeLabel(grade) {
  const verdicts = substantiveAnswerVerdicts(grade);
  if (!isInconsistent(grade)) return VERDICT_LABELS[grade.verdict];
  const count = verdicts.filter((verdict) => verdict === grade.verdict).length;
  return `${VERDICT_LABELS[grade.verdict]} (${count} of ${verdicts.length} answers)`;
}

function cellFor(audit, grades) {
  if (!grades.length) return 'not asked/failed';
  if (grades.length === 1) {
    const grade = grades[0];
    const parts = [mixedGradeLabel(grade)];
    const flags = answerFlags(audit, grade);
    if (grade.wrong_brand) parts.push('wrong-brand');
    if (flags.off_brand && grade.verdict !== 'off_brand') parts.push(`off-brand answer ×${flags.off_brand}`);
    if (flags.lookalike_brand_citation) parts.push(`look-alike citation ×${flags.lookalike_brand_citation}`);
    return parts.join('; ');
  }
  const counts = emptyCounts();
  let wrongBrand = 0;
  let offBrandAnswers = 0;
  let lookalikeCitations = 0;
  for (const grade of grades) {
    counts[grade.verdict] += 1;
    if (grade.wrong_brand) wrongBrand += 1;
    const flags = answerFlags(audit, grade);
    offBrandAnswers += flags.off_brand;
    lookalikeCitations += flags.lookalike_brand_citation;
  }
  const active = Object.entries(counts).filter(([, count]) => count > 0);
  const parts = [active
    .map(([verdict, count]) => `${VERDICT_LABELS[verdict]}${grades.length > 1 ? ` ×${count}` : ''}`)
    .join('; ')];
  for (const grade of grades.filter(isInconsistent)) {
    parts.push(`${grade.run_id}: ${mixedGradeLabel(grade)}`);
  }
  if (wrongBrand) parts.push(`wrong-brand ×${wrongBrand}`);
  if (offBrandAnswers) parts.push(`off-brand answer ×${offBrandAnswers}`);
  if (lookalikeCitations) parts.push(`look-alike citation ×${lookalikeCitations}`);
  return parts.join('; ');
}

function escapeCell(value) {
  return String(value).replaceAll('|', '\\|').replaceAll('\n', ' ');
}

function scoreLine(label, score) {
  const accuracy = percent(score.correct, score.graded);
  const coverage = percent(score.graded, score.asked);
  return `${label}: ${accuracy} accuracy (${score.correct}/${score.graded} graded); ${coverage} coverage (${score.graded}/${score.asked} asked); ${score.silent} silent; ${score.off_brand} off-brand; ${score.not_asked_or_engine_failed} not asked/failed`;
}

function consistentLine(label, score) {
  return `${label}: ${percent(score.consistent_correct, score.consistent_graded)} consistent accuracy (${score.consistent_correct}/${score.consistent_graded} graded); ${score.inconsistent_cells} inconsistent cell${score.inconsistent_cells === 1 ? '' : 's'}`;
}

export function renderMarkdown(audit) {
  const summary = summarizeAudit(audit);
  const out = [];
  const runCount = new Set(summary.grades.map((grade) => grade.run_id)).size;
  const engineLabels = summary.engines.map((engine) => engine.label).join(', ') || 'not recorded';
  const modelLabels = summary.by_engine
    .map((engine) => `${engine.label}: ${engine.models.length ? engine.models.join(', ') : 'not recorded'}`)
    .join('; ');
  const captureLanes = [...new Set((audit.runs || []).map((run) => run.capture).filter(Boolean))].join(', ') || 'not recorded';
  out.push(`**${scoreLine('Overall', summary.overall)}. Date: ${audit.audited_at || 'not recorded'}. Engines: ${engineLabels}. Models: ${modelLabels}. Runs: ${runCount}. Capture: ${captureLanes}. Country: ${audit.country || 'not recorded'}.**`);
  out.push('');
  out.push(`**${consistentLine('Overall', summary.overall)}.**`);
  out.push('');
  out.push(`Audit: ${audit.audited_at || 'date not recorded'} · Brand: ${audit.brand || 'not recorded'} · Runs: ${runCount}`);
  out.push('');
  out.push('### Engine scores');
  out.push('');
  for (const engine of summary.by_engine) {
    const models = engine.models.length ? engine.models.join(', ') : 'model not recorded';
    out.push(`- ${scoreLine(engine.label, engine)}; ${consistentLine('consistent', engine)} · ${models} · ${engine.run_count} run${engine.run_count === 1 ? '' : 's'}`);
  }
  out.push('');
  out.push('### Fact-by-engine grid');
  out.push('');
  out.push(`| Fact | ${summary.engines.map((engine) => escapeCell(engine.label)).join(' | ')} |`);
  out.push(`| --- | ${summary.engines.map(() => '---').join(' | ')} |`);
  for (const fact of summary.facts) {
    const cells = summary.engines.map((engine) => cellFor(audit,
      summary.grades.filter((grade) => grade.fact_id === fact.id && grade.engine === engine.id),
    ));
    out.push(`| ${escapeCell(`${fact.id}: ${fact.text}`)} | ${cells.map(escapeCell).join(' | ')} |`);
  }
  return `${out.join('\n')}\n`;
}

function readInput(path) {
  if (!path || path === '-') return readFileSync(0, 'utf8');
  return readFileSync(path, 'utf8');
}

function main() {
  const path = process.argv[2];
  if (!path || path === '--help' || path === '-h') {
    process.stderr.write('usage: node score.mjs <graded-audit.json|->\n');
    process.exit(path ? 0 : 2);
  }
  try {
    process.stdout.write(renderMarkdown(JSON.parse(readInput(path))));
  } catch (error) {
    process.stderr.write(`error: ${error?.message || error}\n`);
    process.exit(1);
  }
}

if (process.argv[1] && resolve(process.argv[1]) === fileURLToPath(import.meta.url)) main();
