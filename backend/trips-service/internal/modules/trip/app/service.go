package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ilpaka/landmark_app/backend/trips-service/internal/modules/trip/domain"
	"github.com/ilpaka/landmark_app/backend/trips-service/internal/modules/trip/ports"
)

type Service struct {
	Store   ports.Store
	OSRMUrl string // e.g. "http://router.project-osrm.org"
}

func New(s ports.Store, osrmURL string) *Service { return &Service{Store: s, OSRMUrl: osrmURL} }

func (s *Service) CreateDraft(ctx context.Context, ownerID uuid.UUID, title string) (*domain.Trip, error) {
	if title == "" {
		return nil, domain.ErrBadRequest
	}
	t := domain.Trip{
		ID:        uuid.New(),
		OwnerID:   ownerID,
		Title:     title,
		Status:    domain.TripDraft,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	return s.Store.InsertTrip(ctx, t)
}

func (s *Service) List(ctx context.Context, ownerID uuid.UUID, status string, cursor string, limit int) ([]domain.Trip, string, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.Store.ListTrips(ctx, ownerID, status, cursor, limit)
}

func (s *Service) Get(ctx context.Context, ownerID, tripID uuid.UUID) (*domain.Trip, error) {
	t, err := s.Store.GetTrip(ctx, tripID)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, domain.ErrNotFound
	}
	if t.OwnerID != ownerID {
		return nil, domain.ErrForbidden
	}
	return t, nil
}

func (s *Service) Update(ctx context.Context, ownerID, tripID uuid.UUID, updates map[string]any) (*domain.Trip, error) {
	if _, err := s.Get(ctx, ownerID, tripID); err != nil {
		return nil, err
	}
	return s.Store.UpdateTrip(ctx, tripID, updates)
}

func (s *Service) Delete(ctx context.Context, ownerID, tripID uuid.UUID) error {
	if _, err := s.Get(ctx, ownerID, tripID); err != nil {
		return err
	}
	return s.Store.SetTripStatus(ctx, tripID, domain.TripArchived)
}

func (s *Service) AddStop(ctx context.Context, ownerID, tripID uuid.UUID, stop domain.TripStop) (*domain.TripStop, error) {
	if _, err := s.Get(ctx, ownerID, tripID); err != nil {
		return nil, err
	}
	stop.ID = uuid.New()
	stop.TripID = tripID
	stop.CreatedAt = time.Now()
	return s.Store.InsertStop(ctx, stop)
}

func (s *Service) ListStops(ctx context.Context, ownerID, tripID uuid.UUID) ([]domain.TripStop, error) {
	if _, err := s.Get(ctx, ownerID, tripID); err != nil {
		return nil, err
	}
	return s.Store.ListStops(ctx, tripID)
}

func (s *Service) MarkVisited(ctx context.Context, ownerID uuid.UUID, stopID uuid.UUID, at time.Time) error {
	return s.Store.MarkStopVisited(ctx, stopID, at)
}

func (s *Service) ListPublic(ctx context.Context, excludeOwnerID uuid.UUID, cursor string, limit int) ([]domain.Trip, string, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.Store.ListPublicTrips(ctx, excludeOwnerID, cursor, limit)
}

type LatLng struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

func (s *Service) GetRoute(ctx context.Context, ownerID, tripID uuid.UUID) ([]LatLng, error) {
	stops, err := s.Store.ListStops(ctx, tripID)
	if err != nil {
		return nil, err
	}
	if len(stops) < 2 {
		coords := make([]LatLng, len(stops))
		for i, st := range stops {
			coords[i] = LatLng{Lat: st.Latitude, Lng: st.Longitude}
		}
		return coords, nil
	}

	coordParts := make([]string, len(stops))
	for i, st := range stops {
		coordParts[i] = fmt.Sprintf("%f,%f", st.Longitude, st.Latitude)
	}
	osrmBase := s.OSRMUrl
	if osrmBase == "" {
		osrmBase = "http://router.project-osrm.org"
	}
	url := fmt.Sprintf("%s/route/v1/driving/%s?overview=full&geometries=geojson",
		osrmBase, strings.Join(coordParts, ";"))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// Fallback: straight-line waypoints
		coords := make([]LatLng, len(stops))
		for i, st := range stops {
			coords[i] = LatLng{Lat: st.Latitude, Lng: st.Longitude}
		}
		return coords, nil
	}
	defer resp.Body.Close()

	var osrmResp struct {
		Routes []struct {
			Geometry struct {
				Coordinates [][]float64 `json:"coordinates"`
			} `json:"geometry"`
		} `json:"routes"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&osrmResp); err != nil || len(osrmResp.Routes) == 0 {
		coords := make([]LatLng, len(stops))
		for i, st := range stops {
			coords[i] = LatLng{Lat: st.Latitude, Lng: st.Longitude}
		}
		return coords, nil
	}

	coords := osrmResp.Routes[0].Geometry.Coordinates
	result := make([]LatLng, len(coords))
	for i, c := range coords {
		result[i] = LatLng{Lng: c[0], Lat: c[1]}
	}
	return result, nil
}
