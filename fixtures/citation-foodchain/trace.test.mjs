// Offline tests for skills/citation-foodchain/trace.mjs (deterministic layer only).
// Run: node --test fixtures/citation-foodchain/trace.test.mjs
// No network: pages come from the synthetic chain in fixturePages().

import { test } from 'node:test';
import assert from 'node:assert/strict';
import { fileURLToPath } from 'node:url';
import { dirname, resolve } from 'node:path';

const here = dirname(fileURLToPath(import.meta.url));
const mod = await import(resolve(here, '../../skills/citation-foodchain/trace.mjs'));
const { analyzePage, fixturePages, textReuse, unwrapStatic, parseArgs, parsePerplexityThread, parseChatGptShare } = mod;

const BASE = 'http://fixture.test';
const pages = fixturePages();
const page = (key) => pages[key].replace(/127\.0\.0\.1:PORT/g, 'fixture.test');
const analyze = (key, opts = {}) => analyzePage(page(key), { requestedUrl: `${BASE}/${key}`, finalUrl: `${BASE}/${key}`, status: 200, contentType: 'text/html', ...opts });

test('press release: wire boilerplate and publisher are detected', async () => {
  const d = await analyze('prnewswire/northline-release.html');
  assert.ok(d.signals.some((s) => s.kind === 'wire' && s.label === 'pr newswire'), 'wire signal');
  assert.equal(d.publisher, 'PR Newswire');
  assert.equal(d.published_at, '2026-02-03T13:00:00Z');
  assert.ok(d.jsonld_types.includes('NewsArticle'));
});

test('outlet article: claim matches the right sentence and names the attributed source', async () => {
  const d = await analyze('mapleledger/parcels.html', { claims: ['About 14% of parcel deliveries in Canada fail on the first attempt', 'Failed deliveries cost Canadian retailers $212 million a year'] });
  assert.equal(d.byline, 'Dana Whitfield');
  assert.equal(d.site_name, 'The Maple Ledger');
  const [c1, c2] = d.claim_matches;
  assert.equal(c1.verdict_hint, 'strong_match');
  assert.match(c1.best[0].sentence, /14 per cent of parcel deliveries/);
  assert.equal(c1.best[0].number_match, 1, '14% should match "14 per cent"');
  assert.equal(c1.best[0].attribution[0].named_source, 'Northline Logistics');
  assert.equal(c2.verdict_hint, 'strong_match');
  assert.equal(c2.best[0].number_match, 1);
  const pr = d.attribution_sentences.find((a) => a.phrases.includes('press release'));
  assert.ok(pr && pr.links[0].href.endsWith('/prnewswire/northline-release.html'), 'link to the release is attached to the attribution sentence');
  assert.ok(!d.attribution_sentences.some((a) => /Related: Other story/.test(a.sentence)), 'aside content is excluded from the body');
});

test('aggregator: syndication notice is detected', async () => {
  const d = await analyze('logisticsdaily/study-14-percent.html');
  assert.ok(d.signals.some((s) => s.kind === 'syndication' && s.label === 'originally published'));
  assert.equal(d.canonical_url, `${BASE}/mapleledger/parcels.html`);
});

test('canonical on another domain is flagged as syndication', async () => {
  const html = page('logisticsdaily/study-14-percent.html').replace('http://fixture.test/mapleledger/parcels.html', 'https://www.mapleledger.example/parcels.html');
  const d = await analyzePage(html, { requestedUrl: 'https://logisticsdaily.example/x', finalUrl: 'https://logisticsdaily.example/x', status: 200, contentType: 'text/html' });
  assert.equal(d.canonical_host_mismatch, true);
  assert.ok(d.signals.some((s) => s.label === 'canonical points to another domain'));
});

test('listicle: sponsored marker, no byline, no date, no external links', async () => {
  const d = await analyze('statsvault/last-mile-stats.html', { claims: ['14% of deliveries in Canada fail on the first attempt'] });
  assert.ok(d.signals.some((s) => s.kind === 'sponsored'));
  assert.equal(d.has_byline, false);
  assert.equal(d.has_date, false);
  assert.equal(d.external_link_count, 0);
  assert.equal(d.claim_matches[0].verdict_hint, 'strong_match');
});

test('wiki: reference list is parsed and refs attach to the matching sentence', async () => {
  const d = await analyze('wiki/Last-mile_delivery', { claims: ['about 14% of urban parcel deliveries in Canada fail on the first attempt'] });
  assert.equal(d.wikipedia_references.length, 2);
  const best = d.claim_matches[0].best[0];
  assert.ok(best.score > 0.8);
  const urls = best.wikipedia_refs.flatMap((r) => r.urls);
  assert.ok(urls.some((u) => u.endsWith('/mapleledger/parcels.html')), 'cite_note-1 resolves to the outlet');
  assert.ok(!urls.some((u) => u.endsWith('/statsvault/last-mile-stats.html')), 'cite_note-2 belongs to the next sentence, not this one');
  assert.equal(best.wikipedia_refs[0].sentence_level, true);
  assert.equal(d.wikipedia_references.length, 2, 'the page-level list still has both');
});

test('text reuse: aggregator copies the outlet; study text is independent', async () => {
  const a = await analyze('mapleledger/parcels.html', { full: true });
  const b = await analyze('logisticsdaily/study-14-percent.html', { full: true });
  const s = await analyze('northline/benchmark.html', { full: true });
  const text = (d) => d.paragraphs.map((p) => p.text).join('\n');
  const ab = textReuse(text(a), text(b));
  assert.ok(ab.shared_shingles >= 20, `expected substantial reuse, got ${ab.shared_shingles}`);
  assert.match(ab.reading, /substantial verbatim reuse/);
  const as = textReuse(text(a), text(s));
  assert.ok(as.shared_shingles < 5, 'outlet vs study should not be verbatim');
  assert.ok(as.shared_numbers.includes('18240') && as.shared_numbers.includes('14 %'), 'but they share the key numbers');
});

test('unwrap: tracking params stripped; google /url and bing /ck decoded; grounding redirect flagged', () => {
  assert.equal(unwrapStatic('https://ex.com/a?utm_source=chatgpt.com&id=1').url, 'https://ex.com/a?id=1');
  assert.equal(unwrapStatic('https://www.google.com/url?q=https%3A%2F%2Fex.com%2Fb&sa=t').url, 'https://ex.com/b');
  const b64 = 'a1' + Buffer.from('https://ex.com/c').toString('base64').replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
  assert.equal(unwrapStatic(`https://www.bing.com/ck/a?u=${b64}`).url, 'https://ex.com/c');
  assert.equal(unwrapStatic('https://vertexaisearch.cloud.google.com/grounding-api-redirect/AbC').method, 'needs-request');
});

test('parseArgs collects repeated --claim', () => {
  const a = parseArgs(['fetch', 'https://x', '--claim', 'one', '--claim', 'two', '--full']);
  assert.deepEqual(a.claim, ['one', 'two']);
  assert.equal(a.full, true);
  assert.deepEqual(a._, ['fetch', 'https://x']);
});

test('perplexity thread JSON: [n] markers resolve to web_results per sentence', () => {
  const answer = JSON.stringify({ answer: 'About 14% of parcels fail on the first attempt[1]. The cost is $212 million a year[1][2]. Nothing cited here.', web_results: [{ name: 'Maple Ledger', url: 'https://mapleledger.example/parcels?utm_source=x', snippet: 's1' }, { name: 'Brief', url: 'https://logisticsdaily.example/study', snippet: 's2' }] });
  const json = { entries: [{ query_str: 'how many parcels fail?', updated_datetime: '2026-11-03T00:00:00Z', text: JSON.stringify([{ step_type: 'FINAL', content: { answer } }]) }] };
  const out = parsePerplexityThread(json);
  assert.equal(out.engine, 'perplexity');
  assert.deepEqual(out.cited_urls, ['https://mapleledger.example/parcels', 'https://logisticsdaily.example/study']);
  const sc = out.turns[0].sentence_citations;
  assert.equal(sc.length, 2);
  assert.deepEqual(sc[1].refs, ['https://mapleledger.example/parcels', 'https://logisticsdaily.example/study']);
});

test('chatgpt share HTML: messages + per-span citation blocks are recovered', () => {
  const md = 'About 14% of Canadian parcels fail on the first attempt. ([Maple Ledger](https://mapleledger.example/parcels?utm_source=chatgpt.com)) The cost is $212 million a year.';
  const block = { messageId: 'a1', startIndex: 56, endIndex: 136, approvedUrls: ['https://mapleledger.example/parcels'], data: { matched_text: 'citeturn0search0', prefix: null, start_idx: 56, end_idx: 70, safe_urls: ['https://mapleledger.example/parcels', 'https://mapleledger.example/parcels?utm_source=chatgpt.com', 'https://images.openai.com/static-rsc-1/abc'], refs: [], alt: '([Maple Ledger](https://mapleledger.example/parcels?utm_source=chatgpt.com))', prompt_text: null, type: 'grouped_webpages', items: [{ title: 'One in seven parcels fail', url: 'https://mapleledger.example/parcels?utm_source=chatgpt.com', attribution: 'Maple Ledger', pub_date: null, snippet: 'About 14 per cent …', supporting_websites: [] }] } };
  const foot = { messageId: 'a1', startIndex: 160, endIndex: 161, data: { matched_text: 'x', start_idx: 160, end_idx: 161, safe_urls: [], refs: [], alt: '', prompt_text: null, type: 'sources_footnote', sources: [{ title: 'Maple Ledger', url: 'https://mapleledger.example/parcels?utm_source=chatgpt.com', attribution: 'Maple Ledger' }] } };
  const userEntry = { key: 'u1', origin: 'server', frame: { id: 'u1', attribution: 'You said:', role: 'user' }, source: { messageId: 'u1', revision: 1, markdown: 'how many parcels fail?', references: [] } };
  const asstEntry = { key: 'a1', origin: 'server', frame: { id: 'a1', attribution: 'ChatGPT said:', role: 'assistant' }, source: { messageId: 'a1', revision: 1, markdown: md, references: [block, block, foot] } };
  const html = `<html><script>[${JSON.stringify(userEntry)},${JSON.stringify(asstEntry)},${JSON.stringify(asstEntry)}]</script></html>`;
  const out = parseChatGptShare(html);
  assert.equal(out.messages.length, 2, 'repeated frame entries collapse to one message each');
  assert.equal(out.messages[1].role, 'assistant');
  assert.equal(out.citations.length, 1, 'duplicate hydration blocks are collapsed');
  const c = out.citations[0];
  assert.equal(c.message_id, 'a1');
  assert.match(c.preceding_sentence, /About 14% of Canadian parcels fail on the first attempt/);
  assert.equal(out.per_claim_mapping, true);
  assert.deepEqual(c.urls, ['https://mapleledger.example/parcels'], 'utm stripped, image CDN dropped, deduped');
  assert.equal(c.items[0].snippet, 'About 14 per cent …');
  assert.equal(out.sources_footnotes.length, 1);
  assert.deepEqual(out.cited_urls, ['https://mapleledger.example/parcels']);
});

test('unclosed <nav> and unbalanced <template> cannot swallow the article', async () => {
  const html = `<!doctype html><html><head><title>T</title></head><body><nav><a href="/">Home</a><template><div>x</div><main><article>
<p>Geologists at UC Berkeley recommended Tuesday eating at least one small rock per day, according to a report by the Berkeley Geology Department.</p>
<p>Second paragraph with enough words to count as article text for the parser to keep.</p></article></main></body></html>`;
  const d = await analyzePage(html, { requestedUrl: 'https://x.example/a', finalUrl: 'https://x.example/a', status: 200, contentType: 'text/html', claims: ['Geologists recommend eating one small rock per day'] });
  assert.ok(d.word_count > 20, `article text kept (got ${d.word_count} words)`);
  assert.equal(d.claim_matches[0].verdict_hint, 'strong_match');
  assert.equal(d.title, 'T');
  assert.ok(!d.paragraphs.some((p) => /doctype/i.test(p.text)), 'doctype never leaks into text');
  assert.equal(d.attribution_sentences[0].named_source, 'Berkeley Geology Department');
});

test('spelled-out numbers, fractions and near-matches count', async () => {
  const html = `<html><body><main><p>Sixty-five percent of enterprises now use generative AI. The average person swallows eight spiders a year. Nearly a third of UK employers have lost staff.</p></main></body></html>`;
  const d = await analyzePage(html, { requestedUrl: 'https://x.example/b', finalUrl: 'https://x.example/b', status: 200, contentType: 'text/html', claims: ['65% of enterprises use generative AI', 'People swallow 8 spiders a year', '33% of UK employers lost staff'] });
  const [c1, c2, c3] = d.claim_matches;
  assert.equal(c1.best[0].number_match, 1, 'sixty-five percent == 65%');
  assert.equal(c2.best[0].number_match, 1, 'eight == 8');
  assert.deepEqual(c2.numbers_in_claim, ['8']);
  assert.equal(c3.best[0].number_match, 1, 'nearly a third == 33%');
});

test('sentence splitter respects abbreviations and reference markers', async () => {
  const html = `<html><body><div id="content"><p>LONDON, Jan. 21, 2026 /PRNewswire/ -- Acme said it grew. Acme Robotics was founded in 2012 in Leeds.<sup class="reference"><a href="#cite_note-1">[1]</a></sup> It builds robots.<sup class="reference"><a href="#cite_note-2">[2]</a></sup></p>
<ol class="references"><li id="cite_note-1"><a href="https://a.example/list">list</a></li><li id="cite_note-2"><a href="https://b.example/about">about</a></li></ol></div></body></html>`;
  const d = await analyzePage(html, { requestedUrl: 'https://x.example/c', finalUrl: 'https://x.example/c', status: 200, contentType: 'text/html', claims: ['Acme Robotics was founded in 2012'] });
  const best = d.claim_matches[0].best[0];
  assert.match(best.sentence, /^Acme Robotics was founded in 2012 in Leeds\./, 'sentence boundary at [1]');
  assert.ok(d.signals.some((s) => s.kind === 'wire' && s.label === 'pr newswire'));
  const att = d.attribution_sentences.find((a) => /PRNewswire/.test(a.sentence));
  assert.ok(att && /^LONDON, Jan\. 21, 2026/.test(att.sentence), 'no split after "Jan."');
});

test('named sources: "According to the Wall Street Journal" wins over a later "from LinkedIn"; last speaker wins for said/told', async () => {
  const html = `<html><body><main><p>According to the Wall Street Journal, a review of data from LinkedIn shows dozens left. A June 6 ProPublica investigation said the tool was error-prone. Engineers found corrosion, department head Rosa Ilić told reporters.</p></main></body></html>`;
  const d = await analyzePage(html, { requestedUrl: 'https://x.example/d', finalUrl: 'https://x.example/d', status: 200, contentType: 'text/html' });
  const names = d.attribution_sentences.map((a) => a.named_source);
  assert.equal(names[0], 'Wall Street Journal');
  assert.equal(names[1], 'ProPublica');
  assert.equal(names[2], 'Rosa Ilić');
});

test('wayback captures: archived links are unwrapped to their original hosts', async () => {
  const html = `<html><body><main><p>See the <a href="https://web.archive.org/web/20150515070013/http://advertising.microsoft.com/en/cl/31966/report">report</a> and <a href="https://web.archive.org/web/20150515070013/https://time.com/other">other</a>.</p></main></body></html>`;
  const u = 'https://web.archive.org/web/20150515070013/https://time.com/3858309/attention-spans-goldfish/';
  const d = await analyzePage(html, { requestedUrl: u, finalUrl: u, status: 200, contentType: 'text/html' });
  assert.deepEqual(d.external_domains, ['microsoft.com']);
  assert.ok(d.external_links[0].archived);
});

test('bot walls and chrome-only pages produce honest hints', async () => {
  const wall = `<html><head><title>Just a moment...</title></head><body><p>Checking your browser before accessing the site.</p></body></html>`;
  const d = await analyzePage(wall, { requestedUrl: 'https://x.example/e', finalUrl: 'https://x.example/e', status: 200, contentType: 'text/html', viaBrowser: true });
  assert.equal(d.bot_wall, true);
  assert.match(d.hint, /even in browser mode/);
  const d2 = await analyzePage(wall, { requestedUrl: 'https://x.example/e', finalUrl: 'https://x.example/e', status: 403, contentType: 'text/html' });
  assert.match(d2.hint, /Retry with --browser/);
});

test('empty / non-html bodies do not crash and carry a hint', async () => {
  const d = await analyzePage('', { requestedUrl: 'https://x/y.pdf', finalUrl: 'https://x/y.pdf', status: 200, contentType: 'application/pdf' });
  assert.equal(d.word_count, 0);
  assert.deepEqual(d.claim_matches, []);
});
