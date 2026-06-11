package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/ilpaka/landmark_app/backend/places-service/internal/modules/place/app"
	"github.com/ilpaka/landmark_app/backend/places-service/internal/modules/place/domain"
)

// ---------------------------------------------------------------------------
// Hand-rolled mock store
// ---------------------------------------------------------------------------

type mockStore struct {
	// places keyed by id
	places map[uuid.UUID]*domain.Place

	// control knobs
	insertPlaceErr    error
	updatePlaceErr    error
	setPlaceStatusErr error
	insertOutboxErr   error

	// categories
	categories        []domain.Category
	insertCategoryErr error
}

func newMockStore() *mockStore {
	return &mockStore{places: make(map[uuid.UUID]*domain.Place)}
}

func (m *mockStore) ListCategories(_ context.Context) ([]domain.Category, error) {
	return m.categories, nil
}

func (m *mockStore) GetCategoryBySlug(_ context.Context, slug string) (*domain.Category, error) {
	for i := range m.categories {
		if m.categories[i].Slug == slug {
			return &m.categories[i], nil
		}
	}
	return nil, nil
}

func (m *mockStore) InsertCategory(_ context.Context, c domain.Category) (*domain.Category, error) {
	if m.insertCategoryErr != nil {
		return nil, m.insertCategoryErr
	}
	m.categories = append(m.categories, c)
	return &c, nil
}

func (m *mockStore) UpdateCategory(_ context.Context, id uuid.UUID, title, icon, color string, sortOrder int) (*domain.Category, error) {
	for i := range m.categories {
		if m.categories[i].ID == id {
			m.categories[i].Title = title
			m.categories[i].Icon = icon
			m.categories[i].Color = color
			m.categories[i].SortOrder = sortOrder
			return &m.categories[i], nil
		}
	}
	return nil, domain.ErrNotFound
}

func (m *mockStore) DeactivateCategory(_ context.Context, id uuid.UUID) error {
	for i := range m.categories {
		if m.categories[i].ID == id {
			m.categories[i].Active = false
			return nil
		}
	}
	return domain.ErrNotFound
}

func (m *mockStore) ListPlaces(_ context.Context, f domain.ListFilter) ([]domain.Place, error) {
	var out []domain.Place
	for _, p := range m.places {
		out = append(out, *p)
	}
	return out, nil
}

func (m *mockStore) GetPlace(_ context.Context, id uuid.UUID) (*domain.Place, error) {
	p, ok := m.places[id]
	if !ok {
		return nil, nil
	}
	return p, nil
}

func (m *mockStore) InsertPlace(_ context.Context, p domain.Place) (*domain.Place, error) {
	if m.insertPlaceErr != nil {
		return nil, m.insertPlaceErr
	}
	m.places[p.ID] = &p
	return &p, nil
}

func (m *mockStore) UpdatePlace(_ context.Context, id uuid.UUID, updates map[string]any) (*domain.Place, error) {
	if m.updatePlaceErr != nil {
		return nil, m.updatePlaceErr
	}
	p, ok := m.places[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	if v, ok := updates["title"]; ok {
		p.Title = v.(string)
	}
	return p, nil
}

func (m *mockStore) SetPlaceStatus(_ context.Context, id uuid.UUID, status domain.PlaceStatus, extra map[string]any) error {
	if m.setPlaceStatusErr != nil {
		return m.setPlaceStatusErr
	}
	p, ok := m.places[id]
	if !ok {
		return domain.ErrNotFound
	}
	p.Status = status
	return nil
}

func (m *mockStore) SetPlaceCategories(_ context.Context, placeID uuid.UUID, categoryIDs []uuid.UUID) error {
	return nil
}

func (m *mockStore) InsertOutbox(_ context.Context, aggType, evType string, payload map[string]any) error {
	return m.insertOutboxErr
}

// ---------------------------------------------------------------------------
// Helper
// ---------------------------------------------------------------------------

func newServiceWithStore(st *mockStore) *app.Service {
	return app.New(st)
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

// TestService_GetPlace_NotFound verifies that asking for a non-existent id
// returns ErrNotFound.
func TestService_GetPlace_NotFound(t *testing.T) {
	svc := newServiceWithStore(newMockStore())

	_, err := svc.GetPlace(context.Background(), uuid.New(), nil, "")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

// TestService_GetPlace_HiddenFromGuest verifies that a non-published place is
// invisible to unauthenticated callers (userID == nil).
func TestService_GetPlace_HiddenFromGuest(t *testing.T) {
	st := newMockStore()
	authorID := uuid.New()
	p := &domain.Place{
		ID:       uuid.New(),
		Title:    "Draft place",
		Status:   domain.StatusDraft,
		AuthorID: &authorID,
	}
	st.places[p.ID] = p

	svc := newServiceWithStore(st)
	_, err := svc.GetPlace(context.Background(), p.ID, nil, "")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound for guest viewing draft, got %v", err)
	}
}

// TestService_GetPlace_OwnerCanSeeOwnDraft verifies that the author of a draft
// can still retrieve it.
func TestService_GetPlace_OwnerCanSeeOwnDraft(t *testing.T) {
	st := newMockStore()
	authorID := uuid.New()
	p := &domain.Place{
		ID:       uuid.New(),
		Title:    "Draft place",
		Status:   domain.StatusDraft,
		AuthorID: &authorID,
	}
	st.places[p.ID] = p

	svc := newServiceWithStore(st)
	got, err := svc.GetPlace(context.Background(), p.ID, &authorID, "user")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got.ID != p.ID {
		t.Fatalf("returned wrong place")
	}
}

// TestService_CreateDraft_EmptyTitle verifies that an empty title is rejected.
func TestService_CreateDraft_EmptyTitle(t *testing.T) {
	svc := newServiceWithStore(newMockStore())

	_, err := svc.CreateDraft(context.Background(), uuid.New(), "", 0, 0)
	if !errors.Is(err, domain.ErrBadRequest) {
		t.Fatalf("expected ErrBadRequest, got %v", err)
	}
}

// TestService_CreateDraft_Success verifies the happy path.
func TestService_CreateDraft_Success(t *testing.T) {
	svc := newServiceWithStore(newMockStore())
	userID := uuid.New()

	p, err := svc.CreateDraft(context.Background(), userID, "My Place", 48.8, 2.3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Title != "My Place" {
		t.Fatalf("wrong title: %s", p.Title)
	}
	if p.Status != domain.StatusDraft {
		t.Fatalf("expected StatusDraft, got %s", p.Status)
	}
	if p.AuthorID == nil || *p.AuthorID != userID {
		t.Fatalf("wrong author id")
	}
}

// TestService_CreatePlace_PrivatePersistsDetails verifies that a private place
// is published immediately and keeps description/city/country from the form.
func TestService_CreatePlace_PrivatePersistsDetails(t *testing.T) {
	svc := newServiceWithStore(newMockStore())
	userID := uuid.New()
	city := "Москва"
	country := "Россия"

	p, err := svc.CreatePlace(context.Background(), userID, domain.NewPlace{
		Title:       "Секретное место",
		Description: "только моё",
		Latitude:    55.75,
		Longitude:   37.61,
		City:        &city,
		Country:     &country,
		Visibility:  domain.VisibilityPrivate,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Status != domain.StatusPublished || p.PublishedAt == nil {
		t.Fatalf("private place must be published immediately, got %s", p.Status)
	}
	if p.Description != "только моё" {
		t.Fatalf("description lost: %q", p.Description)
	}
	if p.City == nil || *p.City != city {
		t.Fatalf("city lost: %v", p.City)
	}
	if p.Country == nil || *p.Country != country {
		t.Fatalf("country lost: %v", p.Country)
	}
}

// TestService_Submit_Forbidden verifies that a user cannot submit someone else's place.
func TestService_Submit_Forbidden(t *testing.T) {
	st := newMockStore()
	authorID := uuid.New()
	otherUserID := uuid.New()

	p := &domain.Place{
		ID:       uuid.New(),
		Title:    "A place",
		Status:   domain.StatusDraft,
		AuthorID: &authorID,
	}
	st.places[p.ID] = p

	svc := newServiceWithStore(st)
	err := svc.Submit(context.Background(), otherUserID, p.ID)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

// TestService_Submit_NotFound verifies that submitting a non-existent place returns ErrNotFound.
func TestService_Submit_NotFound(t *testing.T) {
	svc := newServiceWithStore(newMockStore())
	err := svc.Submit(context.Background(), uuid.New(), uuid.New())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

// TestService_Submit_Success verifies that the author can submit their own draft.
func TestService_Submit_Success(t *testing.T) {
	st := newMockStore()
	authorID := uuid.New()

	p := &domain.Place{
		ID:       uuid.New(),
		Title:    "A place",
		Status:   domain.StatusDraft,
		AuthorID: &authorID,
	}
	st.places[p.ID] = p

	svc := newServiceWithStore(st)
	err := svc.Submit(context.Background(), authorID, p.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if st.places[p.ID].Status != domain.StatusPendingModeration {
		t.Fatalf("expected status PendingModeration, got %s", st.places[p.ID].Status)
	}
}

// TestService_ApprovePlace_RequiresInternalFlag verifies that ApprovePlace
// calls SetPlaceStatus with StatusPublished (i.e., only an admin path should
// reach it). Here we exercise the happy path and confirm the status transition.
func TestService_ApprovePlace_RequiresInternalFlag(t *testing.T) {
	st := newMockStore()
	authorID := uuid.New()
	adminID := uuid.New()

	p := &domain.Place{
		ID:       uuid.New(),
		Title:    "Pending place",
		Status:   domain.StatusPendingModeration,
		AuthorID: &authorID,
	}
	st.places[p.ID] = p

	svc := newServiceWithStore(st)
	err := svc.ApprovePlace(context.Background(), p.ID, adminID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if st.places[p.ID].Status != domain.StatusPublished {
		t.Fatalf("expected StatusPublished after approval, got %s", st.places[p.ID].Status)
	}
}

// TestService_ApprovePlace_StoreError propagates a store error correctly.
func TestService_ApprovePlace_StoreError(t *testing.T) {
	st := newMockStore()
	st.setPlaceStatusErr = errors.New("db error")

	svc := newServiceWithStore(st)
	err := svc.ApprovePlace(context.Background(), uuid.New(), uuid.New())
	if err == nil {
		t.Fatal("expected error from store, got nil")
	}
}

// TestService_UpdateDraft_Forbidden ensures a user cannot edit another user's draft.
func TestService_UpdateDraft_Forbidden(t *testing.T) {
	st := newMockStore()
	authorID := uuid.New()
	otherID := uuid.New()

	p := &domain.Place{
		ID:       uuid.New(),
		Title:    "Draft",
		Status:   domain.StatusDraft,
		AuthorID: &authorID,
	}
	st.places[p.ID] = p

	svc := newServiceWithStore(st)
	_, err := svc.UpdateDraft(context.Background(), otherID, p.ID, map[string]any{"title": "hacked"})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}
