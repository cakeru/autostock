package service

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cakeru/autostock/internal/invoice/dto"
	"github.com/cakeru/autostock/internal/testutil"
)

func setupService(t *testing.T) (*Service, *pgxpool.Pool, int64, int64) {
	t.Helper()
	pool := testutil.ConnectDB(t)
	branchID := testutil.SeedBranch(t, pool)
	userID := testutil.SeedUser(t, pool, branchID)
	s := NewService(pool)
	return s, pool, branchID, userID
}

func TestGenerateInvoiceNumber(t *testing.T) {
	s, pool, branchID, _ := setupService(t)

	t.Run("returns a valid format", func(t *testing.T) {
		num, err := s.generateInvoiceNumber(context.Background())
		require.NoError(t, err)
		assert.Regexp(t, `^INV-\d{4}-\d{4}$`, num)
	})

	t.Run("increments after inserting an invoice", func(t *testing.T) {
		first, _ := s.generateInvoiceNumber(context.Background())
		_, err := pool.Exec(context.Background(),
			`INSERT INTO invoices (branch_id, invoice_number, status, payment_status, total_usd, exchange_rate, total_khr, paid_amount)
			 VALUES ($1, $2, 'issued', 'unpaid', 0, 4050, 0, 0)`,
			branchID, first)
		require.NoError(t, err)

		second, err := s.generateInvoiceNumber(context.Background())
		require.NoError(t, err)
		assert.NotEqual(t, first, second)
	})
}

func TestCreateInvoice(t *testing.T) {
	s, _, branchID, userID := setupService(t)
	customerID := testutil.SeedCustomer(t, s.pool, branchID)
	prodID := testutil.SeedProduct(t, s.pool, branchID, "Test Tire", "TIR-001", 10, 100)

	t.Run("creates invoice and deducts stock", func(t *testing.T) {
		inv, err := s.Create(context.Background(), branchID, userID, &dto.CreateInvoiceRequest{
			CustomerID:   &customerID,
			Items:        []dto.InvoiceItemReq{{ProductID: &prodID, ItemType: "product", Description: "Tire", Quantity: 2, UnitPriceUSD: 100}},
			ExchangeRate: 4050,
		})
		require.NoError(t, err)
		assert.Equal(t, "issued", inv.Status)
		assert.Equal(t, 200.0, inv.Subtotal)
		assert.Equal(t, 200.0, inv.TotalUSD)

		var stock float64
		s.pool.QueryRow(context.Background(), `SELECT stock_quantity FROM products WHERE id = $1`, prodID).Scan(&stock)
		assert.Equal(t, 8.0, stock)
	})

	t.Run("rejects oversell", func(t *testing.T) {
		_, err := s.Create(context.Background(), branchID, userID, &dto.CreateInvoiceRequest{
			CustomerID:   &customerID,
			Items:        []dto.InvoiceItemReq{{ProductID: &prodID, ItemType: "product", Description: "Tire", Quantity: 99, UnitPriceUSD: 100}},
			ExchangeRate: 4050,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient available stock")
	})

	t.Run("uses exchange rate from settings when not provided", func(t *testing.T) {
		inv, err := s.Create(context.Background(), branchID, userID, &dto.CreateInvoiceRequest{
			CustomerID:   &customerID,
			Items:        []dto.InvoiceItemReq{{ProductID: &prodID, ItemType: "product", Description: "Tire", Quantity: 1, UnitPriceUSD: 100}},
			ExchangeRate: 0,
		})
		require.NoError(t, err)
		assert.Equal(t, 4050.0, inv.ExchangeRate)
	})
}

func TestVoidInvoice(t *testing.T) {
	s, _, branchID, userID := setupService(t)
	customerID := testutil.SeedCustomer(t, s.pool, branchID)
	prodID := testutil.SeedProduct(t, s.pool, branchID, "Void Test", "VOI-001", 5, 50)

	inv, err := s.Create(context.Background(), branchID, userID, &dto.CreateInvoiceRequest{
		CustomerID:   &customerID,
		Items:        []dto.InvoiceItemReq{{ProductID: &prodID, ItemType: "product", Description: "Item", Quantity: 2, UnitPriceUSD: 50}},
		ExchangeRate: 4050,
	})
	require.NoError(t, err)

	t.Run("restores stock on void", func(t *testing.T) {
		var stockBefore float64
		s.pool.QueryRow(context.Background(), `SELECT stock_quantity FROM products WHERE id = $1`, prodID).Scan(&stockBefore)
		assert.Equal(t, 3.0, stockBefore)

		voided, err := s.Void(context.Background(), branchID, inv.ID, userID, "test void")
		require.NoError(t, err)
		assert.Equal(t, "voided", voided.Status)

		var stockAfter float64
		s.pool.QueryRow(context.Background(), `SELECT stock_quantity FROM products WHERE id = $1`, prodID).Scan(&stockAfter)
		assert.Equal(t, 5.0, stockAfter)
	})

	t.Run("rejects double void", func(t *testing.T) {
		_, err := s.Void(context.Background(), branchID, inv.ID, userID, "again")
		require.Error(t, err)
	})
}

func TestUpdateInvoice(t *testing.T) {
	s, _, branchID, userID := setupService(t)
	customerID := testutil.SeedCustomer(t, s.pool, branchID)
	prodID := testutil.SeedProduct(t, s.pool, branchID, "Update Test", "UPD-001", 10, 75)

	inv, err := s.Create(context.Background(), branchID, userID, &dto.CreateInvoiceRequest{
		CustomerID:   &customerID,
		Items:        []dto.InvoiceItemReq{{ProductID: &prodID, ItemType: "product", Description: "Item", Quantity: 1, UnitPriceUSD: 75}},
		ExchangeRate: 4050,
	})
	require.NoError(t, err)

	t.Run("rejects payment exceeding total", func(t *testing.T) {
		_, err := s.RecordPayment(context.Background(), branchID, inv.ID, userID, &dto.RecordPaymentRequest{
			Amount: 999999,
			Method: "cash",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Payment would exceed invoice total")
	})

	t.Run("record payment successfully", func(t *testing.T) {
		p, err := s.RecordPayment(context.Background(), branchID, inv.ID, userID, &dto.RecordPaymentRequest{
			Amount: 75,
			Method: "cash",
		})
		require.NoError(t, err)
		assert.Equal(t, inv.ID, p.InvoiceID)
		assert.Equal(t, 75.0, p.Amount)
		assert.Equal(t, "cash", p.Method)
	})

	t.Run("rejects update on voided invoice", func(t *testing.T) {
		_, err := s.Void(context.Background(), branchID, inv.ID, userID, "void for update test")
		require.NoError(t, err)

		_, err = s.Update(context.Background(), branchID, inv.ID, &dto.UpdateInvoiceRequest{
			PaymentMethod: strPtr("card"),
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Cannot update a voided invoice")
	})
}

func TestRecordPayment(t *testing.T) {
	s, _, branchID, userID := setupService(t)
	customerID := testutil.SeedCustomer(t, s.pool, branchID)
	prodID := testutil.SeedProduct(t, s.pool, branchID, "Pay Test", "PAY-001", 20, 50)

	inv, err := s.Create(context.Background(), branchID, userID, &dto.CreateInvoiceRequest{
		CustomerID:   &customerID,
		Items:        []dto.InvoiceItemReq{{ProductID: &prodID, ItemType: "product", Description: "Item", Quantity: 2, UnitPriceUSD: 50}},
		ExchangeRate: 4050,
	})
	require.NoError(t, err)
	assert.Equal(t, float64(100), inv.TotalUSD)

	t.Run("partial payment updates status to partial", func(t *testing.T) {
		p, err := s.RecordPayment(context.Background(), branchID, inv.ID, userID, &dto.RecordPaymentRequest{
			Amount: 40,
			Method: "cash",
		})
		require.NoError(t, err)
		assert.Equal(t, 40.0, p.Amount)

		detail, err := s.Get(context.Background(), branchID, inv.ID)
		require.NoError(t, err)
		assert.Equal(t, "partial", detail.PaymentStatus)
		assert.Equal(t, 40.0, detail.PaidAmount)
	})

	t.Run("second payment brings to paid", func(t *testing.T) {
		p, err := s.RecordPayment(context.Background(), branchID, inv.ID, userID, &dto.RecordPaymentRequest{
			Amount: 60,
			Method: "card",
		})
		require.NoError(t, err)
		assert.Equal(t, 60.0, p.Amount)

		detail, err := s.Get(context.Background(), branchID, inv.ID)
		require.NoError(t, err)
		assert.Equal(t, "paid", detail.PaymentStatus)
		assert.Equal(t, float64(100), detail.PaidAmount)
	})

	t.Run("rejects payment on fully paid invoice", func(t *testing.T) {
		_, err := s.RecordPayment(context.Background(), branchID, inv.ID, userID, &dto.RecordPaymentRequest{
			Amount: 1,
			Method: "cash",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Payment would exceed invoice total")
	})
}

func TestInvoiceFrozen(t *testing.T) {
	s, _, branchID, userID := setupService(t)
	prodID := testutil.SeedProduct(t, s.pool, branchID, "Frozen Test", "FRZ-001", 10, 50)

	inv, err := s.Create(context.Background(), branchID, userID, &dto.CreateInvoiceRequest{
		Items:        []dto.InvoiceItemReq{{ProductID: &prodID, ItemType: "product", Description: "Item", Quantity: 1, UnitPriceUSD: 50}},
		ExchangeRate: 4050,
	})
	require.NoError(t, err)
	assert.Equal(t, "unpaid", inv.PaymentStatus)

	t.Run("rejects add item on issued invoice", func(t *testing.T) {
		_, err := s.AddItem(context.Background(), branchID, inv.ID, &dto.InvoiceItemReq{
			ProductID: &prodID, ItemType: "product", Description: "Extra", Quantity: 1, UnitPriceUSD: 50,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Cannot modify items on a non-draft invoice")
	})

	t.Run("rejects remove item on issued invoice", func(t *testing.T) {
		var itemID int64
		err := s.pool.QueryRow(context.Background(), `SELECT id FROM invoice_items WHERE invoice_id = $1 LIMIT 1`, inv.ID).Scan(&itemID)
		if err != nil {
			t.Skip("no items to remove")
		}
		err = s.RemoveItem(context.Background(), branchID, itemID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Cannot modify items on a non-draft invoice")
	})
}

func strPtr(s string) *string { return &s }
