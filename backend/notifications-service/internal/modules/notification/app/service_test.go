package app

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/ilpaka/landmark_app/backend/notifications-service/internal/modules/notification/domain"
)

// ---------------------------------------------------------------------------
// Hand-rolled mock store
// ---------------------------------------------------------------------------

type mockStore struct {
	// InsertNotification
	insertedNotification *domain.Notification
	insertErr            error

	// ListNotifications
	notificationsPage *domain.NotificationsPage
	listErr           error

	// UpsertDevice
	upsertedDevice *domain.Device
	upsertDeviceErr error

	// GetPreferences
	preferences    *domain.Preferences
	getPrefsErr    error

	// UpsertPreferences
	upsertPreferences    *domain.Preferences
	upsertPreferencesErr error

	// Capture args
	lastListFilter  domain.ListFilter
	lastUpsertDevice domain.Device
}

func (m *mockStore) InsertNotification(ctx context.Context, in domain.SendInput) (*domain.Notification, error) {
	return m.insertedNotification, m.insertErr
}

func (m *mockStore) ListNotifications(ctx context.Context, f domain.ListFilter) (*domain.NotificationsPage, error) {
	m.lastListFilter = f
	return m.notificationsPage, m.listErr
}

func (m *mockStore) MarkRead(ctx context.Context, id, userID uuid.UUID) error {
	return nil
}

func (m *mockStore) MarkAllRead(ctx context.Context, userID uuid.UUID) error {
	return nil
}

func (m *mockStore) UnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	return 0, nil
}

func (m *mockStore) UpsertDevice(ctx context.Context, d domain.Device) (*domain.Device, error) {
	m.lastUpsertDevice = d
	return m.upsertedDevice, m.upsertDeviceErr
}

func (m *mockStore) DeleteDevice(ctx context.Context, id, userID uuid.UUID) error {
	return nil
}

func (m *mockStore) GetPreferences(ctx context.Context, userID uuid.UUID) (*domain.Preferences, error) {
	return m.preferences, m.getPrefsErr
}

func (m *mockStore) UpsertPreferences(ctx context.Context, p domain.Preferences) (*domain.Preferences, error) {
	return m.upsertPreferences, m.upsertPreferencesErr
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

// Send_EmptyTitle: a SendInput with an empty Title must return ErrBadRequest
// without touching the store.
func TestSend_EmptyTitle(t *testing.T) {
	store := &mockStore{}
	svc := &Service{Store: store}

	_, err := svc.Send(context.Background(), domain.SendInput{
		UserID: uuid.New(),
		Type:   "promo",
		Title:  "",           // invalid
		Body:   "Some body",
	})
	if err != domain.ErrBadRequest {
		t.Errorf("expected ErrBadRequest, got %v", err)
	}
}

// Send_EmptyBody mirrors Send_EmptyTitle but with a missing body.
func TestSend_EmptyBody(t *testing.T) {
	store := &mockStore{}
	svc := &Service{Store: store}

	_, err := svc.Send(context.Background(), domain.SendInput{
		UserID: uuid.New(),
		Title:  "Hello",
		Body:   "", // invalid
	})
	if err != domain.ErrBadRequest {
		t.Errorf("expected ErrBadRequest, got %v", err)
	}
}

// Send_Success: a fully populated SendInput must be forwarded to the store and
// the resulting Notification returned.
func TestSend_Success(t *testing.T) {
	notif := &domain.Notification{
		ID:    uuid.New(),
		Title: "Hello",
		Body:  "World",
	}
	store := &mockStore{insertedNotification: notif}
	svc := &Service{Store: store}

	got, err := svc.Send(context.Background(), domain.SendInput{
		UserID: uuid.New(),
		Title:  "Hello",
		Body:   "World",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || got.ID != notif.ID {
		t.Errorf("expected notification ID %s, got %v", notif.ID, got)
	}
}

// List_DefaultLimit: a Limit of 0 must be normalised to 20 before the store
// is called.
func TestList_DefaultLimit(t *testing.T) {
	store := &mockStore{
		notificationsPage: &domain.NotificationsPage{Items: []domain.Notification{}},
	}
	svc := &Service{Store: store}

	_, err := svc.List(context.Background(), domain.ListFilter{
		UserID: uuid.New(),
		Limit:  0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.lastListFilter.Limit != 20 {
		t.Errorf("expected normalised Limit=20, got %d", store.lastListFilter.Limit)
	}
}

// UpsertDevice_InvalidPlatform: platforms other than "ios" / "android" must
// return ErrBadRequest.
func TestUpsertDevice_InvalidPlatform(t *testing.T) {
	store := &mockStore{}
	svc := &Service{Store: store}

	_, err := svc.UpsertDevice(context.Background(), uuid.New(), "windows", "tok123")
	if err != domain.ErrBadRequest {
		t.Errorf("expected ErrBadRequest for unknown platform, got %v", err)
	}
}

// MarkRead_Delegates: MarkRead must delegate to the store without modification.
func TestMarkRead_Delegates(t *testing.T) {
	svc := &Service{Store: &mockStore{}}
	if err := svc.MarkRead(context.Background(), uuid.New(), uuid.New()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// MarkAllRead_Delegates: MarkAllRead must delegate to the store.
func TestMarkAllRead_Delegates(t *testing.T) {
	svc := &Service{Store: &mockStore{}}
	if err := svc.MarkAllRead(context.Background(), uuid.New()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// UnreadCount_Delegates: UnreadCount must delegate to the store.
func TestUnreadCount_Delegates(t *testing.T) {
	svc := &Service{Store: &mockStore{}}
	n, err := svc.UnreadCount(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0, got %d", n)
	}
}

// UpsertDevice_ValidPlatforms: ios and android must pass through to the store.
func TestUpsertDevice_ValidPlatforms(t *testing.T) {
	dev := &domain.Device{ID: uuid.New(), Platform: "ios"}
	store := &mockStore{upsertedDevice: dev}
	svc := &Service{Store: store}

	for _, platform := range []string{"ios", "android"} {
		got, err := svc.UpsertDevice(context.Background(), uuid.New(), platform, "tok")
		if err != nil {
			t.Fatalf("platform %s: unexpected error: %v", platform, err)
		}
		if got == nil {
			t.Errorf("platform %s: expected non-nil device", platform)
		}
	}
}

// UpdatePreferences_PatchField: patches a single field and expects it reflected.
func TestUpdatePreferences_PatchField(t *testing.T) {
	uid := uuid.New()
	updated := &domain.Preferences{UserID: uid, PushMarketing: true}
	store := &mockStore{
		getPrefsErr:       domain.ErrNotFound, // triggers defaults
		upsertPreferences: updated,
	}
	svc := &Service{Store: store}

	prefs, err := svc.UpdatePreferences(context.Background(), uid, map[string]any{"push_marketing": true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prefs == nil || !prefs.PushMarketing {
		t.Errorf("expected PushMarketing=true, got %v", prefs)
	}
}

// GetPreferences_Defaults: when the store returns ErrNotFound the service
// must synthesise a default Preferences value rather than propagating the error.
func TestGetPreferences_Defaults(t *testing.T) {
	userID := uuid.New()
	store := &mockStore{getPrefsErr: domain.ErrNotFound}
	svc := &Service{Store: store}

	prefs, err := svc.GetPreferences(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prefs == nil {
		t.Fatal("expected non-nil default preferences")
	}
	if prefs.UserID != userID {
		t.Errorf("expected UserID %s, got %s", userID, prefs.UserID)
	}
	// Verify sensible defaults are set.
	if !prefs.PushTripReminders {
		t.Error("expected PushTripReminders=true in defaults")
	}
	if !prefs.EmailWelcome {
		t.Error("expected EmailWelcome=true in defaults")
	}
}
