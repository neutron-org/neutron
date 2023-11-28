package utils

import (
	"fmt"
)

func JoinErrors(parentError error, errs ...error) error {
	// TODO: switch to errors.Join when we bump to golang 1.20
	fullError := fmt.Errorf("errors: %w", parentError)
	for _, err := range errs {
		fullError = fmt.Errorf("%w", err)
	}

	return fullError
}
