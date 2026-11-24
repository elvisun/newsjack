// Offline tests for skills/ai-answer-audit/score.mjs.
// Run: node --test fixtures/ai-answer-audit/score.test.mjs

import { test } from 'node:test';
import assert from 'node:assert/strict';
import { fileURLToPath } from 'node:url';
import { dirname, resolve } from 'node:path';

const here = dirname(fileURLToPath(import.meta.url));
const { renderMarkdown, summarizeAudit } = await import(resolve(here, '../../skills/ai-answer-audit/score.mjs'));

function fixture() {
  return {
    skill: 'ai-answer-audit',
    version: 1,
    brand: 'Pineglass Labs',
    audited_at: '2026-11-24T15:00:00Z',
    facts: [
      { id: 'F1', text: 'Pineglass Labs was founded in 2019.' },
      { id: 'F2', text: 'Plans start at $29 per month.' },
    ],
    engines: [
      { id: 'chatgpt', label: 'ChatGPT', models: ['model-a'] },
      { id: 'gemini', label: 'Gemini' },
    ],
    runs: [
      { id: 'R1', answers: [{ engine: 'gemini', model: 'model-b' }] },
    ],
    grades: [
      { fact_id: 'F1', engine: 'chatgpt', run_id: 'R1', verdict: 'wrong' },
      { fact_id: 'F2', engine: 'chatgpt', run_id: 'R1', verdict: 'correct' },
      { fact_id: 'F1', engine: 'gemini', run_id: 'R1', verdict: 'silent' },
      { fact_id: 'F2', engine: 'gemini', run_id: 'R1', verdict: 'not_asked_or_engine_failed' },
    ],
  };
}

test('accuracy excludes silence and failed engines while coverage reports silence', () => {
  const out = summarizeAudit(fixture());
  assert.equal(out.by_engine[0].accuracy, 0.5);
  assert.equal(out.by_engine[0].coverage, 1);
  assert.equal(out.by_engine[1].accuracy, null);
  assert.equal(out.by_engine[1].coverage, 0);
  assert.equal(out.overall.accuracy, 0.5);
  assert.equal(out.overall.silent, 1);
  assert.equal(out.overall.not_asked_or_engine_failed, 1);
});

test('markdown prints score metadata and a fact-by-engine grid', () => {
  const markdown = renderMarkdown(fixture());
  assert.match(markdown, /Overall: 50\.0% accuracy \(1\/2 graded\); 66\.7% coverage \(2\/3 asked\)/);
  assert.match(markdown, /ChatGPT: 50\.0% accuracy/);
  assert.match(markdown, /Gemini: n\/a accuracy/);
  assert.match(markdown, /\| F1: Pineglass Labs was founded in 2019\. \| wrong \| silent \|/);
  assert.match(markdown, /model-b/);
});

test('repeated runs are counted and shown without hiding variance', () => {
  const audit = fixture();
  audit.runs.push({ id: 'R2', answers: [] });
  audit.grades.push(
    { fact_id: 'F1', engine: 'chatgpt', run_id: 'R2', verdict: 'correct' },
    { fact_id: 'F2', engine: 'chatgpt', run_id: 'R2', verdict: 'correct' },
  );
  const markdown = renderMarkdown(audit);
  assert.match(markdown, /correct ×1; wrong ×1/);
  assert.match(markdown, /ChatGPT: 75\.0% accuracy \(3\/4 graded\)/);
  assert.match(markdown, /2 runs/);
});

test('duplicate fact-engine-run grades are rejected', () => {
  const audit = fixture();
  audit.grades.push({ fact_id: 'F1', engine: 'chatgpt', run_id: 'R1', verdict: 'correct' });
  assert.throws(() => summarizeAudit(audit), /duplicate grade/);
});
