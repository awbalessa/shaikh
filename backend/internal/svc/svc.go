package svc

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/awbalessa/shaikh/backend/internal/dom"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type JWTIssuer struct {
	Secret   []byte
	Issuer   string
	Audience string
	TTL      time.Duration
}

func NewJWTIssuer(audience string, ttl time.Duration) *JWTIssuer {
	return &JWTIssuer{
		Secret:   []byte(os.Getenv("JWT_SECRET")),
		Issuer:   "shaikh-api",
		Audience: audience,
		TTL:      ttl,
	}
}

func (j *JWTIssuer) Sign(userID uuid.UUID) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"iss": j.Issuer,
		"aud": j.Audience,
		"sub": userID.String(),
		"iat": now.Unix(),
		"nbf": now.Unix(),
		"exp": now.Add(j.TTL).UTC(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(j.Secret)
}

type HealthReadinessSvc struct {
	Providers []dom.Provider
}

type CheckResult struct {
	Name   string `json:"name"`
	Status string `json:"status"` // ok | down | skipped
	Error  string `json:"error,omitempty"`
}

func (s *HealthReadinessSvc) CheckReadiness(ctx context.Context) (bool, []CheckResult) {
	results := make([]CheckResult, len(s.Providers))

	var wg sync.WaitGroup
	wg.Add(len(s.Providers))

	for i, p := range s.Providers {
		i, p := i, p // capture
		go func() {
			defer wg.Done()

			cctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
			defer cancel()
			err := p.Ping(cctx)

			cr := CheckResult{Name: p.Name()}

			switch {
			case err == nil:
				cr.Status = "ok"
			case errors.Is(err, dom.ErrNotPingable):
				cr.Status = "skipped"
				cr.Error = dom.ErrNotPingable.Error()
			case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
				cr.Status = "timeout"
				cr.Error = err.Error()
			default:
				cr.Status = "down"
				cr.Error = err.Error()
			}

			results[i] = cr
		}()
	}

	wg.Wait()

	ready := true
	for _, cr := range results {
		if cr.Status == "down" {
			ready = false
			break
		}
	}

	return ready, results
}

type UserSvc struct {
	UserRepo dom.UserRepo
}

func (s *UserSvc) Register(
	ctx context.Context,
	email string,
	password string,
) (*dom.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user, err := s.UserRepo.CreateUser(ctx, uuid.New(), email, string(hash))
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserSvc) Login(
	ctx context.Context,
	email, password string,
) (*dom.User, error) {
	user, err := s.UserRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, err
	}

	return user, nil
}

type SessionSvc struct {
	SessionRepo dom.SessionRepo
}

func (s *SessionSvc) Create(
	ctx context.Context,
	id, userID uuid.UUID,
) (*dom.Session, error) {
	return s.SessionRepo.CreateSession(ctx, id, userID)
}

func (s *SessionSvc) Delete(
	ctx context.Context,
	id, userID uuid.UUID,
) error {
	ok, err := s.SessionRepo.BelongsToUser(ctx, id, userID)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("forbidden resource")
	}

	return s.SessionRepo.DeleteSessionByID(ctx, id)
}
