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
	createdEmail, createdHash string
	byEmail                   User
	byEmailErr                error
}

func (r *repositoryStub) Create(_ context.Context, email, hash string) (User, error) {
	r.createdEmail, r.createdHash = email, hash
	return User{ID: 1, Email: email, PasswordHash: hash, Role: "customer", Status: "active"}, nil
}
func (r *repositoryStub) ByEmail(context.Context, string) (User, error) {
	return r.byEmail, r.byEmailErr
}
func (r *repositoryStub) ByID(context.Context, int64) (User, error) { return User{}, sql.ErrNoRows }

func TestCreateHashesPasswordAndNormalizesEmail(t *testing.T) {
	repo := &repositoryStub{}
	svc := NewService(repo, auth.NewManager("0123456789abcdef0123456789abcdef", time.Hour))
	if _, err := svc.Create(context.Background(), CreateInput{Email: " User@Example.COM ", Password: "password-123"}); err != nil {
		t.Fatal(err)
	}
	if repo.createdEmail != "user@example.com" {
		t.Fatalf("email=%q", repo.createdEmail)
	}
	if repo.createdHash == "password-123" || repo.createdHash == "" {
		t.Fatal("password was not hashed")
	}
}
func TestLoginRejectsUnknownUser(t *testing.T) {
	repo := &repositoryStub{byEmailErr: sql.ErrNoRows}
	svc := NewService(repo, auth.NewManager("0123456789abcdef0123456789abcdef", time.Hour))
	_, err := svc.Login(context.Background(), LoginInput{Email: "none@example.com", Password: "password-123"})
	if !errors.Is(err, domain.ErrUnauthenticated) {
		t.Fatalf("error=%v", err)
	}
}
