package user

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
	"ticketgo/internal/auth"
	"ticketgo/internal/domain"
)

type Service struct {
	repo                       Repository
	tokens                     *auth.Manager
	allowAdminSelfRegistration bool
}

func NewService(repo Repository, tokens *auth.Manager, allowAdminSelfRegistration bool) *Service {
	return &Service{repo: repo, tokens: tokens, allowAdminSelfRegistration: allowAdminSelfRegistration}
}
func (s *Service) Create(ctx context.Context, in CreateInput) (User, error) {
	email := strings.ToLower(strings.TrimSpace(in.Email))
	role := strings.ToLower(strings.TrimSpace(in.Role))
	if role == "" {
		role = "customer"
	}
	if role != "customer" && role != "admin" {
		return User{}, domain.New(domain.ErrInvalid, "role must be customer or admin", nil)
	}
	if role == "admin" && !s.allowAdminSelfRegistration {
		return User{}, domain.New(domain.ErrForbidden, "admin self-registration is disabled", nil)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}
	u, err := s.repo.Create(ctx, email, string(hash), role)
	if isUnique(err) {
		return User{}, domain.New(domain.ErrConflict, "email is already registered", err)
	}
	if err != nil {
		return User{}, err
	}
	return u, nil
}
func (s *Service) Login(ctx context.Context, in LoginInput) (string, error) {
	u, err := s.repo.ByEmail(ctx, strings.TrimSpace(in.Email))
	if errors.Is(err, sql.ErrNoRows) {
		return "", domain.New(domain.ErrUnauthenticated, "invalid email or password", err)
	}
	if err != nil {
		return "", err
	}
	if u.Status != "active" || bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(in.Password)) != nil {
		return "", domain.New(domain.ErrUnauthenticated, "invalid email or password", nil)
	}
	return s.tokens.Issue(u.ID, u.Role)
}
func (s *Service) ByID(ctx context.Context, id int64) (User, error) {
	u, err := s.repo.ByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, domain.New(domain.ErrNotFound, "user not found", err)
	}
	return u, err
}
func isUnique(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
