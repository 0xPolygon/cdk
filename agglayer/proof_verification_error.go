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
	InnerErrors []PPError
}

// String is the implementation of the Error interface
func (p *ProofVerificationError) String() string {
	return fmt.Sprintf("Proof verification error: %v", p.InnerErrors)
}

// Unmarshal unmarshals the data from a map into a ProofVerificationError struct.
func (p *ProofVerificationError) Unmarshal(data interface{}) error {
	getPPErrFn := func(key string, value interface{}) (PPError, error) {
		switch key {
		case VersionMismatchErrorType:
			versionMismatch := &VersionMismatch{}
			if err := versionMismatch.Unmarshal(value); err != nil {
				return nil, err
			}
			return versionMismatch, nil
		case CoreErrorType:
			core := &Core{}
			if err := core.Unmarshal(value); err != nil {
				return nil, err
			}
			return core, nil
		case RecursionErrorType:
			recursion := &Recursion{}
			if err := recursion.Unmarshal(value); err != nil {
				return nil, err
			}
			return recursion, nil
		case PlankErrorType:
			plank := &Plank{}
			if err := plank.Unmarshal(value); err != nil {
				return nil, err
			}
			return plank, nil
		case Groth16ErrorType:
			groth16 := &Groth16{}
			if err := groth16.Unmarshal(value); err != nil {
				return nil, err
			}
			return groth16, nil
		case InvalidPublicValuesErrorType:
			return &InvalidPublicValues{}, nil
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

// VersionMismatch is an error that is returned when the version of the proof is
// different from the version of the core.
type VersionMismatch struct {
	StringError
}

// String is the implementation of the Error interface
func (e *VersionMismatch) String() string {
	return fmt.Sprintf("%s: %s", VersionMismatchErrorType, e.StringError)
}

// Core is an error that is returned when the core machine verification fails.
type Core struct {
	StringError
}

// String is the implementation of the Error interface
func (e *Core) String() string {
	return fmt.Sprintf("%s: Core machine verification error: %s", CoreErrorType, e.StringError)
}

// Recursion is an error that is returned when the recursion verification fails.
type Recursion struct {
	StringError
}

// String is the implementation of the Error interface
func (e *Recursion) String() string {
	return fmt.Sprintf("%s: Recursion verification error: %s", RecursionErrorType, e.StringError)
}

// Plank is an error that is returned when the plank verification fails.
type Plank struct {
	StringError
}

// String is the implementation of the Error interface
func (e *Plank) String() string {
	return fmt.Sprintf("%s: Plank verification error: %s", PlankErrorType, e.StringError)
}

// Groth16 is an error that is returned when the Groth16 verification fails.
type Groth16 struct {
	StringError
}

// String is the implementation of the Error interface
func (e *Groth16) String() string {
	return fmt.Sprintf("%s: Groth16 verification error: %s", Groth16ErrorType, e.StringError)
}

// InvalidPublicValues is an error that is returned when the public values are invalid.
type InvalidPublicValues struct{}

// String is the implementation of the Error interface
func (e *InvalidPublicValues) String() string {
	return fmt.Sprintf("%s: Invalid public values", InvalidPublicValuesErrorType)
}
