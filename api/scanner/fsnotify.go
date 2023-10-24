package scanner

import (
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/scanner_cache"
	"github.com/photoview/photoview/api/utils"
	"gorm.io/gorm"
	"io/fs"
	"k8s.io/utils/inotify"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

const watchFlag = inotify.InCloseWrite | inotify.InMovedTo | inotify.InCreate | inotify.InDelete | inotify.InMovedFrom

func InitFsNotify(db *gorm.DB) error {
	watcher, err := inotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	user := &models.User{}
	db.First(user)
	user.FillAlbums(db)
	for _, album := range user.Albums {
		if album.ParentAlbumID != nil {
			continue
		}
		filepath.WalkDir(album.Path, func(path string, d fs.DirEntry, err error) error {
			if d != nil && d.IsDir() {
				log.Println("adding", path)
				err := watcher.AddWatch(path, watchFlag)
				if err != nil {
					log.Println(err)
				}
			}
			return nil
		})
	}

	go func() {
		for i := 0; i < 2; i++ {
			go worker(watcher, db, user)
		}
	}()

	return nil
}

func worker(watcher *inotify.Watcher, db *gorm.DB, user *models.User) {
	for {
		select {
		case e := <-watcher.Event:
			if strings.HasSuffix(e.Name, "tmp") {
				continue
			}
			if is(e, inotify.InCloseWrite) || is(e, inotify.InMovedTo) {
				createFile(watcher, db, user, e.Name)
			} else if hasAnd(e, inotify.InCreate, inotify.InIsdir) || hasAnd(e, inotify.InMovedTo, inotify.InIsdir) {
				createDir(watcher, db, user, e.Name)
			} else if is(e, inotify.InDelete) {
				deleteFile(watcher, db, user, e.Name)
			} else if hasAnd(e, inotify.InDelete, inotify.InIsdir) || hasAnd(e, inotify.InMovedFrom, inotify.InIsdir) {
				deleteDir(watcher, db, user, e.Name)
			}
		case err := <-watcher.Error:
			log.Println(err)
		}
	}
}

func deferFunc() {
	if e := recover(); e != nil {
		log.Println(string(debug.Stack()))
		return
	}
}

func createFile(watcher *inotify.Watcher, db *gorm.DB, user *models.User, filePath string) {
	defer deferFunc()
	dir := path.Dir(filePath)
	log.Println("Create file", filePath)
	// 删除多余文件
	os.Remove(utils.SwitchBolanghao(filePath))

	if _, err := os.Stat(filePath); err != nil {
		log.Println("Create error not exist,", filePath)
		return
	}
	// 添加新图片
	var album models.Album
	if err := db.Where("path_hash = ?", models.MD5Hash(dir)).Find(&album).Error; err != nil {
		log.Println("Create error album not found,", err)
		return
	}
	db.Model(&album).Update("last_modify_time", time.Now().UTC().Unix())
	media, ok, _ := ScanMedia(db, filePath, album.ID, scanner_cache.MakeAlbumCache())
	if ok {
		err := ProcessSingleMedia(db, media, &album)
		if err != nil {
			log.Printf("ProcessSingleMedia path[%v] err=[%v]\n", filePath, err)
		}
	} else {
		log.Printf("ScanMedia path=[%v] ok=[%v]\n", filePath, ok)
	}
	result := db.First(media)
	if result.Error != nil {
		log.Printf("media insert fail path=[%v] err=[%v]\n", filePath, result.Error)
		createFile(watcher, db, user, filePath)
	}
}

func createDir(watcher *inotify.Watcher, db *gorm.DB, user *models.User, filePath string) {
	defer deferFunc()
	dir := path.Dir(filePath)
	base := path.Base(filePath)
	log.Println("Create dir", filePath)
	watcher.AddWatch(filePath, watchFlag)
	var albumParent models.Album
	if err := db.Where("path_hash = ?", models.MD5Hash(dir)).Find(&albumParent).Error; err != nil {
		return
	}
	st, _ := os.Stat(filePath)
	modTime := int(st.ModTime().UTC().Unix())
	album := &models.Album{
		Title:          base,
		ParentAlbumID:  &albumParent.ID,
		Path:           filePath,
		LastModifyTime: &modTime,
	}
	db.Create(album)
	parentOwners := []models.User{*user}
	db.Model(&album).Association("Owners").Append(parentOwners)
}

func deleteFile(watcher *inotify.Watcher, db *gorm.DB, user *models.User, filePath string) {
	defer deferFunc()
	log.Println("delete file", filePath)
	var media models.Media
	result := db.Where("path_hash = ?", models.MD5Hash(utils.RemoveBolanghao(filePath))).First(&media)
	if result.Error != nil {
		return
	}
	cachePath := path.Join(utils.MediaCachePath(), strconv.Itoa(int(media.AlbumID)), strconv.Itoa(int(media.ID)))
	os.RemoveAll(cachePath)
	db.Delete(media)
}

func deleteDir(watcher *inotify.Watcher, db *gorm.DB, user *models.User, filePath string) {
	defer deferFunc()
	log.Println("delete dir", filePath)
	watcher.RemoveWatch(filePath)
	var album models.Album
	if err := db.Where("path_hash = ?", models.MD5Hash(filePath)).Find(&album).Error; err != nil {
		return
	}
	cachePath := path.Join(utils.MediaCachePath(), strconv.Itoa(int(album.ID)))
	os.RemoveAll(cachePath)
	// Delete old albums from database
	db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("album_id IN (?)", album.ID).Delete(&models.UserAlbums{}).Error; err != nil {
			return err
		}

		if err := tx.Delete(album).Error; err != nil {
			return err
		}

		return nil
	})
}

func is(event *inotify.Event, expected uint32) bool {
	if event.Mask == expected {
		return true
	}
	return false
}

func hasOr(event *inotify.Event, expected ...uint32) bool {
	for _, e := range expected {
		if event.Mask&e == e {
			return true
		}
	}
	return false
}

func hasAnd(event *inotify.Event, expected ...uint32) bool {
	for _, e := range expected {
		if event.Mask&e != e {
			return false
		}
	}
	return true
}
