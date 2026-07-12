-- Employee is the HR/payroll identity; a login account (users) is optional
-- and attached via user_id when someone actually needs system access. This
-- lets techs who never log in still be tracked for assignment and pay.
CREATE TABLE employees (
    id BIGSERIAL PRIMARY KEY,
    branch_id BIGINT NOT NULL REFERENCES branches(id),
    user_id BIGINT UNIQUE REFERENCES users(id),
    name VARCHAR(255) NOT NULL,
    position VARCHAR(100),
    phone VARCHAR(50),
    email VARCHAR(255),
    pay_type VARCHAR(20) NOT NULL DEFAULT 'salary' CHECK (pay_type IN ('salary', 'hourly', 'commission', 'hybrid')),
    base_salary NUMERIC(10,2) NOT NULL DEFAULT 0,
    hourly_rate NUMERIC(10,2) NOT NULL DEFAULT 0,
    commission_rate NUMERIC(5,2) NOT NULL DEFAULT 0,
    hire_date DATE,
    notes TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_employees_branch ON employees (branch_id, is_active);

-- Every existing login account becomes an employee profile 1:1, so current
-- staff/technicians already have a record before job assignment moves off users.
INSERT INTO employees (branch_id, user_id, name, position, is_active)
SELECT branch_id, id, COALESCE(NULLIF(full_name, ''), username), role, is_active FROM users;
