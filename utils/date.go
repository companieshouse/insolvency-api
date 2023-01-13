package utils

import (
	"fmt"
	"strings"
	"time"

	"github.com/companieshouse/chs.go/log"
)

// IsValidDate is a helper function to check if the date supplied
// is after today or before a date benchmark
func IsValidDate(dateToTest string, dateToTestAgainst string) (bool, error) {
	layout := "2006-01-02"
	today := time.Now()
	testDate, err := time.Parse(layout, dateToTest)
	if err != nil {
		log.Error(fmt.Errorf("error when parsing dateToTest: [%s]", err))
		return false, err
	}

	// Retrieve only the dateToTest portion of the dateToTestAgainst datetime string
	if idx := strings.Index(dateToTestAgainst, " "); idx != -1 {
		dateToTestAgainst = dateToTestAgainst[:idx]
	}

	dateBenchmark, err := time.Parse(layout, dateToTestAgainst)
	if err != nil {
		log.Error(fmt.Errorf("error when parsing dateToTestAgainst to dateToTest: [%s]", err))
		return false, err
	}

	// Check if testDate is in the future
	if today.Before(testDate) {
		return false, nil
	}

	// Check if testDate is before company was incorporated
	if testDate.Before(dateBenchmark) {
		return false, nil
	}

	return true, nil
}
