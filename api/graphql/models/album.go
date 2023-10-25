package models

import (
	"crypto/md5"
	"encoding/hex"

	"gorm.io/gorm"
)

type Album struct {
	Model
	Title         string `gorm:"not null"`
	ParentAlbumID *int   `gorm:"index"`
	ParentAlbum   *Album `gorm:"constraint:OnDelete:SET NULL;"`
	// OwnerID       int `gorm:"not null"`
	// Owner         User
	Owners             []User `gorm:"many2many:user_albums;constraint:OnDelete:CASCADE;"`
	Path               string `gorm:"not null"`
	PathHash           string `gorm:"unique"`
	CoverID            *int
	LastModifyTime     *int
	LastLastModifyTime *int
}

func (a *Album) FilePath() string {
	return a.Path
}

func (a *Album) BeforeSave(tx *gorm.DB) (err error) {
	hash := md5.Sum([]byte(a.Path))
	a.PathHash = hex.EncodeToString(hash[:])
	return nil
}

// GetChildren performs a recursive query to get all the children of the album.
// An optional filter can be provided that can be used to modify the query on the children.
func (a *Album) GetChildren(db *gorm.DB, filter func(*gorm.DB) *gorm.DB) (children []*Album, err error) {
	return GetChildrenFromAlbums(db, filter, []int{a.ID})
}

func GetChildrenFromAlbums(db *gorm.DB, filter func(*gorm.DB) *gorm.DB, albumIDs []int) (children []*Album, err error) {
	query := db.Model(&Album{}).Table("sub_albums")

	if filter != nil {
		query = filter(query)
	}

	err = db.Raw(`
	WITH recursive sub_albums AS (
		SELECT * FROM albums AS root WHERE id IN (?)
		UNION ALL
		SELECT child.* FROM albums AS child JOIN sub_albums ON child.parent_album_id = sub_albums.id
	)

	?
	`, albumIDs, query).Find(&children).Error

	return children, err
}

func (a *Album) GetParents(db *gorm.DB, filter func(*gorm.DB) *gorm.DB) (parents []*Album, err error) {
	return GetParentsFromAlbums(db, filter, a.ID)
}

func GetParentsFromAlbums(db *gorm.DB, filter func(*gorm.DB) *gorm.DB, albumID int) (parents []*Album, err error) {
	query := db.Model(&Album{}).Table("super_albums")

	if filter != nil {
		query = filter(query)
	}

	err = db.Raw(`
	WITH recursive super_albums AS (
		SELECT * FROM albums AS leaf WHERE id = ?
		UNION ALL
		SELECT parent.* from albums AS parent JOIN super_albums ON parent.id = super_albums.parent_album_id
	)

	?
	`, albumID, query).Find(&parents).Error

	return parents, err
}

func (a *Album) Thumbnail(db *gorm.DB) (*Media, error) {
	var media Media

	if a.CoverID == nil {
		if err := db.Raw(`
			SELECT media.* 
			FROM media 
			JOIN (
					SELECT a1.id FROM albums AS a1 WHERE a1.id = ?
					UNION
					SELECT a2.id FROM albums AS a2 
					INNER JOIN albums AS a1 ON a2.parent_album_id = a1.id
					WHERE a1.id = ?
				) sub_albums ON media.album_id = sub_albums.id 
			JOIN media_urls ON media_urls.media_id = media.id 
			ORDER BY media.id DESC
			LIMIT 1
		`, a.ID, a.ID).Find(&media).Error; err != nil {
			return nil, err
		}
	} else {
		if err := db.Where("id = ?", a.CoverID).Find(&media).Error; err != nil {
			return nil, err
		}
	}

	return &media, nil
}
