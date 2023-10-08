package main

import (
	"errors"
	"flag"
	"os"
	"text/template"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/sudneo/swed2beancount/models"
	"github.com/sudneo/swed2beancount/parsers"
	"github.com/sudneo/swed2beancount/utils"
)

func init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
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

func Run(config models.Config) error {
	if config.DefaultAccount == "" {
		log.Error("The Default account should be provided either through command-line flag or through the config file.")
		panic("No default account")
	}
	defaultAccount := models.Account{Name: config.DefaultAccount}
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
	transactions, err := parsers.ParseCSV(config.CSVFile, config.CSVType)
	if err != nil {
		return err
	}
	log.WithFields(log.Fields{
		"Transactions": len(transactions),
	}).Info("Found and parsed transactions")
	if config.StartDate != "" && config.EndDate != "" {
		transactions = FilterTransactionsByDate(config.InternalStartDate, config.InternalEndDate, transactions)
		if len(transactions) == 0 {
			log.Warning("The Start/End date specified in the configuration file are filtering out all the transactions.")
		}
	}
	transactions = ProcessTransactions(transactions, defaultAccount, m)
	err = GenerateOutput(config.OutputFile, transactions)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	configPtr := flag.String("config", "config.yaml", "The config file to use")
	accountPtr := flag.String("a", "", "The default beancount account to use for the CSV transactions")
	verbosePtr := flag.Bool("v", false, "Set logging to Debug level")

	flag.Parse()

	if *verbosePtr {
		log.SetLevel(log.DebugLevel)
	}
	c, err := models.ReadConfig(*configPtr)
	if err != nil {
		log.Error(err)
		return
	}
	// The default account flag should override the config one if it exists
	if *accountPtr != "" {
		c.DefaultAccount = *accountPtr
	}
	err = Run(c)

	if err != nil {
		log.Error(err)
	}
}
