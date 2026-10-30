package sproduct

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/Dopesonic/catalog-service/internal/app/entity"
	"github.com/Dopesonic/catalog-service/internal/app/repository/mocks"
	"github.com/Dopesonic/catalog-service/internal/pkg/testutil"
)

type createProductSuite struct {
	suite.Suite
	srv          *srv
	productRepo  *mocks.MockProduct
	categoryRepo *mocks.MockCategory
	ctx          context.Context
}

func (s *createProductSuite) SetupTest() {
	s.ctx = context.Background()
	s.productRepo = mocks.NewMockProduct(s.T())
	s.categoryRepo = mocks.NewMockCategory(s.T())
	s.srv = &srv{
		repoProduct:  s.productRepo,
		repoCategory: s.categoryRepo,
	}
}

func TestCreateProductSuite(t *testing.T) {
	suite.Run(t, new(createProductSuite))
}

func (s *createProductSuite) TestNewService() {
	actual := NewService(s.productRepo, s.categoryRepo)

	expected := &srv{
		repoProduct:  s.productRepo,
		repoCategory: s.categoryRepo,
	}

	s.Equal(expected, actual)
}

func (s *createProductSuite) TestCreate() {
	type args struct {
		req entity.RequestProductCreate
	}
	type want struct {
		err error
	}

	categoryGUID := uuid.Must(uuid.NewV4())

	repoErr := errors.New("repository failure")

	testCases := []struct {
		name    string
		args    args
		want    want
		prepare func(args args)
	}{
		{
			name: "success",
			args: args{
				req: entity.RequestProductCreate{
					Name:         "Test Product",
					Description:  testutil.PtrString("A test product"),
					Price:        1000,
					CategoryGUID: categoryGUID,
				},
			},
			want: want{err: nil},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					InsideTx(s.ctx, mock.AnythingOfType("func(context.Context) error")).
					RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					}).
					Once()

				s.productRepo.EXPECT().
					List(s.ctx, &args.req.Name, (*uuid.UUID)(nil), (*int64)(nil), (*int64)(nil)).
					Return([]entity.Product{}, nil).
					Once()

				s.categoryRepo.EXPECT().
					GetByGUIDs(s.ctx, []uuid.UUID{args.req.CategoryGUID}).
					Return([]entity.Category{{GUID: categoryGUID}}, nil).
					Once()

				s.productRepo.EXPECT().
					Create(s.ctx, mock.MatchedBy(func(p entity.Product) bool {
						return p.Name == args.req.Name &&
							p.Description == args.req.Description &&
							p.Price == args.req.Price &&
							p.CategoryGUID == args.req.CategoryGUID
					})).
					Return(nil).
					Once()
			},
		},
		{
			name: "already exists",
			args: args{
				req: entity.RequestProductCreate{
					Name:         "Existing Product",
					Price:        500,
					CategoryGUID: categoryGUID,
				},
			},
			want: want{err: entity.ErrAlreadyExists},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					InsideTx(s.ctx, mock.AnythingOfType("func(context.Context) error")).
					RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					}).
					Once()

				s.productRepo.EXPECT().
					List(s.ctx, &args.req.Name, (*uuid.UUID)(nil), (*int64)(nil), (*int64)(nil)).
					Return([]entity.Product{{Name: "Existing Product"}}, nil).
					Once()
			},
		},
		{
			name: "category not found",
			args: args{
				req: entity.RequestProductCreate{
					Name:         "New Product",
					Price:        1000,
					CategoryGUID: categoryGUID,
				},
			},
			want: want{err: entity.ErrNotFound},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					InsideTx(s.ctx, mock.AnythingOfType("func(context.Context) error")).
					RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					}).
					Once()

				s.productRepo.EXPECT().
					List(s.ctx, &args.req.Name, (*uuid.UUID)(nil), (*int64)(nil), (*int64)(nil)).
					Return([]entity.Product{}, nil).
					Once()

				s.categoryRepo.EXPECT().
					GetByGUIDs(s.ctx, []uuid.UUID{args.req.CategoryGUID}).
					Return([]entity.Category{}, nil).
					Once()
			},
		},
		{
			name: "list error",
			args: args{
				req: entity.RequestProductCreate{
					Name:         "Test Product",
					Price:        1000,
					CategoryGUID: categoryGUID,
				},
			},
			want: want{err: repoErr},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					InsideTx(
						s.ctx,
						mock.AnythingOfType("func(context.Context) error"),
					).
					RunAndReturn(func(
						ctx context.Context,
						fn func(context.Context) error,
					) error {
						return fn(ctx)
					}).
					Once()

				s.productRepo.EXPECT().
					List(
						s.ctx,
						&args.req.Name,
						(*uuid.UUID)(nil),
						(*int64)(nil),
						(*int64)(nil),
					).
					Return(nil, repoErr).
					Once()
			},
		},
		{
			name: "category lookup error",
			args: args{
				req: entity.RequestProductCreate{
					Name:         "Test Product",
					Price:        1000,
					CategoryGUID: categoryGUID,
				},
			},
			want: want{err: repoErr},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					InsideTx(
						s.ctx,
						mock.AnythingOfType("func(context.Context) error"),
					).
					RunAndReturn(func(
						ctx context.Context,
						fn func(context.Context) error,
					) error {
						return fn(ctx)
					}).
					Once()

				s.productRepo.EXPECT().
					List(
						s.ctx,
						&args.req.Name,
						(*uuid.UUID)(nil),
						(*int64)(nil),
						(*int64)(nil),
					).
					Return([]entity.Product{}, nil).
					Once()

				s.categoryRepo.EXPECT().
					GetByGUIDs(s.ctx, []uuid.UUID{args.req.CategoryGUID}).
					Return(nil, repoErr).
					Once()
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()
			tc.prepare(tc.args)

			result, err := s.srv.Create(s.ctx, tc.args.req)

			if tc.want.err != nil {
				s.ErrorIs(err, tc.want.err)
				s.Empty(result.GUID)
			} else {
				s.NoError(err)
				s.NotEmpty(result.GUID)
				s.Equal(tc.args.req.Name, result.Name)
				s.Equal(tc.args.req.Description, result.Description)
				s.Equal(tc.args.req.Price, result.Price)
				s.Equal(tc.args.req.CategoryGUID, result.CategoryGUID)
			}
		})
	}
}

type getByGUIDsProductSuite struct {
	suite.Suite
	srv          *srv
	productRepo  *mocks.MockProduct
	categoryRepo *mocks.MockCategory
	ctx          context.Context
}

func (s *getByGUIDsProductSuite) SetupTest() {
	s.ctx = context.Background()
	s.productRepo = mocks.NewMockProduct(s.T())
	s.categoryRepo = mocks.NewMockCategory(s.T())
	s.srv = &srv{
		repoProduct:  s.productRepo,
		repoCategory: s.categoryRepo,
	}
}

func TestGetByGUIDsProductSuite(t *testing.T) {
	suite.Run(t, new(getByGUIDsProductSuite))
}

func (s *getByGUIDsProductSuite) TestGetByGUIDs() {
	type args struct {
		guids []uuid.UUID
	}
	type want struct {
		products []entity.Product
	}

	product := entity.Product{
		GUID:  uuid.Must(uuid.NewV4()),
		Name:  "Test Product",
		Price: 1000,
	}
	missingGUID := uuid.Must(uuid.NewV4())

	testCases := []struct {
		name    string
		args    args
		want    want
		prepare func(args args)
	}{
		{
			name: "single product found",
			args: args{
				guids: []uuid.UUID{product.GUID},
			},
			want: want{
				products: []entity.Product{product},
			},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					GetByGUIDs(s.ctx, args.guids).
					Return([]entity.Product{product}, nil).
					Once()
			},
		},
		{
			name: "not found returns empty slice",
			args: args{
				guids: []uuid.UUID{missingGUID},
			},
			want: want{
				products: []entity.Product{},
			},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					GetByGUIDs(s.ctx, args.guids).
					Return([]entity.Product{}, nil).
					Once()
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()
			tc.prepare(tc.args)

			result, err := s.srv.GetByGUIDs(s.ctx, tc.args.guids)

			s.Require().NoError(err)
			s.Equal(tc.want.products, result)
		})
	}
}

type deleteProductSuite struct {
	suite.Suite
	srv          *srv
	productRepo  *mocks.MockProduct
	categoryRepo *mocks.MockCategory
	ctx          context.Context
}

func (s *deleteProductSuite) SetupTest() {
	s.ctx = context.Background()
	s.productRepo = mocks.NewMockProduct(s.T())
	s.categoryRepo = mocks.NewMockCategory(s.T())
	s.srv = &srv{
		repoProduct:  s.productRepo,
		repoCategory: s.categoryRepo,
	}
}

func TestDeleteProductSuite(t *testing.T) {
	suite.Run(t, new(deleteProductSuite))
}

func (s *deleteProductSuite) TestDelete() {
	type args struct {
		guid uuid.UUID
	}
	type want struct {
		err error
	}

	guid := uuid.Must(uuid.NewV4())
	deleteErr := errors.New("delete failed")
	repoErr := errors.New("repository failure")

	testCases := []struct {
		name    string
		args    args
		want    want
		prepare func(args args)
	}{
		{
			name: "success",
			args: args{guid: guid},
			want: want{err: nil},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					GetByGUIDs(s.ctx, []uuid.UUID{args.guid}).
					Return([]entity.Product{{GUID: args.guid}}, nil).
					Once()

				s.productRepo.EXPECT().
					Delete(s.ctx, args.guid).
					Return(nil).
					Once()
			},
		},
		{
			name: "not found",
			args: args{guid: guid},
			want: want{err: entity.ErrNotFound},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					GetByGUIDs(s.ctx, []uuid.UUID{args.guid}).
					Return([]entity.Product{}, nil).
					Once()
			},
		},
		{
			name: "delete error",
			args: args{guid: guid},
			want: want{err: deleteErr},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					GetByGUIDs(s.ctx, []uuid.UUID{args.guid}).
					Return([]entity.Product{{GUID: args.guid}}, nil).
					Once()

				s.productRepo.EXPECT().
					Delete(s.ctx, args.guid).
					Return(deleteErr).
					Once()
			},
		},
		{
			name: "product lookup error",
			args: args{guid: guid},
			want: want{err: repoErr},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					GetByGUIDs(s.ctx, []uuid.UUID{args.guid}).
					Return(nil, repoErr).
					Once()
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			s.productRepo.EXPECT().
				InsideTx(
					s.ctx,
					mock.AnythingOfType("func(context.Context) error"),
				).
				RunAndReturn(func(
					ctx context.Context,
					fn func(context.Context) error,
				) error {
					return fn(ctx)
				}).
				Once()

			tc.prepare(tc.args)

			err := s.srv.Delete(s.ctx, tc.args.guid)

			if tc.want.err != nil {
				s.ErrorIs(err, tc.want.err)
			} else {
				s.NoError(err)
			}
		})
	}
}

type listProductsSuite struct {
	suite.Suite
	srv          *srv
	productRepo  *mocks.MockProduct
	categoryRepo *mocks.MockCategory
	ctx          context.Context
}

func (s *listProductsSuite) SetupTest() {
	s.ctx = context.Background()
	s.productRepo = mocks.NewMockProduct(s.T())
	s.categoryRepo = mocks.NewMockCategory(s.T())
	s.srv = &srv{
		repoProduct:  s.productRepo,
		repoCategory: s.categoryRepo,
	}
}

func TestListProductSuite(t *testing.T) {
	suite.Run(t, new(listProductsSuite))
}

func (s *listProductsSuite) TestList() {
	type args struct {
		req entity.RequestProductList
	}
	type want struct {
		products []entity.Product
	}

	categoryGUID := uuid.Must(uuid.NewV4())
	minPrice := int64(50)
	maxPrice := int64(2000)

	products := []entity.Product{
		{
			Name:         "First Product",
			Price:        1000,
			CategoryGUID: categoryGUID,
		},
		{
			Name:         "Second Product",
			Price:        1500,
			CategoryGUID: categoryGUID,
		},
	}

	testCases := []struct {
		name    string
		args    args
		want    want
		prepare func(args args)
	}{
		{
			name: "success",
			args: args{
				req: entity.RequestProductList{
					CategoryGUID: &categoryGUID,
					MinPrice:     &minPrice,
					MaxPrice:     &maxPrice,
				},
			},
			want: want{products: products},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					List(
						s.ctx,
						(*string)(nil),
						args.req.CategoryGUID,
						args.req.MinPrice,
						args.req.MaxPrice,
					).
					Return(products, nil).
					Once()
			},
		},
		{
			name: "empty result",
			args: args{
				req: entity.RequestProductList{},
			},
			want: want{
				products: []entity.Product{},
			},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					List(
						s.ctx,
						(*string)(nil),
						args.req.CategoryGUID,
						args.req.MinPrice,
						args.req.MaxPrice,
					).
					Return([]entity.Product{}, nil).
					Once()
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()
			tc.prepare(tc.args)

			result, err := s.srv.List(s.ctx, tc.args.req)

			s.Require().NoError(err)
			s.Len(result, len(tc.want.products))
			s.Equal(tc.want.products, result)
		})
	}
}

type updateProductSuite struct {
	suite.Suite
	srv          *srv
	productRepo  *mocks.MockProduct
	categoryRepo *mocks.MockCategory
	ctx          context.Context
}

func (s *updateProductSuite) SetupTest() {
	s.ctx = context.Background()
	s.productRepo = mocks.NewMockProduct(s.T())
	s.categoryRepo = mocks.NewMockCategory(s.T())
	s.srv = &srv{
		repoProduct:  s.productRepo,
		repoCategory: s.categoryRepo,
	}
}

func TestUpdateProductSuite(t *testing.T) {
	suite.Run(t, new(updateProductSuite))
}

func (s *updateProductSuite) TestUpdate() {
	type args struct {
		guid uuid.UUID
		req  entity.RequestProductUpdate
	}

	type want struct {
		product entity.Product
		err     error
	}

	oldTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	repoErr := errors.New("repository failure")

	current := entity.Product{
		ID:           17,
		GUID:         uuid.Must(uuid.NewV4()),
		Name:         "Old Name",
		Description:  testutil.PtrString("Old description"),
		Price:        1000,
		CategoryGUID: uuid.Must(uuid.NewV4()),
		CreatedAt:    oldTime,
		UpdatedAt:    oldTime,
	}

	newCategoryGUID := uuid.Must(uuid.NewV4())
	otherProductGUID := uuid.Must(uuid.NewV4())

	fullReq := entity.RequestProductUpdate{
		Name:         "New Name",
		Description:  testutil.PtrString("New description"),
		Price:        2000,
		CategoryGUID: newCategoryGUID,
	}

	fullProduct := current
	fullProduct.Name = fullReq.Name
	fullProduct.Description = fullReq.Description
	fullProduct.Price = fullReq.Price
	fullProduct.CategoryGUID = fullReq.CategoryGUID

	partialReq := entity.RequestProductUpdate{
		Name: "Only New Name",
	}

	partialProduct := current
	partialProduct.Name = partialReq.Name

	testCases := []struct {
		name    string
		args    args
		want    want
		prepare func(args args)
	}{
		{
			name: "full update",
			args: args{guid: current.GUID, req: fullReq},
			want: want{product: fullProduct},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					GetByGUIDs(s.ctx, []uuid.UUID{args.guid}).
					Return([]entity.Product{current}, nil).
					Once()

				s.productRepo.EXPECT().
					List(
						s.ctx,
						&args.req.Name,
						(*uuid.UUID)(nil),
						(*int64)(nil),
						(*int64)(nil),
					).
					Return([]entity.Product{}, nil).
					Once()

				s.categoryRepo.EXPECT().
					GetByGUIDs(s.ctx, []uuid.UUID{args.req.CategoryGUID}).
					Return([]entity.Category{{GUID: newCategoryGUID}}, nil).
					Once()
			},
		},
		{
			name: "partial update",
			args: args{guid: current.GUID, req: partialReq},
			want: want{product: partialProduct},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					GetByGUIDs(s.ctx, []uuid.UUID{args.guid}).
					Return([]entity.Product{current}, nil).
					Once()

				s.productRepo.EXPECT().
					List(
						s.ctx,
						&args.req.Name,
						(*uuid.UUID)(nil),
						(*int64)(nil),
						(*int64)(nil),
					).
					Return([]entity.Product{}, nil).
					Once()
			},
		},
		{
			name: "not found",
			args: args{guid: current.GUID},
			want: want{err: entity.ErrNotFound},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					GetByGUIDs(s.ctx, []uuid.UUID{args.guid}).
					Return([]entity.Product{}, nil).
					Once()
			},
		},
		{
			name: "duplicate name",
			args: args{
				guid: current.GUID,
				req: entity.RequestProductUpdate{
					Name: "Busy Name",
				},
			},
			want: want{err: entity.ErrAlreadyExists},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					GetByGUIDs(s.ctx, []uuid.UUID{args.guid}).
					Return([]entity.Product{current}, nil).
					Once()

				s.productRepo.EXPECT().
					List(
						s.ctx,
						&args.req.Name,
						(*uuid.UUID)(nil),
						(*int64)(nil),
						(*int64)(nil),
					).
					Return([]entity.Product{{
						GUID: otherProductGUID,
						Name: args.req.Name,
					}}, nil).
					Once()
			},
		},
		{
			name: "category not found",
			args: args{
				guid: current.GUID,
				req: entity.RequestProductUpdate{
					CategoryGUID: newCategoryGUID,
				},
			},
			want: want{err: entity.ErrNotFound},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					GetByGUIDs(s.ctx, []uuid.UUID{args.guid}).
					Return([]entity.Product{current}, nil).
					Once()

				s.categoryRepo.EXPECT().
					GetByGUIDs(s.ctx, []uuid.UUID{args.req.CategoryGUID}).
					Return([]entity.Category{}, nil).
					Once()
			},
		},
		{
			name: "product lookup error",
			args: args{
				guid: current.GUID,
			},
			want: want{err: repoErr},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					GetByGUIDs(s.ctx, []uuid.UUID{args.guid}).
					Return(nil, repoErr).
					Once()
			},
		},
		{
			name: "list error",
			args: args{
				guid: current.GUID,
				req: entity.RequestProductUpdate{
					Name: "New Name",
				},
			},
			want: want{err: repoErr},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					GetByGUIDs(s.ctx, []uuid.UUID{args.guid}).
					Return([]entity.Product{current}, nil).
					Once()

				s.productRepo.EXPECT().
					List(
						s.ctx,
						&args.req.Name,
						(*uuid.UUID)(nil),
						(*int64)(nil),
						(*int64)(nil),
					).
					Return(nil, repoErr).
					Once()
			},
		},
		{
			name: "category lookup error",
			args: args{
				guid: current.GUID,
				req: entity.RequestProductUpdate{
					CategoryGUID: newCategoryGUID,
				},
			},
			want: want{err: repoErr},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					GetByGUIDs(s.ctx, []uuid.UUID{args.guid}).
					Return([]entity.Product{current}, nil).
					Once()

				s.categoryRepo.EXPECT().
					GetByGUIDs(s.ctx, []uuid.UUID{args.req.CategoryGUID}).
					Return(nil, repoErr).
					Once()
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			s.productRepo.EXPECT().
				InsideTx(
					s.ctx,
					mock.AnythingOfType("func(context.Context) error"),
				).
				RunAndReturn(func(
					ctx context.Context,
					fn func(context.Context) error,
				) error {
					return fn(ctx)
				}).
				Once()

			tc.prepare(tc.args)

			var savedProduct entity.Product

			if tc.want.err == nil {
				s.productRepo.EXPECT().
					Update(s.ctx, mock.AnythingOfType("entity.Product")).
					Run(func(_ context.Context, product entity.Product) {
						savedProduct = product
					}).
					Return(nil).
					Once()
			}
			started := time.Now()

			result, err := s.srv.Update(
				s.ctx,
				tc.args.guid,
				tc.args.req,
			)

			finished := time.Now()

			if tc.want.err != nil {
				s.ErrorIs(err, tc.want.err)
				s.Equal(entity.Product{}, result)
				return
			}

			s.Require().NoError(err)

			s.True(!result.UpdatedAt.Before(started))
			s.True(!result.UpdatedAt.After(finished))

			expected := tc.want.product
			expected.UpdatedAt = result.UpdatedAt

			s.Equal(expected, result)
			s.Equal(result, savedProduct)
		})
	}
}
