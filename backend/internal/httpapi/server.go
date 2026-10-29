package httpapi

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"avatar/backend/internal/model"
	"avatar/backend/internal/repo"
	"avatar/backend/internal/service"
)

// Server exposes the REST API over the configured repo.
type Server struct {
	repo     repo.Repo
	cost     *service.CostEngine
	trace    *service.Traceability
	butchery *service.ButcheryService
	recall   *service.RecallService
	anomaly  *service.AnomalyService
	mux      *http.ServeMux
}

// NewServer builds the HTTP handler with all routes registered.
func NewServer(r repo.Repo) http.Handler {
	s := &Server{
		repo:     r,
		cost:     service.NewCostEngine(r),
		trace:    service.NewTraceability(r),
		butchery: service.NewButcheryService(r),
		recall:   service.NewRecallService(r),
		anomaly:  service.NewAnomalyService(r),
		mux:      http.NewServeMux(),
	}
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
	s.mux.HandleFunc("POST /production-orders", s.handleCreateProductionOrder)
	s.mux.HandleFunc("GET /production-orders", s.handleListProductionOrders)
	s.mux.HandleFunc("GET /production-orders/{id}", s.handleGetProductionOrder)
	s.mux.HandleFunc("GET /trace/forward", s.handleTraceForward)
	s.mux.HandleFunc("GET /trace/backward", s.handleTraceBackward)
	s.mux.HandleFunc("POST /butchery-batches", s.handleCreateButcheryBatch)
	s.mux.HandleFunc("GET /butchery-batches", s.handleListButcheryBatches)
	s.mux.HandleFunc("GET /butchery-batches/{id}", s.handleGetButcheryBatch)
	s.mux.HandleFunc("POST /recalls", s.handleCreateRecall)
	s.mux.HandleFunc("GET /recalls/{id}", s.handleGetRecall)
	s.mux.HandleFunc("GET /markets", s.handleListMarkets)
	s.mux.HandleFunc("POST /sales", s.handleCreateSale)
	s.mux.HandleFunc("GET /sales", s.handleListSales)
	s.mux.HandleFunc("POST /sync/push", s.handleSyncPush)
	s.mux.HandleFunc("GET /sync/pull", s.handleSyncPull)
	s.mux.HandleFunc("GET /anomalies", s.handleAnomalies)
	s.mux.HandleFunc("POST /ocr/invoice", s.handleOCRInvoice)
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
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

// productionOrderRequest is the payload for POST /production-orders.
type productionOrderRequest struct {
	model.ProductionOrder
	Inputs  []model.ProductionOrderInput  `json:"inputs"`
	Outputs []model.ProductionOrderOutput `json:"outputs"`
}

func (s *Server) handleCreateProductionOrder(w http.ResponseWriter, r *http.Request) {
	var req productionOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	created, err := s.repo.CreateProductionOrder(r.Context(), req.ProductionOrder, req.Inputs, req.Outputs)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) handleGetProductionOrder(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	summary, err := s.cost.ProductionOrderCost(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, err)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

func (s *Server) handleTraceForward(w http.ResponseWriter, r *http.Request) {
	lotCode := r.URL.Query().Get("lot")
	if lotCode == "" {
		writeError(w, http.StatusBadRequest, errors.New("lot query parameter is required"))
		return
	}
	affected, err := s.trace.Forward(r.Context(), lotCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, err)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, affected)
}

func (s *Server) handleTraceBackward(w http.ResponseWriter, r *http.Request) {
	lotCode := r.URL.Query().Get("lot")
	if lotCode == "" {
		writeError(w, http.StatusBadRequest, errors.New("lot query parameter is required"))
		return
	}
	sources, err := s.trace.Backward(r.Context(), lotCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, err)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, sources)
}

// butcheryBatchRequest is the payload for POST /butchery-batches.
type butcheryBatchRequest struct {
	Batch model.ButcheryBatch  `json:"batch"`
	Items []model.ButcheryItem `json:"items"`
}

func (s *Server) handleCreateButcheryBatch(w http.ResponseWriter, r *http.Request) {
	var req butcheryBatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	result, err := s.butchery.CreateBatch(r.Context(), req.Batch, req.Items)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

func (s *Server) handleGetButcheryBatch(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	result, err := s.butchery.GetBatch(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, err)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// recallRequest is the payload for POST /recalls.
type recallRequest struct {
	SourceLotCode string `json:"source_lot_code"`
	Reason        string `json:"reason"`
}

func (s *Server) handleCreateRecall(w http.ResponseWriter, r *http.Request) {
	var req recallRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if req.SourceLotCode == "" {
		writeError(w, http.StatusBadRequest, errors.New("source_lot_code is required"))
		return
	}
	result, err := s.recall.Create(r.Context(), req.SourceLotCode, req.Reason)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, err)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

func (s *Server) handleGetRecall(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	result, err := s.recall.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, err)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// saleRequest is the payload for POST /sales and sync push.
type saleRequest struct {
	model.Sale
	Lines []model.SaleLine `json:"lines"`
}

func (s *Server) handleCreateSale(w http.ResponseWriter, r *http.Request) {
	var req saleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	created, err := s.recall.CreateSale(r.Context(), req.Sale, req.Lines)
	if err != nil {
		if errors.Is(err, service.ErrLotBlocked) {
			writeError(w, http.StatusConflict, err)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) handleSyncPush(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Sales []saleRequest `json:"sales"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	created := make([]model.Sale, 0, len(req.Sales))
	for _, saleReq := range req.Sales {
		sale, err := s.recall.CreateSale(r.Context(), saleReq.Sale, saleReq.Lines)
		if err != nil {
			if errors.Is(err, service.ErrLotBlocked) {
				writeError(w, http.StatusConflict, err)
				return
			}
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		created = append(created, sale)
	}
	writeJSON(w, http.StatusOK, map[string]any{"created": created})
}

func (s *Server) handleSyncPull(w http.ResponseWriter, r *http.Request) {
	blocked, err := s.repo.ListBlockedLotCodes(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	products, err := s.repo.ListProducts(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"blocked_lot_codes": blocked,
		"products":          products,
	})
}

func (s *Server) handleAnomalies(w http.ResponseWriter, r *http.Request) {
	anomalies, err := s.anomaly.Detect(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, anomalies)
}

// ocrRequest is the payload for POST /ocr/invoice.
type ocrRequest struct {
	Text string `json:"text"`
}

func (s *Server) handleOCRInvoice(w http.ResponseWriter, r *http.Request) {
	var req ocrRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	items := service.ParseInvoice(req.Text)
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Server) handleListProductionOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := s.repo.ListProductionOrders(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, orders)
}

func (s *Server) handleListButcheryBatches(w http.ResponseWriter, r *http.Request) {
	batches, err := s.repo.ListButcheryBatches(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, batches)
}

func (s *Server) handleListMarkets(w http.ResponseWriter, r *http.Request) {
	markets, err := s.repo.ListMarkets(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, markets)
}

func (s *Server) handleListSales(w http.ResponseWriter, r *http.Request) {
	sales, err := s.repo.ListSales(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, sales)
}
