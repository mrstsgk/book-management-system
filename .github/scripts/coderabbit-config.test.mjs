import assert from 'node:assert/strict';
import { readFileSync, existsSync } from 'node:fs';
import test from 'node:test';
import config from '../../.coderabbit.config.ts';

const resolve = (labels, isDraft = false) => config({
  invocation: { source: 'review' },
  pr: { labels, isDraft },
});

for (const [label, profile, drafts] of [
  ['review:design', 'assertive', true],
  ['review:standard', 'chill', false],
  ['review:critical', 'assertive', false],
]) {
  test(`${label}: profile and Draft policy`, () => {
    for (const isDraft of [true, false]) {
      const result = resolve(['unrelated', label], isDraft);
      assert.equal(result.reviews.profile, profile);
      assert.equal(result.reviews.auto_review.enabled, true);
      assert.equal(result.reviews.auto_review.drafts, drafts);
      assert.equal(result.reviews.auto_apply_labels, false);
    }
  });
}

test('review:skip disables automatic review including Drafts', () => {
  for (const isDraft of [true, false]) {
    const result = resolve(['unrelated', 'review:skip'], isDraft);
    assert.equal(result.reviews.auto_review.enabled, false);
    assert.equal(result.reviews.auto_review.drafts, false);
    assert.equal(result.reviews.auto_apply_labels, false);
  }
});

test('missing, unknown and all conflicting combinations do not start automatic review', () => {
  const labels = ['review:design', 'review:standard', 'review:critical', 'review:skip'];
  for (const selection of [[], ['review:unknown'], ...labels.flatMap((a, i) => labels.slice(i + 1).map(b => [a, b])), labels]) {
    assert.equal(resolve(selection).reviews.auto_review.enabled, false);
    assert.equal(resolve(selection).reviews.auto_review.drafts, false);
  }
});

test('null PR context is safe; local CLI does not infer human classification', () => {
  assert.equal(config({ invocation: { source: 'review' }, pr: null }).reviews.auto_review.enabled, false);
  const local = config({ invocation: { source: 'cli' }, pr: null });
  assert.equal(local.reviews.profile, 'chill');
  assert.equal(local.reviews.auto_review.enabled, true);
});

test('transition from design to implementation updates both profile and guidance', () => {
  const design = resolve(['review:design']).reviews;
  const standard = resolve(['review:standard']).reviews;
  const critical = resolve(['review:critical']).reviews;
  assert.match(design.path_instructions[0].instructions, /Do not request production implementation/);
  assert.match(standard.path_instructions[0].instructions, /Propose changes to an agreed design only/);
  assert.doesNotMatch(standard.path_instructions[0].instructions, /Do not request production implementation/);
  assert.match(critical.path_instructions[0].instructions, /concurrency, idempotency/);
  assert.deepEqual(standard.path_filters, design.path_filters);
  assert.deepEqual(standard.auto_review.base_branches, ['develop', 'main']);
});

test('TypeScript config is not shadowed and PR template matches the policy copy', () => {
  assert.equal(existsSync('.coderabbit.yaml'), false);
  assert.equal(existsSync('.coderabbit.yml'), false);
  const template = readFileSync('.github/pull_request_template.md', 'utf8').trim();
  const policy = readFileSync('docs/rules/git-pull-request.md', 'utf8');
  assert.equal(policy.split('````markdown\n')[1].split('\n````')[0].trim(), template);
});
