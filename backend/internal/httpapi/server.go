package httpapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"avatar/backend/internal/config"
	"avatar/backend/internal/ml"
	"avatar/backend/internal/model"
	"avatar/backend/internal/repo"
	"avatar/backend/internal/service"
)

// Server exposes the REST API over the configured repo.
type Server struct {
	repo     repo.Repo
	ml       *ml.Client
	cost     *service.CostEngine
	trace    *service.Traceability
	butchery *service.ButcheryService
	recall   *service.RecallService
	anomaly  *service.AnomalyService
	reports  *service.ReportService
	mux      *http.ServeMux
}

// NewServer builds the HTTP handler with all routes registered and the ML
// service address taken from configuration.
const contentTypeHeader = "Content-Type"

func NewServer(r repo.Repo) http.Handler {
	return NewServerWithML(r, ml.NewClient(config.Load().MLBaseURL))
}

// NewServerWithML is used by tests and by deployments that inject a client.
func NewServerWithML(r repo.Repo, mlClient *ml.Client) http.Handler {
	s := &Server{
		repo:     r,
		ml:       mlClient,
		cost:     service.NewCostEngine(r),
		trace:    service.NewTraceability(r),
		butchery: service.NewButcheryService(r),
		recall:   service.NewRecallService(r),
		anomaly:  service.NewAnomalyService(r),
		reports:  service.NewReportService(r),
		mux:      http.NewServeMux(),
	}
	s.mux.HandleFunc("GET /health", s.handleHealth)
	s.mux.HandleFunc("GET /{$}", s.handleRoot)
	s.mux.HandleFunc("GET /materials", s.handleListMaterials)
	s.mux.HandleFunc("POST /materials", s.handleCreateMaterial)
	s.mux.HandleFunc("GET /products", s.handleListProducts)
	s.mux.HandleFunc("POST /products", s.handleCreateProduct)
	s.mux.HandleFunc("PUT /products/{id}", s.handleUpdateProductShelfLife)
	s.mux.HandleFunc("GET /lots", s.handleListLots)
	s.mux.HandleFunc("POST /lots", s.handleCreateLot)
	s.mux.HandleFunc("GET /lots/{code}", s.handleGetLot)
	s.mux.HandleFunc("GET /lots/{code}/movements", s.handleListMovements)
	s.mux.HandleFunc("POST /stock-movements", s.handleAddStockMovement)
	s.mux.HandleFunc("POST /production-orders", s.handleCreateProductionOrder)
	s.mux.HandleFunc("POST /production-orders/{id}/complete", s.handleCompleteProductionOrder)
	s.mux.HandleFunc("POST /production-orders/{id}/reject", s.handleRejectProductionOrder)
	s.mux.HandleFunc("GET /production-orders", s.handleListProductionOrders)
	s.mux.HandleFunc("GET /production-orders/{id}", s.handleGetProductionOrder)
	s.mux.HandleFunc("GET /production-orders/{id}/inputs", s.handleListProductionOrderInputs)
	s.mux.HandleFunc("GET /production-orders/{id}/outputs", s.handleListProductionOrderOutputs)
	s.mux.HandleFunc("GET /trace/forward", s.handleTraceForward)
	s.mux.HandleFunc("GET /trace/backward", s.handleTraceBackward)
	s.mux.HandleFunc("POST /butchery-batches", s.handleCreateButcheryBatch)
	s.mux.HandleFunc("GET /butchery-batches", s.handleListButcheryBatches)
	s.mux.HandleFunc("GET /butchery-batches/{id}", s.handleGetButcheryBatch)
	s.mux.HandleFunc("POST /recalls", s.handleCreateRecall)
	s.mux.HandleFunc("GET /recalls/{id}", s.handleGetRecall)
	s.mux.HandleFunc("POST /recalls/{id}/acknowledge", s.handleAcknowledgeRecall)
	s.mux.HandleFunc("GET /markets", s.handleListMarkets)
	s.mux.HandleFunc("POST /redistributions", s.handleCreateRedistribution)
	s.mux.HandleFunc("GET /redistributions", s.handleListRedistributions)
	s.mux.HandleFunc("POST /sales", s.handleCreateSale)
	s.mux.HandleFunc("GET /sales", s.handleListSales)
	s.mux.HandleFunc("POST /sync/push", s.handleSyncPush)
	s.mux.HandleFunc("GET /sync/pull", s.handleSyncPull)
	s.mux.HandleFunc("GET /anomalies", s.handleAnomalies)
	s.mux.HandleFunc("GET /reports/summary", s.handleReportSummary)
	s.mux.HandleFunc("POST /ocr/invoice", s.handleOCRInvoice)
	s.mux.HandleFunc("GET /ml/health", s.handleMLHealth)
	s.mux.HandleFunc("GET /ml/models", s.handleMLModels)
	s.mux.HandleFunc("GET /ml/report", s.handleMLReport)
	s.mux.HandleFunc("POST /ml/predict/yield", s.handleMLYieldLegacy)
	s.mux.HandleFunc("POST /ml/predict/butchery", s.handleMLButcheryLegacy)
	s.mux.HandleFunc("POST /ml/predict/v2/yield", s.handleMLYieldV2)
	s.mux.HandleFunc("POST /ml/predict/v2/butchery", s.handleMLButcheryV2)
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", contentTypeHeader)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	s.mux.ServeHTTP(w, r)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleRoot makes the base URL useful in a browser instead of returning 404.
func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"service": "avatar-backend",
		"status":  "ok",
		"version": "1.0.0",
		"endpoints": []string{
			"/health", "/ml/health", "/ml/models", "/ml/report",
			"/sync/pull", "/sync/push", "/reports/summary",
			"/materials", "/products", "/lots", "/sales",
			"/production-orders", "/butchery-batches", "/recalls/{id}",
		},
	})
}

func (s *Server) handleListMaterials(w http.ResponseWriter, r *http.Request) {
	items, err := s.repo.ListMaterials(r.Context())
	if err != nil {
		writeRepoError(w, err)
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
		writeRepoError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) handleListProducts(w http.ResponseWriter, r *http.Request) {
	items, err := s.repo.ListProducts(r.Context())
	if err != nil {
		writeRepoError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) handleUpdateProductShelfLife(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	var body struct {
		ShelfLifeDays int `json:"shelf_life_days"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := s.repo.UpdateProductShelfLife(r.Context(), id, body.ShelfLifeDays); err != nil {
		writeRepoError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"id":              id,
		"shelf_life_days": body.ShelfLifeDays,
	})
}

func (s *Server) handleCreateProduct(w http.ResponseWriter, r *http.Request) {
	var p model.Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	created, err := s.repo.CreateProduct(r.Context(), p)
	if err != nil {
		writeRepoError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) handleListLots(w http.ResponseWriter, r *http.Request) {
	items, err := s.repo.ListLots(r.Context(), r.URL.Query().Get("item_type"))
	if err != nil {
		writeRepoError(w, err)
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
		writeRepoError(w, err)
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
		writeRepoError(w, err)
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
		writeRepoError(w, err)
		return
	}
	movements, err := s.repo.ListStockMovements(r.Context(), l.ID)
	if err != nil {
		writeRepoError(w, err)
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
		writeRepoError(w, err)
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
		writeRepoError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) handleCompleteProductionOrder(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	var body struct {
		FinishedAt *string `json:"finished_at"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil && !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	completed, err := s.repo.CompleteProductionOrder(r.Context(), id, body.FinishedAt)
	if err != nil {
		writeRepoError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, completed)
}

func (s *Server) handleRejectProductionOrder(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	var body struct {
		Reason *string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil && !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	rejected, err := s.repo.RejectProductionOrder(r.Context(), id, body.Reason)
	if err != nil {
		writeRepoError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rejected)
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
		writeRepoError(w, err)
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
		writeRepoError(w, err)
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
		writeRepoError(w, err)
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
		writeRepoError(w, err)
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
		writeRepoError(w, err)
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
		writeRepoError(w, err)
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
		writeRepoError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleAcknowledgeRecall(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	var body struct {
		MarketCode string `json:"market_code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	result, err := s.recall.Acknowledge(r.Context(), id, body.MarketCode)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUnknownRecallMarket):
			writeError(w, http.StatusBadRequest, err)
		case errors.Is(err, sql.ErrNoRows):
			writeError(w, http.StatusNotFound, err)
		default:
			writeRepoError(w, err)
		}
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
		writeRepoError(w, err)
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
			writeRepoError(w, err)
			return
		}
		created = append(created, sale)
	}
	writeJSON(w, http.StatusOK, map[string]any{"created": created})
}

func (s *Server) handleSyncPull(w http.ResponseWriter, r *http.Request) {
	blocked, err := s.repo.ListBlockedLotCodes(r.Context())
	if err != nil {
		writeRepoError(w, err)
		return
	}
	products, err := s.repo.ListProducts(r.Context())
	if err != nil {
		writeRepoError(w, err)
		return
	}
	lots, err := s.repo.ListLots(r.Context(), "")
	if err != nil {
		writeRepoError(w, err)
		return
	}
	markets, err := s.repo.ListMarkets(r.Context())
	if err != nil {
		writeRepoError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"server_time":       time.Now().UTC().Format(time.RFC3339),
		"blocked_lot_codes": blocked,
		"products":          products,
		"lots":              lots,
		"markets":           markets,
	})
}

func (s *Server) handleAnomalies(w http.ResponseWriter, r *http.Request) {
	anomalies, err := s.anomaly.Detect(r.Context())
	if err != nil {
		writeRepoError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, anomalies)
}

func (s *Server) handleReportSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := s.reports.Summary(r.Context())
	if err != nil {
		writeRepoError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, summary)
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
		writeRepoError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, orders)
}

func (s *Server) handleListProductionOrderInputs(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	inputs, err := s.repo.ListProductionOrderInputs(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, inputs)
}

func (s *Server) handleListProductionOrderOutputs(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, errors.New("invalid production order id"))
		return
	}
	if _, err := s.repo.GetProductionOrder(r.Context(), id); err != nil {
		writeRepoError(w, err)
		return
	}
	outputs, err := s.repo.ListProductionOrderOutputs(r.Context(), id)
	if err != nil {
		writeRepoError(w, err)
		return
	}
	if outputs == nil {
		outputs = []model.ProductionOrderOutput{}
	}
	writeJSON(w, http.StatusOK, outputs)
}

func (s *Server) handleListButcheryBatches(w http.ResponseWriter, r *http.Request) {
	batches, err := s.repo.ListButcheryBatches(r.Context())
	if err != nil {
		writeRepoError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, batches)
}

func (s *Server) handleListMarkets(w http.ResponseWriter, r *http.Request) {
	markets, err := s.repo.ListMarkets(r.Context())
	if err != nil {
		writeRepoError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, markets)
}

func (s *Server) handleListSales(w http.ResponseWriter, r *http.Request) {
	sales, err := s.repo.ListSales(r.Context())
	if err != nil {
		writeRepoError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, sales)
}

// redistributionRequest is the payload for POST /redistributions.
type redistributionRequest struct {
	model.Redistribution
}

func (s *Server) handleCreateRedistribution(w http.ResponseWriter, r *http.Request) {
	var req redistributionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	created, err := s.repo.CreateRedistribution(r.Context(), req.Redistribution)
	if err != nil {
		writeRepoError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) handleListRedistributions(w http.ResponseWriter, r *http.Request) {
	lotID, err := strconv.ParseInt(r.URL.Query().Get("lot_id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, errors.New("lot_id query parameter is required"))
		return
	}
	redistributions, err := s.repo.ListRedistributions(r.Context(), lotID)
	if err != nil {
		writeRepoError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, redistributions)
}

// --- ML service proxy -------------------------------------------------------

func (s *Server) requireML(w http.ResponseWriter) bool {
	if s.ml == nil {
		writeError(w, http.StatusServiceUnavailable, errors.New("ml client not configured"))
		return false
	}
	return true
}

func (s *Server) handleMLHealth(w http.ResponseWriter, r *http.Request) {
	if !s.requireML(w) {
		return
	}
	raw, err := s.ml.Health(r.Context())
	writeMLResult(w, raw, err)
}

func (s *Server) handleMLModels(w http.ResponseWriter, r *http.Request) {
	if !s.requireML(w) {
		return
	}
	raw, err := s.ml.Models(r.Context())
	writeMLResult(w, raw, err)
}

func (s *Server) handleMLReport(w http.ResponseWriter, r *http.Request) {
	if !s.requireML(w) {
		return
	}
	raw, err := s.ml.Report(r.Context())
	writeMLResult(w, raw, err)
}

func (s *Server) handleMLYieldLegacy(w http.ResponseWriter, r *http.Request) {
	s.handleMLPredict(w, r, "/predict/yield")
}

func (s *Server) handleMLButcheryLegacy(w http.ResponseWriter, r *http.Request) {
	s.handleMLPredict(w, r, "/predict/butchery")
}

func (s *Server) handleMLYieldV2(w http.ResponseWriter, r *http.Request) {
	s.handleMLPredict(w, r, "/predict/v2/yield")
}

func (s *Server) handleMLButcheryV2(w http.ResponseWriter, r *http.Request) {
	s.handleMLPredict(w, r, "/predict/v2/butchery")
}

func (s *Server) handleMLPredict(w http.ResponseWriter, r *http.Request, path string) {
	if !s.requireML(w) {
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	raw, err := s.ml.Predict(r.Context(), path, body)
	writeMLResult(w, raw, err)
}

func writeMLResult(w http.ResponseWriter, raw []byte, err error) {
	if err == nil {
		w.Header().Set(contentTypeHeader, "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(raw)
		return
	}
	var statusErr *ml.StatusError
	if errors.As(err, &statusErr) {
		w.Header().Set(contentTypeHeader, "application/json")
		w.WriteHeader(statusErr.Status)
		_, _ = w.Write(statusErr.Body)
		return
	}
	if errors.Is(err, context.DeadlineExceeded) {
		writeError(w, http.StatusGatewayTimeout, err)
		return
	}
	writeError(w, http.StatusBadGateway, err)
}
