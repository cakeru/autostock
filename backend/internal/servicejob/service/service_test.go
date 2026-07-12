package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cakeru/autostock/internal/servicejob/dto"
	"github.com/cakeru/autostock/internal/testutil"
)

func setupService(t *testing.T) (*Service, int64, int64) {
	t.Helper()
	pool := testutil.ConnectDB(t)
	branchID := testutil.SeedBranch(t, pool)
	userID := testutil.SeedUser(t, pool, branchID)
	s := NewService(pool)
	return s, branchID, userID
}

func TestGenerateJobNumber(t *testing.T) {
	s, branchID, userID := setupService(t)

	t.Run("returns a valid format", func(t *testing.T) {
		num, err := s.generateJobNumber(context.Background())
		require.NoError(t, err)
		assert.Regexp(t, `^JOB-\d{4}-\d{4}$`, num)
	})

	t.Run("increments after inserting a job", func(t *testing.T) {
		first, _ := s.generateJobNumber(context.Background())
		customerID := testutil.SeedCustomer(t, s.pool, branchID)
		_, err := s.Create(context.Background(), branchID, userID, &dto.CreateServiceJobRequest{
			CustomerID:   &customerID,
			Description:  "Test job",
		})
		require.NoError(t, err)

		second, err := s.generateJobNumber(context.Background())
		require.NoError(t, err)
		assert.NotEqual(t, first, second)
	})
}

func TestStatusTransitions(t *testing.T) {
	s, branchID, userID := setupService(t)
	customerID := testutil.SeedCustomer(t, s.pool, branchID)
	job, err := s.Create(context.Background(), branchID, userID, &dto.CreateServiceJobRequest{
		CustomerID:  &customerID,
		Description: "Transition test",
	})
	require.NoError(t, err)

	t.Run("pending to in_progress", func(t *testing.T) {
		updated, err := s.Update(context.Background(), branchID, job.ID, &dto.UpdateServiceJobRequest{
			Status: strPtr("in_progress"),
		})
		require.NoError(t, err)
		assert.Equal(t, "in_progress", updated.Status)
	})

	t.Run("in_progress to completed", func(t *testing.T) {
		completed, err := s.Complete(context.Background(), branchID, job.ID)
		require.NoError(t, err)
		assert.Equal(t, "completed", completed.Status)
	})

	t.Run("cannot revert completed to pending", func(t *testing.T) {
		_, err := s.Update(context.Background(), branchID, job.ID, &dto.UpdateServiceJobRequest{
			Status: strPtr("pending"),
		})
		require.Error(t, err)
	})

	t.Run("cannot complete an already completed job", func(t *testing.T) {
		_, err := s.Complete(context.Background(), branchID, job.ID)
		require.Error(t, err)
	})
}

func strPtr(s string) *string { return &s }
