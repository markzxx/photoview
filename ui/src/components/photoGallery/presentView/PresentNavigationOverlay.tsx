import React, { useState, useRef, useEffect } from 'react'
import styled from 'styled-components'
import { debounce, DebouncedFn } from '../../../helpers/utils'
import { closePresentModeAction, GalleryAction } from '../mediaGalleryReducer'

import { useSwipeable } from 'react-swipeable'

import ExitIcon from './icons/Exit'
import NextIcon from './icons/Next'
import PrevIcon from './icons/Previous'
const StyledOverlayContainer = styled.div`
  width: 100%;
  height: 100%;
  position: relative;
`

const OverlayButton = styled.button`
  width: 64px;
  height: 64px;
  background: none;
  border: none;
  outline: none;
  cursor: pointer;
  position: absolute;

  & svg {
    width: 32px;
    height: 32px;
    overflow: visible !important;
  }

  & svg path {
    stroke: Gray;
    stroke-width: 10;
  }
`

const ExitButton = styled(OverlayButton)`
  height: 20%;
  width: 25%;
    & svg {
    margin: auto;
    width: 48px;
    height: 64px;
  }
`

const NavigationButton = styled(OverlayButton)<{ align: 'left' | 'right' }>`
  height: 20%;
  width: 20%;
  top: 40%;

  ${({ align: float }) => (float == 'left' ? 'left: 0;' : null)}
  ${({ align: float }) => (float == 'right' ? 'right: 0;' : null)}

  & svg {
    margin: auto;
    width: 48px;
    height: 64px;
  }
`

type PresentNavigationOverlayProps = {
  children?: React.ReactChild
  dispatchMedia: React.Dispatch<GalleryAction>
  disableSaveCloseInHistory?: boolean
}

const PresentNavigationOverlay = ({
  children,
  dispatchMedia,
  disableSaveCloseInHistory,
}: PresentNavigationOverlayProps) => {
  // const [hide, setHide] = useState(true)
  // const onMouseMove = useRef<null | DebouncedFn<() => void>>(null)

  // useEffect(() => {
  //   onMouseMove.current = debounce(
  //     () => {
  //       setHide(hide => !hide)
  //     },
  //     2000,
  //     true
  //   )
  //
  //   return () => {
  //     onMouseMove.current?.cancel()
  //   }
  // }, [])

  const handlers = useSwipeable({
    // onSwipedLeft: () => dispatchMedia({ type: 'nextImage' }),
    // onSwipedRight: () => dispatchMedia({ type: 'previousImage' }),
    // preventScrollOnSwipe: true,
    // trackTouch: false,
  })

  return (
    <StyledOverlayContainer
      data-testid="present-overlay"
      onMouseMove={() => {
        // onMouseMove.current && onMouseMove.current()
      }}
    >
    <div {...handlers}>
      {children}
      <NavigationButton
        aria-label="Previous image"
        // className={hide ? 'hide' : undefined}
        align="left"
        onClick={() => dispatchMedia({ type: 'previousImage' })}
      >
        <PrevIcon />
      </NavigationButton>
      <NavigationButton
        aria-label="Next image"
        // className={hide ? 'hide' : undefined}
        align="right"
        onClick={() => dispatchMedia({ type: 'nextImage' })}
      >
        <NextIcon />
      </NavigationButton>
      <ExitButton
        aria-label="Exit presentation mode"
        // className={hide ? 'hide' : undefined}
        onClick={() => {
          if (disableSaveCloseInHistory === true) {
            dispatchMedia({ type: 'closePresentMode' })
          } else {
            closePresentModeAction({ dispatchMedia })
          }
        }}
      >
        <ExitIcon />
      </ExitButton>
    </div>
    </StyledOverlayContainer>
  )
}

export default PresentNavigationOverlay
