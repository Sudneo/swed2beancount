package parsers

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/sudneo/swed2beancount/models"
)

// ParseCSV parses a Swedbank CSV statement export.
func ParseCSV(file string) ([]models.Transaction, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("reading statement: %w", err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.Comma = ';'
	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("parsing statement %s: %w", file, err)
	}
	transactions, err := parseSwedbank(records)
	if err != nil {
		return nil, fmt.Errorf("parsing statement %s: %w", file, err)
	}
	return transactions, nil
}

func parseSwedbank(records [][]string) ([]models.Transaction, error) {
	var transactions []models.Transaction
	for _, record := range records[2:] {
		log.WithFields(log.Fields{
			"Date":        record[2],
			"Beneficiary": record[3],
			"Details":     record[4],
			"Amount":      record[5],
			"Currency":    record[6],
			"Type":        record[7],
		}).Debug("New Transaction")
		t := models.Transaction{
			Type:        "debit",
			Beneficiary: record[3],
			Description: record[4],
			Currency:    record[6],
		}
		if record[7] == "K" {
			t.Type = "credit"
		}
		var err error
		t.Date, err = time.Parse("02.01.2006", record[2])
		if err != nil {
			return nil, fmt.Errorf("invalid date %q: %w", record[2], err)
		}
		t.Amount, err = strconv.ParseFloat(strings.Replace(record[5], ",", ".", 1), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid amount %q: %w", record[5], err)
		}
		if t.Description == "Turnover" || t.Description == "closing balance" || t.Description == "Accrued interest" {
			log.Debug("Skipping meta-transaction")
			continue
		}
		transactions = append(transactions, t)
	}
	return transactions, nil
}
