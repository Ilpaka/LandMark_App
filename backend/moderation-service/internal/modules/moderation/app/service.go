package app

import (
	"context"
	"github.com/google/uuid"
	"github.com/ilpaka/landmark_app/backend/moderation-service/internal/modules/moderation/domain"
	"github.com/ilpaka/landmark_app/backend/moderation-service/internal/modules/moderation/ports"
)

type Service struct {
	Store  ports.Store
	Places ports.PlacesClient
}

func (s *Service) SubmitItem(ctx context.Context, targetType domain.TargetType, targetID, submittedBy uuid.UUID) (*domain.QueueItem, error) {
	existing, err := s.Store.GetItemByTarget(ctx, targetType, targetID)
	if err == nil && existing.Status == domain.StatusPending {
		return existing, domain.ErrConflict
	}
	return s.Store.InsertItem(ctx, targetType, targetID, submittedBy)
}

func (s *Service) ListPending(ctx context.Context, f domain.ListFilter) (*domain.QueuePage, error) {
	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 20
	}
	st := domain.StatusPending
	f.Status = &st
	return s.Store.ListPending(ctx, f)
}

func (s *Service) Approve(ctx context.Context, moderatorID, itemID uuid.UUID) error {
	item, err := s.Store.GetItem(ctx, itemID)
	if err != nil {
		return err
	}
	if item.Status != domain.StatusPending {
		return domain.ErrBadRequest
	}
	if item.TargetType == domain.TargetPlace {
		if err := s.Places.ApprovePlace(ctx, item.TargetID); err != nil {
			return err
		}
	}
	return s.Store.UpdateDecision(ctx, itemID, domain.StatusApproved, moderatorID, nil)
}

func (s *Service) Reject(ctx context.Context, moderatorID, itemID uuid.UUID, note string) error {
	item, err := s.Store.GetItem(ctx, itemID)
	if err != nil {
		return err
	}
	if item.Status != domain.StatusPending {
		return domain.ErrBadRequest
	}
	if item.TargetType == domain.TargetPlace {
		if err := s.Places.RejectPlace(ctx, item.TargetID, note); err != nil {
			return err
		}
	}
	return s.Store.UpdateDecision(ctx, itemID, domain.StatusRejected, moderatorID, &note)
}

func (s *Service) GetStats(ctx context.Context) (*domain.ModerationStats, error) {
	return s.Store.GetStats(ctx)
}
