package service

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/cakeru/autostock/internal/inventory/dto"
)

type ImportResult struct {
	TotalRows int              `json:"total_rows"`
	Created   int              `json:"created"`
	Updated   int              `json:"updated"`
	Failed    int              `json:"failed"`
	Errors    []ImportRowError `json:"errors,omitempty"`
}

type ImportRowError struct {
	Row     int    `json:"row"`
	SKU     string `json:"sku,omitempty"`
	Message string `json:"message"`
}

// maxImportRows caps a single upload so a bad file (or one meant for another
// system) can't tie up the request indefinitely.
const maxImportRows = 2000

// ImportCSV bulk-creates or updates products from a CSV file. Rows are matched
// to existing products by SKU within the branch: a known SKU updates that
// product (name/prices/specs — never stock_quantity, which stays governed by
// Receive/Adjust so the batch ledger doesn't drift), an unknown SKU creates a
// new product with opening stock if given. Each row is independent so one bad
// row doesn't sink the rest of the file.
func (s *Service) ImportCSV(ctx context.Context, branchID, userID int64, r io.Reader) (*ImportResult, error) {
	reader := csv.NewReader(r)
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true

	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("read header row: %w", err)
	}
	colIndex := make(map[string]int, len(header))
	for i, h := range header {
		colIndex[strings.ToLower(strings.TrimSpace(h))] = i
	}
	if _, ok := colIndex["sku"]; !ok {
		return nil, fmt.Errorf("CSV must have a 'sku' column")
	}
	if _, ok := colIndex["name"]; !ok {
		return nil, fmt.Errorf("CSV must have a 'name' column")
	}

	get := func(row []string, col string) string {
		idx, ok := colIndex[col]
		if !ok || idx >= len(row) {
			return ""
		}
		return strings.TrimSpace(row[idx])
	}

	result := &ImportResult{}
	rowNum := 1 // header is row 1
	for {
		row, readErr := reader.Read()
		if readErr == io.EOF {
			break
		}
		rowNum++
		if readErr != nil {
			result.Failed++
			result.Errors = append(result.Errors, ImportRowError{Row: rowNum, Message: readErr.Error()})
			continue
		}
		// Skip fully blank lines (trailing newline, spreadsheet padding, etc).
		if isBlankRow(row) {
			continue
		}

		result.TotalRows++
		if result.TotalRows > maxImportRows {
			result.Failed++
			result.Errors = append(result.Errors, ImportRowError{Row: rowNum, Message: fmt.Sprintf("stopped after %d rows — split the file", maxImportRows)})
			break
		}

		sku := get(row, "sku")
		name := get(row, "name")
		if sku == "" {
			result.Failed++
			result.Errors = append(result.Errors, ImportRowError{Row: rowNum, Message: "missing sku"})
			continue
		}

		productType := strings.ToLower(get(row, "type"))
		if productType == "" {
			productType = "part"
		}
		if productType != "tire" && productType != "part" && productType != "labor" && productType != "consumable" {
			result.Failed++
			result.Errors = append(result.Errors, ImportRowError{Row: rowNum, SKU: sku, Message: fmt.Sprintf("invalid type %q (must be tire, part, labor, or consumable)", productType)})
			continue
		}

		buyPrice, buyErr := parseFloatOr(get(row, "buy_price"), 0)
		sellPrice, sellErr := parseFloatOr(get(row, "sell_price"), 0)
		stockQty, stockErr := parseFloatOr(get(row, "stock_quantity"), 0)
		minAlert, minErr := parseFloatOr(get(row, "min_stock_alert"), 5)
		if buyErr != nil || sellErr != nil || stockErr != nil || minErr != nil {
			result.Failed++
			result.Errors = append(result.Errors, ImportRowError{Row: rowNum, SKU: sku, Message: "invalid number in buy_price, sell_price, stock_quantity, or min_stock_alert"})
			continue
		}

		existingID, found, findErr := s.findSKUID(ctx, branchID, sku)
		if findErr != nil {
			result.Failed++
			result.Errors = append(result.Errors, ImportRowError{Row: rowNum, SKU: sku, Message: "lookup failed: " + findErr.Error()})
			continue
		}

		if found {
			req := &dto.UpdateProductRequest{
				Name:        nullIfEmpty(name),
				Barcode:     nullIfEmpty(get(row, "barcode")),
				Description: nullIfEmpty(get(row, "description")),
				Category:    nullIfEmpty(get(row, "category")),
				Unit:        nullIfEmpty(get(row, "unit")),
				Location:    nullIfEmpty(get(row, "location")),
				TireSize:    nullIfEmpty(get(row, "tire_size")),
				TireBrand:   nullIfEmpty(get(row, "tire_brand")),
				TireModel:   nullIfEmpty(get(row, "tire_model")),
				TirePattern: nullIfEmpty(get(row, "tire_pattern")),
				DOTCode:     nullIfEmpty(get(row, "dot_code")),
				LoadIndex:   nullIfEmpty(get(row, "load_index")),
				SpeedRating: nullIfEmpty(get(row, "speed_rating")),
				TireType:    nullIfEmpty(get(row, "tire_type")),
			}
			if buyPrice > 0 {
				req.BuyPrice = &buyPrice
			}
			if sellPrice > 0 {
				req.SellPrice = &sellPrice
			}
			if get(row, "min_stock_alert") != "" {
				req.MinStockAlert = &minAlert
			}
			if _, err := s.Update(ctx, branchID, existingID, req); err != nil {
				result.Failed++
				result.Errors = append(result.Errors, ImportRowError{Row: rowNum, SKU: sku, Message: "update failed: " + err.Error()})
				continue
			}
			result.Updated++
		} else {
			if name == "" {
				result.Failed++
				result.Errors = append(result.Errors, ImportRowError{Row: rowNum, SKU: sku, Message: "missing name (required for a new product)"})
				continue
			}
			req := &dto.CreateProductRequest{
				Type: productType, SKU: sku, Barcode: get(row, "barcode"), Name: name,
				Description: get(row, "description"), Category: get(row, "category"),
				BuyPrice: buyPrice, SellPrice: sellPrice, StockQuantity: stockQty, MinStockAlert: minAlert,
				Unit: get(row, "unit"), Location: get(row, "location"),
				TireSize: get(row, "tire_size"), TireBrand: get(row, "tire_brand"), TireModel: get(row, "tire_model"),
				TirePattern: get(row, "tire_pattern"), DOTCode: get(row, "dot_code"),
				LoadIndex: get(row, "load_index"), SpeedRating: get(row, "speed_rating"), TireType: get(row, "tire_type"),
			}
			if _, err := s.Create(ctx, branchID, userID, req); err != nil {
				result.Failed++
				result.Errors = append(result.Errors, ImportRowError{Row: rowNum, SKU: sku, Message: "create failed: " + err.Error()})
				continue
			}
			result.Created++
		}
	}

	return result, nil
}

func (s *Service) findSKUID(ctx context.Context, branchID int64, sku string) (int64, bool, error) {
	var id int64
	err := s.pool.QueryRow(ctx,
		`SELECT id FROM products WHERE branch_id = $1 AND sku = $2 AND is_active = true`,
		branchID, sku).Scan(&id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, false, nil
		}
		return 0, false, err
	}
	return id, true, nil
}

func isBlankRow(row []string) bool {
	for _, v := range row {
		if strings.TrimSpace(v) != "" {
			return false
		}
	}
	return true
}

func parseFloatOr(s string, def float64) (float64, error) {
	if s == "" {
		return def, nil
	}
	return strconv.ParseFloat(s, 64)
}

func parseIntOr(s string, def int) (int, error) {
	if s == "" {
		return def, nil
	}
	return strconv.Atoi(s)
}
