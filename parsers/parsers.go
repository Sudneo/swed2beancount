package parsers

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/sudneo/swed2beancount/models"
	"github.com/sudneo/swed2beancount/utils"
	log "github.com/sirupsen/logrus"
)

func ParseCSV(file string, csvType string) ([]models.Transaction, error) {
	var transactions []models.Transaction
	records, err := utils.ReadCsvFile(file)
	if err != nil {
		log.Error("Failed to parse CSV")
		return transactions, err
	}
	switch csvType {
	case "swedbank":
		transactions, err := parseSwedbank(records)
		if err != nil {
			return transactions, err
		}
		return transactions, nil
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
		t := models.Transaction{}
		if record[7] == "K" {
			t.Type = "credit"
		} else {
			t.Type = "debit"
		}
		layout := "02.01.2006"
		d, err := time.Parse(layout, record[2])
		if err != nil {
			log.WithFields(log.Fields{
				"Date":   record[2],
				"Error":  err,
				"Layout": layout,
			}).Error("Failed to parse date")
			return transactions, errors.New("Failed to parse CSV")
		}
		t.Date = d
		t.Beneficiary = record[3]
		t.Description = record[4]
		correctedAmount := strings.Replace(record[5], ",", ".", 1)
		t.Amount, err = strconv.ParseFloat(correctedAmount, 64)
		if err != nil {
			log.WithFields(log.Fields{
				"Amount": record[5],
				"Error":  err,
			}).Error("Failed to parse amount")
			return transactions, errors.New("Failed to parse CSV")
		}
		t.Currency = record[6]
		if t.Description == "Turnover" || t.Description == "closing balance" || t.Description == "Accrued interest" {
			log.Debug("Skipping meta-transaction")
			continue
		}
		transactions = append(transactions, t)
	}
	return transactions, nil
}
