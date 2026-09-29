type RatingStarsProps = {
  rating: number
  size?: 'sm' | 'lg'
}

export function RatingStars({ rating, size = 'sm' }: RatingStarsProps) {
  return (
    // 星の記号は読み上げで意味をなさないため、まとめて1つの画像として評価を読ませる
    <span
      role="img"
      aria-label={`評価 ${rating} / 5`}
      className={`text-brand-700 ${size === 'lg' ? 'text-[22px] tracking-[2px]' : 'text-sm tracking-[1px]'}`}
    >
      {'★'.repeat(rating) + '☆'.repeat(5 - rating)}
    </span>
  )
}
