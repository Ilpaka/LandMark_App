package app_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/ilpaka/landmark_app/backend/profile-service/internal/modules/profile/app"
	"github.com/ilpaka/landmark_app/backend/profile-service/internal/modules/profile/domain"
)

// ---------------------------------------------------------------------------
// Hand-rolled mock store
// ---------------------------------------------------------------------------

type mockStore struct {
	profiles map[uuid.UUID]*domain.Profile
	privacy  map[uuid.UUID]*domain.PrivacySettings

	getProfileErr    error
	upsertProfileErr error
	nicknameExists   bool
	nicknameExistsErr error
	getPrivacyErr    error
	upsertPrivacyErr error
}

func newMockStore() *mockStore {
	return &mockStore{
		profiles: make(map[uuid.UUID]*domain.Profile),
		privacy:  make(map[uuid.UUID]*domain.PrivacySettings),
	}
}

func (m *mockStore) GetProfile(_ context.Context, userID uuid.UUID) (*domain.Profile, error) {
	if m.getProfileErr != nil {
		return nil, m.getProfileErr
	}
	p, ok := m.profiles[userID]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return p, nil
}

func (m *mockStore) UpsertProfile(_ context.Context, p domain.Profile) (*domain.Profile, error) {
	if m.upsertProfileErr != nil {
		return nil, m.upsertProfileErr
	}
	m.profiles[p.UserID] = &p
	return &p, nil
}

func (m *mockStore) NicknameExists(_ context.Context, _ string, _ uuid.UUID) (bool, error) {
	return m.nicknameExists, m.nicknameExistsErr
}

func (m *mockStore) GetPrivacy(_ context.Context, userID uuid.UUID) (*domain.PrivacySettings, error) {
	if m.getPrivacyErr != nil {
		return nil, m.getPrivacyErr
	}
	priv, ok := m.privacy[userID]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return priv, nil
}

func (m *mockStore) UpsertPrivacy(_ context.Context, s domain.PrivacySettings) (*domain.PrivacySettings, error) {
	if m.upsertPrivacyErr != nil {
		return nil, m.upsertPrivacyErr
	}
	m.privacy[s.UserID] = &s
	return &s, nil
}

func newSvc(st *mockStore) *app.Service {
	return &app.Service{Store: st}
}

// ---------------------------------------------------------------------------
// GetProfile tests
// ---------------------------------------------------------------------------

func TestGetProfile_Existing(t *testing.T) {
	st := newMockStore()
	uid := uuid.New()
	st.profiles[uid] = &domain.Profile{UserID: uid, Nickname: "alice", DisplayName: "Alice"}

	p, err := newSvc(st).GetProfile(context.Background(), uid)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Nickname != "alice" {
		t.Errorf("got nickname %q, want %q", p.Nickname, "alice")
	}
}

func TestGetProfile_NotFound(t *testing.T) {
	st := newMockStore()
	_, err := newSvc(st).GetProfile(context.Background(), uuid.New())
	if err != domain.ErrNotFound {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// UpdateProfile tests
// ---------------------------------------------------------------------------

func TestUpdateProfile_NewUser_Bootstrap(t *testing.T) {
	st := newMockStore()
	uid := uuid.New()
	name := "Bob"

	p, err := newSvc(st).UpdateProfile(context.Background(), uid, domain.UpdateProfileInput{DisplayName: &name})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.DisplayName != "Bob" {
		t.Errorf("got %q, want %q", p.DisplayName, "Bob")
	}
}

func TestUpdateProfile_NicknameConflict(t *testing.T) {
	st := newMockStore()
	st.nicknameExists = true
	uid := uuid.New()
	nick := "taken"

	_, err := newSvc(st).UpdateProfile(context.Background(), uid, domain.UpdateProfileInput{Nickname: &nick})
	if err != domain.ErrConflict {
		t.Fatalf("want ErrConflict, got %v", err)
	}
}

func TestUpdateProfile_EmptyNickname_BadRequest(t *testing.T) {
	st := newMockStore()
	uid := uuid.New()
	empty := "   "

	_, err := newSvc(st).UpdateProfile(context.Background(), uid, domain.UpdateProfileInput{Nickname: &empty})
	if err != domain.ErrBadRequest {
		t.Fatalf("want ErrBadRequest, got %v", err)
	}
}

func TestUpdateProfile_EmptyDisplayName_BadRequest(t *testing.T) {
	st := newMockStore()
	uid := uuid.New()
	empty := ""

	_, err := newSvc(st).UpdateProfile(context.Background(), uid, domain.UpdateProfileInput{DisplayName: &empty})
	if err != domain.ErrBadRequest {
		t.Fatalf("want ErrBadRequest, got %v", err)
	}
}

func TestUpdateProfile_StoreError(t *testing.T) {
	st := newMockStore()
	uid := uuid.New()
	st.upsertProfileErr = domain.ErrConflict
	name := "Bob"

	_, err := newSvc(st).UpdateProfile(context.Background(), uid, domain.UpdateProfileInput{DisplayName: &name})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ---------------------------------------------------------------------------
// GetPrivacy tests
// ---------------------------------------------------------------------------

func TestGetPrivacy_Defaults_WhenNotFound(t *testing.T) {
	st := newMockStore()
	uid := uuid.New()

	priv, err := newSvc(st).GetPrivacy(context.Background(), uid)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if priv.ProfileVisibility != "public" {
		t.Errorf("got %q, want public", priv.ProfileVisibility)
	}
}

func TestGetPrivacy_Existing(t *testing.T) {
	st := newMockStore()
	uid := uuid.New()
	st.privacy[uid] = &domain.PrivacySettings{UserID: uid, ProfileVisibility: "private"}

	priv, err := newSvc(st).GetPrivacy(context.Background(), uid)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if priv.ProfileVisibility != "private" {
		t.Errorf("got %q, want private", priv.ProfileVisibility)
	}
}

// ---------------------------------------------------------------------------
// UpdatePrivacy tests
// ---------------------------------------------------------------------------

func TestUpdatePrivacy_PatchField(t *testing.T) {
	st := newMockStore()
	uid := uuid.New()
	vis := "friends"

	priv, err := newSvc(st).UpdatePrivacy(context.Background(), uid, domain.UpdatePrivacyInput{TripsVisibility: &vis})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if priv.TripsVisibility != "friends" {
		t.Errorf("got %q, want friends", priv.TripsVisibility)
	}
}

func TestUpdatePrivacy_UpsertError(t *testing.T) {
	st := newMockStore()
	uid := uuid.New()
	st.upsertPrivacyErr = domain.ErrNotFound
	vis := "private"

	_, err := newSvc(st).UpdatePrivacy(context.Background(), uid, domain.UpdatePrivacyInput{ProfileVisibility: &vis})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
