package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/ilpaka/landmark_app/backend/journal-service/internal/modules/journal/domain"
)

type Store interface {
	InsertEntry(ctx context.Context, e domain.Entry) (*domain.Entry, error)
	GetEntry(ctx context.Context, id uuid.UUID) (*domain.Entry, error)
	ListEntries(ctx context.Context, authorID uuid.UUID, tripID *uuid.UUID, cursor string, limit int) ([]domain.Entry, string, error)
	UpdateEntry(ctx context.Context, id uuid.UUID, updates map[string]any) (*domain.Entry, error)
	DeleteEntry(ctx context.Context, id uuid.UUID) error
	AddMedia(ctx context.Context, entryID, mediaID uuid.UUID, sortOrder int) error
	RemoveMedia(ctx context.Context, entryID, mediaID uuid.UUID) error
	ListMedia(ctx context.Context, entryID uuid.UUID) ([]uuid.UUID, error)

	// Reactions
	UpsertReaction(ctx context.Context, r domain.Reaction) error
	DeleteReaction(ctx context.Context, entryID, authorID uuid.UUID, emoji string) error
	ListReactionCounts(ctx context.Context, entryID uuid.UUID) ([]domain.ReactionCount, error)
	GetUserReactions(ctx context.Context, entryID, authorID uuid.UUID) ([]string, error)
}
