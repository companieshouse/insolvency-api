package utils

import (
	"fmt"
	"strings"
	"time"

	"github.com/companieshouse/chs.go/log"
)

// IsValidDate is a helper function to check if the date supplied
// is after today or before the company was incorporated
func IsValidDate(date string, incorporatedOn string) (bool, error) {
	layout := "2006-01-02"
	today := time.Now()
	suppliedDate, err := time.Parse(layout, date)
	if err != nil {
		log.Error(fmt.Errorf("error when parsing appointedOn to date: [%s]", err))
		return false, err
	}

	// Retrieve only the date portion of the incorporatedOn datetime string
	if idx := strings.Index(incorporatedOn, " "); idx != -1 {
		incorporatedOn = incorporatedOn[:idx]
	}

	incorporatedDate, err := time.Parse(layout, incorporatedOn)
	if err != nil {
		log.Error(fmt.Errorf("error when parsing incorporatedOn to date: [%s]", err))
		return false, err
	}

	// Check if suppliedDate is in the future
	if today.Before(suppliedDate) {
		return false, nil
	}

	// Check if suppliedDate is before company was incorporated
	if suppliedDate.Before(incorporatedDate) {
		return false, nil
	}

	return true, nil
}
