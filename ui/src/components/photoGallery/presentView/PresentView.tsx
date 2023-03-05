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
  bottom: 5%;
`

const PresentDeleteButton = styled.button`
  background: none;
  border: none;
  outline: none;
  cursor: pointer;
  position: absolute;
  right: 10%;
  bottom: 5%;
`

const PresentTitle = styled.div`
  width: 100%;
  position: absolute;
  margin-left: auto;
  margin-right: auto;
  text-align: center;
  top: 0;
  font-size: 30px;
  -webkit-text-stroke: 1px black;
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
        // viewBox="0 0 19 17"
        // version="1.1"
      >
        <path
          d="m90.93581,14.56901q3.76401,0 6.21061,1.03818t3.88948,2.59544t2.00747,3.37407t0.5646,3.37407q0,0.72672 -0.06273,1.14199t-0.06273,0.72672l0,0.62291l-8.28082,0l0,54.50422q0,2.07635 -1.00374,3.99698t-2.88574,3.37407t-4.57954,2.3359t-6.21061,0.88245l-55.95826,0q-3.26214,0 -6.14788,-0.83054t-4.95594,-2.3359t-3.26214,-3.63361t-1.19194,-4.82752l0,-53.46605l-7.77895,0q-0.12547,-0.10382 -0.12547,-0.51909q-0.12547,-0.51909 -0.12547,-3.21834q0,-1.34963 0.7528,-3.01071t2.25841,-3.06262t3.88948,-2.3359t5.64601,-0.93436l11.16656,0l0,-6.33287q0,-2.69926 2.25841,-4.61988t5.52055,-1.92063l39.27116,0q4.39134,0 6.08515,1.92063t1.6938,4.61988l0,6.43669q2.63481,0.10382 5.64601,0.10382l5.77148,0zm-58.4676,0l39.27116,0l0,-6.54051l-39.27116,0l0,6.54051zm-3.88948,65.71652q4.01494,0 4.01494,-4.25652l0,-48.17135l-7.77895,0l0,48.17135q0,2.18017 0.81554,3.21834t2.94847,1.03818zm23.71326,-0.10382q2.13294,0 2.88574,-0.98627t0.7528,-3.16644l0,-48.17135l-7.77895,0l0,48.17135q0,4.1527 4.14041,4.1527zm23.58779,-0.20764q2.25841,0 3.01121,-0.98627t0.7528,-3.16644l0,-47.96372l-7.90442,0l0,47.96372q0,4.1527 4.14041,4.1527z"
          strokeWidth="100"
          color="gray"
          fill="currentColor"
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
            e.stopPropagation()
            toggleFavoriteAction({
                media: activeMedia,
                markFavorite
            })
          }}
      />
      <PresentDeleteIcon
        onClick={e => {
          e.stopPropagation()
          var r = confirm("确认删除")
          if (r)
          {
            toggleDeleteAction({
              media: activeMedia,
              deleteMutation
            })
            dispatchMedia({ type: 'deleteMedia' })
          }
        }}
      />
      <PresentTitle>
        {activeMedia.title.match(/[0-9]+/i) ? activeMedia.title.match(/[0-9]+/i)[0] : activeMedia.title}
      </PresentTitle>
    </StyledContainer>
  )
}

export default PresentView
