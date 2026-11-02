// YAML で使っている制約の種類ごとに、Orval が zod へ正しく変換しているかを 1 組ずつ確かめる
// （docs/rules/testing.md「フロントエンド: 何をテストするか」の入力チェック）。
// 制約の値が仕様どおりかは BE の Presentation のテストが担保するので、ここでは種類だけを見る。
import { describe, expect, it } from 'vitest'
import {
  LoginBody,
  ReadBookParams,
  RegisterBookBody,
  SearchBooksQueryParams,
} from './generated/zod/api'

const validBook = {
  isbn: '9784873118703',
  summary: 'まとめ',
  comment: '感想',
  rating: 3,
}

describe('生成された zod が OpenAPI の制約の種類を検証する', () => {
  it.each([
    ['required: 値があれば通る', {}, true],
    ['required: 欠けたら弾く', { summary: undefined }, false],
    ['minLength: 1 文字は通る', { summary: 'あ' }, true],
    ['minLength: 0 文字は弾く', { summary: '' }, false],
    ['maxLength: 上限の 100 文字は通る', { summary: 'あ'.repeat(100) }, true],
    ['maxLength: 101 文字は弾く', { summary: 'あ'.repeat(101) }, false],
    ['minimum: 下限の 1 は通る', { rating: 1 }, true],
    ['minimum: 0 は弾く', { rating: 0 }, false],
    ['maximum: 上限の 5 は通る', { rating: 5 }, true],
    ['maximum: 6 は弾く', { rating: 6 }, false],
    [
      'maxItems: 上限の 10 件は通る',
      { tagIds: Array.from({ length: 10 }, (_, i) => i + 1) },
      true,
    ],
    [
      'maxItems: 11 件は弾く',
      { tagIds: Array.from({ length: 11 }, (_, i) => i + 1) },
      false,
    ],
    ['type: 整数は通る', { rating: 2 }, true],
    ['type: 小数は弾く', { rating: 2.5 }, false],
  ])('%s', (_, override, ok) => {
    expect(
      RegisterBookBody.safeParse({ ...validBook, ...override }).success,
    ).toBe(ok)
  })

  // 文字数は BE（utf8.RuneCountInString）と同じくコードポイントで数える。zod 4 の .max() はコードポイントで数えるので、
  // 補助平面の文字（UTF-16 で 2 単位）も 1 文字として上限ちょうどまで通ることを確かめる
  it.each([
    ['maxLength: 絵文字 100 文字は通る', '📚'.repeat(100), true],
    ['maxLength: 絵文字 101 文字は弾く', '📚'.repeat(101), false],
  ])('%s', (_, summary, ok) => {
    expect(RegisterBookBody.safeParse({ ...validBook, summary }).success).toBe(
      ok,
    )
  })

  it.each([
    ['minimum: path の ID は 1 なら通る', 1, true],
    ['minimum: path の ID は 0 なら弾く', 0, false],
  ])('%s', (_, id, ok) => {
    expect(ReadBookParams.safeParse({ id }).success).toBe(ok)
  })

  it.each([
    ['default: limit を省略すると既定値になる', {}, true],
    ['maximum: limit は 100 なら通る', { limit: 100 }, true],
    ['maximum: limit は 101 なら弾く', { limit: 101 }, false],
  ])('%s', (_, query, ok) => {
    expect(SearchBooksQueryParams.safeParse(query).success).toBe(ok)
  })

  it.each([
    [
      'required: ID とパスワードがあれば通る',
      { id: 'admin', password: 'x' },
      true,
    ],
    ['required: パスワードが欠けたら弾く', { id: 'admin' }, false],
  ])('%s', (_, body, ok) => {
    expect(LoginBody.safeParse(body).success).toBe(ok)
  })
})
