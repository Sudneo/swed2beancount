package models

import (
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

type Transaction struct {
	Type          string
	CreditAccount string
	DebitAccount  string
	Amount        float64
	Description   string
	Currency      string
	Beneficiary   string
	Date          time.Time
}

func containsFold(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// Categorize returns the account and description override of the first
// matching mapping, or empty strings if no mapping matches.
func (t Transaction) Categorize(m Mappings) (string, string) {
	for _, mapping := range m {
		var matched bool
		switch mapping.Field {
		case "beneficiary":
			matched = containsFold(t.Beneficiary, mapping.Contains)
		case "details":
			matched = containsFold(t.Description, mapping.Contains)
		default:
			matched = containsFold(t.Beneficiary, mapping.Contains) || containsFold(t.Description, mapping.Contains)
		}
		if matched {
			log.WithFields(log.Fields{
				"Match": mapping.Account,
				"Field": mapping.Field,
			}).Debug("Found match for transaction")
			return mapping.Account, mapping.DescriptionOverride
		}
	}
	return "", ""
}
