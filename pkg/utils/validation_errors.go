package CustomUtils

import (
	"regexp"
	"uuid"
)

type ValidationError struct{ Msg string }

func (e *ValidationError) Error() string { return e.Msg }
func Invalid(msg string) error           { return &ValidationError{Msg: msg} }

func ValidateUUIDs(ids ...string) error {
	for _, id := range ids {
		if id == "" {
			return Invalid("The id cannot be an empty")
		}
		_, err := uuid.Parse(id)
		if err != nil {
			return Invalid("Not a valid id- " + id)
		}
	}
	return nil
}

func ValidateVRNs(numbers ...string) error {
	standardRegex := regexp.MustCompile(`^[A-Z]{2}[ -]?[0-9]{1,2}[ -]?[A-Z]{1,2}[ -]?[0-9]{4}$`)
	bhSeriesRegex := regexp.MustCompile(`^[0-9]{2}BH[ -]?[0-9]{4}[ -]?[A-HJ-NP-Z]{1,2}$`)

	for _, vrn := range numbers {
		if !(standardRegex.MatchString(vrn) || bhSeriesRegex.MatchString(vrn)) {
			return Invalid("Not a valid VRN- " + vrn)
		}
	}
	return nil
}

func ValidateMobileNumbers(numbers ...string) error {
	pattern := regexp.MustCompile(`^[6-9]\d{9}$`)

	for _, mobNo := range numbers {
		if mobNo == "" {
			return Invalid("Mobile Number cannot be empty")
		}
		if !(pattern.MatchString(mobNo)) {
			return Invalid("Not a valid Mobile Number- " + mobNo)
		}
	}
	return nil
}
