package models

import (
	"errors"
	"io/ioutil"
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type Mappings []Mapping

type Mapping struct {
	Type                string  `yaml:"type"`
	Contains            string  `yaml:"contains"`
	Field               string  `yaml:"field"`
	DateBegin           string  `yaml:"date_begin"`
	DateEnd             string  `yaml:"date_end"`
	Amount              float64 `yaml:"exact_amount"`
	Account             string  `yaml:"account"`
	CreditOnly          bool    `yaml:"credit_only"`
	DebitOnly           bool    `yaml:"debit_only"`
	DescriptionOverride string  `yaml:"desc_override"`
	InternalDateBegin   time.Time
	InternalDateEnd     time.Time
}

func ReadMapping(file string) (Mappings, error) {
	var m Mappings
	yamlFile, err := ioutil.ReadFile(file)
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
		// Validate type
		if mapping.Type != "text" && mapping.Type != "date" && mapping.Type != "amount" {
			log.WithFields(log.Fields{
				"Field": "Type",
				"Value": mapping.Type,
			}).Error("Unrecognized value in mapping")
			return errors.New("Failed to validate mappings")
		}
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
		if mapping.Type == "date" {
			// Validate dates
			layout := "2006-01-02"
			d, err := time.Parse(layout, mapping.DateBegin)
			if err != nil {
				log.WithFields(log.Fields{
					"Field": "DateBegin",
					"Value": mapping.DateBegin,
				}).Error("Invalid date specified")
				return errors.New("Failed to validate mappings")
			}
			mapping.InternalDateBegin = d
			d, err = time.Parse(layout, mapping.DateEnd)
			if err != nil {
				log.WithFields(log.Fields{
					"Field": "DateEnd",
					"Value": mapping.DateEnd,
				}).Error("Invalid date specified")
				return errors.New("Failed to validate mappings")
			}
			mapping.InternalDateEnd = d
		}
		if mapping.Type == "text" && mapping.Contains == "" {
			log.WithFields(log.Fields{
				"Field": "Contains",
				"Value": mapping.Contains,
			}).Error("Cannot use empty 'contains' in a text mapping")
			return errors.New("Failed to validate mappings")
		} else if mapping.Type == "date" && (mapping.DateBegin == "" || mapping.DateEnd == "") {
			log.WithFields(log.Fields{
				"Field": "Dates",
			}).Error("Cannot use empty 'dates' in a date mapping")
			return errors.New("Failed to validate mappings")
		}

	}
	return nil
}
