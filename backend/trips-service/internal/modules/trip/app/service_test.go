package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ilpaka/landmark_app/backend/trips-service/internal/modules/trip/app"
	"github.com/ilpaka/landmark_app/backend/trips-service/internal/modules/trip/domain"
)

// ---------------------------------------------------------------------------
// Hand-rolled mock store
// ---------------------------------------------------------------------------

type mockStore struct {
	trips map[uuid.UUID]*domain.Trip
	stops map[uuid.UUID]*domain.TripStop

	insertTripErr  error
	updateTripErr  error
	insertStopErr  error
}

func newMockStore() *mockStore {
	return &mockStore{
		trips: make(map[uuid.UUID]*domain.Trip),
		stops: make(map[uuid.UUID]*domain.TripStop),
	}
}

func (m *mockStore) InsertTrip(_ context.Context, t domain.Trip) (*domain.Trip, error) {
	if m.insertTripErr != nil {
		return nil, m.insertTripErr
	}
	m.trips[t.ID] = &t
	return &t, nil
}

func (m *mockStore) ListTrips(_ context.Context, ownerID uuid.UUID, status, cursor string, limit int) ([]domain.Trip, string, error) {
	var out []domain.Trip
	for _, t := range m.trips {
		if t.OwnerID == ownerID {
			out = append(out, *t)
		}
	}
	return out, "", nil
}

func (m *mockStore) GetTrip(_ context.Context, id uuid.UUID) (*domain.Trip, error) {
	t, ok := m.trips[id]
	if !ok {
		return nil, nil
	}
	return t, nil
}

func (m *mockStore) UpdateTrip(_ context.Context, id uuid.UUID, updates map[string]any) (*domain.Trip, error) {
	if m.updateTripErr != nil {
		return nil, m.updateTripErr
	}
	t, ok := m.trips[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	if v, ok := updates["title"]; ok {
		t.Title = v.(string)
	}
	return t, nil
}

func (m *mockStore) SetTripStatus(_ context.Context, id uuid.UUID, status domain.TripStatus) error {
	t, ok := m.trips[id]
	if !ok {
		return domain.ErrNotFound
	}
	t.Status = status
	return nil
}

func (m *mockStore) InsertStop(_ context.Context, s domain.TripStop) (*domain.TripStop, error) {
	if m.insertStopErr != nil {
		return nil, m.insertStopErr
	}
	m.stops[s.ID] = &s
	return &s, nil
}

func (m *mockStore) ListStops(_ context.Context, tripID uuid.UUID) ([]domain.TripStop, error) {
	var out []domain.TripStop
	for _, s := range m.stops {
		if s.TripID == tripID {
			out = append(out, *s)
		}
	}
	return out, nil
}

func (m *mockStore) UpdateStop(_ context.Context, id uuid.UUID, updates map[string]any) (*domain.TripStop, error) {
	s, ok := m.stops[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return s, nil
}

func (m *mockStore) DeleteStop(_ context.Context, id uuid.UUID) error {
	if _, ok := m.stops[id]; !ok {
		return domain.ErrNotFound
	}
	delete(m.stops, id)
	return nil
}

func (m *mockStore) MarkStopVisited(_ context.Context, id uuid.UUID, at time.Time) error {
	s, ok := m.stops[id]
	if !ok {
		return domain.ErrNotFound
	}
	s.VisitedAt = &at
	return nil
}

func (m *mockStore) ListPublicTrips(_ context.Context, _ uuid.UUID, _ string, _ int) ([]domain.Trip, string, error) {
	return []domain.Trip{}, "", nil
}

// ---------------------------------------------------------------------------
// Helper
// ---------------------------------------------------------------------------

func newService(st *mockStore) *app.Service {
	return app.New(st, "")
}

// seedTrip inserts a trip owned by ownerID into the store and returns it.
func seedTrip(st *mockStore, ownerID uuid.UUID, status domain.TripStatus) *domain.Trip {
	t := &domain.Trip{
		ID:        uuid.New(),
		OwnerID:   ownerID,
		Title:     "Test Trip",
		Status:    status,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	st.trips[t.ID] = t
	return t
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

// TestService_CreateTrip_Success verifies the happy-path for draft creation.
func TestService_CreateTrip_Success(t *testing.T) {
	svc := newService(newMockStore())
	ownerID := uuid.New()

	trip, err := svc.CreateDraft(context.Background(), ownerID, "Road Trip 2025")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if trip.Title != "Road Trip 2025" {
		t.Fatalf("wrong title: %s", trip.Title)
	}
	if trip.OwnerID != ownerID {
		t.Fatalf("wrong owner id")
	}
	if trip.Status != domain.TripDraft {
		t.Fatalf("expected TripDraft, got %s", trip.Status)
	}
}

// TestService_CreateTrip_EmptyTitle verifies that an empty title returns ErrBadRequest.
func TestService_CreateTrip_EmptyTitle(t *testing.T) {
	svc := newService(newMockStore())

	_, err := svc.CreateDraft(context.Background(), uuid.New(), "")
	if !errors.Is(err, domain.ErrBadRequest) {
		t.Fatalf("expected ErrBadRequest, got %v", err)
	}
}

// TestService_GetTrip_NotFound verifies that requesting a missing trip returns ErrNotFound.
func TestService_GetTrip_NotFound(t *testing.T) {
	svc := newService(newMockStore())
	ownerID := uuid.New()

	_, err := svc.Get(context.Background(), ownerID, uuid.New())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

// TestService_GetTrip_Success verifies that the owner can retrieve their own trip.
func TestService_GetTrip_Success(t *testing.T) {
	st := newMockStore()
	ownerID := uuid.New()
	trip := seedTrip(st, ownerID, domain.TripDraft)

	svc := newService(st)
	got, err := svc.Get(context.Background(), ownerID, trip.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != trip.ID {
		t.Fatalf("returned wrong trip")
	}
}

// TestService_UpdateTrip_Forbidden verifies that a different user cannot update someone else's trip.
func TestService_UpdateTrip_Forbidden(t *testing.T) {
	st := newMockStore()
	ownerID := uuid.New()
	otherUserID := uuid.New()
	trip := seedTrip(st, ownerID, domain.TripDraft)

	svc := newService(st)
	_, err := svc.Update(context.Background(), otherUserID, trip.ID, map[string]any{"title": "hacked"})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

// TestService_UpdateTrip_Success verifies the owner can update their trip.
func TestService_UpdateTrip_Success(t *testing.T) {
	st := newMockStore()
	ownerID := uuid.New()
	trip := seedTrip(st, ownerID, domain.TripDraft)

	svc := newService(st)
	updated, err := svc.Update(context.Background(), ownerID, trip.ID, map[string]any{"title": "New Title"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Title != "New Title" {
		t.Fatalf("title not updated, got: %s", updated.Title)
	}
}

// TestService_AddStop_TripNotFound verifies that adding a stop to a non-existent trip returns ErrNotFound.
func TestService_AddStop_TripNotFound(t *testing.T) {
	svc := newService(newMockStore())
	ownerID := uuid.New()

	stop := domain.TripStop{
		Title:     "Some Stop",
		Latitude:  48.8,
		Longitude: 2.3,
	}
	_, err := svc.AddStop(context.Background(), ownerID, uuid.New(), stop)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

// TestService_AddStop_Forbidden verifies that a non-owner cannot add stops.
func TestService_AddStop_Forbidden(t *testing.T) {
	st := newMockStore()
	ownerID := uuid.New()
	otherID := uuid.New()
	trip := seedTrip(st, ownerID, domain.TripDraft)

	svc := newService(st)
	stop := domain.TripStop{
		Title:     "Stop 1",
		Latitude:  48.8,
		Longitude: 2.3,
	}
	_, err := svc.AddStop(context.Background(), otherID, trip.ID, stop)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

// TestService_AddStop_Success verifies the owner can add a stop to their trip.
func TestService_AddStop_Success(t *testing.T) {
	st := newMockStore()
	ownerID := uuid.New()
	trip := seedTrip(st, ownerID, domain.TripDraft)

	svc := newService(st)
	stop := domain.TripStop{
		Title:     "Stop 1",
		Latitude:  48.8,
		Longitude: 2.3,
	}
	created, err := svc.AddStop(context.Background(), ownerID, trip.ID, stop)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.TripID != trip.ID {
		t.Fatalf("stop has wrong tripID")
	}
	if created.ID == uuid.Nil {
		t.Fatalf("stop ID was not assigned")
	}
}

// TestService_Delete_Forbidden verifies that a non-owner cannot delete a trip.
func TestService_Delete_Forbidden(t *testing.T) {
	st := newMockStore()
	ownerID := uuid.New()
	otherID := uuid.New()
	trip := seedTrip(st, ownerID, domain.TripDraft)

	svc := newService(st)
	err := svc.Delete(context.Background(), otherID, trip.ID)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

// TestService_Delete_Success verifies that the owner can delete (archive) their trip.
func TestService_Delete_Success(t *testing.T) {
	st := newMockStore()
	ownerID := uuid.New()
	trip := seedTrip(st, ownerID, domain.TripDraft)

	svc := newService(st)
	err := svc.Delete(context.Background(), ownerID, trip.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if st.trips[trip.ID].Status != domain.TripArchived {
		t.Fatalf("expected TripArchived, got %s", st.trips[trip.ID].Status)
	}
}
