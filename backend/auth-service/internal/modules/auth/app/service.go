package app

import (
	"log/slog"
	"net/mail"
	"strings"
	"time"

	redisx "github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/adapters/redis"
	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/crypto"
	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/domain"
	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/ports"
)

// Service orchestrates auth use cases.
type Service struct {
	Store      ports.Store
	PhoneOTP   ports.PhoneOTPChallengeStore
	Redis      *redisx.Client
	Hasher     *crypto.Argon2idHasher
	JWT        ports.JWTSigner
	Profile    ports.ProfileClient
	Mail       ports.EmailSender
	SMS        ports.SMSSender
	Pepper     string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
	VerifyTTL  time.Duration
	ResetTTL   time.Duration
	Log        *slog.Logger

	dummyHash string
}

func NewService(
	store ports.Store,
	phoneOTP ports.PhoneOTPChallengeStore,
	rdb *redisx.Client,
	hasher *crypto.Argon2idHasher,
	jwt ports.JWTSigner,
	profile ports.ProfileClient,
	mail ports.EmailSender,
	sms ports.SMSSender,
	pepper string,
	accessTTL, refreshTTL, verifyTTL, resetTTL time.Duration,
	log *slog.Logger,
) (*Service, error) {
	dh, err := hasher.Hash("argon2id-timing-placeholder-secret")
	if err != nil {
		return nil, err
	}
	return &Service{
		Store:      store,
		PhoneOTP:   phoneOTP,
		Redis:      rdb,
		Hasher:     hasher,
		JWT:        jwt,
		Profile:    profile,
		Mail:       mail,
		SMS:        sms,
		Pepper:     pepper,
		AccessTTL:  accessTTL,
		RefreshTTL: refreshTTL,
		VerifyTTL:  verifyTTL,
		ResetTTL:   resetTTL,
		Log:        log,
		dummyHash:  dh,
	}, nil
}

func normalizeEmail(email string) (string, string, error) {
	addr, err := mail.ParseAddress(strings.TrimSpace(email))
	if err != nil {
		return "", "", domain.ErrBadRequest
	}
	norm := strings.ToLower(addr.Address)
	return addr.Address, norm, nil
}

func ptrTime(t time.Time) *time.Time { return &t }

func domainFromEmail(norm string) string {
	i := strings.LastIndex(norm, "@")
	if i < 0 {
		return ""
	}
	return norm[i+1:]
}
