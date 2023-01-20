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

// ValidateDate validate date is in a valid date format
func ValidateDate(dateToTest string) (time.Time, error) {
	layout := "2006-01-02"
	var validDateStr string = dateToTest

	if idx := strings.Index(dateToTest, " "); idx != -1 {
		validDateStr = dateToTest[:idx]
	}

	validDate, err := time.Parse(layout, validDateStr)
	if err != nil {
		log.Error(fmt.Errorf("error when parsing dateToTest: [%s]", err))
		return time.Time{}, err
	}

	return validDate, nil
}

// IsDateInFuture check if the date supplied is after today
func IsDateInFuture(dateToTest string) (bool, error) {
	validDate, err := ValidateDate(dateToTest)
	if err != nil {
		return false, err
	}

	today := time.Now()

	return today.Before(validDate), nil
}

// IsDateBeforeDate check if before date is before after date
func IsDateBeforeDate(beforeDate string, afterDate string) (bool, error) {
	validBeforeDate, err := ValidateDate(beforeDate)
	if err != nil {
		return false, err
	}

	validAfterDate, err := ValidateDate(afterDate)
	if err != nil {
		return false, err
	}

	return validBeforeDate.Before(validAfterDate), nil
}
