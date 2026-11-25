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
    country: 'CA',
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
      {
        id: 'R1',
        capture: 'medialyst',
        answers: [
          { question_id: 'Q1', engine: 'chatgpt', model: 'model-a', status: 'ok' },
          { question_id: 'Q2', engine: 'chatgpt', model: 'model-a', status: 'ok' },
          { question_id: 'Q1', engine: 'gemini', model: 'model-b', status: 'ok' },
        ],
      },
    ],
    grades: [
      { fact_id: 'F1', engine: 'chatgpt', run_id: 'R1', verdict: 'wrong', answer_verdicts: { Q1: 'wrong' } },
      { fact_id: 'F2', engine: 'chatgpt', run_id: 'R1', verdict: 'correct', answer_verdicts: { Q2: 'correct' } },
      { fact_id: 'F1', engine: 'gemini', run_id: 'R1', verdict: 'silent', answer_verdicts: { Q1: 'silent' } },
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
  assert.equal(out.overall.consistent_accuracy, 0.5);
  assert.equal(out.overall.inconsistent_cells, 0);
  assert.equal(out.overall.silent, 1);
  assert.equal(out.overall.off_brand, 0);
  assert.equal(out.overall.not_asked_or_engine_failed, 1);
});

test('markdown prints score metadata and a fact-by-engine grid', () => {
  const markdown = renderMarkdown(fixture());
  assert.match(markdown, /Overall: 50\.0% accuracy \(1\/2 graded\); 66\.7% coverage \(2\/3 asked\); 1 silent; 0 off-brand/);
  assert.match(markdown, /Date: 2026-11-24T15:00:00Z\. Engines: ChatGPT, Gemini\. Models: ChatGPT: model-a; Gemini: model-b\. Runs: 1\. Capture: medialyst\. Country: CA/);
  assert.match(markdown, /Overall: 50\.0% consistent accuracy \(1\/2 graded\); 0 inconsistent cells/);
  assert.match(markdown, /ChatGPT: 50\.0% accuracy/);
  assert.match(markdown, /Gemini: n\/a accuracy/);
  assert.match(markdown, /\| F1: Pineglass Labs was founded in 2019\. \| wrong \| silent \|/);
  assert.match(markdown, /model-b/);
});

test('repeated runs are counted and shown without hiding variance', () => {
  const audit = fixture();
  audit.runs.push({ id: 'R2', capture: 'paste', answers: [] });
  audit.grades.push(
    { fact_id: 'F1', engine: 'chatgpt', run_id: 'R2', verdict: 'correct', answer_verdicts: { Q1: 'correct' } },
    { fact_id: 'F2', engine: 'chatgpt', run_id: 'R2', verdict: 'correct', answer_verdicts: { Q2: 'correct' } },
  );
  const markdown = renderMarkdown(audit);
  assert.match(markdown, /correct ×1; wrong ×1/);
  assert.match(markdown, /ChatGPT: 75\.0% accuracy \(3\/4 graded\)/);
  assert.match(markdown, /2 runs/);
});

test('mixed answers retain the serious verdict and add consistent accuracy', () => {
  const audit = fixture();
  const grade = audit.grades.find((item) => item.fact_id === 'F1' && item.engine === 'chatgpt');
  grade.verdict = 'outdated_or_partial';
  grade.answer_verdicts = {
    Q1: 'correct',
    Q2: 'correct',
    Q3: 'correct',
    Q4: 'correct',
    Q5: 'outdated_or_partial',
  };
  const out = summarizeAudit(audit);
  assert.equal(out.by_engine[0].accuracy, 0.5);
  assert.equal(out.by_engine[0].consistent_accuracy, 1);
  assert.equal(out.by_engine[0].inconsistent_cells, 1);
  const markdown = renderMarkdown(audit);
  assert.match(markdown, /outdated\/partial \(1 of 5 answers\)/);
  assert.match(markdown, /Overall: 100\.0% consistent accuracy \(2\/2 graded\); 1 inconsistent cell/);
});

test('off-brand answers are separate from silence and citation look-alikes are flagged', () => {
  const audit = fixture();
  const geminiAnswer = audit.runs[0].answers.find((answer) => answer.engine === 'gemini');
  geminiAnswer.off_brand = true;
  const geminiGrade = audit.grades.find((grade) => grade.fact_id === 'F1' && grade.engine === 'gemini');
  geminiGrade.verdict = 'off_brand';
  geminiGrade.answer_verdicts = { Q1: 'off_brand' };
  const chatgptAnswer = audit.runs[0].answers.find((answer) => answer.engine === 'chatgpt' && answer.question_id === 'Q2');
  chatgptAnswer.lookalike_brand_citation = true;

  const out = summarizeAudit(audit);
  assert.equal(out.overall.silent, 0);
  assert.equal(out.overall.off_brand, 1);
  const markdown = renderMarkdown(audit);
  assert.match(markdown, /\| F1: Pineglass Labs was founded in 2019\. \| wrong \| off-brand \|/);
  assert.match(markdown, /correct; look-alike citation ×1/);
});

test('answer verdicts must agree with the most-serious cell verdict', () => {
  const audit = fixture();
  audit.grades[0].verdict = 'correct';
  assert.throws(() => summarizeAudit(audit), /must use most-serious verdict wrong/);
});

test('duplicate fact-engine-run grades are rejected', () => {
  const audit = fixture();
  audit.grades.push({ fact_id: 'F1', engine: 'chatgpt', run_id: 'R1', verdict: 'correct' });
  assert.throws(() => summarizeAudit(audit), /duplicate grade/);
});
