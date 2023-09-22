package types

import (
	"fmt"
)

func ValidateEpochIdentifierInterface(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	err := ValidateEpochIdentifierString(v)

	return err
}

func ValidateEpochIdentifierString(s string) error {
	if s == "" {
		return fmt.Errorf("empty distribution epoch identifier: %+v", s)
	}

	return nil
}
