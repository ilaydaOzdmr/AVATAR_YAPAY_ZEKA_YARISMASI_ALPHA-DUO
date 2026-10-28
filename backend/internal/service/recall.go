package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"avatar/backend/internal/model"
	"avatar/backend/internal/repo"
)

// ErrLotBlocked is returned when a sale targets a lot blocked by a recall.
var ErrLotBlocked = errors.New("lot is blocked by a recall")

// RecallResult is a recall event plus its affected product lots (with markets).
type RecallResult struct {
	Event         model.RecallEvent `json:"event"`
	SourceLotCode string            `json:"source_lot_code"`
	AffectedLots  []AffectedLot     `json:"affected_lots"`
	Message       string            `json:"message"`
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

// Create blocks every product lot affected by sourceLotCode and returns the result.
func (s *RecallService) Create(ctx context.Context, sourceLotCode, reason string) (*RecallResult, error) {
	sourceLot, err := s.repo.GetLotByCode(ctx, sourceLotCode)
	if err != nil {
		return nil, err
	}

	affected, err := s.trace.Forward(ctx, sourceLotCode)
	if err != nil {
		return nil, err
	}

	event, err := s.repo.CreateRecallEvent(ctx, model.RecallEvent{
		Code:        fmt.Sprintf("REC-%s", time.Now().Format("20060102-150405")),
		Reason:      reason,
		SourceLotID: sourceLot.ID,
	})
	if err != nil {
		return nil, err
	}

	for _, lot := range affected {
		blockedLot, err := s.repo.GetLotByCode(ctx, lot.LotCode)
		if err != nil {
			return nil, err
		}
		if _, err := s.repo.CreateRecallBlock(ctx, model.RecallBlock{
			RecallEventID: event.ID,
			LotID:         blockedLot.ID,
		}); err != nil {
			return nil, err
		}
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
	result := &RecallResult{
		Event:         event,
		SourceLotCode: sourceLotCode,
		AffectedLots:  affected,
	}
	marketCount := 0
	for _, lot := range affected {
		marketCount += len(lot.Markets)
	}
	result.Message = fmt.Sprintf(
		"Geri çağırma: %s lotundan etkilenen ürünler %d mamul lotunda tespit edildi ve satışı engellendi. Etkilenen market sayısı: %d.",
		sourceLotCode, len(affected), marketCount)
	return result, nil
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
