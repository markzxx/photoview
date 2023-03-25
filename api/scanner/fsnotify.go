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
	"strconv"
	"strings"
	"time"
)

func InitFsNotify(db *gorm.DB) error {
	watcher, err := inotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	user := models.User{}
	db.First(&user)
	user.FillAlbums(db)
	for _, album := range user.Albums {
		if album.ParentAlbumID != nil {
			continue
		}
		filepath.WalkDir(album.Path, func(path string, d fs.DirEntry, err error) error {
			if d.IsDir() {
				log.Println("adding", path)
				err := watcher.Watch(path)
				if err != nil {
					log.Println(err)
				}
			}
			return nil
		})
	}

	go func() {
		defer watcher.Close()
		for {
			select {
			case e := <-watcher.Event:
				dir := path.Dir(e.Name)
				base := path.Base(e.Name)
				if strings.HasSuffix(e.Name, "tmp") {
					continue
				}
				if is(e, inotify.InCloseWrite) || is(e, inotify.InMovedTo) {
					log.Println("Create file", e.Name)
					// 删除多余文件
					if strings.HasSuffix(dir, "精修图片") {
						files, _ := os.ReadDir(dir)
						count := 0
						var deletes []string
						for _, file := range files {
							if strings.HasPrefix(file.Name(), base[0:6]) {
								count += 1
								if strings.Contains(base, "云修") {
									deletes = append(deletes, e.Name)
								}
							}
						}
						if count >= 2 {
							for _, d := range deletes {
								os.Remove(d)
							}
						}
					} else {
						os.Remove(utils.SwitchBolanghao(e.Name))
					}

					if _, err := os.Stat(e.Name); err != nil {
						continue
					}
					// 添加新图片
					var album models.Album
					if err := db.Where("path_hash = ?", models.MD5Hash(dir)).Find(&album).Error; err != nil {
						continue
					}
					db.Model(&album).Update("last_modify_time", time.Now().UTC().Unix())
					media, _, _ := ScanMedia(db, e.Name, album.ID, scanner_cache.MakeAlbumCache())
					ProcessSingleMedia(db, media)
				} else if hasAnd(e, inotify.InCreate, inotify.InIsdir) || hasAnd(e, inotify.InMovedTo, inotify.InIsdir) {
					log.Println("Create dir", e.Name)
					watcher.Watch(e.Name)
					var albumParent models.Album
					if err := db.Where("path_hash = ?", models.MD5Hash(dir)).Find(&albumParent).Error; err != nil {
						continue
					}
					st, _ := os.Stat(e.Name)
					modTime := int(st.ModTime().UTC().Unix())
					album := &models.Album{
						Title:          base,
						ParentAlbumID:  &albumParent.ID,
						Path:           e.Name,
						LastModifyTime: &modTime,
					}
					db.Create(album)
					parentOwners := []models.User{user}
					db.Model(&album).Association("Owners").Append(parentOwners)
				} else if is(e, inotify.InDelete) || is(e, inotify.InMovedFrom) {
					log.Println("delete file", e.Name)
					var media models.Media
					result := db.Where("path_hash = ?", models.MD5Hash(utils.RemoveBolanghao(e.Name))).First(&media)
					if result.Error != nil {
						continue
					}
					cachePath := path.Join(utils.MediaCachePath(), strconv.Itoa(int(media.AlbumID)), strconv.Itoa(int(media.ID)))
					os.RemoveAll(cachePath)
					db.Delete(media)
				} else if hasAnd(e, inotify.InDelete, inotify.InIsdir) || hasAnd(e, inotify.InMovedFrom, inotify.InIsdir) {
					log.Println("delete dir", e.Name)
					watcher.RemoveWatch(e.Name)
					var album models.Album
					if err := db.Where("path_hash = ?", models.MD5Hash(dir)).Find(&album).Error; err != nil {
						continue
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
			case err := <-watcher.Error:
				log.Println(err)
			}
		}
	}()

	return nil
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
