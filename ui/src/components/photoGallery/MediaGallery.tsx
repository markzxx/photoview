import React, { useContext } from 'react'
import styled from 'styled-components'
import { MediaThumbnail, MediaPlaceholder } from './MediaThumbnail'
import PresentView from './presentView/PresentView'
import {
  openPresentModeAction,
  PhotoGalleryAction,
  MediaGalleryState,
} from './mediaGalleryReducer'
import {
  toggleFavoriteAction,
  useMarkFavoriteMutation,
} from './photoGalleryMutations'

import { gql } from '@apollo/client'

const Gallery = styled.div`
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  min-height: 200px;
  position: relative;
  margin: -4px;

  @media (max-width: 1000px) {
    /* Compensate for tab bar on mobile */
    margin-bottom: 76px;
  }
`

export const PhotoFiller = styled.div`
  height: 200px;
  flex-grow: 999999;
`

export const MEDIA_GALLERY_FRAGMENT = gql`
  fragment MediaGalleryFields on Media {
    id
    type
    title
    blurhash
    thumbnail {
      url
      width
      height
    }
    highRes {
      url
    }
    original {
      url
    }
    videoWeb {
      url
    }
    favorite
  }
`

type MediaGalleryProps = {
  loading: boolean
  mediaState: MediaGalleryState
  dispatchMedia: React.Dispatch<PhotoGalleryAction>
}

const MediaGallery = ({ mediaState, dispatchMedia }: MediaGalleryProps) => {
  const [markFavorite] = useMarkFavoriteMutation()

  const { media, activeIndex, presenting } = mediaState
  let active = media[activeIndex]

  let mediaElements = []
  if (media) {
    mediaElements = media.map((media, index) => {
      const active = activeIndex == index

      return (
        <MediaThumbnail
          key={media.id}
          media={media}
          active={active}
          clickFavorite={() => {
            toggleFavoriteAction({
              media,
              markFavorite,
            })
          }}
          clickPresent={() => {
            openPresentModeAction({ dispatchMedia, activeIndex: index })
          }}
        />
      )
    })
  } else {
    for (let i = 0; i < 6; i++) {
      mediaElements.push(<MediaPlaceholder key={i} />)
    }
  }

  return (
    <>
      <Gallery data-testid="photo-gallery-wrapper">
        {mediaElements}
        <PhotoFiller />
      </Gallery>
      {presenting && (
        <PresentView
          activeMedia={active}
          dispatchMedia={dispatchMedia}
        />
      )}
    </>
  )
}

export default MediaGallery
