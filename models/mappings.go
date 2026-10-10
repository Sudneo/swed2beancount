package models

import (
	"bytes"
	"errors"
	"os"

	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v3"
)

type Mappings []Mapping

type Mapping struct {
	Contains            string `yaml:"contains"`
	Field               string `yaml:"field"`
	Account             string `yaml:"account"`
	CreditOnly          bool   `yaml:"credit_only"`
	DebitOnly           bool   `yaml:"debit_only"`
	DescriptionOverride string `yaml:"desc_override"`
}

func ReadMapping(file string) (Mappings, error) {
	var m Mappings
	yamlFile, err := os.ReadFile(file)
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Error("Could not read mapping file")
		return m, errors.New("Failed to read mapping file")
	}
	m, err = parseMappings(yamlFile)
	return m, err
}

func parseMappings(yamlData []byte) (Mappings, error) {
	var m Mappings
	// Reject unknown keys, so that rules using removed options
	// (e.g., the old "type", "amount" or "date_begin") are reported.
	decoder := yaml.NewDecoder(bytes.NewReader(yamlData))
	decoder.KnownFields(true)
	err := decoder.Decode(&m)
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Error("Could not parse mapping file.")
		return m, errors.New("Failed to parse mapping file")
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
		if mapping.Contains == "" {
			log.WithFields(log.Fields{
				"Field": "Contains",
				"Value": mapping.Contains,
			}).Error("Cannot use empty 'contains' in a mapping")
			return errors.New("Failed to validate mappings")
		}
	}
	return nil
}
