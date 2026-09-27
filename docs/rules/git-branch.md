# ブランチ運用ルール

本プロジェクトのブランチ運用ルール。Git Flow ベースで、`develop` に開発を集約し、リリース時に `main` へマージする。

## 永続ブランチ（直接コミット禁止）

- `main`: 本番。リリース時のみ `develop` か `hotfix/*` からマージ
- `develop`: 開発統合先。すべての作業ブランチはここから分岐・ここへマージ

## 作業ブランチ（プレフィックス必須）

| プレフィックス | 用途 | 分岐元 / マージ先 |
|--------------|------|------------------|
| `feat/` | 新機能・UI 追加 | `develop` → `develop` |
| `fix/` | バグ修正（非緊急） | `develop` → `develop` |
| `refactor/` | 内部構造改善（挙動不変） | `develop` → `develop` |
| `chore/` | CI / 設定 / 依存更新など | `develop` → `develop` |
| `hotfix/` | 本番の緊急修正 | `main` → `main` + `develop` |

⚠️ **`feature/` ではなく `feat/`**。

⚠️ **`test/` プレフィックスは使わない**。テスト追加は対象機能の `feat/` / `fix/` ブランチに同梱する（コミットメッセージは `test:` プレフィックス）。

## 命名規則

```
<type>/<short-description>
```

- 小文字 + kebab-case
- 英語で簡潔に内容を表す
- **ブランチ名に Issue 番号を含めない**（`issue-83` 等）。人が見て何の作業かわかる名前にする。Issue との紐付けは PR 本文で行う

✅ `chore/setup-golangci-lint` / `refactor/cleanup-book-usecase` / `feat/author-book-list`
❌ `fix/bug`（不明確）/ `feature/foo`（プレフィックス違反）/ `feat/issue-83-xxx`（Issue 番号）

## 通常フロー

1. `develop` から作業ブランチを切る
2. 実装 → PR を `develop` へ
3. Copilot レビュー + 人間レビュー通過後にマージ

## 緊急修正フロー（hotfix）

1. `main` から `hotfix/*` を切る
2. PR を `main` へ作成しマージ → 本番反映
3. 同内容を `develop` にも反映（取りこぼし防止）

関連: [コミットメッセージ規約](./git-commit-message.md) / [PR 運用ルール](./git-pull-request.md)