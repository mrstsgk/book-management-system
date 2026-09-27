/**
 * Web 共通 UI（デジタル庁 DS 準拠）。
 * 必要なコンポーネントを上流から取り込み、ここ経由でアプリが参照する。
 * 上流: https://github.com/digital-go-jp/design-system-example-components-react
 * （npm に dist が無いため zip/丸ごと vendor せず、使う部品だけ取り込む）
 */
export {
  Button,
  buttonBaseStyle,
  buttonSizeStyle,
  buttonVariantStyle,
  type ButtonProps,
  type ButtonSize,
  type ButtonVariant,
} from './Button'
