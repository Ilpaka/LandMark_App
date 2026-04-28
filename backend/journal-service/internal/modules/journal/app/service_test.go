package app

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/ilpaka/landmark_app/backend/journal-service/internal/modules/journal/domain"
)

// ---------------------------------------------------------------------------
// Hand-rolled mock store
// ---------------------------------------------------------------------------

type mockStore struct {
	// InsertEntry
	insertedEntry    *domain.Entry
	insertErr        error

	// GetEntry
	storedEntry    *domain.Entry
	getErr         error

	// UpdateEntry
	updatedEntry    *domain.Entry
	updateErr       error

	// DeleteEntry
	deleteErr error

	// ListEntries
	listEntries []domain.Entry
	listCursor  string
	listErr     error

	// Capture args
	lastDeleteID uuid.UUID
	lastUpdateID uuid.UUID
}

func (m *mockStore) InsertEntry(ctx context.Context, e domain.Entry) (*domain.Entry, error) {
	if m.insertedEntry != nil {
		return m.insertedEntry, m.insertErr
	}
	// Echo the entry back so callers can inspect what was stored.
	return &e, m.insertErr
}

func (m *mockStore) GetEntry(ctx context.Context, id uuid.UUID) (*domain.Entry, error) {
	return m.storedEntry, m.getErr
}

func (m *mockStore) ListEntries(ctx context.Context, authorID uuid.UUID, tripID *uuid.UUID, cursor string, limit int) ([]domain.Entry, string, error) {
	return m.listEntries, m.listCursor, m.listErr
}

func (m *mockStore) UpdateEntry(ctx context.Context, id uuid.UUID, updates map[string]any) (*domain.Entry, error) {
	m.lastUpdateID = id
	return m.updatedEntry, m.updateErr
}

func (m *mockStore) DeleteEntry(ctx context.Context, id uuid.UUID) error {
	m.lastDeleteID = id
	return m.deleteErr
}

func (m *mockStore) AddMedia(ctx context.Context, entryID, mediaID uuid.UUID, sortOrder int) error {
	return nil
}

func (m *mockStore) RemoveMedia(ctx context.Context, entryID, mediaID uuid.UUID) error {
	return nil
}

func (m *mockStore) ListMedia(ctx context.Context, entryID uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}

func (m *mockStore) UpsertReaction(_ context.Context, _ domain.Reaction) error { return nil }
func (m *mockStore) DeleteReaction(_ context.Context, _, _ uuid.UUID, _ string) error { return nil }
func (m *mockStore) ListReactionCounts(_ context.Context, _ uuid.UUID) ([]domain.ReactionCount, error) {
	return []domain.ReactionCount{}, nil
}
func (m *mockStore) GetUserReactions(_ context.Context, _, _ uuid.UUID) ([]string, error) {
	return []string{}, nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func strPtr(s string) *string { return &s }

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

// CreateEntry_Success: a valid entry is created; the service assigns a new ID
// and sets AuthorID from the caller.
func TestCreateEntry_Success(t *testing.T) {
	store := &mockStore{}
	svc := New(store)

	authorID := uuid.New()
	input := domain.Entry{Body: "Great hike today"}
	got, err := svc.Create(context.Background(), authorID, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.AuthorID != authorID {
		t.Errorf("AuthorID: want %s, got %s", authorID, got.AuthorID)
	}
	if got.ID == uuid.Nil {
		t.Error("expected a non-nil ID to be assigned")
	}
}

// CreateEntry_EmptyTitle: the journal service does not enforce a non-empty
// title at the service layer (Title is a *string); creation must succeed.
func TestCreateEntry_EmptyTitle(t *testing.T) {
	store := &mockStore{}
	svc := New(store)

	// Title field is *string; passing an empty string pointer is valid per domain.
	empty := ""
	input := domain.Entry{Title: &empty, Body: "Some body"}
	got, err := svc.Create(context.Background(), uuid.New(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil entry")
	}
}

// GetEntry_NotFound: when the store returns ErrNotFound the service must
// propagate it.
func TestGetEntry_NotFound(t *testing.T) {
	store := &mockStore{getErr: domain.ErrNotFound}
	svc := New(store)

	_, err := svc.Get(context.Background(), uuid.New(), uuid.New())
	if err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// GetEntry_HiddenFromOther: when an entry exists but the caller is not the
// author, the service must return ErrForbidden.
func TestGetEntry_HiddenFromOther(t *testing.T) {
	authorID := uuid.New()
	callerID := uuid.New()
	entryID := uuid.New()

	store := &mockStore{
		storedEntry: &domain.Entry{ID: entryID, AuthorID: authorID},
	}
	svc := New(store)

	_, err := svc.Get(context.Background(), callerID, entryID)
	if err != domain.ErrForbidden {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

// UpdateEntry_Forbidden: a caller who is not the author receives ErrForbidden
// when attempting to update the entry.
func TestUpdateEntry_Forbidden(t *testing.T) {
	authorID := uuid.New()
	callerID := uuid.New()
	entryID := uuid.New()

	store := &mockStore{
		storedEntry: &domain.Entry{ID: entryID, AuthorID: authorID},
	}
	svc := New(store)

	_, err := svc.Update(context.Background(), callerID, entryID, map[string]any{"body": "hacked"})
	if err != domain.ErrForbidden {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

// DeleteEntry_Forbidden: a caller who is not the author receives ErrForbidden
// when attempting to delete the entry.
func TestDeleteEntry_Forbidden(t *testing.T) {
	authorID := uuid.New()
	callerID := uuid.New()
	entryID := uuid.New()

	store := &mockStore{
		storedEntry: &domain.Entry{ID: entryID, AuthorID: authorID},
	}
	svc := New(store)

	err := svc.Delete(context.Background(), callerID, entryID)
	if err != domain.ErrForbidden {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

// DeleteEntry_Success: the correct author can delete their own entry without
// error; the store's DeleteEntry is called with the right ID.
func TestDeleteEntry_Success(t *testing.T) {
	authorID := uuid.New()
	entryID := uuid.New()

	store := &mockStore{
		storedEntry: &domain.Entry{ID: entryID, AuthorID: authorID},
	}
	svc := New(store)

	err := svc.Delete(context.Background(), authorID, entryID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.lastDeleteID != entryID {
		t.Errorf("expected DeleteEntry to be called with %s, got %s", entryID, store.lastDeleteID)
	}
}

// AddReaction_EmptyEmoji: empty emoji must be rejected with ErrBadRequest.
func TestAddReaction_EmptyEmoji(t *testing.T) {
	svc := New(&mockStore{storedEntry: &domain.Entry{ID: uuid.New()}})
	err := svc.AddReaction(context.Background(), uuid.New(), uuid.New(), "")
	if err != domain.ErrBadRequest {
		t.Errorf("expected ErrBadRequest, got %v", err)
	}
}

// AddReaction_EntryNotFound: reacting to a missing entry returns ErrNotFound.
func TestAddReaction_EntryNotFound(t *testing.T) {
	store := &mockStore{storedEntry: nil}
	svc := New(store)
	err := svc.AddReaction(context.Background(), uuid.New(), uuid.New(), "👍")
	if err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// AddReaction_Success: a valid reaction on an existing entry succeeds.
func TestAddReaction_Success(t *testing.T) {
	entryID := uuid.New()
	store := &mockStore{storedEntry: &domain.Entry{ID: entryID}}
	svc := New(store)
	if err := svc.AddReaction(context.Background(), uuid.New(), entryID, "❤️"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ListReactions_Success: counts and mine slices are returned without error.
func TestListReactions_Success(t *testing.T) {
	entryID := uuid.New()
	store := &mockStore{storedEntry: &domain.Entry{ID: entryID}}
	svc := New(store)
	counts, mine, err := svc.ListReactions(context.Background(), entryID, uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if counts == nil || mine == nil {
		t.Error("expected non-nil slices")
	}
}
