import React, { useEffect } from 'react'
import styled, { createGlobalStyle } from 'styled-components'
import PresentNavigationOverlay from './PresentNavigationOverlay'
import PresentMedia from './PresentMedia'
import { closePresentModeAction, GalleryAction } from '../mediaGalleryReducer'
import { MediaGalleryFields } from '../__generated__/MediaGalleryFields'
import {
  toggleDeleteAction,
  toggleFavoriteAction,
  useDeleteMediaMutation,
  useMarkFavoriteMutation
} from "../photoGalleryMutations";

const StyledContainer = styled.div`
  position: fixed;
  width: 100vw;
  height: 100vh;
  background-color: black;
  color: white;
  top: 0;
  left: 0;
  z-index: 100;
`

const PreventScroll = createGlobalStyle`
  * {
    overflow: hidden !important;
  }
`

const PresentFavoriteButton = styled.button`
  background: none;
  border: none;
  outline: none;
  cursor: pointer;
  position: absolute;
  left: 10%;
  top: 80%;
`

const PresentDeleteButton = styled.button`
  background: none;
  border: none;
  outline: none;
  cursor: pointer;
  position: absolute;
  left: 85%;
  top: 80%;
`

type FavoriteIconProps = {
  favorite: boolean
  onClick(e: React.MouseEvent<HTMLButtonElement, MouseEvent>): void
}

const PresentFavoriteIcon = ({ favorite, onClick }: FavoriteIconProps) => {
  return (
      <PresentFavoriteButton
          onClick={onClick}
          style={{
            opacity: favorite ? '0.75' : '0.5'
          }}
      >
        <svg
            className="text-white m-auto mt-1"
            width="100px"
            height="100px"
            viewBox="0 0 19 17"
            version="1.1"
        >
          <path
              d="M13.999086,1 C15.0573371,1 16.0710089,1.43342987 16.8190212,2.20112483 C17.5765039,2.97781012 18,4.03198704 18,5.13009709 C18,6.22820714 17.5765039,7.28238406 16.8188574,8.05923734 L16.8188574,8.05923734 L15.8553647,9.04761889 L9.49975689,15.5674041 L3.14414912,9.04761889 L2.18065643,8.05923735 C1.39216493,7.2503776 0.999999992,6.18971057 1,5.13009711 C1.00000001,4.07048366 1.39216496,3.00981663 2.18065647,2.20095689 C2.95931483,1.40218431 3.97927681,1.00049878 5.00042783,1.00049878 C6.02157882,1.00049878 7.04154078,1.4021843 7.82019912,2.20095684 L7.82019912,2.20095684 L9.4997569,3.92390079 L11.1794784,2.20078881 C11.9271631,1.43342987 12.9408349,1 13.999086,1 L13.999086,1 Z"
              fill={favorite ? 'currentColor' : 'none'}
              stroke="currentColor"
              strokeWidth={favorite ? '0' : '2'}
              color="red"
          ></path>
        </svg>
      </PresentFavoriteButton>
  )
}

type DeleteIconProps = {
  onClick(e: React.MouseEvent<HTMLButtonElement, MouseEvent>): void
}

const PresentDeleteIcon = ({ onClick }: DeleteIconProps) => {
  return (
    <PresentDeleteButton
      onClick={onClick}
    >
      <svg
        className="text-white m-auto mt-1"
        width="100px"
        height="100px"
        viewBox="0 0 19 17"
        version="1.1"
      >
        <path
          d="m0,7.99248l0,0c0,-4.6944 3.80559,-8.49998 8.5,-8.49998l0,0c2.25437,0 4.4164,0.89553 6.01041,2.4896c1.59407,1.59406 2.4896,3.75608 2.4896,6.01039l0,0c0,4.69446 -3.80556,8.50002 -8.50001,8.50002l0,0c-4.69441,0 -8.5,-3.80556 -8.5,-8.50002l0,0zm13.72557,3.80247l0,0c1.87142,-2.5719 1.59324,-6.1231 -0.65585,-8.37216c-2.24909,-2.24908 -5.80029,-2.52726 -8.37213,-0.65582l9.02798,9.02798zm-10.4511,-7.60486c-1.87143,2.57189 -1.59326,6.12309 0.65582,8.37212c2.24907,2.24909 5.80028,2.52727 8.37211,0.65585l-9.02793,-9.02797l0,0z"          fill='currentColor'
          strokeWidth="2"
          color="gray"
        ></path>
      </svg>
    </PresentDeleteButton>
  )
}

type PresentViewProps = {
  className?: string
  imageLoaded?(): void
  activeMedia: MediaGalleryFields
  dispatchMedia: React.Dispatch<GalleryAction>
  disableSaveCloseInHistory?: boolean
}

const PresentView = ({
  className,
  imageLoaded,
  activeMedia,
  dispatchMedia,
  disableSaveCloseInHistory,
}: PresentViewProps) => {
  useEffect(() => {
    const keyDownEvent = (e: KeyboardEvent) => {
      if (e.key == 'ArrowRight') {
        e.stopPropagation()
        dispatchMedia({ type: 'nextImage' })
      }

      if (e.key == 'ArrowLeft') {
        e.stopPropagation()
        dispatchMedia({ type: 'previousImage' })
      }

      if (e.key == 'Escape') {
        e.stopPropagation()

        if (disableSaveCloseInHistory === true) {
          dispatchMedia({ type: 'closePresentMode' })
        } else {
          closePresentModeAction({ dispatchMedia })
        }
      }
    }

    document.addEventListener('keydown', keyDownEvent)

    return function cleanup() {
      document.removeEventListener('keydown', keyDownEvent)
    }
  })
    const [markFavorite] = useMarkFavoriteMutation()
    const [deleteMutation] = useDeleteMediaMutation()

    return (
    <StyledContainer className={className}>
      <PreventScroll />
      <PresentNavigationOverlay
        dispatchMedia={dispatchMedia}
        disableSaveCloseInHistory
      >
        <PresentMedia media={activeMedia} imageLoaded={imageLoaded} />
      </PresentNavigationOverlay>
      <PresentFavoriteIcon
          favorite={activeMedia.favorite}
          onClick={e => {
            toggleFavoriteAction({
                media: activeMedia,
                markFavorite
            })
            e.stopPropagation()
          }}
      />
      <PresentDeleteIcon
        onClick={e => {
          toggleDeleteAction({
            media: activeMedia,
            deleteMutation
          })
          dispatchMedia({ type: 'deleteMedia' })
          e.stopPropagation()
        }}
      />
    </StyledContainer>
  )
}

export default PresentView
