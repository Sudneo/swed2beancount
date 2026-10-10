package parsers

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/sudneo/swed2beancount/models"
)

// Swedbank statement row types.
const (
	rowOpeningBalance = "10"
	rowTransaction    = "20"
	rowClosingBalance = "86"
)

// ParseStatements parses all the Swedbank CSV statements matching the given
// paths or glob patterns.
func ParseStatements(patterns []string) ([]models.Transaction, []models.Balance, error) {
	var transactions []models.Transaction
	var balances []models.Balance
	for _, pattern := range patterns {
		files, err := filepath.Glob(pattern)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid statement pattern %q: %w", pattern, err)
		}
		if len(files) == 0 {
			return nil, nil, fmt.Errorf("no statement matches %q", pattern)
		}
		for _, file := range files {
			t, b, err := ParseCSV(file)
			if err != nil {
				return nil, nil, err
			}
			log.WithFields(log.Fields{
				"File":         file,
				"Transactions": len(t),
				"Balances":     len(b),
			}).Info("Parsed statement")
			transactions = append(transactions, t...)
			balances = append(balances, b...)
		}
	}
	return transactions, balances, nil
}

// ParseCSV parses a Swedbank CSV statement export.
func ParseCSV(file string) ([]models.Transaction, []models.Balance, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, nil, fmt.Errorf("reading statement: %w", err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.Comma = ';'
	records, err := r.ReadAll()
	if err != nil {
		return nil, nil, fmt.Errorf("parsing statement %s: %w", file, err)
	}
	transactions, balances, err := parseSwedbank(records)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing statement %s: %w", file, err)
	}
	return transactions, balances, nil
}

func parseSwedbank(records [][]string) ([]models.Transaction, []models.Balance, error) {
	var transactions []models.Transaction
	var balances []models.Balance
	if len(records) == 0 {
		return nil, nil, nil
	}
	// Skip the header
	for i, record := range records[1:] {
		if len(record) < 8 {
			return nil, nil, fmt.Errorf("row %d: expected at least 8 fields, got %d", i+2, len(record))
		}
		log.WithFields(log.Fields{
			"Account":     record[0],
			"RowType":     record[1],
			"Date":        record[2],
			"Beneficiary": record[3],
			"Details":     record[4],
			"Amount":      record[5],
			"Currency":    record[6],
			"Type":        record[7],
		}).Debug("New row")
		rowType := record[1]
		if rowType != rowTransaction && rowType != rowOpeningBalance && rowType != rowClosingBalance {
			log.Debug("Skipping meta-row")
			continue
		}
		date, err := time.Parse("02.01.2006", record[2])
		if err != nil {
			return nil, nil, fmt.Errorf("row %d: invalid date %q: %w", i+2, record[2], err)
		}
		amount, err := parseAmount(record[5])
		if err != nil {
			return nil, nil, fmt.Errorf("row %d: invalid amount %q: %w", i+2, record[5], err)
		}
		credit := record[7] == "K"

		switch rowType {
		case rowTransaction:
			t := models.Transaction{
				ClientAccount: record[0],
				Type:          "debit",
				Date:          date,
				Beneficiary:   record[3],
				Description:   record[4],
				Amount:        amount,
				Currency:      record[6],
			}
			if credit {
				t.Type = "credit"
			}
			transactions = append(transactions, t)
		case rowOpeningBalance, rowClosingBalance:
			if !credit {
				amount = -amount
			}
			// Beancount checks balances at the beginning of the day, so the
			// closing balance of a day is asserted on the following one.
			if rowType == rowClosingBalance {
				date = date.AddDate(0, 0, 1)
			}
			balances = append(balances, models.Balance{
				ClientAccount: record[0],
				Date:          date,
				Amount:        amount,
				Currency:      record[6],
			})
		}
	}
	return transactions, balances, nil
}

func parseAmount(s string) (models.Cents, error) {
	f, err := strconv.ParseFloat(strings.Replace(s, ",", ".", 1), 64)
	if err != nil {
		return 0, err
	}
	return models.Cents(math.Round(f * 100)), nil
}
