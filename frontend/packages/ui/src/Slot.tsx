import {
  Children,
  cloneElement,
  type HTMLAttributes,
  isValidElement,
  type ReactNode,
} from 'react'

type SlotProps = HTMLAttributes<HTMLElement> & {
  children?: ReactNode
}

/** @internal デジタル庁 DS Slot（asChild 用） */
// 子の要素へ props を合成するのが Slot の役割そのもので、cloneElement・Children の代わりがない
/* eslint-disable @eslint-react/no-clone-element, @eslint-react/no-children-count, @eslint-react/no-children-only */
export function Slot(props: SlotProps) {
  const { children, ...rest } = props

  if (isValidElement(children)) {
    const child = children as React.ReactElement<{ className?: string }>
    return cloneElement(child, {
      ...rest,
      ...child.props,
      className: `${rest.className ?? ''} ${child.props.className ?? ''}`,
    })
  }

  if (Children.count(children) > 1) {
    Children.only(null)
  }

  return null
}
/* eslint-enable @eslint-react/no-clone-element, @eslint-react/no-children-count, @eslint-react/no-children-only */
