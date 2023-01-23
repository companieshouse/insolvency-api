package utils

import (
	"fmt"
	"strings"
	"time"

	"github.com/companieshouse/chs.go/log"
)

// IsDateBetweenIncorporationAndNow is a helper function to check if the date supplied
// is in the future or before the incorporation date
func IsDateBetweenIncorporationAndNow(dateToTest string, incorporationDate string) (bool, error) {
	validDateToTest, err := ValidateDate(dateToTest)
	if err != nil {
		return false, err
	}

	validIncorporationDate, err := ValidateDate(incorporationDate)
	if err != nil {
		return false, err
	}

	today := time.Now()

	// Check if date is in the future
	if today.Before(validDateToTest) {
		return false, nil
	}

	// Check if date is before company was incorporated
	if validDateToTest.Before(validIncorporationDate) {
		return false, nil
	}

	return true, nil
}

// ValidateDate check date is in a valid date format
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

// IsDateBeforeDate check if before date is earlier than after date
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
