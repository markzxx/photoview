package scanner

import (
	"context"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/media_encoding"
	"github.com/photoview/photoview/api/scanner/scanner_cache"
	"github.com/photoview/photoview/api/scanner/scanner_task"
	"github.com/photoview/photoview/api/utils"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"log"
	"os"
	"path"
)

func ScanMedia(tx *gorm.DB, mediaPath string, albumId int, cache *scanner_cache.AlbumScannerCache) (*models.Media, bool, error) {
	mediaName := path.Base(mediaPath)

	// Check if media already exists
	{
		var media []*models.Media

		result := tx.Where("path_hash = ?", models.MD5Hash(utils.RemoveSymbol(mediaPath))).Find(&media)

		if result.Error != nil {
			return nil, false, errors.Wrap(result.Error, "scan media fetch from database")
		}

		if result.RowsAffected > 0 {
			// log.Printf("Media already scanned: %s\n", mediaPath)
			if media[0].Title != mediaName {
				media[0].Title = mediaName
				media[0].Path = mediaPath
				tx.Save(&media[0])
			}
			return media[0], false, nil
		}
	}

	log.Printf("Scanning media: %s\n", mediaPath)

	mediaType, err := cache.GetMediaType(mediaPath)
	if err != nil {
		return nil, false, errors.Wrap(err, "could determine if media was photo or video")
	}

	var mediaTypeText models.MediaType

	if mediaType.IsVideo() {
		mediaTypeText = models.MediaTypeVideo
	} else {
		mediaTypeText = models.MediaTypePhoto
	}

	stat, err := os.Stat(mediaPath)
	if err != nil {
		return nil, false, err
	}

	media := models.Media{
		Title:    mediaName,
		Path:     mediaPath,
		AlbumID:  albumId,
		Type:     mediaTypeText,
		DateShot: stat.ModTime(),
	}

	if err := tx.Create(&media).Error; err != nil {
		return nil, false, errors.Wrap(err, "could not insert media into database")
	}

	return &media, true, nil
}

// ProcessSingleMedia processes a single media, might be used to reprocess media with corrupted cache
// Function waits for processing to finish before returning.
func ProcessSingleMedia(db *gorm.DB, media *models.Media, album *models.Album) error {
	album_cache := scanner_cache.MakeAlbumCache()

	if album == nil {
		album = &models.Album{}
		if err := db.Model(media).Association("Album").Find(&album); err != nil {
			return err
		}
	}

	media_data := media_encoding.NewEncodeMediaData(media)

	task_context := scanner_task.NewTaskContext(context.Background(), db, album, album_cache)
	if err := scanMedia(task_context, media, &media_data, 0, 1); err != nil {
		return errors.Wrap(err, "single media scan")
	}

	return nil
}
