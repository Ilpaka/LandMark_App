package app

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/ilpaka/landmark_app/backend/favorites-service/internal/modules/favorite/domain"
)

// ---------------------------------------------------------------------------
// Hand-rolled mock store
// ---------------------------------------------------------------------------

type mockStore struct {
	// GetOrCreateDefaultCollection
	defaultCollection    *domain.Collection
	defaultCollectionErr error

	// GetCollection
	getCollection    *domain.Collection
	getCollectionErr error

	// InsertCollection
	insertCollection    *domain.Collection
	insertCollectionErr error

	// ListFavorites
	favoritesPage    *domain.FavoritesPage
	favoritesPageErr error

	// UpsertFavorite
	upsertErr error

	// DeleteFavorite
	deleteErr error

	// Capture arguments for assertion
	lastUpsertFav  domain.Favorite
	lastListFilter domain.ListFavoritesFilter
}

func (m *mockStore) GetDefaultCollection(ctx context.Context, userID uuid.UUID) (*domain.Collection, error) {
	return m.defaultCollection, m.defaultCollectionErr
}

func (m *mockStore) GetOrCreateDefaultCollection(ctx context.Context, userID uuid.UUID) (*domain.Collection, error) {
	return m.defaultCollection, m.defaultCollectionErr
}

func (m *mockStore) ListCollections(ctx context.Context, userID uuid.UUID) ([]domain.Collection, error) {
	return nil, nil
}

func (m *mockStore) GetCollection(ctx context.Context, id uuid.UUID) (*domain.Collection, error) {
	return m.getCollection, m.getCollectionErr
}

func (m *mockStore) InsertCollection(ctx context.Context, userID uuid.UUID, title, color string) (*domain.Collection, error) {
	return m.insertCollection, m.insertCollectionErr
}

func (m *mockStore) UpdateCollection(ctx context.Context, id uuid.UUID, title, color string) (*domain.Collection, error) {
	return nil, nil
}

func (m *mockStore) DeleteCollection(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockStore) ListFavorites(ctx context.Context, f domain.ListFavoritesFilter) (*domain.FavoritesPage, error) {
	m.lastListFilter = f
	return m.favoritesPage, m.favoritesPageErr
}

func (m *mockStore) GetFavorite(ctx context.Context, userID uuid.UUID, tt domain.TargetType, targetID uuid.UUID) (*domain.Favorite, error) {
	return nil, nil
}

func (m *mockStore) UpsertFavorite(ctx context.Context, fav domain.Favorite) error {
	m.lastUpsertFav = fav
	return m.upsertErr
}

func (m *mockStore) DeleteFavorite(ctx context.Context, userID uuid.UUID, tt domain.TargetType, targetID uuid.UUID) error {
	return m.deleteErr
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

// AddFavorite_DefaultCollection: when collectionID is nil the service must
// call GetOrCreateDefaultCollection and attach the returned ID to the favorite.
func TestAddFavorite_DefaultCollection(t *testing.T) {
	colID := uuid.New()
	store := &mockStore{
		defaultCollection: &domain.Collection{ID: colID, UserID: uuid.New()},
	}
	svc := &Service{Store: store}

	userID := uuid.New()
	targetID := uuid.New()
	err := svc.AddFavorite(context.Background(), userID, domain.TargetPlace, targetID, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if store.lastUpsertFav.CollectionID == nil {
		t.Fatal("expected CollectionID to be set, got nil")
	}
	if *store.lastUpsertFav.CollectionID != colID {
		t.Errorf("expected CollectionID %s, got %s", colID, *store.lastUpsertFav.CollectionID)
	}
}

// AddFavorite_CollectionOwnershipCheck: when a collectionID is supplied but
// belongs to a different user, AddFavorite must return ErrForbidden.
func TestAddFavorite_CollectionOwnershipCheck(t *testing.T) {
	ownerID := uuid.New()
	callerID := uuid.New() // different user
	colID := uuid.New()

	store := &mockStore{
		getCollection: &domain.Collection{ID: colID, UserID: ownerID},
	}
	svc := &Service{Store: store}

	err := svc.AddFavorite(context.Background(), callerID, domain.TargetTrip, uuid.New(), &colID, nil)
	if err != domain.ErrForbidden {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

// RemoveFavorite_Success: RemoveFavorite must delegate straight to the store
// and propagate nil on success.
func TestRemoveFavorite_Success(t *testing.T) {
	store := &mockStore{}
	svc := &Service{Store: store}

	err := svc.RemoveFavorite(context.Background(), uuid.New(), domain.TargetPlace, uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ListFavorites_Pagination: a Limit of 0 must be normalised to 20 before the
// store is called.
func TestListFavorites_Pagination(t *testing.T) {
	store := &mockStore{
		favoritesPage: &domain.FavoritesPage{Items: []domain.Favorite{}},
	}
	svc := &Service{Store: store}

	filter := domain.ListFavoritesFilter{
		UserID: uuid.New(),
		Limit:  0, // invalid — should be clamped to 20
	}
	_, err := svc.ListFavorites(context.Background(), filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.lastListFilter.Limit != 20 {
		t.Errorf("expected Limit=20 after normalisation, got %d", store.lastListFilter.Limit)
	}
}

// CreateCollection_EmptyTitle: an empty title must return ErrBadRequest
// without touching the store.
func TestCreateCollection_EmptyTitle(t *testing.T) {
	store := &mockStore{}
	svc := &Service{Store: store}

	col, err := svc.CreateCollection(context.Background(), uuid.New(), "", "#fff")
	if err != domain.ErrBadRequest {
		t.Errorf("expected ErrBadRequest, got %v", err)
	}
	if col != nil {
		t.Error("expected nil collection on error")
	}
}

// CreateCollection_DefaultColor: empty color string must be replaced with the
// default teal colour before calling the store.
func TestCreateCollection_DefaultColor(t *testing.T) {
	uid := uuid.New()
	created := &domain.Collection{ID: uuid.New(), UserID: uid, Title: "My", Color: "#0E7C7B"}
	store := &mockStore{insertCollection: created}
	svc := &Service{Store: store}

	col, err := svc.CreateCollection(context.Background(), uid, "My", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if col.Color != "#0E7C7B" {
		t.Errorf("expected default color #0E7C7B, got %q", col.Color)
	}
}

// UpdateCollection_Forbidden: a collection belonging to another user must
// return ErrForbidden.
func TestUpdateCollection_Forbidden(t *testing.T) {
	ownerID := uuid.New()
	callerID := uuid.New()
	colID := uuid.New()

	store := &mockStore{
		getCollection: &domain.Collection{ID: colID, UserID: ownerID, Title: "X"},
	}
	svc := &Service{Store: store}

	_, err := svc.UpdateCollection(context.Background(), callerID, colID, "Y", "")
	if err != domain.ErrForbidden {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

// DeleteCollection_Forbidden: deleting another user's collection must return
// ErrForbidden.
func TestDeleteCollection_Forbidden(t *testing.T) {
	ownerID := uuid.New()
	colID := uuid.New()
	store := &mockStore{
		getCollection: &domain.Collection{ID: colID, UserID: ownerID},
	}
	svc := &Service{Store: store}

	err := svc.DeleteCollection(context.Background(), uuid.New(), colID)
	if err != domain.ErrForbidden {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

// DeleteCollection_Default: deleting a default collection must return
// ErrBadRequest.
func TestDeleteCollection_Default(t *testing.T) {
	uid := uuid.New()
	colID := uuid.New()
	store := &mockStore{
		getCollection: &domain.Collection{ID: colID, UserID: uid, IsDefault: true},
	}
	svc := &Service{Store: store}

	err := svc.DeleteCollection(context.Background(), uid, colID)
	if err != domain.ErrBadRequest {
		t.Errorf("expected ErrBadRequest, got %v", err)
	}
}

// AddFavorite_InvalidTargetType: an unknown TargetType must return ErrBadRequest.
func TestAddFavorite_InvalidTargetType(t *testing.T) {
	svc := &Service{Store: &mockStore{}}
	err := svc.AddFavorite(context.Background(), uuid.New(), "unknown", uuid.New(), nil, nil)
	if err != domain.ErrBadRequest {
		t.Errorf("expected ErrBadRequest, got %v", err)
	}
}
