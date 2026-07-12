package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cakeru/autostock/internal/inventory/dto"
	"github.com/cakeru/autostock/internal/testutil"
)

func setupService(t *testing.T) (*Service, int64) {
	t.Helper()
	pool := testutil.ConnectDB(t)
	branchID := testutil.SeedBranch(t, pool)
	s := NewService(pool)
	return s, branchID
}

func TestReceiveStock(t *testing.T) {
	s, branchID := setupService(t)
	userID := testutil.SeedUser(t, s.pool, branchID)
	prodID := testutil.SeedProduct(t, s.pool, branchID, "Receive Test", "RCV-001", 5, 100)

	t.Run("increases stock and records movement", func(t *testing.T) {
		p, err := s.ReceiveStock(context.Background(), branchID, prodID, userID, &dto.ReceiveStockRequest{
			Quantity: 10,
			UnitCost: 80,
		})
		require.NoError(t, err)
		assert.Equal(t, 15.0, p.StockQuantity)
		assert.Equal(t, 80.0, p.BuyPrice)

		var movementQty float64
		err = s.pool.QueryRow(context.Background(),
			`SELECT quantity_change FROM stock_movements WHERE product_id = $1 AND reason = 'received'`, prodID).Scan(&movementQty)
		require.NoError(t, err)
		assert.Equal(t, 10.0, movementQty)
	})
}

func TestAdjustStock(t *testing.T) {
	s, branchID := setupService(t)
	userID := testutil.SeedUser(t, s.pool, branchID)
	prodID := testutil.SeedProduct(t, s.pool, branchID, "Adjust Test", "ADJ-001", 10, 50)

	t.Run("adjusts stock with positive delta", func(t *testing.T) {
		p, err := s.AdjustStock(context.Background(), branchID, prodID, userID, &dto.AdjustStockRequest{
			QuantityChange: 5,
			Reason:         "restock",
		})
		require.NoError(t, err)
		assert.Equal(t, 15.0, p.StockQuantity)
	})

	t.Run("adjusts stock with negative delta", func(t *testing.T) {
		p, err := s.AdjustStock(context.Background(), branchID, prodID, userID, &dto.AdjustStockRequest{
			QuantityChange: -3,
			Reason:         "damaged",
		})
		require.NoError(t, err)
		assert.Equal(t, 12.0, p.StockQuantity)
	})

	t.Run("adjusts stock by a fractional amount", func(t *testing.T) {
		p, err := s.AdjustStock(context.Background(), branchID, prodID, userID, &dto.AdjustStockRequest{
			QuantityChange: -2.5,
			Reason:         "poured 2.5L",
		})
		require.NoError(t, err)
		assert.Equal(t, 9.5, p.StockQuantity)
	})

	t.Run("rejects over-adjustment below zero", func(t *testing.T) {
		_, err := s.AdjustStock(context.Background(), branchID, prodID, userID, &dto.AdjustStockRequest{
			QuantityChange: -99,
			Reason:         "over adjustment",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Cannot adjust stock below 0")
	})
}
