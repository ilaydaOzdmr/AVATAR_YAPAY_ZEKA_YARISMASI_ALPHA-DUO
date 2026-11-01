package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"avatar/backend/internal/model"
	"avatar/backend/internal/repo"
)

// ErrLotBlocked is returned when a sale targets a lot blocked by a recall.
var ErrLotBlocked = repo.ErrLotBlocked

// ErrUnknownRecallMarket is returned when a confirmation does not match any
// market affected by the recall.
var ErrUnknownRecallMarket = errors.New("market is not affected by this recall")

// RecallResult is a recall event plus its affected product lots (with markets).
type RecallResult struct {
	Event            model.RecallEvent `json:"event"`
	SourceLotCode    string            `json:"source_lot_code"`
	AffectedLots     []AffectedLot     `json:"affected_lots"`
	ConfirmedMarkets []string          `json:"confirmed_markets"`
	Message          string            `json:"message"`
}

// RecallService creates recalls and blocks affected lots.
type RecallService struct {
	repo  repo.Repo
	trace *Traceability
}

// NewRecallService returns a RecallService backed by the given repository.
func NewRecallService(r repo.Repo) *RecallService {
	return &RecallService{repo: r, trace: NewTraceability(r)}
}

// Create blocks every lot affected by sourceLotCode and returns the result.
func (s *RecallService) Create(ctx context.Context, sourceLotCode, reason string) (*RecallResult, error) {
	sourceLot, err := s.repo.GetLotByCode(ctx, sourceLotCode)
	if err != nil {
		return nil, err
	}

	// Repeating a recall for the same lot returns the existing event with its
	// persisted recipient confirmations instead of creating a duplicate.
	existing, err := s.repo.GetLatestRecallBySourceLot(ctx, sourceLot.ID)
	if err == nil {
		affected, err := s.trace.Forward(ctx, sourceLotCode)
		if err != nil {
			return nil, err
		}
		result, err := s.buildResult(ctx, existing, sourceLotCode, affected)
		if err != nil {
			return nil, err
		}
		result.Message = "Bu lot için aktif geri çağırma zaten var. " + result.Message
		return result, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	affected, err := s.trace.Forward(ctx, sourceLotCode)
	if err != nil {
		return nil, err
	}

	blocks := []model.RecallBlock{{LotID: sourceLot.ID}}
	seen := map[int64]bool{sourceLot.ID: true}
	for _, lot := range affected {
		blockedLot, err := s.repo.GetLotByCode(ctx, lot.LotCode)
		if err != nil {
			return nil, err
		}
		if seen[blockedLot.ID] {
			continue
		}
		seen[blockedLot.ID] = true
		blocks = append(blocks, model.RecallBlock{LotID: blockedLot.ID})
	}

	// One transaction writes the event, the blocks and lots.status='blocked'.
	event, err := s.repo.CreateRecall(ctx, model.RecallEvent{
		Code:        fmt.Sprintf("REC-%s", time.Now().Format("20060102-150405")),
		Reason:      reason,
		SourceLotID: sourceLot.ID,
	}, blocks)
	if err != nil {
		return nil, err
	}

	return s.buildResult(ctx, event, sourceLotCode, affected)
}

// Get reconstructs the recall result for a stored recall event.
func (s *RecallService) Get(ctx context.Context, id int64) (*RecallResult, error) {
	event, err := s.repo.GetRecallEvent(ctx, id)
	if err != nil {
		return nil, err
	}
	sourceLot, err := s.repo.GetLot(ctx, event.SourceLotID)
	if err != nil {
		return nil, err
	}
	affected, err := s.trace.Forward(ctx, sourceLot.Code)
	if err != nil {
		return nil, err
	}
	return s.buildResult(ctx, event, sourceLot.Code, affected)
}

func (s *RecallService) buildResult(ctx context.Context, event model.RecallEvent, sourceLotCode string, affected []AffectedLot) (*RecallResult, error) {
	confirmed, err := s.repo.ListRecallAcknowledgements(ctx, event.SourceLotID)
	if err != nil {
		return nil, err
	}
	result := &RecallResult{
		Event:            event,
		SourceLotCode:    sourceLotCode,
		AffectedLots:     affected,
		ConfirmedMarkets: confirmed,
	}
	marketCount := 0
	subBuyerCount := 0
	for _, lot := range affected {
		marketCount += len(lot.Markets)
		subBuyerCount += len(lot.SubBuyers)
	}
	result.Message = fmt.Sprintf(
		"Geri çağırma: %s lotundan etkilenen ürünler %d mamul lotunda tespit edildi ve satışı engellendi. Etkilenen market sayısı: %d, ikinci kademe alıcı sayısı: %d.",
		sourceLotCode, len(affected), marketCount, subBuyerCount)
	return result, nil
}

// Acknowledge persists a market's recall confirmation for the affected source
// lot. Confirmations are shared by every recall event for the same lot and do
// not lift the sales block.
func (s *RecallService) Acknowledge(ctx context.Context, recallID int64, marketCode string) (*RecallResult, error) {
	event, err := s.repo.GetRecallEvent(ctx, recallID)
	if err != nil {
		return nil, err
	}
	sourceLot, err := s.repo.GetLot(ctx, event.SourceLotID)
	if err != nil {
		return nil, err
	}
	affected, err := s.trace.Forward(ctx, sourceLot.Code)
	if err != nil {
		return nil, err
	}

	marketCode = strings.TrimSpace(marketCode)
	valid := false
	for _, lot := range affected {
		for _, market := range lot.Markets {
			if market.MarketCode == marketCode {
				valid = true
				break
			}
		}
	}
	if !valid {
		return nil, ErrUnknownRecallMarket
	}

	if err := s.repo.AcknowledgeRecallMarket(ctx, event.SourceLotID, marketCode); err != nil {
		return nil, err
	}
	return s.buildResult(ctx, event, sourceLot.Code, affected)
}

// CreateSale validates that no line targets a blocked lot, then persists the sale.
func (s *RecallService) CreateSale(ctx context.Context, sale model.Sale, lines []model.SaleLine) (model.Sale, error) {
	for _, line := range lines {
		blocked, err := s.repo.IsLotBlocked(ctx, line.LotID)
		if err != nil {
			return model.Sale{}, err
		}
		if blocked {
			return model.Sale{}, ErrLotBlocked
		}
	}
	return s.repo.CreateSale(ctx, sale, lines)
}
