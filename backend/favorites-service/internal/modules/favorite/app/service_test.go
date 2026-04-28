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
	lastUpsertFav domain.Favorite
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
