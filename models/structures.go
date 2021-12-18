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
		// Text Matching
		if mapping.Type == "text" {
			matchBeneficiary := stringMatch(t.Beneficiary, mapping.Contains)
			matchDetails := stringMatch(t.Description, mapping.Contains)
			if mapping.Field == "beneficiary" && matchBeneficiary {
				// Match beneficiary text
				log.WithFields(log.Fields{
					"Match": mapping.Account,
					"Type":  "Beneficiary",
				}).Debug("Found match for transaction")
				return mapping.Account, mapping.DescriptionOverride
			}
			if mapping.Field == "details" && matchDetails {
				log.WithFields(log.Fields{
					"Match": mapping.Account,
					"Type":  "Details",
				}).Debug("Found match for transaction")
				return mapping.Account, mapping.DescriptionOverride
			} else {
				if matchDetails || matchBeneficiary {
					log.WithFields(log.Fields{
						"Match": mapping.Account,
						"Type":  "Text",
					}).Debug("Found match for transaction")
					return mapping.Account, mapping.DescriptionOverride
				}
			}
		}
		// Amount Matching
		if mapping.Type == "amount" {
			if t.Amount == mapping.Amount {
				log.WithFields(log.Fields{
					"Match": mapping.Account,
					"Type":  "Amount",
				}).Debug("Found match for transaction")
				return mapping.Account, mapping.DescriptionOverride
			}
		}
		// DateMatching
		if mapping.Type == "date" {
			if t.Date.After(mapping.InternalDateBegin) && t.Date.Before(mapping.InternalDateEnd) {
				log.WithFields(log.Fields{
					"Match": mapping.Account,
					"Type":  "Date",
				}).Debug("Found match for transaction")
				return mapping.Account, mapping.DescriptionOverride
			}
		}
	}
	return matchAccount, descriptionOverride
}
