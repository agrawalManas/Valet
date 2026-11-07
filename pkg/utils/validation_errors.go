package CustomUtils

import "uuid"

type ValidationError struct{ Msg string }

func (e *ValidationError) Error() string { return e.Msg }
func Invalid(msg string) error           { return &ValidationError{Msg: msg} }

func ValidateUUIDS(ids ...string) error {
	for _, id := range ids {
		_, err := uuid.Parse(id)
		if err != nil {
			return Invalid("Not a valid id- " + id)
		}
	}
	return nil
}
