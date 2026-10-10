package models

import (
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// Cents is a monetary amount in hundredths, to avoid float rounding issues.
type Cents int64

func (c Cents) String() string {
	sign := ""
	if c < 0 {
		sign, c = "-", -c
	}
	return fmt.Sprintf("%s%d.%02d", sign, c/100, c%100)
}

type Transaction struct {
	// ClientAccount is the IBAN of the account the statement belongs to.
	ClientAccount string
	Type          string
	CreditAccount string
	DebitAccount  string
	Amount        Cents
	Description   string
	Currency      string
	Beneficiary   string
	Date          time.Time
}

// Balance is a balance assertion. Date is the date of the beancount
// directive, which is checked at the beginning of that day.
type Balance struct {
	ClientAccount string
	Account       string
	Date          time.Time
	Amount        Cents
	Currency      string
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
