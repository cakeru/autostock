package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/cakeru/autostock/internal/domain"
	"github.com/cakeru/autostock/internal/employee/dto"
)

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

const selectEmployee = `
	SELECT e.id, e.user_id, COALESCE(u.username, ''), e.name, COALESCE(e.position, ''),
	       COALESCE(e.phone, ''), COALESCE(e.email, ''), e.pay_type,
	       e.base_salary, e.hourly_rate, e.commission_rate,
	       COALESCE(e.hire_date::text, ''), COALESCE(e.notes, ''), e.is_active, e.created_at
	FROM employees e
	LEFT JOIN users u ON u.id = e.user_id
`

func scanEmployee(row pgx.Row) (*dto.EmployeeResponse, error) {
	var e dto.EmployeeResponse
	if err := row.Scan(&e.ID, &e.UserID, &e.Username, &e.Name, &e.Position,
		&e.Phone, &e.Email, &e.PayType, &e.BaseSalary, &e.HourlyRate, &e.CommissionRate,
		&e.HireDate, &e.Notes, &e.IsActive, &e.CreatedAt); err != nil {
		return nil, err
	}
	return &e, nil
}

func (s *Service) List(ctx context.Context, branchID int64) ([]dto.EmployeeResponse, error) {
	rows, err := s.pool.Query(ctx, selectEmployee+`
		WHERE e.branch_id = $1 AND e.is_active ORDER BY e.name`, branchID)
	if err != nil {
		return nil, fmt.Errorf("list employees: %w", err)
	}
	defer rows.Close()

	out := []dto.EmployeeResponse{}
	for rows.Next() {
		emp, err := scanEmployee(rows)
		if err != nil {
			return nil, fmt.Errorf("scan employee: %w", err)
		}
		out = append(out, *emp)
	}
	return out, nil
}

func (s *Service) Get(ctx context.Context, branchID, id int64) (*dto.EmployeeResponse, error) {
	emp, err := scanEmployee(s.pool.QueryRow(ctx, selectEmployee+`
		WHERE e.id = $1 AND e.branch_id = $2`, id, branchID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get employee: %w", err)
	}
	return emp, nil
}

func (s *Service) Create(ctx context.Context, branchID int64, req *dto.CreateEmployeeRequest) (*dto.EmployeeResponse, error) {
	payType := req.PayType
	if payType == "" {
		payType = "salary"
	}

	var id int64
	err := s.pool.QueryRow(ctx, `
		INSERT INTO employees (branch_id, name, position, phone, email, pay_type, base_salary, hourly_rate, commission_rate, hire_date, notes)
		VALUES ($1, $2, NULLIF($3,''), NULLIF($4,''), NULLIF($5,''), $6, $7, $8, $9, NULLIF($10,'')::date, NULLIF($11,''))
		RETURNING id`,
		branchID, req.Name, req.Position, req.Phone, req.Email, payType,
		req.BaseSalary, req.HourlyRate, req.CommissionRate, req.HireDate, req.Notes).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("create employee: %w", err)
	}
	return s.Get(ctx, branchID, id)
}

func (s *Service) Update(ctx context.Context, branchID, id int64, req *dto.UpdateEmployeeRequest) (*dto.EmployeeResponse, error) {
	var hireDate *string
	if req.HireDate != nil {
		hireDate = req.HireDate
	}
	tag, err := s.pool.Exec(ctx, `
		UPDATE employees SET
			name = COALESCE($1, name),
			position = COALESCE($2, position),
			phone = COALESCE($3, phone),
			email = COALESCE($4, email),
			pay_type = COALESCE($5, pay_type),
			base_salary = COALESCE($6, base_salary),
			hourly_rate = COALESCE($7, hourly_rate),
			commission_rate = COALESCE($8, commission_rate),
			hire_date = CASE WHEN $9::text IS NOT NULL THEN NULLIF($9::text, '')::date ELSE hire_date END,
			notes = COALESCE($10, notes),
			updated_at = NOW()
		WHERE id = $11 AND branch_id = $12`,
		req.Name, req.Position, req.Phone, req.Email, req.PayType,
		req.BaseSalary, req.HourlyRate, req.CommissionRate, hireDate, req.Notes, id, branchID)
	if err != nil {
		return nil, fmt.Errorf("update employee: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil, domain.ErrNotFound
	}
	return s.Get(ctx, branchID, id)
}

func (s *Service) Deactivate(ctx context.Context, branchID, id int64) error {
	tag, err := s.pool.Exec(ctx, `UPDATE employees SET is_active = false, updated_at = NOW() WHERE id = $1 AND branch_id = $2`, id, branchID)
	if err != nil {
		return fmt.Errorf("deactivate employee: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// CreateAccount gives an existing employee profile a login account. The
// employee row (pay info, job assignment history) is untouched — only
// user_id gets linked.
func (s *Service) CreateAccount(ctx context.Context, branchID, employeeID int64, req *dto.CreateAccountRequest) (*dto.EmployeeResponse, error) {
	var existingUserID *int64
	var name string
	err := s.pool.QueryRow(ctx, `SELECT user_id, name FROM employees WHERE id = $1 AND branch_id = $2`, employeeID, branchID).
		Scan(&existingUserID, &name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("check employee: %w", err)
	}
	if existingUserID != nil {
		return nil, &domain.AppError{Code: "ALREADY_HAS_ACCOUNT", Message: "This employee already has a login account", Status: 400}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	perms := req.Permissions
	if perms == nil {
		perms = []string{}
	}
	permissionsJSON, _ := json.Marshal(perms)

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var userID int64
	err = tx.QueryRow(ctx, `
		INSERT INTO users (branch_id, username, password_hash, full_name, role, permissions)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`,
		branchID, req.Username, string(hash), name, req.Role, permissionsJSON).Scan(&userID)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, &domain.AppError{Code: "USERNAME_TAKEN", Message: "That username is already in use", Status: 400}
		}
		return nil, fmt.Errorf("create user: %w", err)
	}

	if _, err := tx.Exec(ctx, `UPDATE employees SET user_id = $1, updated_at = NOW() WHERE id = $2`, userID, employeeID); err != nil {
		return nil, fmt.Errorf("link account: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return s.Get(ctx, branchID, employeeID)
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
