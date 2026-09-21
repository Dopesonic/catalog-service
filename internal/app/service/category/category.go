package scategory

import (
	"context"
	"time"

	"github.com/gofrs/uuid"

	"github.com/Dopesonic/catalog-service/internal/app/entity"
	"github.com/Dopesonic/catalog-service/internal/app/repository"
	"github.com/Dopesonic/catalog-service/internal/app/service"
)

type srv struct {
	repoCategory repository.Category
	repoProduct  repository.Product
}

func NewService(repoCategory repository.Category, repoProduct repository.Product) service.Category {
	return &srv{
		repoCategory: repoCategory,
		repoProduct:  repoProduct,
	}
}
