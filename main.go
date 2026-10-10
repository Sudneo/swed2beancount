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
)

const uncategorizedAccount = "UNCATEGORIZED_ACCOUNT"

var outputTemplate = template.Must(template.New("output").Parse(`{{ range . }}
{{ .Date.Format "2006-01-02" }} * "{{ .Description }}"
	{{ .CreditAccount }} {{ .Amount }} {{ .Currency }}
	{{ .DebitAccount }} -{{ .Amount }} {{ .Currency }} 
{{ end }}
`))

func processTransactions(transactions []models.Transaction, defaultAccount string, mappings models.Mappings) []models.Transaction {
	processed := make([]models.Transaction, 0, len(transactions))
	for _, t := range transactions {
		account, desc := t.Categorize(mappings)
		if account == "" {
			account = uncategorizedAccount
		}
		if desc != "" {
			t.Description = desc
		}
		if t.Type == "debit" {
			t.DebitAccount, t.CreditAccount = defaultAccount, account
		} else {
			t.DebitAccount, t.CreditAccount = account, defaultAccount
		}
		processed = append(processed, t)
	}
	return processed
}

// filterTransactionsByDate keeps the transactions within [start, end].
// A zero start or end leaves that side of the range open.
func filterTransactionsByDate(start, end time.Time, transactions []models.Transaction) []models.Transaction {
	var filtered []models.Transaction
	for _, t := range transactions {
		if (start.IsZero() || !t.Date.Before(start)) && (end.IsZero() || !t.Date.After(end)) {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

func generateOutput(outfile string, transactions []models.Transaction) error {
	f, err := os.Create(outfile)
	if err != nil {
		return err
	}
	if err := outputTemplate.Execute(f, transactions); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

func run(config models.Config) error {
	if config.DefaultAccount == "" {
		return errors.New("the default account should be provided either through the -a flag or through the config file")
	}
	accounts, err := models.ReadAccounts(config.BeancountLedger)
	if err != nil {
		return err
	}
	log.WithField("Accounts", len(accounts)).Info("Found beancount accounts")
	m, err := models.ReadMapping(config.MappingFile)
	if err != nil {
		return err
	}
	if err := m.Validate(accounts); err != nil {
		return err
	}
	log.Info("Mappings validated")
	transactions, err := parsers.ParseCSV(config.CSVFile)
	if err != nil {
		return err
	}
	log.WithField("Transactions", len(transactions)).Info("Found and parsed transactions")
	transactions = filterTransactionsByDate(config.StartDate.Time, config.EndDate.Time, transactions)
	if len(transactions) == 0 {
		log.Warning("No transactions to write. Check the from_date/until_date range.")
	}
	return generateOutput(config.OutputFile, processTransactions(transactions, config.DefaultAccount, m))
}

func main() {
	configFile := flag.String("config", "config.yaml", "The config file to use")
	account := flag.String("a", "", "The default beancount account to use for the CSV transactions")
	verbose := flag.Bool("v", false, "Set logging to Debug level")
	flag.Parse()

	if *verbose {
		log.SetLevel(log.DebugLevel)
	}
	c, err := models.ReadConfig(*configFile)
	if err != nil {
		log.Fatal(err)
	}
	// The default account flag overrides the config one
	if *account != "" {
		c.DefaultAccount = *account
	}
	if err := run(c); err != nil {
		log.Fatal(err)
	}
}
