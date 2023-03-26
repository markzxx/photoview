package routes

import (
	"github.com/pkg/errors"
	"net/http"

	"github.com/photoview/photoview/api/graphql/auth"
	"github.com/photoview/photoview/api/graphql/models"
	"gorm.io/gorm"
)

func authenticateMedia(media *models.Media, db *gorm.DB, r *http.Request) (success bool, responseMessage string, responseStatus int, errorMessage error) {
	if media == nil {
		return false, "internal server error", http.StatusInternalServerError, errors.New("media not found")
	}
	user := auth.UserFromContext(r.Context())

	if user != nil {
		var album models.Album
		if err := db.First(&album, media.AlbumID).Error; err != nil {
			return false, "internal server error", http.StatusInternalServerError, err
		}

		ownsAlbum, err := user.OwnsAlbum(db, &album)
		if err != nil {
			return false, "internal server error", http.StatusInternalServerError, err
		}

		if !ownsAlbum {
			return false, "invalid credentials", http.StatusForbidden, nil
		}
	}

	return true, "success", http.StatusAccepted, nil
}

func authenticateAlbum(album *models.Album, db *gorm.DB, r *http.Request) (success bool, responseMessage string, responseStatus int, errorMessage error) {
	user := auth.UserFromContext(r.Context())

	if user != nil {
		ownsAlbum, err := user.OwnsAlbum(db, album)
		if err != nil {
			return false, "internal server error", http.StatusInternalServerError, err
		}

		if !ownsAlbum {
			return false, "invalid credentials", http.StatusForbidden, nil
		}
	}

	return true, "success", http.StatusAccepted, nil
}
