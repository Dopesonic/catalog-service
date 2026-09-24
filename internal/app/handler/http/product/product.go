package hproduct

import (
	"encoding/json"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"

	"github.com/Dopesonic/catalog-service/internal/app/entity"
	rhandler "github.com/Dopesonic/catalog-service/internal/app/handler/http"
	"github.com/Dopesonic/catalog-service/internal/app/service"
	"github.com/Dopesonic/catalog-service/internal/pkg/http/httph"
)

type handler struct {
	srv service.Product
}

func NewHandler(srv service.Product) rhandler.Product {
	return &handler{srv: srv}
}

func (h *handler) Create(w http.ResponseWriter, r *http.Request) {
	var req entity.RequestProductCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httph.HandleError(w, entity.ErrIncorrectParameters)
		return
	}
	if err := req.Validate(); err != nil {
		httph.HandleError(w, err)
		return
	}
	product, err := h.srv.Create(r.Context(), req)
	if err != nil {
		httph.HandleError(w, err)
		return
	}

	resp := entity.ResponseProductCreate{
		GUID:         product.GUID,
		Name:         product.Name,
		Description:  product.Description,
		Price:        product.Price,
		CategoryGUID: product.CategoryGUID,
		CreatedAt:    product.CreatedAt,
	}

	httph.SendJSON(w, http.StatusCreated, resp)
}

func (h *handler) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	guid, err := uuid.FromString(vars["guid"])
	if err != nil {
		httph.HandleError(w, entity.ErrIncorrectParameters)
		return
	}

	var req entity.RequestProductUpdate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httph.HandleError(w, entity.ErrIncorrectParameters)
		return
	}

	if err := req.Validate(); err != nil {
		httph.HandleError(w, err)
		return
	}

	updatedProd, err := h.srv.Update(r.Context(), guid, req)
	if err != nil {
		httph.HandleError(w, err)
		return
	}

	response := entity.ResponseProductUpdate{
		GUID:         updatedProd.GUID,
		Name:         updatedProd.Name,
		Description:  updatedProd.Description,
		Price:        updatedProd.Price,
		CategoryGUID: updatedProd.CategoryGUID,
		CreatedAt:    updatedProd.CreatedAt,
		UpdatedAt:    updatedProd.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

func (h *handler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	guid, err := uuid.FromString(vars["guid"])
	if err != nil {
		httph.HandleError(w, entity.ErrIncorrectParameters)
		return
	}

	if err := h.srv.Delete(r.Context(), guid); err != nil {
		httph.HandleError(w, err)
		return
	}

	httph.SendEmpty(w, http.StatusOK)
}

func (h *handler) List(w http.ResponseWriter, r *http.Request) {
	var req entity.RequestProductList
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httph.HandleError(w, entity.ErrIncorrectParameters)
		return
	}

	products, err := h.srv.List(r.Context(), req)
	if err != nil {
		httph.HandleError(w, err)
		return
	}

	listItems := make([]entity.ResponseProductListItem, 0, len(products))
	for _, p := range products {
		item := entity.ResponseProductListItem{
			GUID:         p.GUID,
			Name:         p.Name,
			Description:  p.Description,
			Price:        p.Price,
			CategoryGUID: p.CategoryGUID,
			CreatedAt:    p.CreatedAt,
			UpdatedAt:    p.UpdatedAt,
		}
		listItems = append(listItems, item)
	}

	response := entity.ResponseProductList{
		Data: listItems,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}
