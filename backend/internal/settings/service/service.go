package service

import (
	"context"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cakeru/autostock/internal/settings/dto"
)

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

func (s *Service) GetSettings(ctx context.Context, branchID int64) (*dto.SettingsResponse, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT key, value FROM settings WHERE branch_id = $1 OR branch_id IS NULL`, branchID)
	if err != nil {
		return nil, fmt.Errorf("query settings: %w", err)
	}
	defer rows.Close()

	settings := &dto.SettingsResponse{
		InvoicePrefix:      "INV",
		LowStockThreshold:  5,
		ExchangeRateUSDKHR: 4050,
		ShopName:           "K&S Wheel-Tyre",
	}

	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, fmt.Errorf("scan setting: %w", err)
		}

		switch key {
		case "exchange_rate_usd_khr":
			if v, err := strconv.ParseFloat(value, 64); err == nil {
				settings.ExchangeRateUSDKHR = v
			}
		case "tax_rate_percent":
			if v, err := strconv.ParseFloat(value, 64); err == nil {
				settings.TaxRatePercent = v
			}
		case "tax_enabled":
			settings.TaxEnabled = value == "true"
		case "invoice_prefix":
			settings.InvoicePrefix = value
		case "low_stock_threshold":
			if v, err := strconv.Atoi(value); err == nil {
				settings.LowStockThreshold = v
			}
		case "sale_packages":
			settings.SalePackages = value
		case "labor_presets":
			settings.LaborPresets = value
		case "fee_presets":
			settings.FeePresets = value
		case "payment_methods":
			settings.PaymentMethods = value
		case "shop_name":
			if value != "" {
				settings.ShopName = value
			}
		case "shop_address":
			settings.ShopAddress = value
		case "shop_phone":
			settings.ShopPhone = value
		case "shop_email":
			settings.ShopEmail = value
		case "feature_batch_scan":
			settings.FeatureBatchScan = value == "true"
		}
	}

	return settings, nil
}

func (s *Service) UpdateSetting(ctx context.Context, branchID int64, key, value string) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO settings (branch_id, key, value)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (branch_id, key) DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()`,
		branchID, key, value)
	return err
}

func (s *Service) GetExchangeRate(ctx context.Context, branchID int64) (*dto.ExchangeRateResponse, error) {
	var value string
	err := s.pool.QueryRow(ctx,
		`SELECT value FROM settings WHERE branch_id = $1 AND key = 'exchange_rate_usd_khr'`, branchID).Scan(&value)
	if err != nil {
		return &dto.ExchangeRateResponse{Rate: 4050}, nil
	}

	rate, _ := strconv.ParseFloat(value, 64)
	if rate == 0 {
		rate = 4050
	}

	return &dto.ExchangeRateResponse{Rate: rate}, nil
}
