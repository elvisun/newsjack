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
  'not_asked_or_engine_failed',
]);

const VERDICT_LABELS = {
  correct: 'correct',
  wrong: 'wrong',
  outdated_or_partial: 'outdated/partial',
  silent: 'silent',
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
  return { correct: 0, wrong: 0, outdated_or_partial: 0, silent: 0, not_asked_or_engine_failed: 0 };
}

function scoreCounts(counts) {
  const graded = counts.correct + counts.wrong + counts.outdated_or_partial;
  const asked = graded + counts.silent;
  return {
    ...counts,
    graded,
    asked,
    accuracy: graded ? counts.correct / graded : null,
    coverage: asked ? graded / asked : null,
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
  }));

  const factIds = new Set(facts.map((fact) => fact.id));
  const engineIds = new Set(engines.map((engine) => engine.id));
  const keys = new Set();
  for (const grade of grades) {
    if (!factIds.has(grade.fact_id)) throw new Error(`grade references unknown fact ${grade.fact_id}`);
    if (!engineIds.has(grade.engine)) throw new Error(`grade references unknown engine ${grade.engine}`);
    if (!grade.run_id) throw new Error(`grade for ${grade.fact_id}/${grade.engine} has no run_id`);
    if (!VERDICTS.has(grade.verdict)) throw new Error(`unknown verdict ${grade.verdict}`);
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
      ...scoreCounts(counts),
      ...modelAndRunMetadata(audit, engine.id, engineGrades, engine),
    });
  }

  return {
    audit,
    facts,
    engines,
    grades,
    by_engine: byEngine,
    overall: scoreCounts(overallCounts),
  };
}

function cellFor(grades) {
  if (!grades.length) return 'not asked/failed';
  const counts = emptyCounts();
  let wrongBrand = 0;
  for (const grade of grades) {
    counts[grade.verdict] += 1;
    if (grade.wrong_brand) wrongBrand += 1;
  }
  const active = Object.entries(counts).filter(([, count]) => count > 0);
  const text = active
    .map(([verdict, count]) => `${VERDICT_LABELS[verdict]}${grades.length > 1 ? ` ×${count}` : ''}`)
    .join('; ');
  return wrongBrand ? `${text}; wrong-brand ×${wrongBrand}` : text;
}

function escapeCell(value) {
  return String(value).replaceAll('|', '\\|').replaceAll('\n', ' ');
}

function scoreLine(label, score) {
  const accuracy = percent(score.correct, score.graded);
  const coverage = percent(score.graded, score.asked);
  return `${label}: ${accuracy} accuracy (${score.correct}/${score.graded} graded); ${coverage} coverage (${score.graded}/${score.asked} asked); ${score.silent} silent; ${score.not_asked_or_engine_failed} not asked/failed`;
}

export function renderMarkdown(audit) {
  const summary = summarizeAudit(audit);
  const out = [];
  out.push(`**${scoreLine('Overall', summary.overall)}.**`);
  const runCount = new Set(summary.grades.map((grade) => grade.run_id)).size;
  out.push('');
  out.push(`Audit: ${audit.audited_at || 'date not recorded'} · Brand: ${audit.brand || 'not recorded'} · Runs: ${runCount}`);
  out.push('');
  out.push('### Engine scores');
  out.push('');
  for (const engine of summary.by_engine) {
    const models = engine.models.length ? engine.models.join(', ') : 'model not recorded';
    out.push(`- ${scoreLine(engine.label, engine)} · ${models} · ${engine.run_count} run${engine.run_count === 1 ? '' : 's'}`);
  }
  out.push('');
  out.push('### Fact-by-engine grid');
  out.push('');
  out.push(`| Fact | ${summary.engines.map((engine) => escapeCell(engine.label)).join(' | ')} |`);
  out.push(`| --- | ${summary.engines.map(() => '---').join(' | ')} |`);
  for (const fact of summary.facts) {
    const cells = summary.engines.map((engine) => cellFor(
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
