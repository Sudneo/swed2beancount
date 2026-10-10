package main

import (
	"cmp"
	"flag"
	"fmt"
	"os"
	"slices"
	"text/template"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/sudneo/swed2beancount/models"
	"github.com/sudneo/swed2beancount/parsers"
)

const uncategorizedAccount = "UNCATEGORIZED_ACCOUNT"

var outputTemplate = template.Must(template.New("output").Parse(`{{ range .Transactions }}
{{ .Date.Format "2006-01-02" }} * "{{ .Description }}"
	{{ .CreditAccount }} {{ .Amount }} {{ .Currency }}
	{{ .DebitAccount }} -{{ .Amount }} {{ .Currency }}
{{ end }}
{{ range .Balances }}{{ .Date.Format "2006-01-02" }} balance {{ .Account }} {{ .Amount }} {{ .Currency }}
{{ end }}`))

func resolveAccount(accounts map[string]string, iban string) (string, error) {
	account, ok := accounts[iban]
	if !ok {
		return "", fmt.Errorf("no account configured for %s", iban)
	}
	return account, nil
}

func processTransactions(transactions []models.Transaction, accounts map[string]string, mappings models.Mappings) ([]models.Transaction, error) {
	processed := make([]models.Transaction, 0, len(transactions))
	for _, t := range transactions {
		own, err := resolveAccount(accounts, t.ClientAccount)
		if err != nil {
			return nil, err
		}
		account, desc := t.Categorize(mappings)
		if account == "" {
			account = uncategorizedAccount
		}
		if desc != "" {
			t.Description = desc
		}
		if t.Type == "debit" {
			t.DebitAccount, t.CreditAccount = own, account
		} else {
			t.DebitAccount, t.CreditAccount = account, own
		}
		processed = append(processed, t)
	}
	return processed, nil
}

// removeDuplicates drops the transactions with the same date, amount and
// accounts as a previous one from a different statement account. This is the
// case of transfers between own accounts, which appear in both statements.
func removeDuplicates(transactions []models.Transaction) []models.Transaction {
	var unique []models.Transaction
	for _, t := range transactions {
		if slices.ContainsFunc(unique, func(u models.Transaction) bool {
			return u.ClientAccount != t.ClientAccount && u.Date.Equal(t.Date) && u.Amount == t.Amount &&
				u.DebitAccount == t.DebitAccount && u.CreditAccount == t.CreditAccount
		}) {
			log.WithFields(log.Fields{
				"Date":        t.Date.Format("2006-01-02"),
				"Description": t.Description,
				"Amount":      t.Amount,
			}).Info("Skipping duplicate transaction")
			continue
		}
		unique = append(unique, t)
	}
	return unique
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

// processBalances resolves the beancount account of the balances and keeps
// the ones on the edges of [start, end]. Since balances are checked at the
// beginning of the day, the closing balance of end is asserted on end+1.
func processBalances(start, end time.Time, balances []models.Balance, accounts map[string]string) ([]models.Balance, error) {
	if !end.IsZero() {
		end = end.AddDate(0, 0, 1)
	}
	var processed []models.Balance
	for _, b := range balances {
		account, err := resolveAccount(accounts, b.ClientAccount)
		if err != nil {
			return nil, err
		}
		b.Account = account
		if (start.IsZero() || !b.Date.Before(start)) && (end.IsZero() || !b.Date.After(end)) {
			processed = append(processed, b)
		}
	}
	slices.SortStableFunc(processed, func(a, b models.Balance) int {
		return cmp.Or(a.Date.Compare(b.Date), cmp.Compare(a.Account, b.Account))
	})
	return processed, nil
}

func generateOutput(outfile string, transactions []models.Transaction, balances []models.Balance) error {
	f, err := os.Create(outfile)
	if err != nil {
		return err
	}
	data := struct {
		Transactions []models.Transaction
		Balances     []models.Balance
	}{transactions, balances}
	if err := outputTemplate.Execute(f, data); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

func run(config models.Config) error {
	accounts, err := models.ReadAccounts(config.BeancountLedger)
	if err != nil {
		return err
	}
	log.WithField("Accounts", len(accounts)).Info("Found beancount accounts")
	if err := config.Validate(accounts); err != nil {
		return err
	}
	m, err := models.ReadMapping(config.MappingFile)
	if err != nil {
		return err
	}
	if err := m.Validate(accounts); err != nil {
		return err
	}
	log.Info("Mappings validated")
	transactions, balances, err := parsers.ParseStatements(config.Statements)
	if err != nil {
		return err
	}
	log.WithField("Transactions", len(transactions)).Info("Found and parsed transactions")
	transactions, err = processTransactions(transactions, config.Accounts, m)
	if err != nil {
		return err
	}
	transactions = filterTransactionsByDate(config.StartDate.Time, config.EndDate.Time, transactions)
	if len(transactions) == 0 {
		log.Warning("No transactions to write. Check the from_date/until_date range.")
	}
	transactions = removeDuplicates(transactions)
	slices.SortStableFunc(transactions, func(a, b models.Transaction) int {
		return a.Date.Compare(b.Date)
	})
	balances, err = processBalances(config.StartDate.Time, config.EndDate.Time, balances, config.Accounts)
	if err != nil {
		return err
	}
	return generateOutput(config.OutputFile, transactions, balances)
}

func main() {
	configFile := flag.String("config", "config.yaml", "The config file to use")
	verbose := flag.Bool("v", false, "Set logging to Debug level")
	flag.Parse()

	if *verbose {
		log.SetLevel(log.DebugLevel)
	}
	c, err := models.ReadConfig(*configFile)
	if err != nil {
		log.Fatal(err)
	}
	if err := run(c); err != nil {
		log.Fatal(err)
	}
}
