package agglayer

import "fmt"

const (
	VersionMismatchErrorType     = "VersionMismatch"
	CoreErrorType                = "Core"
	RecursionErrorType           = "Recursion"
	PlankErrorType               = "Plank"
	Groth16ErrorType             = "Groth16"
	InvalidPublicValuesErrorType = "InvalidPublicValues"
)

// ProofVerificationError is an error that is returned when verifying a proof
type ProofVerificationError struct {
	InnerErrors []error
}

// String is the implementation of the Error interface
func (p *ProofVerificationError) Error() string {
	return fmt.Sprintf("Proof verification error: %v", p.InnerErrors)
}

// Unmarshal unmarshals the data from a map into a ProofVerificationError struct.
func (p *ProofVerificationError) Unmarshal(data interface{}) error {
	getPPErrFn := func(key string, value interface{}) (error, error) {
		switch key {
		case VersionMismatchErrorType:
			versionMismatch := &VersionMismatchError{}
			if err := versionMismatch.Unmarshal(value); err != nil {
				return nil, err
			}
			return versionMismatch, nil
		case CoreErrorType:
			core := &CoreError{}
			if err := core.Unmarshal(value); err != nil {
				return nil, err
			}
			return core, nil
		case RecursionErrorType:
			recursion := &RecursionError{}
			if err := recursion.Unmarshal(value); err != nil {
				return nil, err
			}
			return recursion, nil
		case PlankErrorType:
			plank := &PlankError{}
			if err := plank.Unmarshal(value); err != nil {
				return nil, err
			}
			return plank, nil
		case Groth16ErrorType:
			groth16 := &Groth16Error{}
			if err := groth16.Unmarshal(value); err != nil {
				return nil, err
			}
			return groth16, nil
		case InvalidPublicValuesErrorType:
			return &InvalidPublicValuesError{}, nil
		default:
			return nil, fmt.Errorf("unknown proof verification error type: %v", key)
		}
	}

	getAndAddInnerErrorFn := func(key string, value interface{}) error {
		ppErr, err := getPPErrFn(key, value)
		if err != nil {
			return err
		}

		if ppErr != nil {
			p.InnerErrors = append(p.InnerErrors, ppErr)
		}

		return nil
	}

	dataMap, ok := data.(map[string]interface{})
	if !ok {
		// it can be a single error
		return getAndAddInnerErrorFn(data.(string), nil) //nolint:forcetypeassert
	}

	for key, value := range dataMap {
		if err := getAndAddInnerErrorFn(key, value); err != nil {
			return err
		}
	}

	return nil
}

// StringError is an error that is inherited by other errors that expect a string
// field in the data.
type StringError string

// Unmarshal unmarshals the data from an interface{} into a StringError.
func (e *StringError) Unmarshal(data interface{}) error {
	str, ok := data.(string)
	if !ok {
		return fmt.Errorf("expected string for StringError, got %T", data)
	}
	*e = StringError(str)
	return nil
}

// VersionMismatchError is an error that is returned when the version of the proof is
// different from the version of the core.
type VersionMismatchError struct {
	StringError
}

// String is the implementation of the Error interface
func (e *VersionMismatchError) String() string {
	return fmt.Sprintf("%s: %s", VersionMismatchErrorType, e.StringError)
}

func (e *VersionMismatchError) Error() string {
	return e.String()
}

// CoreError is an error that is returned when the core machine verification fails.
type CoreError struct {
	StringError
}

// String is the implementation of the Error interface
func (e *CoreError) Error() string {
	return fmt.Sprintf("%s: Core machine verification error: %s", CoreErrorType, e.StringError)
}

// RecursionError is an error that is returned when the recursion verification fails.
type RecursionError struct {
	StringError
}

// String is the implementation of the Error interface
func (e *RecursionError) Error() string {
	return fmt.Sprintf("%s: Recursion verification error: %s", RecursionErrorType, e.StringError)
}

// PlankError is an error that is returned when the plank verification fails.
type PlankError struct {
	StringError
}

// String is the implementation of the Error interface
func (e *PlankError) Error() string {
	return fmt.Sprintf("%s: Plank verification error: %s", PlankErrorType, e.StringError)
}

// Groth16Error is an error that is returned when the Groth16Error verification fails.
type Groth16Error struct {
	StringError
}

// String is the implementation of the Error interface
func (e *Groth16Error) Error() string {
	return fmt.Sprintf("%s: Groth16 verification error: %s", Groth16ErrorType, e.StringError)
}

// InvalidPublicValuesError is an error that is returned when the public values are invalid.
type InvalidPublicValuesError struct{}

// String is the implementation of the Error interface
func (e *InvalidPublicValuesError) Error() string {
	return fmt.Sprintf("%s: Invalid public values", InvalidPublicValuesErrorType)
}
