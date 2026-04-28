package app_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/ilpaka/landmark_app/backend/moderation-service/internal/modules/moderation/app"
	"github.com/ilpaka/landmark_app/backend/moderation-service/internal/modules/moderation/domain"
)

// ---------------------------------------------------------------------------
// Hand-rolled mocks
// ---------------------------------------------------------------------------

type mockStore struct {
	items map[uuid.UUID]*domain.QueueItem

	insertErr         error
	getItemErr        error
	getByTargetErr    error
	getByTargetItem   *domain.QueueItem
	listPendingResult *domain.QueuePage
	listPendingErr    error
	updateDecisionErr error
}

func newMockStore() *mockStore {
	return &mockStore{items: make(map[uuid.UUID]*domain.QueueItem)}
}

func (m *mockStore) InsertItem(_ context.Context, tt domain.TargetType, targetID, submittedBy uuid.UUID) (*domain.QueueItem, error) {
	if m.insertErr != nil {
		return nil, m.insertErr
	}
	item := &domain.QueueItem{
		ID:          uuid.New(),
		TargetType:  tt,
		TargetID:    targetID,
		SubmittedBy: submittedBy,
		Status:      domain.StatusPending,
	}
	m.items[item.ID] = item
	return item, nil
}

func (m *mockStore) GetItem(_ context.Context, id uuid.UUID) (*domain.QueueItem, error) {
	if m.getItemErr != nil {
		return nil, m.getItemErr
	}
	item, ok := m.items[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return item, nil
}

func (m *mockStore) GetItemByTarget(_ context.Context, _ domain.TargetType, _ uuid.UUID) (*domain.QueueItem, error) {
	if m.getByTargetErr != nil {
		return nil, m.getByTargetErr
	}
	return m.getByTargetItem, nil
}

func (m *mockStore) ListPending(_ context.Context, _ domain.ListFilter) (*domain.QueuePage, error) {
	if m.listPendingErr != nil {
		return nil, m.listPendingErr
	}
	if m.listPendingResult != nil {
		return m.listPendingResult, nil
	}
	return &domain.QueuePage{}, nil
}

func (m *mockStore) UpdateDecision(_ context.Context, id uuid.UUID, status domain.ItemStatus, moderatorID uuid.UUID, note *string) error {
	if m.updateDecisionErr != nil {
		return m.updateDecisionErr
	}
	if item, ok := m.items[id]; ok {
		item.Status = status
		item.ModeratorID = &moderatorID
		item.Note = note
	}
	return nil
}

type mockPlaces struct {
	approveErr error
	rejectErr  error
}

func (mp *mockPlaces) ApprovePlace(_ context.Context, _ uuid.UUID) error {
	return mp.approveErr
}

func (mp *mockPlaces) RejectPlace(_ context.Context, _ uuid.UUID, _ string) error {
	return mp.rejectErr
}

func newSvc(st *mockStore, pl *mockPlaces) *app.Service {
	return &app.Service{Store: st, Places: pl}
}

// ---------------------------------------------------------------------------
// SubmitItem tests
// ---------------------------------------------------------------------------

func TestSubmitItem_New(t *testing.T) {
	st := newMockStore()
	st.getByTargetErr = domain.ErrNotFound

	item, err := newSvc(st, &mockPlaces{}).SubmitItem(context.Background(), domain.TargetPlace, uuid.New(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Status != domain.StatusPending {
		t.Errorf("got status %q, want pending", item.Status)
	}
}

func TestSubmitItem_DuplicatePending_Conflict(t *testing.T) {
	st := newMockStore()
	existing := &domain.QueueItem{ID: uuid.New(), Status: domain.StatusPending}
	st.getByTargetItem = existing

	_, err := newSvc(st, &mockPlaces{}).SubmitItem(context.Background(), domain.TargetPlace, uuid.New(), uuid.New())
	if err != domain.ErrConflict {
		t.Fatalf("want ErrConflict, got %v", err)
	}
}

func TestSubmitItem_ResubmitRejected_OK(t *testing.T) {
	st := newMockStore()
	existing := &domain.QueueItem{ID: uuid.New(), Status: domain.StatusRejected}
	st.getByTargetItem = existing

	item, err := newSvc(st, &mockPlaces{}).SubmitItem(context.Background(), domain.TargetPlace, uuid.New(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Status != domain.StatusPending {
		t.Errorf("got status %q, want pending", item.Status)
	}
}

// ---------------------------------------------------------------------------
// ListPending tests
// ---------------------------------------------------------------------------

func TestListPending_DefaultLimit(t *testing.T) {
	st := newMockStore()
	_, err := newSvc(st, &mockPlaces{}).ListPending(context.Background(), domain.ListFilter{Limit: 0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Approve tests
// ---------------------------------------------------------------------------

func TestApprove_OK(t *testing.T) {
	st := newMockStore()
	pl := &mockPlaces{}
	svc := newSvc(st, pl)

	item, _ := st.InsertItem(context.Background(), domain.TargetPlace, uuid.New(), uuid.New())

	if err := svc.Approve(context.Background(), uuid.New(), item.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if st.items[item.ID].Status != domain.StatusApproved {
		t.Errorf("expected approved, got %q", st.items[item.ID].Status)
	}
}

func TestApprove_AlreadyApproved_BadRequest(t *testing.T) {
	st := newMockStore()
	item, _ := st.InsertItem(context.Background(), domain.TargetPlace, uuid.New(), uuid.New())
	st.items[item.ID].Status = domain.StatusApproved

	err := newSvc(st, &mockPlaces{}).Approve(context.Background(), uuid.New(), item.ID)
	if err != domain.ErrBadRequest {
		t.Fatalf("want ErrBadRequest, got %v", err)
	}
}

func TestApprove_PlacesClientError(t *testing.T) {
	st := newMockStore()
	pl := &mockPlaces{approveErr: domain.ErrNotFound}
	svc := newSvc(st, pl)

	item, _ := st.InsertItem(context.Background(), domain.TargetPlace, uuid.New(), uuid.New())

	err := svc.Approve(context.Background(), uuid.New(), item.ID)
	if err == nil {
		t.Fatal("expected error from places client, got nil")
	}
}

// ---------------------------------------------------------------------------
// Reject tests
// ---------------------------------------------------------------------------

func TestReject_OK(t *testing.T) {
	st := newMockStore()
	pl := &mockPlaces{}
	svc := newSvc(st, pl)

	item, _ := st.InsertItem(context.Background(), domain.TargetPlace, uuid.New(), uuid.New())

	if err := svc.Reject(context.Background(), uuid.New(), item.ID, "spam"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if st.items[item.ID].Status != domain.StatusRejected {
		t.Errorf("expected rejected, got %q", st.items[item.ID].Status)
	}
}

func TestReject_ItemNotFound(t *testing.T) {
	st := newMockStore()

	err := newSvc(st, &mockPlaces{}).Reject(context.Background(), uuid.New(), uuid.New(), "reason")
	if err != domain.ErrNotFound {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}
