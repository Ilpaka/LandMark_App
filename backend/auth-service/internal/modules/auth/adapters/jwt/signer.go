package jwt

import (
	"context"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"time"

	jose "github.com/go-jose/go-jose/v4"
	"github.com/google/uuid"

	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/domain"
	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/ports"
)

// Signer implements ports.JWTSigner with RS256.
type Signer struct {
	priv     *rsa.PrivateKey
	pub      *rsa.PublicKey
	keyID    string
	issuer   string
	audience string
}

func LoadSigner(privatePath, publicPath, kid, issuer, audience string) (*Signer, error) {
	privPEM, err := os.ReadFile(privatePath)
	if err != nil {
		return nil, fmt.Errorf("read private key: %w", err)
	}
	pubPEM, err := os.ReadFile(publicPath)
	if err != nil {
		return nil, fmt.Errorf("read public key: %w", err)
	}
	priv, err := parseRSAPrivateKey(privPEM)
	if err != nil {
		return nil, fmt.Errorf("parse private: %w", err)
	}
	pub, err := parseRSAPublicKey(pubPEM)
	if err != nil {
		return nil, fmt.Errorf("parse public: %w", err)
	}
	if kid == "" {
		sum := sha256.Sum256(pubPEM)
		kid = base64.RawURLEncoding.EncodeToString(sum[:8])
	}
	return &Signer{priv: priv, pub: pub, keyID: kid, issuer: issuer, audience: audience}, nil
}

func parseRSAPrivateKey(pemBytes []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("no pem block")
	}
	switch block.Type {
	case "RSA PRIVATE KEY":
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	case "PRIVATE KEY":
		k, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		pk, ok := k.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("not rsa private key")
		}
		return pk, nil
	default:
		return nil, fmt.Errorf("unsupported pem type %q", block.Type)
	}
}

func parseRSAPublicKey(pemBytes []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("no pem block")
	}
	switch block.Type {
	case "RSA PUBLIC KEY":
		return x509.ParsePKCS1PublicKey(block.Bytes)
	case "PUBLIC KEY":
		k, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		pk, ok := k.(*rsa.PublicKey)
		if !ok {
			return nil, errors.New("not rsa public key")
		}
		return pk, nil
	default:
		return nil, fmt.Errorf("unsupported pem type %q", block.Type)
	}
}

func (s *Signer) signerOpts() *jose.SignerOptions {
	opts := jose.SignerOptions{}
	opts.WithType("JWT")
	opts.WithHeader("kid", s.keyID)
	return &opts
}

func (s *Signer) SignAccess(ctx context.Context, userID, sessionID uuid.UUID, role domain.Role, ttl time.Duration) (token, jti string, exp time.Time, err error) {
	_ = ctx
	jti = uuid.NewString()
	now := time.Now().UTC()
	exp = now.Add(ttl)
	sig, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: s.priv}, s.signerOpts())
	if err != nil {
		return "", "", time.Time{}, err
	}
	cl := map[string]any{
		"sub":  userID.String(),
		"sid":  sessionID.String(),
		"role": string(role),
		"jti":  jti,
		"iat":  now.Unix(),
		"exp":  exp.Unix(),
		"typ":  "access",
	}
	if s.issuer != "" {
		cl["iss"] = s.issuer
	}
	if s.audience != "" {
		cl["aud"] = s.audience
	}
	raw, err := json.Marshal(cl)
	if err != nil {
		return "", "", time.Time{}, err
	}
	tok, err := sig.Sign(raw)
	if err != nil {
		return "", "", time.Time{}, err
	}
	compact, err := tok.CompactSerialize()
	if err != nil {
		return "", "", time.Time{}, err
	}
	return compact, jti, exp, nil
}

func (s *Signer) ParseAccess(ctx context.Context, token string) (ports.AccessClaims, error) {
	_ = ctx
	tok, err := jose.ParseSigned(token, []jose.SignatureAlgorithm{jose.RS256})
	if err != nil {
		return ports.AccessClaims{}, err
	}
	raw, err := tok.Verify(s.pub)
	if err != nil {
		return ports.AccessClaims{}, err
	}
	var cl map[string]any
	if err := json.Unmarshal(raw, &cl); err != nil {
		return ports.AccessClaims{}, err
	}
	sub, _ := cl["sub"].(string)
	sid, _ := cl["sid"].(string)
	roleStr, _ := cl["role"].(string)
	jti, _ := cl["jti"].(string)
	expF, _ := cl["exp"].(float64)
	iatF, _ := cl["iat"].(float64)
	if sub == "" || sid == "" || jti == "" {
		return ports.AccessClaims{}, errors.New("missing claims")
	}
	uid, err := uuid.Parse(sub)
	if err != nil {
		return ports.AccessClaims{}, err
	}
	sessID, err := uuid.Parse(sid)
	if err != nil {
		return ports.AccessClaims{}, err
	}
	return ports.AccessClaims{
		Sub:      uid,
		Session:  sessID,
		Role:     domain.Role(roleStr),
		JTI:      jti,
		Expires:  time.Unix(int64(expF), 0).UTC(),
		IssuedAt: time.Unix(int64(iatF), 0).UTC(),
	}, nil
}

func (s *Signer) JWKS(ctx context.Context) ([]byte, error) {
	_ = ctx
	jwk := jose.JSONWebKey{Key: s.pub, KeyID: s.keyID, Algorithm: string(jose.RS256), Use: "sig"}
	set := jose.JSONWebKeySet{Keys: []jose.JSONWebKey{jwk}}
	return json.Marshal(set)
}
