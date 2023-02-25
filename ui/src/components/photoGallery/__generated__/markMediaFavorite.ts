/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: markMediaFavorite
// ====================================================

import {albumQuery_album} from "../../../Pages/AlbumPage/__generated__/albumQuery";

export interface markMediaFavorite_favoriteMedia {
  __typename: "Media";
  id: string;
  favorite: boolean;
}

export interface markMediaFavorite {
  /**
   * Mark or unmark a media as being a favorite
   */
  favoriteMedia: markMediaFavorite_favoriteMedia;
}

export interface markMediaFavoriteVariables {
  mediaId: string;
  favorite: boolean;
}

export interface deleteMedia {
  /**
   * Delete a media from filesystem and database
   */
  Album: albumQuery_album;
}

export interface deleteMediaVariables {
  mediaId: string;
}