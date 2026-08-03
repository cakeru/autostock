package testutil

import (
	"context"
	"fmt"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

var userSeq int64

func ConnectDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://autostock:autostock123@localhost:5433/autostock?sslmode=disable"
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	require.NoError(t, err, "failed to connect to database")

	t.Cleanup(func() {
		pool.Close()
	})

	return pool
}

func SeedBranch(t *testing.T, pool *pgxpool.Pool) int64 {
	t.Helper()

	var id int64
	err := pool.QueryRow(context.Background(),
		`INSERT INTO branches (name, address, phone) VALUES ('Test Branch', 'Test Address', '123') RETURNING id`).Scan(&id)
	require.NoError(t, err, "failed to seed branch")

	t.Cleanup(func() {
		pool.Exec(context.Background(), `DELETE FROM branches WHERE id = $1`, id)
	})

	return id
}

func SeedUser(t *testing.T, pool *pgxpool.Pool, branchID int64) int64 {
	t.Helper()

	seq := atomic.AddInt64(&userSeq, 1)
	username := fmt.Sprintf("testuser_%d_%d", time.Now().UnixNano(), seq)

	var id int64
	err := pool.QueryRow(context.Background(),
		`INSERT INTO users (branch_id, username, password_hash, full_name, role, permissions) VALUES ($1, $2, '$2a$12$NbStBtjAPLePbJsK4v9c/.GJUme2amTx48imqIx8FNg6kWf45QyyO', 'Test User', 'admin', '[]') RETURNING id`,
		branchID, username).Scan(&id)
	require.NoError(t, err, "failed to seed user")

	t.Cleanup(func() {
		pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, id)
	})

	return id
}

func SeedProduct(t *testing.T, pool *pgxpool.Pool, branchID int64, name string, sku string, stockQty int32, price float64) int64 {
	t.Helper()

	if name == "" {
		name = "Test Product"
	}
	if sku == "" {
		sku = fmt.Sprintf("TST-%d", stockQty)
	}

	var id int64
	err := pool.QueryRow(context.Background(),
		`INSERT INTO products (branch_id, type, sku, name, sell_price, stock_quantity, unit) VALUES ($1, 'part', $2, $3, $4, $5, 'piece') RETURNING id`,
		branchID, sku, name, price, stockQty).Scan(&id)
	require.NoError(t, err, "failed to seed product")

	t.Cleanup(func() {
		pool.Exec(context.Background(), `DELETE FROM products WHERE id = $1`, id)
	})

	return id
}

func SeedCustomer(t *testing.T, pool *pgxpool.Pool, branchID int64) int64 {
	t.Helper()

	var id int64
	err := pool.QueryRow(context.Background(),
		`INSERT INTO customers (branch_id, name, phone) VALUES ($1, 'Test Customer', '123456789') RETURNING id`,
		branchID).Scan(&id)
	require.NoError(t, err, "failed to seed customer")

	t.Cleanup(func() {
		pool.Exec(context.Background(), `DELETE FROM customers WHERE id = $1`, id)
	})

	return id
}

func SeedVehicle(t *testing.T, pool *pgxpool.Pool, customerID int64) int64 {
	t.Helper()

	var id int64
	err := pool.QueryRow(context.Background(),
		`INSERT INTO vehicles (branch_id, customer_id, plate_number, make, distance_unit)
		 VALUES ((SELECT branch_id FROM customers WHERE id = $1), $1, $2, 'Test Make', 'km')
		 RETURNING id`,
		customerID, fmt.Sprintf("TST-%d-%d", customerID, time.Now().UnixNano())).Scan(&id)
	require.NoError(t, err, "failed to seed vehicle")

	t.Cleanup(func() {
		pool.Exec(context.Background(), `DELETE FROM vehicles WHERE id = $1`, id)
	})
	return id
}

func SeedServiceJob(t *testing.T, pool *pgxpool.Pool, branchID, customerID int64, status string) int64 {
	t.Helper()

	var id int64
	err := pool.QueryRow(context.Background(),
		`INSERT INTO service_jobs (branch_id, customer_id, job_number, status) VALUES ($1, $2, 'TST-0001', $3) RETURNING id`,
		branchID, customerID, status).Scan(&id)
	require.NoError(t, err, "failed to seed service job")

	t.Cleanup(func() {
		pool.Exec(context.Background(), `DELETE FROM service_jobs WHERE id = $1`, id)
	})

	return id
}
