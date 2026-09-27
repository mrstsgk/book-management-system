# @book-management/ui

デジタル庁デザインシステムの **Tailwind テーマ**と、利用する React 部品を Web アプリ共通で提供する。

- テーマ: `@digital-go-jp/tailwind-theme-plugin`（npm）
- 部品: 上流ソースから必要なものだけ取り込み（現状 `Button`）
- 上流: https://github.com/digital-go-jp/design-system-example-components-react  
  （npm に dist が無いため、リポジトリ丸ごと vendor / zip はしない）

追加部品が必要になったら:

1. 公式 usage を読む（https://design.digital.go.jp/dads/ …/usage/）
2. 上流ソースから該当コンポーネントを `src/` に取り込む
3. `index.ts` で export し、アプリ側 Storybook に状態を載せる
4. PR 本文に usage URL を1行書く
