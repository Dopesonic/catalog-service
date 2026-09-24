package sproduct

import (
	"context"
	"time"

	"github.com/gofrs/uuid"

	"github.com/Dopesonic/catalog-service/internal/app/entity"
	"github.com/Dopesonic/catalog-service/internal/app/repository"
	"github.com/Dopesonic/catalog-service/internal/app/service"
)

type srv struct {
	repoProduct  repository.Product
	repoCategory repository.Category
}

func NewService(repoProduct repository.Product, repoCategory repository.Category) service.Product {
	return &srv{
		repoProduct:  repoProduct,
		repoCategory: repoCategory,
	}
}

func (s *srv) Create(ctx context.Context, req entity.RequestProductCreate) (entity.Product, error) {
	existing, err := s.repoProduct.List(ctx, &req.Name, nil)
	if err != nil {
		return entity.Product{}, err
	}
	if len(existing) > 0 {
		return entity.Product{}, entity.ErrAlreadyExists
	}

	cats, err := s.repoCategory.GetByGUIDs(ctx, []uuid.UUID{req.CategoryGUID})
	if err != nil {
		return entity.Product{}, err
	}
	if len(cats) == 0 {
		return entity.Product{}, entity.ErrNotFound
	}

	now := time.Now()

	product := entity.Product{
		GUID:         uuid.Must(uuid.NewV4()),
		Name:         req.Name,
		Description:  req.Description,
		Price:        req.Price,
		CategoryGUID: req.CategoryGUID,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.repoProduct.Create(ctx, product); err != nil {
		return entity.Product{}, err
	}

	return product, nil
}

func (s *srv) GetByGUIDs(ctx context.Context, guid []uuid.UUID) ([]entity.Product, error) {
	return s.repoProduct.GetByGUIDs(ctx, guid)
}

func (s *srv) Update(ctx context.Context, guid uuid.UUID, req entity.RequestProductUpdate) (entity.Product, error) {
	prods, err := s.repoProduct.GetByGUIDs(ctx, []uuid.UUID{guid})
	if err != nil {
		return entity.Product{}, err
	}
	if len(prods) == 0 {
		return entity.Product{}, entity.ErrNotFound
	}

	currentProd := prods[0]

	if req.Name != "" {
		existingProds, err := s.repoProduct.List(ctx, &req.Name, nil)
		if err != nil {
			return entity.Product{}, err
		}
		for _, p := range existingProds {
			if p.GUID != guid {
				return entity.Product{}, entity.ErrAlreadyExists
			}
		}
		currentProd.Name = req.Name
	}

	if req.Description != nil {
		currentProd.Description = req.Description
	}
	if req.Price > 0 {
		currentProd.Price = req.Price
	}

	if req.CategoryGUID != uuid.Nil {
		cats, err := s.repoCategory.GetByGUIDs(ctx, []uuid.UUID{req.CategoryGUID})
		if err != nil {
			return entity.Product{}, err
		}
		if len(cats) == 0 {
			return entity.Product{}, entity.ErrNotFound
		}
		currentProd.CategoryGUID = req.CategoryGUID
	}

	currentProd.UpdatedAt = time.Now()

	if err := s.repoProduct.Update(ctx, currentProd); err != nil {
		return entity.Product{}, err
	}

	return currentProd, nil
}

func (s *srv) Delete(ctx context.Context, guid uuid.UUID) error {
	prods, err := s.repoProduct.GetByGUIDs(ctx, []uuid.UUID{guid})
	if err != nil {
		return err
	}
	if len(prods) == 0 {
		return entity.ErrNotFound
	}

	return s.repoProduct.Delete(ctx, guid)
}

func (s *srv) List(ctx context.Context, req entity.RequestProductList) ([]entity.Product, error) {
	return s.repoProduct.List(ctx, nil, req.CategoryGUID)
}
