package user

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"ticketgo/internal/auth"
	"ticketgo/internal/domain"
)

type repositoryStub struct {
	createdEmail, createdHash, createdRole string
	byEmail                                User
	byEmailErr                             error
}

func (r *repositoryStub) Create(_ context.Context, email, hash, role string) (User, error) {
	r.createdEmail, r.createdHash, r.createdRole = email, hash, role
	return User{ID: 1, Email: email, PasswordHash: hash, Role: role, Status: "active"}, nil
}
func (r *repositoryStub) ByEmail(context.Context, string) (User, error) {
	return r.byEmail, r.byEmailErr
}
func (r *repositoryStub) ByID(context.Context, int64) (User, error) { return User{}, sql.ErrNoRows }

func TestCreateHashesPasswordAndNormalizesEmail(t *testing.T) {
	repo := &repositoryStub{}
	svc := NewService(repo, auth.NewManager("0123456789abcdef0123456789abcdef", time.Hour), false)
	if _, err := svc.Create(context.Background(), CreateInput{Email: " User@Example.COM ", Password: "password-123"}); err != nil {
		t.Fatal(err)
	}
	if repo.createdEmail != "user@example.com" {
		t.Fatalf("email=%q", repo.createdEmail)
	}
	if repo.createdHash == "password-123" || repo.createdHash == "" {
		t.Fatal("password was not hashed")
	}
	if repo.createdRole != "customer" {
		t.Fatalf("role=%q, want customer", repo.createdRole)
	}
}
func TestLoginRejectsUnknownUser(t *testing.T) {
	repo := &repositoryStub{byEmailErr: sql.ErrNoRows}
	svc := NewService(repo, auth.NewManager("0123456789abcdef0123456789abcdef", time.Hour), false)
	_, err := svc.Login(context.Background(), LoginInput{Email: "none@example.com", Password: "password-123"})
	if !errors.Is(err, domain.ErrUnauthenticated) {
		t.Fatalf("error=%v", err)
	}
}

func TestCreateAdminRespectsDemoFlag(t *testing.T) {
	manager := auth.NewManager("0123456789abcdef0123456789abcdef", time.Hour)
	repo := &repositoryStub{}
	disabled := NewService(repo, manager, false)
	_, err := disabled.Create(context.Background(), CreateInput{Email: "admin@example.com", Password: "password-123", Role: "admin"})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("disabled error=%v, want forbidden", err)
	}

	enabled := NewService(repo, manager, true)
	u, err := enabled.Create(context.Background(), CreateInput{Email: "admin@example.com", Password: "password-123", Role: "admin"})
	if err != nil {
		t.Fatal(err)
	}
	if u.Role != "admin" || repo.createdRole != "admin" {
		t.Fatalf("user role=%q repository role=%q", u.Role, repo.createdRole)
	}
}

func TestCreateRejectsInvalidRole(t *testing.T) {
	svc := NewService(&repositoryStub{}, auth.NewManager("0123456789abcdef0123456789abcdef", time.Hour), true)
	_, err := svc.Create(context.Background(), CreateInput{Email: "user@example.com", Password: "password-123", Role: "owner"})
	if !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("error=%v, want invalid", err)
	}
}
