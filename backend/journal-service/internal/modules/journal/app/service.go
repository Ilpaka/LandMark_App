package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ilpaka/landmark_app/backend/journal-service/internal/modules/journal/domain"
	"github.com/ilpaka/landmark_app/backend/journal-service/internal/modules/journal/ports"
)

type Service struct{ Store ports.Store }

func New(s ports.Store) *Service { return &Service{Store: s} }

func (s *Service) Create(ctx context.Context, authorID uuid.UUID, input domain.Entry) (*domain.Entry, error) {
	input.ID = uuid.New()
	input.AuthorID = authorID
	input.CreatedAt = time.Now()
	input.UpdatedAt = time.Now()
	if input.OccurredAt.IsZero() {
		input.OccurredAt = time.Now()
	}
	return s.Store.InsertEntry(ctx, input)
}

func (s *Service) Get(ctx context.Context, authorID, entryID uuid.UUID) (*domain.Entry, error) {
	e, err := s.Store.GetEntry(ctx, entryID)
	if err != nil {
		return nil, err
	}
	if e == nil {
		return nil, domain.ErrNotFound
	}
	if e.AuthorID != authorID {
		return nil, domain.ErrForbidden
	}
	return e, nil
}

func (s *Service) List(ctx context.Context, authorID uuid.UUID, tripID *uuid.UUID, cursor string, limit int) ([]domain.Entry, string, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	return s.Store.ListEntries(ctx, authorID, tripID, cursor, limit)
}

func (s *Service) Update(ctx context.Context, authorID, entryID uuid.UUID, updates map[string]any) (*domain.Entry, error) {
	if _, err := s.Get(ctx, authorID, entryID); err != nil {
		return nil, err
	}
	return s.Store.UpdateEntry(ctx, entryID, updates)
}

func (s *Service) Delete(ctx context.Context, authorID, entryID uuid.UUID) error {
	if _, err := s.Get(ctx, authorID, entryID); err != nil {
		return err
	}
	return s.Store.DeleteEntry(ctx, entryID)
}

func (s *Service) AddMedia(ctx context.Context, authorID, entryID, mediaID uuid.UUID, sortOrder int) error {
	if _, err := s.Get(ctx, authorID, entryID); err != nil {
		return err
	}
	return s.Store.AddMedia(ctx, entryID, mediaID, sortOrder)
}

func (s *Service) RemoveMedia(ctx context.Context, authorID, entryID, mediaID uuid.UUID) error {
	if _, err := s.Get(ctx, authorID, entryID); err != nil {
		return err
	}
	return s.Store.RemoveMedia(ctx, entryID, mediaID)
}

func (s *Service) ListMedia(ctx context.Context, authorID, entryID uuid.UUID) ([]uuid.UUID, error) {
	if _, err := s.Get(ctx, authorID, entryID); err != nil {
		return nil, err
	}
	return s.Store.ListMedia(ctx, entryID)
}

// Reactions — any authenticated user may react (not just the author).

func (s *Service) AddReaction(ctx context.Context, userID, entryID uuid.UUID, emoji string) error {
	if emoji == "" {
		return domain.ErrBadRequest
	}
	e, err := s.Store.GetEntry(ctx, entryID)
	if err != nil || e == nil {
		return domain.ErrNotFound
	}
	return s.Store.UpsertReaction(ctx, domain.Reaction{
		EntryID: entryID, AuthorID: userID, Emoji: emoji,
	})
}

func (s *Service) RemoveReaction(ctx context.Context, userID, entryID uuid.UUID, emoji string) error {
	return s.Store.DeleteReaction(ctx, entryID, userID, emoji)
}

func (s *Service) ListReactions(ctx context.Context, entryID, callerID uuid.UUID) ([]domain.ReactionCount, []string, error) {
	counts, err := s.Store.ListReactionCounts(ctx, entryID)
	if err != nil {
		return nil, nil, err
	}
	mine, err := s.Store.GetUserReactions(ctx, entryID, callerID)
	if err != nil {
		return nil, nil, err
	}
	return counts, mine, nil
}
