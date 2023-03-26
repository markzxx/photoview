package resolvers

import (
	"context"
	"fmt"
	"github.com/photoview/photoview/api/utils"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"

	"github.com/photoview/photoview/api/database/drivers"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/periodic_scanner"
	"github.com/photoview/photoview/api/scanner/scanner_queue"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func (r *mutationResolver) ScanAll(ctx context.Context) (*models.ScannerResult, error) {
	err := scanner_queue.AddAllToQueue(true)
	if err != nil {
		return nil, err
	}

	startMessage := "Scanner started"

	return &models.ScannerResult{
		Finished: false,
		Success:  true,
		Message:  &startMessage,
	}, nil
}

func (r *mutationResolver) ScanUser(ctx context.Context, userID int) (*models.ScannerResult, error) {
	var user models.User
	if err := r.DB(ctx).First(&user, userID).Error; err != nil {
		return nil, errors.Wrap(err, "get user from database")
	}

	scanner_queue.AddUserToQueue(&user, true)

	startMessage := "Scanner started"
	return &models.ScannerResult{
		Finished: false,
		Success:  true,
		Message:  &startMessage,
	}, nil
}

func (r *mutationResolver) SetPeriodicScanInterval(ctx context.Context, interval int) (int, error) {
	db := r.DB(ctx)
	if interval < 0 {
		return 0, errors.New("interval must be 0 or above")
	}

	if err := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Model(&models.SiteInfo{}).Update("periodic_scan_interval", interval).Error; err != nil {
		return 0, err
	}

	var siteInfo models.SiteInfo
	if err := db.First(&siteInfo).Error; err != nil {
		return 0, err
	}

	periodic_scanner.ChangePeriodicScanInterval(time.Duration(siteInfo.PeriodicScanInterval) * time.Second)

	return siteInfo.PeriodicScanInterval, nil
}

func (r *mutationResolver) SetScannerConcurrentWorkers(ctx context.Context, workers int) (int, error) {
	db := r.DB(ctx)
	if workers < 1 {
		return 0, errors.New("concurrent workers must at least be 1")
	}

	if workers > 1 && drivers.DatabaseDriverFromEnv() == drivers.SQLITE {
		return 0, errors.New("multiple workers not supported for SQLite databases")
	}

	if err := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Model(&models.SiteInfo{}).Update("concurrent_workers", workers).Error; err != nil {
		return 0, err
	}

	var siteInfo models.SiteInfo
	if err := db.First(&siteInfo).Error; err != nil {
		return 0, err
	}

	scanner_queue.ChangeScannerConcurrentWorkers(siteInfo.ConcurrentWorkers)

	return siteInfo.ConcurrentWorkers, nil
}

func (r *mutationResolver) MarkModify(ctx context.Context, modPath string) (int, error) {
	db := r.DB(ctx)

	for modPath != "/" {
		var album models.Album
		err := db.Where("path_hash = ?", models.MD5Hash(modPath)).First(&album).Error
		if err == nil {
			db.Model(&album).Update("last_modify_time", time.Now().UTC().Unix())
		}
		modPath = path.Dir(modPath)
	}

	return 0, nil
}

const finalDir = "/data/文档同步/双向同步/客户照片修图"

func (r *mutationResolver) MakeFinalDir(ctx context.Context, albumID int) (int, error) {
	album := &models.Album{}
	db := r.DB(ctx)
	db.First(album, albumID)
	if album.Title == "精修图片" {
		parent := &models.Album{}
		db.First(parent, album.ParentAlbumID)
		album = parent
		log.Println("find parent dir", album.Title)
	}
	newRootPath := path.Join(finalDir, album.Title)
	newOriginPath := path.Join(newRootPath, "原图")
	os.MkdirAll(newOriginPath, os.ModePerm)
	dirContent, _ := ioutil.ReadDir(album.Path)
	for _, item := range dirContent {
		if item.IsDir() {
			CopyDir(path.Join(album.Path, item.Name()), path.Join(newRootPath, item.Name()))
		} else {
			CopyFile(path.Join(album.Path, item.Name()), path.Join(newOriginPath, item.Name()))
		}
	}

	return 0, nil
}

func (r *mutationResolver) MarkRetouchFile(ctx context.Context, albumID int) (int, error) {
	album := &models.Album{}
	db := r.DB(ctx)
	db.First(album, albumID)

	var medias []models.Media
	db.Where("album_id = ?", albumID).Find(&medias)
	for _, media := range medias {
		var favorite models.UserMediaData
		db.Where("media_id = ?", media.ID).First(&favorite)
		if favorite.Favorite {
			os.Rename(media.Path, utils.RemoveBolanghao(media.Path))
		} else {
			os.Rename(media.Path, utils.AddBolanghao(media.Path))
		}
	}

	return 0, nil
}

// File copies a single file from src to dst
func CopyFile(src, dst string) error {
	var err error
	var srcfd *os.File
	var dstfd *os.File
	var srcinfo os.FileInfo

	if srcfd, err = os.Open(src); err != nil {
		return err
	}
	defer srcfd.Close()

	if dstfd, err = os.Create(dst); err != nil {
		return err
	}
	defer dstfd.Close()

	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return err
	}
	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}
	return os.Chmod(dst, srcinfo.Mode())
}

// Dir copies a whole directory recursively
func CopyDir(src string, dst string) error {
	var err error
	var fds []os.FileInfo
	var srcinfo os.FileInfo

	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}

	if err = os.MkdirAll(dst, srcinfo.Mode()); err != nil {
		return err
	}

	if fds, err = ioutil.ReadDir(src); err != nil {
		return err
	}
	for _, fd := range fds {
		srcfp := path.Join(src, fd.Name())
		dstfp := path.Join(dst, fd.Name())

		if fd.IsDir() {
			if err = CopyDir(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		} else {
			if err = CopyFile(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}
