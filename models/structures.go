package models

import (
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/text/language"
	"golang.org/x/text/search"
)

type Account struct {
	Name string
}

type Transaction struct {
	Type          string
	CreditAccount Account
	DebitAccount  Account
	Amount        float64
	Description   string
	Currency      string
	Beneficiary   string
	Date          time.Time
}

func stringMatch(str string, substr string) bool {
	m := search.New(language.Und, search.IgnoreCase)
	start, _ := m.IndexString(str, substr)
	if start == -1 {
		return false
	}
	return true
}

func (t *Transaction) Categorize(m Mappings) (string, string) {

	// Returns the account of the first mapping match, if no match, returns an empty account
	matchAccount := ""
	descriptionOverride := ""
	for _, mapping := range m {
		matched := false
		switch mapping.Field {
		case "beneficiary":
			matched = stringMatch(t.Beneficiary, mapping.Contains)
		case "details":
			matched = stringMatch(t.Description, mapping.Contains)
		default:
			matched = stringMatch(t.Beneficiary, mapping.Contains) || stringMatch(t.Description, mapping.Contains)
		}
		if matched {
			log.WithFields(log.Fields{
				"Match": mapping.Account,
				"Field": mapping.Field,
			}).Debug("Found match for transaction")
			return mapping.Account, mapping.DescriptionOverride
		}
	}
	return matchAccount, descriptionOverride
}
