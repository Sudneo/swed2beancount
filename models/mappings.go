package models

import (
	"bytes"
	"fmt"
	"os"
	"slices"

	yaml "gopkg.in/yaml.v3"
)

type Mappings []Mapping

type Mapping struct {
	Contains            string `yaml:"contains"`
	Field               string `yaml:"field"`
	Account             string `yaml:"account"`
	DescriptionOverride string `yaml:"desc_override"`
}

func ReadMapping(file string) (Mappings, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("reading mappings: %w", err)
	}
	m, err := parseMappings(data)
	if err != nil {
		return nil, fmt.Errorf("parsing mappings %s: %w", file, err)
	}
	return m, nil
}

func parseMappings(yamlData []byte) (Mappings, error) {
	var m Mappings
	// Reject unknown keys, so that rules using removed options
	// (e.g., the old "type", "amount" or "date_begin") are reported.
	decoder := yaml.NewDecoder(bytes.NewReader(yamlData))
	decoder.KnownFields(true)
	if err := decoder.Decode(&m); err != nil {
		return nil, err
	}
	return m, nil
}

func (m Mappings) Validate(accounts []string) error {
	for i, mapping := range m {
		switch {
		case mapping.Contains == "":
			return fmt.Errorf("mapping %d: 'contains' cannot be empty", i+1)
		case mapping.Field != "" && mapping.Field != "beneficiary" && mapping.Field != "details":
			return fmt.Errorf("mapping %d: unrecognized field %q (want beneficiary or details)", i+1, mapping.Field)
		case !slices.Contains(accounts, mapping.Account):
			return fmt.Errorf("mapping %d: account %q is not opened in the ledger", i+1, mapping.Account)
		}
	}
	return nil
}
