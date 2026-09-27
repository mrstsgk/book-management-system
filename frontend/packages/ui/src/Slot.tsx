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
