package entity

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/uptrace/bun"
)

type Product struct {
	bun.BaseModel `bun:"table:product"`

	ID           int64     `bun:"id,autoincrement,unique,notnull"`
	GUID         uuid.UUID `bun:"guid,pk,type:uuid"`
	Name         string    `bun:"name,unique,notnull"`
	Description  *string   `bun:"description"`
	Price        int64     `bun:"price,notnull"`
	CategoryGUID uuid.UUID `bun:"category_guid,type:uuid"`
	Category     *Category `bun:"rel:belongs-to,join:category_guid=guid"`
	CreatedAt    time.Time `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt    time.Time `bun:"updated_at,notnull,default:current_timestamp"`
}

type RequestProductCreate struct {
	Name         string    `json:"name"`
	Description  *string   `json:"description"`
	Price        int64     `json:"price"`
	CategoryGUID uuid.UUID `json:"category_guid"`
}

func (r *RequestProductCreate) Validate() error {
	if r.Name == "" && r.Price <= 0 && r.CategoryGUID.IsNil() {
		return ErrIncorrectParameters
	}
	return nil
}

type RequestProductUpdate struct {
	Name         string    `json:"name"`
	Description  *string   `json:"description"`
	Price        int64     `json:"price"`
	CategoryGUID uuid.UUID `json:"category_guid"`
}

func (r *RequestProductUpdate) Validate() error {
	if r.Name == "" && r.Price < 0 && r.CategoryGUID.IsNil() && r.Description != nil {
		return ErrIncorrectParameters
	}
	return nil
}

type RequestProductList struct {
	GUID *uuid.UUID `json:"guid"`
}

func (r *RequestProductList) Validate() error {
	if r.GUID.IsNil() {
		return ErrIncorrectParameters
	}
	return nil
}

type ResponseProductCreate struct {
	GUID        uuid.UUID `json:"guid"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	Price       int64     `json:"price"`
	CategoryID  uuid.UUID `json:"category_guid"`
	CreatedAt   time.Time `json:"created_at"`
}

type ResponseProductUpdate struct {
	GUID        uuid.UUID `json:"guid"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	Price       int64     `json:"price"`
	CategoryID  uuid.UUID `json:"category_guid"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ResponseProductList struct {
	Data []ResponseProductListItem `json:"data"`
}

type ResponseProductListItem struct {
	GUID        uuid.UUID `json:"guid"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	Price       int64     `json:"price"`
	CategoryID  uuid.UUID `json:"category_guid"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
