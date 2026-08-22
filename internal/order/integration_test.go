package order

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"ticketgo/internal/domain"
	"ticketgo/internal/inventory"
)

func integrationDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.PingContext(context.Background()); err != nil {
		db.Close()
		t.Skipf("PostgreSQL unavailable: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

type fixture struct{ userID, itemID, activityID int64 }

func seed(t *testing.T, db *sql.DB, total int64) fixture {
	t.Helper()
	ctx := context.Background()
	suffix := time.Now().UnixNano()
	var f fixture
	if err := db.QueryRowContext(ctx, `INSERT INTO users(email,password_hash) VALUES($1,'hash') RETURNING id`, fmt.Sprintf("phase1-%d@example.test", suffix)).Scan(&f.userID); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `INSERT INTO items(name,price_cents) VALUES('integration item',1000) RETURNING id`).Scan(&f.itemID); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `INSERT INTO activities(item_id,name,price_cents,starts_at,ends_at,status) VALUES($1,'integration activity',800,CURRENT_TIMESTAMP-INTERVAL '1 hour',CURRENT_TIMESTAMP+INTERVAL '1 hour','active') RETURNING id`, f.itemID).Scan(&f.activityID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO inventories(activity_id,total,available) VALUES($1,$2,$2)`, f.activityID, total); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Exec(`DELETE FROM seckill_records WHERE activity_id=$1`, f.activityID)
		db.Exec(`DELETE FROM orders WHERE activity_id=$1`, f.activityID)
		db.Exec(`DELETE FROM inventories WHERE activity_id=$1`, f.activityID)
		db.Exec(`DELETE FROM activities WHERE id=$1`, f.activityID)
		db.Exec(`DELETE FROM items WHERE id=$1`, f.itemID)
		db.Exec(`DELETE FROM users WHERE id=$1`, f.userID)
	})
	return f
}

func TestIntegrationSeckillCommitDuplicateCancelAndRollback(t *testing.T) {
	db := integrationDB(t)
	f := seed(t, db, 3)
	svc := NewService(db, NewRepository(db), inventory.NewRepository())
	ctx := context.Background()
	o, err := svc.Seckill(ctx, f.userID, f.activityID, 2)
	if err != nil {
		t.Fatal(err)
	}
	assertInventory(t, db, f.activityID, 3, 1, 2)
	if _, err = svc.Seckill(ctx, f.userID, f.activityID, 1); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("duplicate error=%v", err)
	}
	assertInventory(t, db, f.activityID, 3, 1, 2)
	cancelled, err := svc.Cancel(ctx, f.userID, o.ID)
	if err != nil {
		t.Fatal(err)
	}
	if cancelled.Status != "cancelled" {
		t.Fatalf("status=%s", cancelled.Status)
	}
	assertInventory(t, db, f.activityID, 3, 3, 0)

	f2 := seed(t, db, 2)
	faultSvc := NewService(db, NewRepository(db), inventory.NewRepository())
	faultSvc.afterInventoryUpdate = func() error { return errors.New("forced failure") }
	if _, err = faultSvc.Seckill(ctx, f2.userID, f2.activityID, 1); err == nil {
		t.Fatal("fault injection unexpectedly committed")
	}
	assertInventory(t, db, f2.activityID, 2, 2, 0)
	var orders int
	if err = db.QueryRow(`SELECT COUNT(*) FROM orders WHERE activity_id=$1`, f2.activityID).Scan(&orders); err != nil || orders != 0 {
		t.Fatalf("orders=%d err=%v", orders, err)
	}

	f3 := seed(t, db, 2)
	constraintSvc := NewService(db, NewRepository(db), inventory.NewRepository())
	constraintSvc.beforeOrderInsert = func(o *Order) { o.Quantity = 0 }
	if _, err = constraintSvc.Seckill(ctx, f3.userID, f3.activityID, 1); err == nil {
		t.Fatal("invalid order unexpectedly committed")
	}
	assertInventory(t, db, f3.activityID, 2, 2, 0)
}

func TestIntegrationForeignKeyConstraint(t *testing.T) {
	db := integrationDB(t)
	_, err := db.Exec(`INSERT INTO activities(item_id,name,price_cents,starts_at,ends_at,status) VALUES(9223372036854775807,'invalid fk',1,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP+INTERVAL '1 hour','draft')`)
	if err == nil {
		t.Fatal("activity with missing item unexpectedly inserted")
	}
}

func assertInventory(t *testing.T, db *sql.DB, activityID, total, available, sold int64) {
	t.Helper()
	var gotTotal, gotAvailable, gotSold int64
	if err := db.QueryRow(`SELECT total,available,sold FROM inventories WHERE activity_id=$1`, activityID).Scan(&gotTotal, &gotAvailable, &gotSold); err != nil {
		t.Fatal(err)
	}
	if gotTotal != total || gotAvailable != available || gotSold != sold {
		t.Fatalf("inventory=(%d,%d,%d), want=(%d,%d,%d)", gotTotal, gotAvailable, gotSold, total, available, sold)
	}
}
