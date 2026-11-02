// FormField のヒント・エラー・文字数を入力に結ぶ aria-describedby。
// 無い要素の id は読み上げで無視されるので、常に3つとも並べてよい
export function describedBy(id: string): string {
  return `${id}-hint ${id}-error ${id}-count`
}
