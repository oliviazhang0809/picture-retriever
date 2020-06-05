package models

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
	"math/rand"
	"time"
)

func NewImageFactory(db *sqlx.DB) *ImageFactory {
	imageFactory := &ImageFactory{}
	imageFactory.db = db
	imageFactory.table = "image_factory"
	imageFactory.hasID = true

	return imageFactory
}

type ImageRow struct {
	ID    int64  `db:"id"`
	Email string `db:"category"`
	URL   string `db:"url"`
}

type ImageFactory struct {
	Base
}

func (u *ImageFactory) imageRowFromSqlResult(tx *sqlx.Tx, sqlResult sql.Result) (*ImageRow, error) {
	imageId, err := sqlResult.LastInsertId()
	if err != nil {
		return nil, err
	}

	return u.GetById(tx, imageId)
}

// AllImages returns all image rows.
func (u *ImageFactory) AllImages(tx *sqlx.Tx) ([]*ImageRow, error) {
	images := []*ImageRow{}
	query := fmt.Sprintf("SELECT * FROM %v", u.table)
	err := u.db.Select(&images, query)

	return images, err
}

// GetById returns record by id.
func (u *ImageFactory) GetById(tx *sqlx.Tx, id int64) (*ImageRow, error) {
	image := &ImageRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE id=?", u.table)
	err := u.db.Get(image, query, id)

	return image, err
}

// GetByCategory returns record by email.
func (u *ImageFactory) GetByCategoryLike(tx *sqlx.Tx, category string) (*ImageRow, error) {
	images := []*ImageRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE category LIKE ", u.table) + "\"" + category + "%\""

	err := u.db.Select(&images, query)

	if images == nil || len(images) == 0 {
		return nil, err
	}

	rand.Seed(time.Now().Unix())
	image := images[rand.Intn(len(images))]

	return image, err
}

// GetByCategory returns record by email.
func (u *ImageFactory) GetByCategory(tx *sqlx.Tx, category string) (*ImageRow, error) {
	image := &ImageRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE category=?", u.table)
	err := u.db.Get(image, query, category)

	return image, err
}

// GetByCategory returns record by email but checks password first.
func (u *ImageFactory) GetImageByCategoryAndURL(tx *sqlx.Tx, category, url string) (*ImageRow, error) {
	image, err := u.GetByCategory(tx, category)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(image.URL), []byte(url))
	if err != nil {
		return nil, err
	}

	return image, err
}

// Save create a new record of image.
func (u *ImageFactory) Save(tx *sqlx.Tx, category, url string) (*ImageRow, error) {
	if category == "" {
		return nil, errors.New("Category cannot be blank.")
	}
	if url == "" {
		return nil, errors.New("URL cannot be blank.")
	}

	hashedURL, err := bcrypt.GenerateFromPassword([]byte(url), 5)
	if err != nil {
		return nil, err
	}

	data := make(map[string]interface{})
	data["category"] = category
	data["url"] = hashedURL

	sqlResult, err := u.InsertIntoTable(tx, data)
	if err != nil {
		return nil, err
	}

	return u.imageRowFromSqlResult(tx, sqlResult)
}

// UpdateCategoryAndUrlById updates category and url.
func (u *ImageFactory) UpdateCategoryAndUrlById(tx *sqlx.Tx, imageID int64, category, url string) (*ImageRow, error) {
	data := make(map[string]interface{})

	if category != "" {
		data["category"] = category
	}

	if url != "" {
		hashedURL, err := bcrypt.GenerateFromPassword([]byte(url), 5)
		if err != nil {
			return nil, err
		}

		data["url"] = hashedURL
	}

	if len(data) > 0 {
		_, err := u.UpdateByID(tx, data, imageID)
		if err != nil {
			return nil, err
		}
	}

	return u.GetById(tx, imageID)
}
