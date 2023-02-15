package utils

import (
	"encoding/json"
)

// ConvertStringToMapObjectAndStringList is a helper function to convert string to map object
func ConvertStringToMapObjectAndStringList(mapString string) (map[string]string, []string, error) {
	var mapResource map[string]string
	var stringPractitionerIdsArray []string

	err := json.Unmarshal([]byte(mapString), &mapResource)
	if err != nil {
		return nil, nil, err
	}

	for key := range mapResource {
		stringPractitionerIdsArray = append(stringPractitionerIdsArray, key)
	}

	return mapResource, stringPractitionerIdsArray, nil
}

// ConvertMapToString is a helper function to convert map[string]string to string
func ConvertMapToString(mapString map[string]string) (string, error) {
	stringPractitionerLinks, err := json.Marshal(mapString)
	if err != nil {
		return "", err
	}

	return string(stringPractitionerLinks), nil
}

// ConvertMapToStringArray is a helper function to convert map[string]string to string array
func ConvertMapToStringArray(mapString map[string]string) []string {
	var practitionerIds []string
	for practitionerId := range mapString {
		practitionerIds = append(practitionerIds, practitionerId)
	}

	return practitionerIds
}
