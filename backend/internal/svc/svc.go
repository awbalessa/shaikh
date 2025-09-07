package svc

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/awbalessa/shaikh/backend/internal/dom"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type JWTIssuer struct {
	secret   []byte
	issuer   string
	audience string
	TTL      time.Duration
}

func NewJWTIssuer(ttl time.Duration) *JWTIssuer {
	return &JWTIssuer{
		secret:   []byte(os.Getenv("JWT_SECRET")),
		issuer:   "shaikh-api",
		audience: "shaikh-api",
		TTL:      ttl,
	}
}

func (j *JWTIssuer) Sign(userID uuid.UUID, role string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"iss":  j.issuer,
		"aud":  j.audience,
		"sub":  userID.String(),
		"role": role,
		"iat":  now.Unix(),
		"nbf":  now.Unix(),
		"exp":  now.Add(j.TTL).Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	str, err := t.SignedString(j.secret)
	if err != nil {
		return "", dom.NewTaggedError(dom.ErrInternal, err)
	}

	return str, nil
}

type AuthSvc struct {
	jwt     *JWTIssuer
	refresh dom.RefreshTokenRepo
	user    dom.UserRepo
}

func BuildAuthSvc(iss *JWTIssuer, re dom.RefreshTokenRepo) *AuthSvc {
	return &AuthSvc{jwt: iss, refresh: re}
}

func (a *AuthSvc) JWT() *JWTIssuer {
	return a.jwt
}

func (a *AuthSvc) IssueTokens(ctx context.Context, u *dom.User) (string, string, error) {
	var role = map[bool]string{
		true:  "admin",
		false: "user",
	}[u.IsAdmin]

	acc, err := a.jwt.Sign(u.ID, role)
	if err != nil {
		return "", "", err
	}

	ref, err := a.refresh.CreateRefreshToken(ctx, u.ID, 60*24*time.Hour)
	if err != nil {
		return "", "", err
	}

	return acc, ref, nil
}

func (a *AuthSvc) Refresh(ctx context.Context, rawRefresh string) (string, string, error) {
	userID, err := a.refresh.ValidateAndRotate(ctx, rawRefresh)
	if err != nil {
		return "", "", err
	}

	u, err := a.user.GetUserByID(ctx, userID)
	if err != nil {
		return "", "", err
	}

	return a.IssueTokens(ctx, u)
}

func (a *AuthSvc) Revoke(ctx context.Context, rawRefresh string) error {
	return a.refresh.Revoke(ctx, rawRefresh)
}

func (a *AuthSvc) RevokeAll(ctx context.Context, userID uuid.UUID) error {
	return a.refresh.RevokeAll(ctx, userID)
}

type HealthReadinessSvc struct {
	Probes []dom.Probe
}

func BuildHealthReadinessSvc(probes []dom.Probe) *HealthReadinessSvc {
	return &HealthReadinessSvc{Probes: probes}
}

type CheckResult struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
	Code   string `json:"code,omitempty"`
}

func (s *HealthReadinessSvc) CheckReadiness(ctx context.Context) (bool, []CheckResult) {
	results := make([]CheckResult, len(s.Probes))

	var wg sync.WaitGroup
	wg.Add(len(s.Probes))

	for i, p := range s.Probes {
		i, p := i, p
		go func() {
			defer wg.Done()

			cctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
			defer cancel()
			err := p.Ping(cctx)

			cr := CheckResult{Name: p.Name()}
			if err == nil {
				cr.Status = "ok"
				results[i] = cr
				return
			}
			derr := dom.ToDomainError(err)
			switch derr.Code {
			case dom.CodeNotPingable:
				cr.Status = "skipped"
				cr.Error = derr.Message
			case dom.CodeTimeout:
				cr.Status = "timeout"
				cr.Error = derr.Message
			case dom.CodeUnavailable, dom.CodeInternal:
				cr.Status = "down"
				cr.Error = derr.Message
			default:
				cr.Status = "down"
				cr.Error = derr.Message
			}

			results[i] = cr
		}()
	}

	wg.Wait()

	ready := true
	for _, cr := range results {
		if cr.Status == "down" || cr.Status == "timeout" {
			ready = false
			break
		}
	}

	return ready, results
}

type UserSvc struct {
	UserRepo dom.UserRepo
}

func BuildUserSvc(ur dom.UserRepo) *UserSvc {
	return &UserSvc{UserRepo: ur}
}

func (s *UserSvc) Register(
	ctx context.Context,
	email string,
	password string,
) (*dom.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, dom.ToDomainError(dom.NewTaggedError(dom.ErrInvalidInput, err))
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
		return nil, dom.ToDomainError(dom.NewTaggedError(dom.ErrInvalidInput, err))
	}

	return user, nil
}

func (s *UserSvc) Delete(
	ctx context.Context,
	userID uuid.UUID,
) error {
	return s.UserRepo.DeleteUserByID(ctx, userID)
}

type SessionSvc struct {
	SessionRepo dom.SessionRepo
}

func BuildSessionSvc(sr dom.SessionRepo) *SessionSvc {
	return &SessionSvc{SessionRepo: sr}
}

func (s *SessionSvc) Create(
	ctx context.Context,
	userID uuid.UUID,
) (*dom.Session, error) {
	se, err := s.SessionRepo.CreateSession(ctx, uuid.New(), userID)
	if err != nil {
		return nil, err
	}

	return se, nil
}

func (s *SessionSvc) BelongsToUser(
	ctx context.Context,
	id, userID uuid.UUID,
) (bool, error) {
	se, err := s.SessionRepo.GetSessionByID(ctx, id)
	if err != nil {
		return false, err
	}

	if se.UserID != userID {
		return false, dom.ToDomainError(dom.NewTaggedError(dom.ErrOwnershipViolation, nil))
	}

	return true, nil
}

func (s *SessionSvc) SetArchive(
	ctx context.Context,
	id, userID uuid.UUID,
	archived bool,
) (*dom.Session, error) {
	ok, err := s.BelongsToUser(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, dom.ToDomainError(dom.NewTaggedError(dom.ErrOwnershipViolation, nil))
	}

	var archived_at *time.Time = nil
	if archived {
		archived_at = dom.Ptr(time.Now())
	}

	se, err := s.SessionRepo.UpdateSessionByID(ctx, id, nil, nil, nil, archived_at)
	if err != nil {
		return nil, err
	}

	return se, nil
}

func (s *SessionSvc) Delete(
	ctx context.Context,
	id, userID uuid.UUID,
) error {
	ok, err := s.BelongsToUser(ctx, id, userID)
	if err != nil {
		return err
	}
	if !ok {
		return dom.ToDomainError(dom.NewTaggedError(dom.ErrOwnershipViolation, nil))
	}

	if err := s.SessionRepo.DeleteSessionByID(ctx, id); err != nil {
		return err
	}

	return nil
}
