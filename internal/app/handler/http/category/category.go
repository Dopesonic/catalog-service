package hcategory

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
	srv service.Category
}

func NewHandler(srv service.Category) rhandler.Category {
	return &handler{srv: srv}
}

func (h *handler) Create(w http.ResponseWriter, r *http.Request) {
	var req entity.RequestCategoryCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httph.HandleError(w, entity.ErrIncorrectParameters)
		return
	}
	if err := req.Validate(); err != nil {
		httph.HandleError(w, err)
		return
	}
	category, err := h.srv.Create(r.Context(), req)
	if err != nil {
		httph.HandleError(w, err)
		return
	}

	resp := entity.ResponseCategoryCreate{
		GUID:      category.GUID,
		Name:      category.Name,
		CreatedAt: category.CreatedAt,
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

	var req entity.RequestCategoryUpdate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httph.HandleError(w, entity.ErrIncorrectParameters)
		return
	}

	if err := req.Validate(); err != nil {
		httph.HandleError(w, err)
		return
	}

	updatedCat, err := h.srv.Update(r.Context(), guid, req)
	if err != nil {
		httph.HandleError(w, err)
		return
	}

	response := entity.ResponseCategoryUpdate{
		GUID:      updatedCat.GUID,
		Name:      updatedCat.Name,
		CreatedAt: updatedCat.CreatedAt,
		UpdatedAt: updatedCat.UpdatedAt,
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
	categories, err := h.srv.List(r.Context())
	if err != nil {
		httph.HandleError(w, err)
		return
	}

	listItems := make([]entity.ResponseCategoryListItem, 0, len(categories))
	for _, cat := range categories {
		item := entity.ResponseCategoryListItem{
			GUID:      cat.GUID,
			Name:      cat.Name,
			CreatedAt: cat.CreatedAt,
			UpdatedAt: cat.UpdatedAt,
		}
		listItems = append(listItems, item)
	}

	response := entity.ResponseCategoryList{
		Data: listItems,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}
