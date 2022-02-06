package utils

import (
	"bufio"
	"encoding/csv"
	"errors"
	"os"
	"regexp"

	"git.home.lab/daniele/swed2beancount/models"
	log "github.com/sirupsen/logrus"
)

func GetAccounts(BeancountLedger string) []models.Account {
	var accounts []models.Account
	r := regexp.MustCompile("open (?P<account>[a-zA-Z]+(:([a-zA-Z]*))*)")
	file, err := os.Open(BeancountLedger)
	if err != nil {
		log.WithFields(log.Fields{
			"Error": err,
			"File":  BeancountLedger,
		}).Error("Failed to read Beancount file to source accounts")
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		line := scanner.Text()
		matches := r.FindStringSubmatch(line)
		if matches == nil {
			continue
		} else {
			accountIndex := r.SubexpIndex("account")
			a := models.Account{Name: matches[accountIndex]}
			accounts = append(accounts, a)
		}
	}
	return accounts
}

func ReadCsvFile(filePath string) ([][]string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		log.WithFields(log.Fields{
			"File":  filePath,
			"Error": err,
		}).Error("Unable to read input file")
		return [][]string{}, errors.New("Failed to read CSV")
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	csvReader.Comma = ';'
	records, err := csvReader.ReadAll()
	if err != nil {
		log.WithFields(log.Fields{
			"File":  filePath,
			"Error": err,
		}).Error("Unable to read input file")
		return [][]string{}, errors.New("Failed to parse CSV")
	}
	return records, nil
}
