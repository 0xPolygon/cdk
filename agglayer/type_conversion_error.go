package agglayer

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
)

const (
	MultipleL1InfoRootErrorType            = "MultipleL1InfoRoot"
	MismatchNewLocalExitRootErrorType      = "MismatchNewLocalExitRoot"
	BalanceOverflowErrorType               = "BalanceOverflow"
	BalanceUnderflowErrorType              = "BalanceUnderflow"
	BalanceProofGenerationFailedErrorType  = "BalanceProofGenerationFailed"
	NullifierPathGenerationFailedErrorType = "NullifierPathGenerationFailed"
	L1InfoRootIncorrectErrorType           = "L1InfoRootIncorrect"
)

// TypeConversionError is an error that is returned when verifying a certficate
// before generating its proof.
type TypeConversionError struct {
	InnerErrors []error
}

// String is the implementation of the Error interface
func (p *TypeConversionError) Error() string {
	return fmt.Sprintf("Type conversion error: %v", p.InnerErrors)
}

// Unmarshal unmarshals the data from a map into a ProofGenerationError struct.
func (p *TypeConversionError) Unmarshal(data interface{}) error {
	getPPErrFn := func(key string, value interface{}) (error, error) {
		switch key {
		case MultipleL1InfoRootErrorType:
			p.InnerErrors = append(p.InnerErrors, &MultipleL1InfoRootError{})
		case MismatchNewLocalExitRootErrorType:
			p.InnerErrors = append(p.InnerErrors, NewMismatchNewLocalExitRoot())
		case BalanceOverflowErrorType:
			balanceOverflow := NewBalanceOverflow()
			if err := balanceOverflow.UnmarshalFromMap(value); err != nil {
				return nil, err
			}
			return balanceOverflow, nil
		case BalanceUnderflowErrorType:
			balanceUnderflow := NewBalanceUnderflow()
			if err := balanceUnderflow.UnmarshalFromMap(value); err != nil {
				return nil, err
			}
			return balanceUnderflow, nil
		case BalanceProofGenerationFailedErrorType:
			balanceProofGenerationFailed := NewBalanceProofGenerationFailed()
			if err := balanceProofGenerationFailed.UnmarshalFromMap(value); err != nil {
				return nil, err
			}
			return balanceProofGenerationFailed, nil
		case NullifierPathGenerationFailedErrorType:
			nullifierPathGenerationFailed := NewNullifierPathGenerationFailed()
			if err := nullifierPathGenerationFailed.UnmarshalFromMap(value); err != nil {
				return nil, err
			}
			return nullifierPathGenerationFailed, nil
		case L1InfoRootIncorrectErrorType:
			l1InfoRootIncorrect := &L1InfoRootIncorrectError{}
			if err := l1InfoRootIncorrect.Unmarshal(value); err != nil {
				return nil, err
			}
			return l1InfoRootIncorrect, nil
		default:
			return nil, fmt.Errorf("unknown type conversion error type: %v", key)
		}

		return nil, nil
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

	errorSourceMap, ok := data.(map[string]interface{})
	if !ok {
		// it can be a single error
		return getAndAddInnerErrorFn(data.(string), nil) //nolint:forcetypeassert
	}

	for key, value := range errorSourceMap {
		if err := getAndAddInnerErrorFn(key, value); err != nil {
			return err
		}
	}

	return nil
}

// MultipleL1InfoRootError is an error that is returned when the imported bridge exits
// refer to different L1 info roots.
type MultipleL1InfoRootError struct{}

// String is the implementation of the Error interface
func (e *MultipleL1InfoRootError) Error() string {
	return fmt.Sprintf(`%s: The imported bridge exits should refer to one and the same L1 info root.`,
		MultipleL1InfoRootErrorType)
}

// MissingNewLocalExitRoot is an error that is returned when the certificate refers to
// a new local exit root which differ from the one computed by the agglayer.
type MismatchNewLocalExitRootError struct {
	*DeclaredComputedError
}

func NewMismatchNewLocalExitRoot() *MismatchNewLocalExitRootError {
	return &MismatchNewLocalExitRootError{
		DeclaredComputedError: &DeclaredComputedError{ErrType: MismatchNewLocalExitRootErrorType},
	}
}

// BalanceOverflowError is an error that is returned when the given token balance cannot overflow.
type BalanceOverflowError struct {
	*TokenInfoError
}

// NewBalanceOverflow returns a new BalanceOverflow error.
func NewBalanceOverflow() *BalanceOverflowError {
	return &BalanceOverflowError{
		TokenInfoError: &TokenInfoError{},
	}
}

// String is the implementation of the Error interface
func (e *BalanceOverflowError) Error() string {
	return fmt.Sprintf("%s: The given token balance cannot overflow. %s",
		BalanceOverflowErrorType, e.TokenInfo.String())
}

// BalanceUnderflowError is an error that is returned when the given token balance cannot be negative.
type BalanceUnderflowError struct {
	*TokenInfoError
}

// NewBalanceOverflow returns a new BalanceOverflow error.
func NewBalanceUnderflow() *BalanceUnderflowError {
	return &BalanceUnderflowError{
		TokenInfoError: &TokenInfoError{},
	}
}

// String is the implementation of the Error interface
func (e *BalanceUnderflowError) Error() string {
	return fmt.Sprintf("%s: The given token balance cannot be negative. %s",
		BalanceUnderflowErrorType, e.TokenInfo.String())
}

// SmtError is a type that is inherited by all errors that occur during SMT operations.
type SmtError struct {
	ErrorCode string
	Error     string
}

func (e *SmtError) Unmarshal(data interface{}) error {
	errCode, ok := data.(string)
	if !ok {
		return errors.New("error code is not a string")
	}

	e.ErrorCode = errCode

	switch errCode {
	case "KeyAlreadyPresent":
		e.Error = "trying to insert a key already in the SMT"
	case "KeyNotPresent":
		e.Error = "trying to generate a Merkle proof for a key not in the SMT"
	case "KeyPresent":
		e.Error = "trying to generate a non-inclusion proof for a key present in the SMT"
	case "DepthOutOfBounds":
		e.Error = "depth out of bounds"
	default:
		return fmt.Errorf("unknown SMT error code: %s", errCode)
	}

	return nil
}

// BalanceProofGenerationFailedError is a struct that represents an error that occurs when
// the  balance proof for the given token cannot be generated.
type BalanceProofGenerationFailedError struct {
	*TokenInfoError
	*SmtError
}

func NewBalanceProofGenerationFailed() *BalanceProofGenerationFailedError {
	return &BalanceProofGenerationFailedError{
		TokenInfoError: &TokenInfoError{},
		SmtError:       &SmtError{},
	}
}

// String is the implementation of the Error interface
func (e *BalanceProofGenerationFailedError) Error() string {
	return fmt.Sprintf("%s: The balance proof for the given token cannot be generated. TokenInfo: %s. Error type: %s. %s",
		BalanceProofGenerationFailedErrorType, e.TokenInfo.String(),
		e.SmtError.ErrorCode, e.SmtError.Error)
}

func (e *BalanceProofGenerationFailedError) UnmarshalFromMap(data interface{}) error {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return errNotMap
	}

	if err := e.TokenInfoError.UnmarshalFromMap(dataMap["token"]); err != nil {
		return err
	}

	return e.SmtError.Unmarshal(dataMap["source"])
}

// NullifierPathGenerationFailedError is a struct that represents an error that occurs when
// the nullifier path for the given imported bridge exit cannot be generated..
type NullifierPathGenerationFailedError struct {
	GlobalIndex *GlobalIndex `json:"global_index"`
	*SmtError
}

func NewNullifierPathGenerationFailed() *NullifierPathGenerationFailedError {
	return &NullifierPathGenerationFailedError{
		SmtError: &SmtError{},
	}
}

// String is the implementation of the Error interface
func (e *NullifierPathGenerationFailedError) Error() string {
	return fmt.Sprintf("%s: The nullifier path for the given imported bridge exit cannot be generated. "+
		"GlobalIndex: %s. Error type: %s. %s",
		NullifierPathGenerationFailedErrorType, e.GlobalIndex.String(),
		e.SmtError.ErrorCode, e.SmtError.Error)
}

func (e *NullifierPathGenerationFailedError) UnmarshalFromMap(data interface{}) error {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return errNotMap
	}

	if err := e.SmtError.Unmarshal(dataMap["source"]); err != nil {
		return err
	}

	globalIndexMap, err := convertMapValue[map[string]interface{}](dataMap, "global_index")
	if err != nil {
		return err
	}

	e.GlobalIndex = &GlobalIndex{}
	return e.GlobalIndex.UnmarshalFromMap(globalIndexMap)
}

// L1InfoRootIncorrectError is an error that is returned when the L1 Info Root is invalid or unsettled
type L1InfoRootIncorrectError struct {
	Declared  common.Hash `json:"declared"`
	Retrieved common.Hash `json:"retrieved"`
	LeafCount uint32      `json:"leaf_count"`
}

// String is the implementation of the Error interface
func (e *L1InfoRootIncorrectError) Error() string {
	return fmt.Sprintf("%s: The L1 Info Root is incorrect. Declared: %s, Retrieved: %s, LeafCount: %d",
		L1InfoRootIncorrectErrorType, e.Declared.String(), e.Retrieved.String(), e.LeafCount)
}

// Unmarshal unmarshals the data from a map into a L1InfoRootIncorrect struct.
func (e *L1InfoRootIncorrectError) Unmarshal(data interface{}) error {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return errNotMap
	}

	declared, err := convertMapValue[string](dataMap, "declared")
	if err != nil {
		return err
	}

	retrieved, err := convertMapValue[string](dataMap, "retrieved")
	if err != nil {
		return err
	}

	leafCount, err := convertMapValue[uint32](dataMap, "leaf_count")
	if err != nil {
		return err
	}

	e.Declared = common.HexToHash(declared)
	e.Retrieved = common.HexToHash(retrieved)
	e.LeafCount = leafCount

	return nil
}
