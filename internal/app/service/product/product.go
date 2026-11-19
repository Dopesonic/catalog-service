package sproduct

import (
	"context"
	"time"

	"github.com/gofrs/uuid"

	"github.com/Dopesonic/catalog-service/internal/app/entity"
	"github.com/Dopesonic/catalog-service/internal/app/monitor"
	"github.com/Dopesonic/catalog-service/internal/app/repository"
	"github.com/Dopesonic/catalog-service/internal/app/service"
)

type srv struct {
	repoProduct  repository.Product
	repoCategory repository.Category
	metrics      monitor.ProductMetrics
}

type noopProductMetrics struct{}

type noopProductCreateMetric struct{}

func (noopProductMetrics) Create() monitor.ProductCreateMetric {
	return noopProductCreateMetric{}
}

func (noopProductCreateMetric) Success(int64) {}

func (noopProductCreateMetric) Failed(error) {}

func NewService(
	repoProduct repository.Product,
	repoCategory repository.Category,
	metrics monitor.ProductMetrics,
) service.Product {
	return &srv{
		repoProduct:  repoProduct,
		repoCategory: repoCategory,
		metrics:      metrics,
	}
}

func (s *srv) Create(ctx context.Context, req entity.RequestProductCreate) (entity.Product, error) {
	createMetric := s.metrics.Create()

	var product entity.Product
	err := s.repoProduct.InsideTx(ctx, func(txCtx context.Context) error {
		existing, err := s.repoProduct.List(txCtx, &req.Name, nil, nil, nil)
		if err != nil {
			return err
		}
		if len(existing) > 0 {
			return entity.ErrAlreadyExists
		}

		cats, err := s.repoCategory.GetByGUIDs(txCtx, []uuid.UUID{req.CategoryGUID})
		if err != nil {
			return err
		}
		if len(cats) == 0 {
			return entity.ErrNotFound
		}

		now := time.Now()

		product = entity.Product{
			GUID:         uuid.Must(uuid.NewV4()),
			Name:         req.Name,
			Description:  req.Description,
			Price:        req.Price,
			CategoryGUID: req.CategoryGUID,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		return s.repoProduct.Create(txCtx, product)
	})
	if err != nil {
		createMetric.Failed(err)
		return entity.Product{}, err
	}

	createMetric.Success(product.Price)
	return product, nil
}

func (s *srv) GetByGUIDs(ctx context.Context, guids []uuid.UUID) ([]entity.Product, error) {
	return s.repoProduct.GetByGUIDs(ctx, guids)
}

func (s *srv) Update(ctx context.Context, guid uuid.UUID, req entity.RequestProductUpdate) (entity.Product, error) {
	var currentProd entity.Product
	err := s.repoProduct.InsideTx(ctx, func(txCtx context.Context) error {
		prods, err := s.repoProduct.GetByGUIDs(txCtx, []uuid.UUID{guid})
		if err != nil {
			return err
		}
		if len(prods) == 0 {
			return entity.ErrNotFound
		}

		currentProd = prods[0]

		if req.Name != "" {
			existingProds, err := s.repoProduct.List(txCtx, &req.Name, nil, nil, nil)
			if err != nil {
				return err
			}
			for _, p := range existingProds {
				if p.GUID != guid {
					return entity.ErrAlreadyExists
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
			cats, err := s.repoCategory.GetByGUIDs(txCtx, []uuid.UUID{req.CategoryGUID})
			if err != nil {
				return err
			}
			if len(cats) == 0 {
				return entity.ErrNotFound
			}
			currentProd.CategoryGUID = req.CategoryGUID
		}

		currentProd.UpdatedAt = time.Now()

		return s.repoProduct.Update(txCtx, currentProd)
	})
	if err != nil {
		return entity.Product{}, err
	}

	return currentProd, nil
}

func (s *srv) Delete(ctx context.Context, guid uuid.UUID) error {
	err := s.repoProduct.InsideTx(ctx, func(txCtx context.Context) error {
		prods, err := s.repoProduct.GetByGUIDs(txCtx, []uuid.UUID{guid})
		if err != nil {
			return err
		}
		if len(prods) == 0 {
			return entity.ErrNotFound
		}

		return s.repoProduct.Delete(txCtx, guid)
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *srv) List(ctx context.Context, req entity.RequestProductList) ([]entity.Product, error) {
	return s.repoProduct.List(ctx, nil, req.CategoryGUID, req.MinPrice, req.MaxPrice)
}
