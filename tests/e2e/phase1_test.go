package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"go.uber.org/zap"
	"ticketgo/internal/config"
	"ticketgo/internal/httpapi"
	"ticketgo/pkg/database"
)

func testRouter(t *testing.T) (http.Handler, *database.DB) {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	dbcfg := database.Config{URL: dsn, MaxOpenConns: 5, MaxIdleConns: 2, ConnMaxLifetime: time.Minute, ConnMaxIdleTime: time.Minute, ConnectTimeout: 2 * time.Second, QueryTimeout: time.Second}
	db, err := database.Open(context.Background(), dbcfg)
	if err != nil {
		t.Skipf("PostgreSQL unavailable: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	cfg := config.Config{HTTP: config.HTTPConfig{RequestTimeout: 3 * time.Second}, Database: dbcfg, Auth: config.AuthConfig{JWTSecret: "0123456789abcdef0123456789abcdef", TokenTTL: time.Hour}}
	return httpapi.NewRouter(db, cfg, zap.NewNop()), db
}
func request(t *testing.T, router http.Handler, method, path, token string, body any) (int, map[string]any) {
	t.Helper()
	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	var decoded map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &decoded); err != nil {
		t.Fatalf("decode %s: %v body=%s", path, err, rec.Body.String())
	}
	return rec.Code, decoded
}
func data(body map[string]any) map[string]any { return body["data"].(map[string]any) }

func TestPhase1BusinessLoopAndErrorMapping(t *testing.T) {
	router, db := testRouter(t)
	suffix := time.Now().UnixNano()
	email := fmt.Sprintf("e2e-%d@example.test", suffix)
	ctx := context.Background()
	var userID, itemID, activityID int64
	t.Cleanup(func() {
		db.ExecContext(ctx, `DELETE FROM seckill_records WHERE activity_id=$1`, activityID)
		db.ExecContext(ctx, `DELETE FROM orders WHERE activity_id=$1`, activityID)
		db.ExecContext(ctx, `DELETE FROM inventories WHERE activity_id=$1`, activityID)
		db.ExecContext(ctx, `DELETE FROM activities WHERE id=$1`, activityID)
		db.ExecContext(ctx, `DELETE FROM items WHERE id=$1`, itemID)
		db.ExecContext(ctx, `DELETE FROM users WHERE id=$1`, userID)
	})

	if status, _ := request(t, router, http.MethodGet, "/api/v1/orders", "", nil); status != http.StatusUnauthorized {
		t.Fatalf("unauthorized status=%d", status)
	}
	status, body := request(t, router, http.MethodPost, "/api/v1/users", "", map[string]any{"email": email, "password": "password-123"})
	if status != http.StatusCreated {
		t.Fatalf("create user status=%d body=%v", status, body)
	}
	userID = int64(data(body)["id"].(float64))
	status, body = request(t, router, http.MethodPost, "/api/v1/login", "", map[string]any{"email": email, "password": "password-123"})
	if status != http.StatusOK {
		t.Fatalf("login status=%d body=%v", status, body)
	}
	customerToken := data(body)["access_token"].(string)
	if status, _ = request(t, router, http.MethodPost, "/api/v1/items", customerToken, map[string]any{"name": "x", "price_cents": 100}); status != http.StatusForbidden {
		t.Fatalf("forbidden status=%d", status)
	}
	if _, err := db.ExecContext(ctx, `UPDATE users SET role='admin' WHERE id=$1`, userID); err != nil {
		t.Fatal(err)
	}
	_, body = request(t, router, http.MethodPost, "/api/v1/login", "", map[string]any{"email": email, "password": "password-123"})
	adminToken := data(body)["access_token"].(string)
	status, body = request(t, router, http.MethodPost, "/api/v1/items", adminToken, map[string]any{"name": "Concert ticket", "description": "Phase 1", "price_cents": 10000})
	if status != http.StatusCreated {
		t.Fatalf("create item status=%d body=%v", status, body)
	}
	itemID = int64(data(body)["id"].(float64))
	now := time.Now().UTC()
	status, body = request(t, router, http.MethodPost, "/api/v1/activities", adminToken, map[string]any{"item_id": itemID, "name": "Opening sale", "price_cents": 8000, "starts_at": now.Add(-time.Minute), "ends_at": now.Add(time.Hour), "status": "active", "total": 2})
	if status != http.StatusCreated {
		t.Fatalf("create activity status=%d body=%v", status, body)
	}
	activityID = int64(data(body)["id"].(float64))
	status, body = request(t, router, http.MethodPost, fmt.Sprintf("/api/v1/activities/%d/seckill", activityID), adminToken, map[string]any{"quantity": 1})
	if status != http.StatusCreated {
		t.Fatalf("seckill status=%d body=%v", status, body)
	}
	orderID := int64(data(body)["id"].(float64))
	if status, _ = request(t, router, http.MethodGet, fmt.Sprintf("/api/v1/orders/%d", orderID), adminToken, nil); status != http.StatusOK {
		t.Fatalf("get order status=%d", status)
	}
	if status, _ = request(t, router, http.MethodGet, "/api/v1/orders?limit=101", adminToken, nil); status != http.StatusBadRequest {
		t.Fatalf("pagination status=%d", status)
	}
	if status, _ = request(t, router, http.MethodGet, "/api/v1/orders/999999999", adminToken, nil); status != http.StatusNotFound {
		t.Fatalf("not found status=%d", status)
	}
	if status, _ = request(t, router, http.MethodPost, fmt.Sprintf("/api/v1/orders/%d/cancel", orderID), adminToken, nil); status != http.StatusOK {
		t.Fatalf("cancel status=%d", status)
	}
	if status, _ = request(t, router, http.MethodPost, fmt.Sprintf("/api/v1/activities/%d/seckill", activityID), adminToken, map[string]any{"quantity": 1}); status != http.StatusConflict {
		t.Fatalf("duplicate status=%d", status)
	}
	if _, err := db.ExecContext(ctx, `UPDATE users SET status='disabled' WHERE id=$1`, userID); err != nil {
		t.Fatal(err)
	}
	if status, _ = request(t, router, http.MethodGet, "/api/v1/users/me", adminToken, nil); status != http.StatusForbidden {
		t.Fatalf("disabled user status=%d", status)
	}
}

func TestInternalDatabaseErrorMapsTo500(t *testing.T) {
	router, db := testRouter(t)
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	status, body := request(t, router, http.MethodPost, "/api/v1/users", "", map[string]any{"email": "closed-db@example.test", "password": "password-123"})
	if status != http.StatusInternalServerError || body["code"] != "internal_error" {
		t.Fatalf("status=%d body=%v", status, body)
	}
}
