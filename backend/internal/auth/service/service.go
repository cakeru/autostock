package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cakeru/autostock/internal/auth/dto"
	"github.com/cakeru/autostock/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	pool      *pgxpool.Pool
	jwtSecret string
	jwtExpiry time.Duration
}

func NewService(pool *pgxpool.Pool, jwtSecret string, jwtExpiry time.Duration) *Service {
	return &Service{pool: pool, jwtSecret: jwtSecret, jwtExpiry: jwtExpiry}
}

func (s *Service) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	var user domain.User
	var p string
	err := s.pool.QueryRow(ctx,
		`SELECT id, branch_id, username, COALESCE(email,''), password_hash, COALESCE(full_name,''), role, permissions::text, is_active
		 FROM users WHERE username = $1 AND is_active = true`, req.Username).
		Scan(&user.ID, &user.BranchID, &user.Username, &user.Email,
			&user.PasswordHash, &user.FullName, &user.Role, &p, &user.IsActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrInvalidLogin
		}
		return nil, fmt.Errorf("query user: %w", err)
	}

	json.Unmarshal([]byte(p), &user.Permissions)

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, domain.ErrInvalidLogin
	}

	_, _ = s.pool.Exec(ctx, `UPDATE users SET last_login_at = NOW() WHERE id = $1`, user.ID)

	token, _, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int(s.jwtExpiry.Seconds()),
		User: dto.UserInfo{
			ID:          user.ID,
			Username:    user.Username,
			Email:       user.Email,
			FullName:    user.FullName,
			Role:        user.Role,
			Permissions: user.Permissions,
			BranchID:    user.BranchID,
			IsActive:    user.IsActive,
		},
	}, nil
}

func (s *Service) GetMe(ctx context.Context, userID int64) (*dto.UserInfo, error) {
	var user domain.User
	var p string
	err := s.pool.QueryRow(ctx,
		`SELECT id, branch_id, username, COALESCE(email,''), COALESCE(full_name,''), role, permissions::text, is_active
		 FROM users WHERE id = $1 AND is_active = true`, userID).
		Scan(&user.ID, &user.BranchID, &user.Username, &user.Email,
			&user.FullName, &user.Role, &p, &user.IsActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("query user: %w", err)
	}

	json.Unmarshal([]byte(p), &user.Permissions)

	return &dto.UserInfo{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		FullName:    user.FullName,
		Role:        user.Role,
		Permissions: user.Permissions,
		BranchID:    user.BranchID,
		IsActive:    user.IsActive,
	}, nil
}

func (s *Service) ListUsers(ctx context.Context, branchID int64) ([]dto.UserInfo, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, branch_id, username, COALESCE(email,''), COALESCE(full_name,''), role, permissions::text, is_active
		 FROM users WHERE branch_id = $1 AND is_active = true ORDER BY created_at DESC`, branchID)
	if err != nil {
		return nil, fmt.Errorf("query users: %w", err)
	}
	defer rows.Close()

	var users []dto.UserInfo
	for rows.Next() {
		var u dto.UserInfo
		var p string
		if err := rows.Scan(&u.ID, &u.BranchID, &u.Username, &u.Email,
			&u.FullName, &u.Role, &p, &u.IsActive); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		json.Unmarshal([]byte(p), &u.Permissions)
		users = append(users, u)
	}

	if users == nil {
		users = []dto.UserInfo{}
	}
	return users, nil
}

func (s *Service) CreateUser(ctx context.Context, req *dto.CreateUserRequest, branchID int64) (*dto.UserInfo, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	perms := req.Permissions
	if perms == nil {
		perms = []string{}
	}
	permissionsJSON, _ := json.Marshal(perms)

	var user dto.UserInfo
	var p string
	err = s.pool.QueryRow(ctx,
		`INSERT INTO users (branch_id, username, email, password_hash, full_name, role, permissions)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, branch_id, username, COALESCE(email,''), COALESCE(full_name,''), role, permissions::text, is_active`,
		branchID, req.Username, req.Email, string(hash), req.FullName, req.Role, permissionsJSON).
		Scan(&user.ID, &user.BranchID, &user.Username, &user.Email,
			&user.FullName, &user.Role, &p, &user.IsActive)

	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	json.Unmarshal([]byte(p), &user.Permissions)
	return &user, nil
}

func (s *Service) UpdateUser(ctx context.Context, id int64, req *dto.UpdateUserRequest) (*dto.UserInfo, error) {
	var permissionsJSON []byte
	if req.Permissions != nil {
		permissionsJSON, _ = json.Marshal(req.Permissions)
	}

	var isActive *bool
	if req.IsActive != nil {
		isActive = req.IsActive
	}

	var user dto.UserInfo
	var p string
	err := s.pool.QueryRow(ctx,
		`UPDATE users
		 SET email = COALESCE($1, email),
		     full_name = COALESCE($2, full_name),
		     permissions = CASE WHEN $3::jsonb IS NOT NULL THEN $3 ELSE permissions END,
		     is_active = CASE WHEN $4::bool IS NOT NULL THEN $4 ELSE is_active END,
		     updated_at = NOW()
		 WHERE id = $5
		 RETURNING id, branch_id, username, COALESCE(email,''), COALESCE(full_name,''), role, permissions::text, is_active`,
		req.Email, req.FullName, permissionsJSON, isActive, id).
		Scan(&user.ID, &user.BranchID, &user.Username, &user.Email,
			&user.FullName, &user.Role, &p, &user.IsActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("update user: %w", err)
	}

	json.Unmarshal([]byte(p), &user.Permissions)
	return &user, nil
}

func (s *Service) DeleteUser(ctx context.Context, id int64) error {
	result, err := s.pool.Exec(ctx, `UPDATE users SET is_active = false, updated_at = NOW() WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (s *Service) ChangePassword(ctx context.Context, userID int64, req *dto.ChangePasswordRequest) error {
	var currentHash string
	err := s.pool.QueryRow(ctx, `SELECT password_hash FROM users WHERE id = $1`, userID).Scan(&currentHash)
	if err != nil {
		return domain.ErrNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(currentHash), []byte(req.CurrentPassword)); err != nil {
		return domain.ErrInvalidRequest
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), 12)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	_, err = s.pool.Exec(ctx, `UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2`, string(newHash), userID)
	return err
}

func (s *Service) generateToken(user domain.User) (string, int64, error) {
	exp := time.Now().Add(s.jwtExpiry)
	claims := jwt.MapClaims{
		"user_id":     user.ID,
		"username":    user.Username,
		"role":        user.Role,
		"branch_id":   user.BranchID,
		"permissions": user.Permissions,
		"exp":         exp.Unix(),
		"iat":         time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", 0, fmt.Errorf("sign token: %w", err)
	}

	return tokenStr, exp.Unix(), nil
}
