package httpapi

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"avatar/backend/internal/model"
	"avatar/backend/internal/repo"
)

// Server exposes the REST API over the configured repo.
type Server struct {
	repo repo.Repo
	mux  *http.ServeMux
}

// NewServer builds the HTTP handler with all routes registered.
func NewServer(r repo.Repo) http.Handler {
	s := &Server{repo: r, mux: http.NewServeMux()}
	s.mux.HandleFunc("GET /health", s.handleHealth)
	s.mux.HandleFunc("GET /materials", s.handleListMaterials)
	s.mux.HandleFunc("POST /materials", s.handleCreateMaterial)
	s.mux.HandleFunc("GET /products", s.handleListProducts)
	s.mux.HandleFunc("POST /products", s.handleCreateProduct)
	s.mux.HandleFunc("GET /lots", s.handleListLots)
	s.mux.HandleFunc("POST /lots", s.handleCreateLot)
	s.mux.HandleFunc("GET /lots/{code}", s.handleGetLot)
	s.mux.HandleFunc("GET /lots/{code}/movements", s.handleListMovements)
	s.mux.HandleFunc("POST /stock-movements", s.handleAddStockMovement)
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleListMaterials(w http.ResponseWriter, r *http.Request) {
	items, err := s.repo.ListMaterials(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) handleCreateMaterial(w http.ResponseWriter, r *http.Request) {
	var m model.Material
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	created, err := s.repo.CreateMaterial(r.Context(), m)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) handleListProducts(w http.ResponseWriter, r *http.Request) {
	items, err := s.repo.ListProducts(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) handleCreateProduct(w http.ResponseWriter, r *http.Request) {
	var p model.Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	created, err := s.repo.CreateProduct(r.Context(), p)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) handleListLots(w http.ResponseWriter, r *http.Request) {
	items, err := s.repo.ListLots(r.Context(), r.URL.Query().Get("item_type"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) handleCreateLot(w http.ResponseWriter, r *http.Request) {
	var l model.Lot
	if err := json.NewDecoder(r.Body).Decode(&l); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	created, err := s.repo.CreateLot(r.Context(), l)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) handleGetLot(w http.ResponseWriter, r *http.Request) {
	l, err := s.repo.GetLotByCode(r.Context(), r.PathValue("code"))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, err)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, l)
}

func (s *Server) handleListMovements(w http.ResponseWriter, r *http.Request) {
	l, err := s.repo.GetLotByCode(r.Context(), r.PathValue("code"))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, err)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	movements, err := s.repo.ListStockMovements(r.Context(), l.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, movements)
}

func (s *Server) handleAddStockMovement(w http.ResponseWriter, r *http.Request) {
	var m model.StockMovement
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	created, err := s.repo.AddStockMovement(r.Context(), m)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}
