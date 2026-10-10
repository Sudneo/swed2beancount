package models

import (
	"errors"
	"os"

	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v3"
)

type Mappings []Mapping

type Mapping struct {
	Contains            string  `yaml:"contains"`
	Field               string  `yaml:"field"`
	Amount              float64 `yaml:"exact_amount"`
	Account             string  `yaml:"account"`
	CreditOnly          bool    `yaml:"credit_only"`
	DebitOnly           bool    `yaml:"debit_only"`
	DescriptionOverride string  `yaml:"desc_override"`
}

func ReadMapping(file string) (Mappings, error) {
	var m Mappings
	yamlFile, err := os.ReadFile(file)
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Fatal("Could not read mapping file")
	}
	err = yaml.Unmarshal(yamlFile, &m)
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Fatal("Could not parse mapping file.")
	}
	return m, nil
}

func (m *Mappings) ValidateMappings(accounts []Account) error {
	for _, mapping := range *m {
		// Validate Field
		if mapping.Field != "" && mapping.Field != "beneficiary" && mapping.Field != "details" {
			log.WithFields(log.Fields{
				"Field": "Field",
				"Value": mapping.Field,
			}).Error("Unrecognized value in mapping")
			return errors.New("Failed to validate mappings")
		}
		// Validate account
		validAccount := false
		for _, a := range accounts {
			if mapping.Account == a.Name {
				validAccount = true
			}
		}
		if !validAccount {
			log.WithFields(log.Fields{
				"Field": "Account",
				"Value": mapping.Account,
			}).Error("Unrecognized value in mapping")
			return errors.New("Failed to validate mappings")
		}
		// Validate amount
		if mapping.Amount < 0 {
			log.WithFields(log.Fields{
				"Field": "Amount",
				"Value": mapping.Amount,
			}).Error("Amount can only be a positive number")
			return errors.New("Failed to validate mappings")
		}
		if mapping.Contains == "" {
			log.WithFields(log.Fields{
				"Field": "Contains",
				"Value": mapping.Contains,
			}).Error("Cannot use empty 'contains' in a text mapping")
			return errors.New("Failed to validate mappings")
		}
	}
	return nil
}
