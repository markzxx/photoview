package scanner

import (
	"bufio"
	"container/list"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/scanner_cache"
	"github.com/photoview/photoview/api/scanner/scanner_tasks/cleanup_tasks"
	"github.com/photoview/photoview/api/scanner/scanner_utils"
	"github.com/photoview/photoview/api/utils"
	"github.com/pkg/errors"
	ignore "github.com/sabhiram/go-gitignore"
	"gorm.io/gorm"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sort"
)

func getPhotoviewIgnore(ignorePath string) ([]string, error) {
	var photoviewIgnore []string

	// Open .photoviewignore file, if exists
	photoviewIgnoreFile, err := os.Open(path.Join(ignorePath, ".photoviewignore"))
	if err != nil {
		if os.IsNotExist(err) {
			return photoviewIgnore, nil
		}
		return photoviewIgnore, err
	}

	// Close file on exit
	defer photoviewIgnoreFile.Close()

	// Read and save .photoviewignore data
	scanner := bufio.NewScanner(photoviewIgnoreFile)
	for scanner.Scan() {
		photoviewIgnore = append(photoviewIgnore, scanner.Text())
		log.Printf("Ignore found: %s", scanner.Text())
	}

	return photoviewIgnore, scanner.Err()
}

func FindAlbumsForUser(db *gorm.DB, user *models.User, album_cache *scanner_cache.AlbumScannerCache, scan_all bool) ([]*models.Album, []error) {

	if err := user.FillAlbums(db); err != nil {
		return nil, []error{err}
	}

	userAlbumIDs := make([]int, len(user.Albums))
	for i, album := range user.Albums {
		userAlbumIDs[i] = album.ID
	}

	var userRootAlbums []*models.Album
	if err := db.Where("id IN (?)", userAlbumIDs).Where("parent_album_id IS NULL OR parent_album_id NOT IN (?)", userAlbumIDs).Find(&userRootAlbums).Error; err != nil {
		return nil, []error{err}
	}

	scanErrors := make([]error, 0)

	type scanInfo struct {
		path       string
		parent     *models.Album
		ignore     []string
		modifyTime int
	}

	scanQueue := list.New()

	for _, album := range userRootAlbums {
		log.Printf("my albums %s", album.Path)
		// Check if user album directory exists on the file system
		if file, err := os.Stat(album.Path); err != nil {
			if os.IsNotExist(err) {
				scanErrors = append(scanErrors, errors.Errorf("Album directory for user '%s' does not exist '%s'\n", user.Username, album.Path))
			} else {
				scanErrors = append(scanErrors, errors.Errorf("Could not read album directory for user '%s': %s\n", user.Username, album.Path))
			}
		} else {
			scanQueue.PushBack(scanInfo{
				path:       album.Path,
				parent:     nil,
				ignore:     nil,
				modifyTime: int(file.ModTime().UTC().Unix()),
			})
		}
	}

	userAlbums := make([]*models.Album, 0)
	needScanAlbums := make([]*models.Album, 0)

	for scanQueue.Front() != nil {
		skip := false
		albumInfo := scanQueue.Front().Value.(scanInfo)
		scanQueue.Remove(scanQueue.Front())

		albumPath := albumInfo.path
		albumParent := albumInfo.parent
		albumIgnore := albumInfo.ignore

		// Read path
		dirContent, err := ioutil.ReadDir(albumPath)
		if err != nil {
			scanErrors = append(scanErrors, errors.Wrapf(err, "read directory (%s)", albumPath))
			continue
		}

		sort.SliceStable(dirContent, func(i, j int) bool {
			return dirContent[i].ModTime().Unix() > dirContent[j].ModTime().Unix()
		})

		// Skip this dir if in ignore list
		ignorePaths := ignore.CompileIgnoreLines(albumIgnore...)
		if ignorePaths.MatchesPath(albumPath + "/") {
			log.Printf("Skip, directroy %s is in ignore file", albumPath)
			continue
		}

		// Update ignore dir list
		photoviewIgnore, err := getPhotoviewIgnore(albumPath)
		if err != nil {
			log.Printf("Failed to get ignore file, err = %s", err)
		} else {
			albumIgnore = append(albumIgnore, photoviewIgnore...)
		}

		// Will become new album or album from db
		var album *models.Album

		transErr := db.Transaction(func(tx *gorm.DB) error {
			log.Printf("Scanning directory: %s", albumPath)

			// check if album already exists
			var albumResult []models.Album
			result := tx.Where("path_hash = ?", models.MD5Hash(albumPath)).Find(&albumResult)
			if result.Error != nil {
				return result.Error
			}

			// album does not exist, create new
			if len(albumResult) == 0 {
				albumTitle := path.Base(albumPath)

				var albumParentID *int
				parentOwners := make([]models.User, 0)
				if albumParent != nil {
					albumParentID = &albumParent.ID

					if err := tx.Model(&albumParent).Association("Owners").Find(&parentOwners); err != nil {
						return err
					}
				}

				album = &models.Album{
					Title:          albumTitle,
					ParentAlbumID:  albumParentID,
					Path:           albumPath,
					LastModifyTime: &albumInfo.modifyTime,
				}

				// Store album ignore
				album_cache.InsertAlbumIgnore(albumPath, albumIgnore)

				if err := tx.Create(&album).Error; err != nil {
					return errors.Wrap(err, "insert album into database")
				}

				if err := tx.Model(&album).Association("Owners").Append(parentOwners); err != nil {
					return errors.Wrap(err, "add owners to album")
				}
			} else {
				album = &albumResult[0]
				if !scan_all && album.LastModifyTime != nil && *album.LastModifyTime == albumInfo.modifyTime {
					log.Printf("Skip directory: %s", albumPath)
					skip = true
				} else {
					tx.Model(&album).Update("last_modify_time", albumInfo.modifyTime)
				}

				// Add user as an owner of the album if not already
				var userAlbumOwner []models.User
				if err := tx.Model(&album).Association("Owners").Find(&userAlbumOwner, "user_albums.user_id = ?", user.ID); err != nil {
					return err
				}
				if len(userAlbumOwner) == 0 {
					newUser := models.User{}
					newUser.ID = user.ID
					if err := tx.Model(&album).Association("Owners").Append(&newUser); err != nil {
						return err
					}
				}

				// Update album ignore
				album_cache.InsertAlbumIgnore(albumPath, albumIgnore)
			}

			userAlbums = append(userAlbums, album)
			if !skip {
				needScanAlbums = append(needScanAlbums, album)
			}
			return nil
		})

		if transErr != nil {
			scanErrors = append(scanErrors, errors.Wrap(transErr, "begin database transaction"))
			continue
		}

		// Scan for sub-albums
		for _, item := range dirContent {
			subalbumPath := path.Join(albumPath, item.Name())

			// Skip if directory is hidden
			if path.Base(subalbumPath)[0:1] == "." {
				continue
			}

			isDirSymlink, err := utils.IsDirSymlink(subalbumPath)
			if err != nil {
				scanErrors = append(scanErrors, errors.Wrapf(err, "could not check for symlink target of %s", subalbumPath))
				continue
			}

			if (item.IsDir() || isDirSymlink) && directoryContainsPhotos(subalbumPath, album_cache, albumIgnore) {
				scanQueue.PushBack(scanInfo{
					path:       subalbumPath,
					parent:     album,
					ignore:     albumIgnore,
					modifyTime: int(item.ModTime().UTC().Unix()),
				})
			}
		}
	}

	deleteErrors := cleanup_tasks.DeleteOldUserAlbums(db, userAlbums, user)
	scanErrors = append(scanErrors, deleteErrors...)
	log.Printf("Real scan dir %v", len(needScanAlbums))
	return needScanAlbums, scanErrors
}

func directoryContainsPhotos(rootPath string, cache *scanner_cache.AlbumScannerCache, albumIgnore []string) bool {

	if contains_image := cache.AlbumContainsPhotos(rootPath); contains_image != nil {
		return *contains_image
	}

	scanQueue := list.New()
	scanQueue.PushBack(rootPath)

	scanned_directories := make([]string, 0)

	for scanQueue.Front() != nil {

		dirPath := scanQueue.Front().Value.(string)
		scanQueue.Remove(scanQueue.Front())

		scanned_directories = append(scanned_directories, dirPath)

		// Update ignore dir list
		photoviewIgnore, err := getPhotoviewIgnore(dirPath)
		if err != nil {
			log.Printf("Failed to get ignore file, err = %s", err)
		} else {
			albumIgnore = append(albumIgnore, photoviewIgnore...)
		}
		ignoreEntries := ignore.CompileIgnoreLines(albumIgnore...)

		dirContent, err := ioutil.ReadDir(dirPath)
		if err != nil {
			scanner_utils.ScannerError("Could not read directory (%s): %s\n", dirPath, err.Error())
			return false
		}

		for _, fileInfo := range dirContent {
			filePath := path.Join(dirPath, fileInfo.Name())

			isDirSymlink, err := utils.IsDirSymlink(filePath)
			if err != nil {
				log.Printf("Cannot detect whether %s is symlink to a directory. Pretending it is not", filePath)
				isDirSymlink = false
			}

			if fileInfo.IsDir() || isDirSymlink {
				scanQueue.PushBack(filePath)
			} else {
				if cache.IsPathMedia(filePath) {
					if ignoreEntries.MatchesPath(fileInfo.Name()) {
						log.Printf("Match found %s, continue search for media", fileInfo.Name())
						continue
					}
					log.Printf("Insert Album %s %s, contains photo is true", dirPath, rootPath)
					cache.InsertAlbumPaths(dirPath, rootPath, true)
					return true
				}
			}
		}

	}

	for _, scanned_path := range scanned_directories {
		log.Printf("Insert Album %s, contains photo is false", scanned_path)
		cache.InsertAlbumPath(scanned_path, false)
	}
	return false
}
