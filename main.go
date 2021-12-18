package main

import (
	"encoding/csv"
	"errors"
	"flag"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	"git.home.lab/daniele/swed2beancount/models"
	"git.home.lab/daniele/swed2beancount/utils"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

func readCsvFile(filePath string) ([][]string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		log.WithFields(log.Fields{
			"File":  filePath,
			"Error": err,
		}).Error("Unable to read input file")
		return [][]string{}, errors.New("Failed to read CSV")
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	csvReader.Comma = ';'
	records, err := csvReader.ReadAll()
	if err != nil {
		log.WithFields(log.Fields{
			"File":  filePath,
			"Error": err,
		}).Error("Unable to read input file")
		return [][]string{}, errors.New("Failed to parse CSV")
	}
	return records, nil
}

func ParseCSV(file string) ([]models.Transaction, error) {
	var transactions []models.Transaction
	records, err := readCsvFile(file)
	if err != nil {
		log.Error("Failed to parse CSV")
		return transactions, err
	}
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

func ProcessTransactions(t []models.Transaction, defaultAccount models.Account, mappings models.Mappings) []models.Transaction {
	var processedTransactions []models.Transaction
	for _, transaction := range t {
		match, desc := transaction.Categorize(mappings)
		var matchAccount models.Account
		if match != "" {
			matchAccount.Name = match
		} else {
			matchAccount.Name = "UNCATEGORIZED_ACCOUNT"
		}
		processed := transaction
		if desc != "" {
			processed.Description = desc
		}
		if transaction.Type == "debit" {
			processed.DebitAccount = defaultAccount
			processed.CreditAccount = matchAccount
		} else {
			processed.DebitAccount = matchAccount
			processed.CreditAccount = defaultAccount
		}
		processedTransactions = append(processedTransactions, processed)
	}
	return processedTransactions
}

func FilterTransactionsByDate(startDate time.Time, endDate time.Time, transactions []models.Transaction) []models.Transaction {
	var t []models.Transaction
	for _, transaction := range transactions {
		if (transaction.Date.After(startDate) || transaction.Date.Equal(startDate)) && (transaction.Date.Before(endDate) || transaction.Date.Equal(endDate)) {
			t = append(t, transaction)
		}
	}
	return t
}

func GenerateOutput(outfile string, transactions []models.Transaction) error {
	// 2020-01-03 * "HTB subscription"
	// Expenses:EE:Personal:Study 11.96 EUR
	// Assets:EE:Bank:Personal:Checking -11.96 EUR
	PageData := `{{ range . }}
{{ .Date.Format "2006-01-02" }} * "{{ .Description }}"
	{{ .CreditAccount.Name }} {{ .Amount }} {{ .Currency }}
	{{ .DebitAccount.Name }} -{{ .Amount }} {{ .Currency }} 
{{ end }}
`

	t := template.Must(template.New("Config").Parse(PageData))

	f, err := os.Create(outfile)
	if err != nil {
		log.WithFields(log.Fields{
			"File":  outfile,
			"Error": err,
		}).Error("Failed to create output file")
		return errors.New("Failed to render template: cannot create the output file")
	}
	t.Execute(f, transactions)
	return nil
}

func Run(config models.Config, defaultAccount models.Account) error {
	accounts := utils.GetAccounts(config.BeancountLedger)
	log.WithFields(log.Fields{
		"Accounts": len(accounts),
	}).Info("Found beancount accounts")
	m, err := models.ReadMapping(config.MappingFile)
	if err != nil {
		return err
	}
	m.ValidateMappings(accounts)
	log.Info("Mappings validated")
	transactions, err := ParseCSV(config.CSVFile)
	if err != nil {
		return err
	}
	log.WithFields(log.Fields{
		"Transactions": len(transactions),
	}).Info("Found and parsed transactions")
	transactions = ProcessTransactions(transactions, defaultAccount, m)
	if config.StartDate != "" && config.EndDate != "" {
		transactions = FilterTransactionsByDate(config.InternalStartDate, config.InternalEndDate, transactions)
	}
	err = GenerateOutput(config.OutputFile, transactions)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	configPtr := flag.String("config", "config.yaml", "The config file to use")
	accountPtr := flag.String("a", "Assets:EE:Bank:Personal:Checking", "The default beancount account to use for the CSV transactions")
	verbosePtr := flag.Bool("v", false, "Set logging to Debug level")

	flag.Parse()

	if *verbosePtr {
		log.SetLevel(log.DebugLevel)
	}

	acc := models.Account{Name: *accountPtr}
	c, err := models.ReadConfig(*configPtr)
	if err != nil {
		log.Error(err)
		return
	}
	err = Run(c, acc)

	if err != nil {
		log.Error(err)
	}
}
