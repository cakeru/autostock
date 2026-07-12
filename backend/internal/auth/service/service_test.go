package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cakeru/autostock/internal/auth/dto"
	"github.com/cakeru/autostock/internal/testutil"
)

func TestLogin(t *testing.T) {
	pool := testutil.ConnectDB(t)
	branchID := testutil.SeedBranch(t, pool)

	_, err := pool.Exec(context.Background(),
		`INSERT INTO users (branch_id, username, password_hash, full_name, role, permissions) VALUES ($1, 'logintest', '$2a$12$NbStBtjAPLePbJsK4v9c/.GJUme2amTx48imqIx8FNg6kWf45QyyO', 'Login Test', 'admin', '[]')`,
		branchID)
	require.NoError(t, err)

	t.Cleanup(func() {
		pool.Exec(context.Background(), `DELETE FROM users WHERE username = 'logintest'`)
	})

	s := NewService(pool, "test-secret", 24*time.Hour)

	t.Run("successful login", func(t *testing.T) {
		resp, err := s.Login(context.Background(), &dto.LoginRequest{
			Username: "logintest",
			Password: "admin123",
		})
		require.NoError(t, err)
		assert.NotEmpty(t, resp.AccessToken)
		assert.Equal(t, "logintest", resp.User.Username)
	})

	t.Run("wrong password", func(t *testing.T) {
		_, err := s.Login(context.Background(), &dto.LoginRequest{
			Username: "logintest",
			Password: "wrong",
		})
		require.Error(t, err)
	})

	t.Run("unknown user", func(t *testing.T) {
		_, err := s.Login(context.Background(), &dto.LoginRequest{
			Username: "nobody",
			Password: "x",
		})
		require.Error(t, err)
	})
}

func TestGetMe(t *testing.T) {
	pool := testutil.ConnectDB(t)
	branchID := testutil.SeedBranch(t, pool)
	userID := testutil.SeedUser(t, pool, branchID)

	s := NewService(pool, "test-secret", 24*time.Hour)

	me, err := s.GetMe(context.Background(), userID)
	require.NoError(t, err)
	assert.Equal(t, "admin", me.Role)
	assert.True(t, me.IsActive)
}

func TestCreateUser(t *testing.T) {
	pool := testutil.ConnectDB(t)
	branchID := testutil.SeedBranch(t, pool)

	s := NewService(pool, "test-secret", 24*time.Hour)

	t.Run("creates user with valid data", func(t *testing.T) {
		user, err := s.CreateUser(context.Background(), &dto.CreateUserRequest{
			Username: "newuser",
			Password: "newpass123",
			FullName: "New User",
			Role:     "staff",
		}, branchID)
		require.NoError(t, err)
		assert.Equal(t, "newuser", user.Username)
		assert.Equal(t, "staff", user.Role)
		assert.True(t, user.IsActive)
	})

	t.Run("rejects duplicate username", func(t *testing.T) {
		_, err := s.CreateUser(context.Background(), &dto.CreateUserRequest{
			Username: "newuser",
			Password: "another",
			FullName: "Duplicate",
			Role:     "staff",
		}, branchID)
		require.Error(t, err)
	})
}

func TestListUsers(t *testing.T) {
	pool := testutil.ConnectDB(t)
	branchID := testutil.SeedBranch(t, pool)
	testutil.SeedUser(t, pool, branchID)

	s := NewService(pool, "test-secret", 24*time.Hour)

	users, err := s.ListUsers(context.Background(), branchID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(users), 1)
}
