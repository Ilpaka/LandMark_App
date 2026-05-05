package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ilpaka/landmark_app/backend/places-service/internal/modules/place/domain"
	"github.com/ilpaka/landmark_app/backend/places-service/internal/modules/place/ports"
)

type Service struct {
	Store      ports.Store
	Moderation ports.ModerationClient
}

func New(store ports.Store) *Service { return &Service{Store: store} }

// NewWithModeration собирает сервис с клиентом модерации, чтобы Submit мог
// постaвить место в очередь админ-панели.
func NewWithModeration(store ports.Store, mod ports.ModerationClient) *Service {
	return &Service{Store: store, Moderation: mod}
}

func (s *Service) ListCategories(ctx context.Context) ([]domain.Category, error) {
	return s.Store.ListCategories(ctx)
}

func (s *Service) ListPlaces(ctx context.Context, f domain.ListFilter) ([]domain.Place, error) {
	if f.Limit <= 0 || f.Limit > 200 {
		f.Limit = 200
	}
	return s.Store.ListPlaces(ctx, f)
}

func (s *Service) GetPlace(ctx context.Context, id uuid.UUID, userID *uuid.UUID, role string) (*domain.Place, error) {
	p, err := s.Store.GetPlace(ctx, id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, domain.ErrNotFound
	}
	isOwner := userID != nil && p.AuthorID != nil && *p.AuthorID == *userID
	// Приватные — только владельцу (модерации не подлежат, админу не видны).
	if p.Visibility == domain.VisibilityPrivate {
		if !isOwner {
			return nil, domain.ErrNotFound
		}
		return p, nil
	}
	// Публичные неопубликованные — владельцу или админу.
	if p.Status != domain.StatusPublished {
		if !isOwner && role != "admin" {
			return nil, domain.ErrNotFound
		}
	}
	return p, nil
}

func (s *Service) CreateDraft(ctx context.Context, userID uuid.UUID, title string, lat, lng float64) (*domain.Place, error) {
	return s.CreatePlace(ctx, userID, title, lat, lng, domain.VisibilityPublic)
}

// CreatePlace создаёт место с явной видимостью.
//   - public: status=draft, дальше нужно дернуть Submit → модерация → published.
//   - private: status сразу = published, published_at = now(), показывается
//     только автору (фильтрация в ListPlaces по ViewerID).
func (s *Service) CreatePlace(ctx context.Context, userID uuid.UUID, title string, lat, lng float64, visibility domain.PlaceVisibility) (*domain.Place, error) {
	if title == "" {
		return nil, domain.ErrBadRequest
	}
	if visibility != domain.VisibilityPublic && visibility != domain.VisibilityPrivate {
		return nil, domain.ErrBadRequest
	}
	now := time.Now()
	p := domain.Place{
		ID:         uuid.New(),
		Title:      title,
		Latitude:   lat,
		Longitude:  lng,
		AuthorID:   &userID,
		Source:     "user",
		Visibility: visibility,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if visibility == domain.VisibilityPrivate {
		p.Status = domain.StatusPublished
		p.PublishedAt = &now
	} else {
		p.Status = domain.StatusDraft
	}
	return s.Store.InsertPlace(ctx, p)
}

func (s *Service) UpdateDraft(ctx context.Context, userID, placeID uuid.UUID, updates map[string]any) (*domain.Place, error) {
	p, err := s.Store.GetPlace(ctx, placeID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, domain.ErrNotFound
	}
	if p.AuthorID == nil || *p.AuthorID != userID {
		return nil, domain.ErrForbidden
	}
	if p.Status != domain.StatusDraft {
		return nil, domain.ErrConflict
	}
	return s.Store.UpdatePlace(ctx, placeID, updates)
}

func (s *Service) Submit(ctx context.Context, userID, placeID uuid.UUID) error {
	p, err := s.Store.GetPlace(ctx, placeID)
	if err != nil {
		return err
	}
	if p == nil {
		return domain.ErrNotFound
	}
	if p.AuthorID == nil || *p.AuthorID != userID {
		return domain.ErrForbidden
	}
	// Приватные места модерации не подлежат — они видны только автору.
	if p.Visibility == domain.VisibilityPrivate {
		return domain.ErrBadRequest
	}
	if p.Status != domain.StatusDraft {
		return domain.ErrConflict
	}
	if err := s.Store.SetPlaceStatus(ctx, placeID, domain.StatusPendingModeration, nil); err != nil {
		return err
	}
	// Outbox — для аудита/будущего async-publisher'а.
	_ = s.Store.InsertOutbox(ctx, "place", "place_submitted", map[string]any{
		"place_id":     placeID.String(),
		"author_id":    userID.String(),
		"submitted_at": time.Now(),
	})
	// Синхронно ставим в очередь moderation-service. Без этого админ-панель
	// никогда не увидит место. Идемпотентно на стороне moderation.
	if s.Moderation != nil {
		if err := s.Moderation.SubmitPlace(ctx, placeID, userID); err != nil {
			// Откатываем статус — иначе место «застрянет» в pending без записи
			// в очереди модерации.
			_ = s.Store.SetPlaceStatus(ctx, placeID, domain.StatusDraft, nil)
			return err
		}
	}
	return nil
}

func (s *Service) ApprovePlace(ctx context.Context, placeID, adminID uuid.UUID) error {
	now := time.Now()
	if err := s.Store.SetPlaceStatus(ctx, placeID, domain.StatusPublished, map[string]any{"published_at": now}); err != nil {
		return err
	}
	p, _ := s.Store.GetPlace(ctx, placeID)
	if p != nil && p.AuthorID != nil {
		_ = s.Store.InsertOutbox(ctx, "place", "place_published", map[string]any{
			"place_id":     placeID.String(),
			"author_id":    p.AuthorID.String(),
			"published_at": now,
		})
	}
	return nil
}

func (s *Service) RejectPlace(ctx context.Context, placeID, adminID uuid.UUID, reason string) error {
	if err := s.Store.SetPlaceStatus(ctx, placeID, domain.StatusRejected, map[string]any{"reject_reason": reason}); err != nil {
		return err
	}
	p, _ := s.Store.GetPlace(ctx, placeID)
	if p != nil && p.AuthorID != nil {
		_ = s.Store.InsertOutbox(ctx, "place", "place_rejected", map[string]any{
			"place_id":  placeID.String(),
			"author_id": p.AuthorID.String(),
			"reason":    reason,
		})
	}
	return nil
}

func (s *Service) CreateCategory(ctx context.Context, slug, title, icon, color string, sortOrder int) (*domain.Category, error) {
	c := domain.Category{
		ID:        uuid.New(),
		Slug:      slug,
		Title:     title,
		Icon:      icon,
		Color:     color,
		SortOrder: sortOrder,
		Active:    true,
		CreatedAt: time.Now(),
	}
	return s.Store.InsertCategory(ctx, c)
}

func (s *Service) UpdateCategory(ctx context.Context, id uuid.UUID, title, icon, color string, sortOrder int) (*domain.Category, error) {
	return s.Store.UpdateCategory(ctx, id, title, icon, color, sortOrder)
}

func (s *Service) DeactivateCategory(ctx context.Context, id uuid.UUID) error {
	return s.Store.DeactivateCategory(ctx, id)
}
