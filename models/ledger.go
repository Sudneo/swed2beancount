package models

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
)

// openDirective matches a Beancount "open" directive and captures the account name.
var openDirective = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}\s+open\s+([\p{Lu}][\p{L}\p{N}-]*(?::[\p{Lu}\p{N}][\p{L}\p{N}-]*)+)`)

// ReadAccounts returns the accounts opened in a Beancount ledger.
func ReadAccounts(ledger string) ([]string, error) {
	f, err := os.Open(ledger)
	if err != nil {
		return nil, fmt.Errorf("reading ledger: %w", err)
	}
	defer f.Close()

	var accounts []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if m := openDirective.FindStringSubmatch(scanner.Text()); m != nil {
			accounts = append(accounts, m[1])
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading ledger %s: %w", ledger, err)
	}
	return accounts, nil
}
