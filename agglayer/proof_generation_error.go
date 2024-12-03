package agglayer

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/ethereum/go-ethereum/common"
)

var errNotMap = errors.New("inner error is not a map")

const (
	InvalidSignerErrorType                = "InvalidSigner"
	InvalidPreviousLERErrorType           = "InvalidPreviousLocalExitRoot"
	InvalidPreviousBalanceRootErrorType   = "InvalidPreviousBalanceRoot"
	InvalidPreviousNullifierRootErrorType = "InvalidPreviousNullifierRoot"
	InvalidNewLocalExitRootErrorType      = "InvalidNewLocalExitRoot"
	InvalidNewBalanceRootErrorType        = "InvalidNewBalanceRoot"
	InvalidNewNullifierRootErrorType      = "InvalidNewNullifierRoot"
	InvalidImportedExitsRootErrorType     = "InvalidImportedExitsRoot"
	MismatchImportedExitsRootErrorType    = "MismatchImportedExitsRoot"
	InvalidNullifierPathErrorType         = "InvalidNullifierPath"
	InvalidBalancePathErrorType           = "InvalidBalancePath"
	BalanceOverflowInBridgeExitErrorType  = "BalanceOverflowInBridgeExit"
	BalanceUnderflowInBridgeExitErrorType = "BalanceUnderflowInBridgeExit"
	CannotExitToSameNetworkErrorType      = "CannotExitToSameNetwork"
	InvalidMessageOriginNetworkErrorType  = "InvalidMessageOriginNetwork"
	InvalidL1TokenInfoErrorType           = "InvalidL1TokenInfo"
	MissingTokenBalanceProofErrorType     = "MissingTokenBalanceProof"
	DuplicateTokenBalanceProofErrorType   = "DuplicateTokenBalanceProof"
	InvalidSignatureErrorType             = "InvalidSignature"
	InvalidImportedBridgeExitErrorType    = "InvalidImportedBridgeExit"
	UnknownErrorType                      = "UnknownError"
)

// ProofGenerationError is a struct that represents an error that occurs when generating a proof.
type ProofGenerationError struct {
	GenerationType string
	InnerErrors    []error
}

// String is the implementation of the Error interface
func (p *ProofGenerationError) Error() string {
	return fmt.Sprintf("Proof generation error: %s. %s", p.GenerationType, p.InnerErrors)
}

// Unmarshal unmarshals the data from a map into a ProofGenerationError struct.
func (p *ProofGenerationError) Unmarshal(data interface{}) error {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return errNotMap
	}

	generationType, err := convertMapValue[string](dataMap, "generation_type")
	if err != nil {
		return err
	}

	p.GenerationType = generationType

	getPPErrFn := func(key string, value interface{}) (error, error) {
		switch key {
		case InvalidSignerErrorType:
			invalidSigner := &InvalidSignerError{}
			if err := invalidSigner.UnmarshalFromMap(value); err != nil {
				return nil, err
			}
			return invalidSigner, nil
		case InvalidPreviousLERErrorType:
			invalidPreviousLER := NewInvalidPreviousLocalExitRoot()
			if err := invalidPreviousLER.UnmarshalFromMap(value); err != nil {
				return nil, err
			}
			p.InnerErrors = append(p.InnerErrors, invalidPreviousLER)
		case InvalidPreviousBalanceRootErrorType:
			invalidPreviousBalanceRoot := NewInvalidPreviousBalanceRoot()
			if err := invalidPreviousBalanceRoot.UnmarshalFromMap(value); err != nil {
				return nil, err
			}
			p.InnerErrors = append(p.InnerErrors, invalidPreviousBalanceRoot)
		case InvalidPreviousNullifierRootErrorType:
			invalidPreviousNullifierRoot := NewInvalidPreviousNullifierRoot()
			if err := invalidPreviousNullifierRoot.UnmarshalFromMap(value); err != nil {
				return nil, err
			}
			p.InnerErrors = append(p.InnerErrors, invalidPreviousNullifierRoot)
		case InvalidNewLocalExitRootErrorType:
			invalidNewLocalExitRoot := NewInvalidNewLocalExitRoot()
			if err := invalidNewLocalExitRoot.UnmarshalFromMap(value); err != nil {
				return nil, err
			}
			p.InnerErrors = append(p.InnerErrors, invalidNewLocalExitRoot)
		case InvalidNewBalanceRootErrorType:
			invalidNewBalanceRoot := NewInvalidNewBalanceRoot()
			if err := invalidNewBalanceRoot.UnmarshalFromMap(value); err != nil {
				return nil, err
			}
			p.InnerErrors = append(p.InnerErrors, invalidNewBalanceRoot)
		case InvalidNewNullifierRootErrorType:
			invalidNewNullifierRoot := NewInvalidNewNullifierRoot()
			if err := invalidNewNullifierRoot.UnmarshalFromMap(value); err != nil {
				return nil, err
			}
			p.InnerErrors = append(p.InnerErrors, invalidNewNullifierRoot)
		case InvalidImportedExitsRootErrorType:
			invalidImportedExitsRoot := NewInvalidImportedExitsRoot()
			if err := invalidImportedExitsRoot.UnmarshalFromMap(value); err != nil {
				return nil, err
			}
			p.InnerErrors = append(p.InnerErrors, invalidImportedExitsRoot)
		case MismatchImportedExitsRootErrorType:
			p.InnerErrors = append(p.InnerErrors, &MismatchImportedExitsRootError{})
		case InvalidNullifierPathErrorType:
			p.InnerErrors = append(p.InnerErrors, &InvalidNullifierPathError{})
		case InvalidBalancePathErrorType:
			p.InnerErrors = append(p.InnerErrors, &InvalidBalancePathError{})
		case BalanceOverflowInBridgeExitErrorType:
			p.InnerErrors = append(p.InnerErrors, &BalanceOverflowInBridgeExitError{})
		case BalanceUnderflowInBridgeExitErrorType:
			p.InnerErrors = append(p.InnerErrors, &BalanceUnderflowInBridgeExitError{})
		case CannotExitToSameNetworkErrorType:
			p.InnerErrors = append(p.InnerErrors, &CannotExitToSameNetworkError{})
		case InvalidMessageOriginNetworkErrorType:
			p.InnerErrors = append(p.InnerErrors, &InvalidMessageOriginNetworkError{})
		case InvalidL1TokenInfoErrorType:
			invalidL1TokenInfo := NewInvalidL1TokenInfo()
			if err := invalidL1TokenInfo.UnmarshalFromMap(value); err != nil {
				return nil, err
			}
			p.InnerErrors = append(p.InnerErrors, invalidL1TokenInfo)
		case MissingTokenBalanceProofErrorType:
			missingTokenBalanceProof := NewMissingTokenBalanceProof()
			if err := missingTokenBalanceProof.UnmarshalFromMap(value); err != nil {
				return nil, err
			}
			p.InnerErrors = append(p.InnerErrors, missingTokenBalanceProof)
		case DuplicateTokenBalanceProofErrorType:
			duplicateTokenBalanceProof := NewDuplicateTokenBalanceProof()
			if err := duplicateTokenBalanceProof.UnmarshalFromMap(value); err != nil {
				return nil, err
			}
			p.InnerErrors = append(p.InnerErrors, duplicateTokenBalanceProof)
		case InvalidSignatureErrorType:
			p.InnerErrors = append(p.InnerErrors, &InvalidSignatureError{})
		case InvalidImportedBridgeExitErrorType:
			invalidImportedBridgeExit := &InvalidImportedBridgeExitError{}
			if err := invalidImportedBridgeExit.UnmarshalFromMap(value); err != nil {
				return nil, err
			}
			p.InnerErrors = append(p.InnerErrors, invalidImportedBridgeExit)
		case UnknownErrorType:
			p.InnerErrors = append(p.InnerErrors, &UnknownError{})
		default:
			return nil, fmt.Errorf("unknown proof generation error type: %s", key)
		}

		return nil, nil
	}

	errorSourceMap, err := convertMapValue[map[string]interface{}](dataMap, "source")
	if err != nil {
		// it can be a single error
		errSourceString, err := convertMapValue[string](dataMap, "source")
		if err != nil {
			return err
		}

		ppErr, err := getPPErrFn(errSourceString, nil)
		if err != nil {
			return err
		}

		if ppErr != nil {
			p.InnerErrors = append(p.InnerErrors, ppErr)
		}

		return nil
	}

	// there will always be only one key in the source map
	for key, value := range errorSourceMap {
		ppErr, err := getPPErrFn(key, value)
		if err != nil {
			return err
		}

		if ppErr != nil {
			p.InnerErrors = append(p.InnerErrors, ppErr)
		}
	}

	return nil
}

// InvalidSignerError is a struct that represents an error that occurs when
// the signer of the certificate is invalid, or the hash that was signed was not valid.
type InvalidSignerError struct {
	Declared  common.Address `json:"declared"`
	Recovered common.Address `json:"recovered"`
}

// String is the implementation of the Error interface
func (e *InvalidSignerError) String() string {
	return fmt.Sprintf("%s. Declared: %s, Computed: %s",
		InvalidSignerErrorType, e.Declared.String(), e.Recovered.String())
}

func (e *InvalidSignerError) Error() string {
	return e.String()
}

func (e *InvalidSignerError) UnmarshalFromMap(data interface{}) error {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return errNotMap
	}

	declared, err := convertMapValue[string](dataMap, "declared")
	if err != nil {
		return err
	}

	recovered, err := convertMapValue[string](dataMap, "recovered")
	if err != nil {
		return err
	}

	e.Declared = common.HexToAddress(declared)
	e.Recovered = common.HexToAddress(recovered)

	return nil
}

// DeclaredComputedError is a base struct for errors that have both declared and computed values.
type DeclaredComputedError struct {
	Declared common.Hash `json:"declared"`
	Computed common.Hash `json:"computed"`
	ErrType  string
}

// String is the implementation of the Error interface
func (e *DeclaredComputedError) Error() string {
	return fmt.Sprintf("%s. Declared: %s, Computed: %s",
		e.ErrType, e.Declared.String(), e.Computed.String())
}

// UnmarshalFromMap is the implementation of the Error interface
func (e *DeclaredComputedError) UnmarshalFromMap(data interface{}) error {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return errNotMap
	}

	declared, err := convertMapValue[string](dataMap, "declared")
	if err != nil {
		return err
	}

	computed, err := convertMapValue[string](dataMap, "computed")
	if err != nil {
		return err
	}

	e.Declared = common.HexToHash(declared)
	e.Computed = common.HexToHash(computed)

	return nil
}

// InvalidPreviousLocalExitRootError is a struct that represents an error that occurs when
// the previous local exit root is invalid.
type InvalidPreviousLocalExitRootError struct {
	*DeclaredComputedError
}

func NewInvalidPreviousLocalExitRoot() *InvalidPreviousLocalExitRootError {
	return &InvalidPreviousLocalExitRootError{
		DeclaredComputedError: &DeclaredComputedError{ErrType: InvalidPreviousLERErrorType},
	}
}

// InvalidPreviousBalanceRootError is a struct that represents an error that occurs when
// the previous balance root is invalid.
type InvalidPreviousBalanceRootError struct {
	*DeclaredComputedError
}

func NewInvalidPreviousBalanceRoot() *InvalidPreviousBalanceRootError {
	return &InvalidPreviousBalanceRootError{
		DeclaredComputedError: &DeclaredComputedError{ErrType: InvalidPreviousBalanceRootErrorType},
	}
}

// InvalidPreviousNullifierRootError is a struct that represents an error that occurs when
// the previous nullifier root is invalid.
type InvalidPreviousNullifierRootError struct {
	*DeclaredComputedError
}

func NewInvalidPreviousNullifierRoot() *InvalidPreviousNullifierRootError {
	return &InvalidPreviousNullifierRootError{
		DeclaredComputedError: &DeclaredComputedError{ErrType: InvalidPreviousNullifierRootErrorType},
	}
}

// InvalidNewLocalExitRootError is a struct that represents an error that occurs when
// the new local exit root is invalid.
type InvalidNewLocalExitRootError struct {
	*DeclaredComputedError
}

func NewInvalidNewLocalExitRoot() *InvalidNewLocalExitRootError {
	return &InvalidNewLocalExitRootError{
		DeclaredComputedError: &DeclaredComputedError{ErrType: InvalidNewLocalExitRootErrorType},
	}
}

// InvalidNewBalanceRootError is a struct that represents an error that occurs when
// the new balance root is invalid.
type InvalidNewBalanceRootError struct {
	*DeclaredComputedError
}

func NewInvalidNewBalanceRoot() *InvalidNewBalanceRootError {
	return &InvalidNewBalanceRootError{
		DeclaredComputedError: &DeclaredComputedError{ErrType: InvalidNewBalanceRootErrorType},
	}
}

// InvalidNewNullifierRootError is a struct that represents an error that occurs when
// the new nullifier root is invalid.
type InvalidNewNullifierRootError struct {
	*DeclaredComputedError
}

func NewInvalidNewNullifierRoot() *InvalidNewNullifierRootError {
	return &InvalidNewNullifierRootError{
		DeclaredComputedError: &DeclaredComputedError{ErrType: InvalidNewNullifierRootErrorType},
	}
}

// InvalidImportedExitsRootError is a struct that represents an error that occurs when
// the imported exits root is invalid.
type InvalidImportedExitsRootError struct {
	*DeclaredComputedError
}

func NewInvalidImportedExitsRoot() *InvalidImportedExitsRootError {
	return &InvalidImportedExitsRootError{
		DeclaredComputedError: &DeclaredComputedError{ErrType: InvalidImportedExitsRootErrorType},
	}
}

// MismatchImportedExitsRootError is a struct that represents an error that occurs when
// the commitment to the list of imported bridge exits but the list of imported bridge exits is empty.
type MismatchImportedExitsRootError struct{}

// String is the implementation of the Error interface
func (e *MismatchImportedExitsRootError) Error() string {
	return fmt.Sprintf(`%s: The commitment to the list of imported bridge exits 
	should be Some if and only if this list is non-empty, should be None otherwise.`,
		MismatchImportedExitsRootErrorType)
}

// InvalidNullifierPathError is a struct that represents an error that occurs when
// the provided nullifier path is invalid.
type InvalidNullifierPathError struct{}

// String is the implementation of the Error interface
func (e *InvalidNullifierPathError) Error() string {
	return fmt.Sprintf("%s: The provided nullifier path is invalid", InvalidNullifierPathErrorType)
}

// InvalidBalancePathError is a struct that represents an error that occurs when
// the provided balance path is invalid.
type InvalidBalancePathError struct{}

// String is the implementation of the Error interface
func (e *InvalidBalancePathError) Error() string {
	return fmt.Sprintf("%s: The provided balance path is invalid", InvalidBalancePathErrorType)
}

// BalanceOverflowInBridgeExitError is a struct that represents an error that occurs when
// bridge exit led to balance overflow.
type BalanceOverflowInBridgeExitError struct{}

// String is the implementation of the Error interface
func (e *BalanceOverflowInBridgeExitError) Error() string {
	return fmt.Sprintf("%s: The imported bridge exit led to balance overflow.", BalanceOverflowInBridgeExitErrorType)
}

// BalanceUnderflowInBridgeExitError is a struct that represents an error that occurs when
// bridge exit led to balance underflow.
type BalanceUnderflowInBridgeExitError struct{}

// String is the implementation of the Error interface
func (e *BalanceUnderflowInBridgeExitError) Error() string {
	return fmt.Sprintf("%s: The imported bridge exit led to balance underflow.", BalanceUnderflowInBridgeExitErrorType)
}

// CannotExitToSameNetworkError is a struct that represents an error that occurs when
// the user tries to exit to the same network.
type CannotExitToSameNetworkError struct{}

// String is the implementation of the Error interface
func (e *CannotExitToSameNetworkError) Error() string {
	return fmt.Sprintf("%s: The provided bridge exit goes to the sender’s own network which is not permitted.",
		CannotExitToSameNetworkErrorType)
}

// InvalidMessageOriginNetworkError is a struct that represents an error that occurs when
// the origin network of the message is invalid.
type InvalidMessageOriginNetworkError struct{}

// String is the implementation of the Error interface
func (e *InvalidMessageOriginNetworkError) Error() string {
	return fmt.Sprintf("%s: The origin network of the message is invalid.", InvalidMessageOriginNetworkErrorType)
}

// TokenInfoError is a struct inherited by other errors that have a TokenInfo field.
type TokenInfoError struct {
	TokenInfo *TokenInfo `json:"token_info"`
	isNested  bool
}

func (e *TokenInfoError) UnmarshalFromMap(data interface{}) error {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return errNotMap
	}

	var (
		err          error
		tokenInfoMap map[string]interface{}
	)

	if e.isNested {
		tokenInfoMap, err = convertMapValue[map[string]interface{}](dataMap, "TokenInfo")
		if err != nil {
			return err
		}
	} else {
		tokenInfoMap = dataMap
	}

	originNetwork, err := convertMapValue[uint32](tokenInfoMap, "origin_network")
	if err != nil {
		return err
	}

	originAddress, err := convertMapValue[string](tokenInfoMap, "origin_token_address")
	if err != nil {
		return err
	}

	e.TokenInfo = &TokenInfo{
		OriginNetwork:      originNetwork,
		OriginTokenAddress: common.HexToAddress(originAddress),
	}

	return nil
}

func (e *TokenInfoError) Error() string {
	return fmt.Sprintf("token info %v, is nested %t", e.TokenInfo.String(), e.isNested)
}

// InvalidL1TokenInfoError is a struct that represents an error that occurs when
// the L1 token info is invalid.
type InvalidL1TokenInfoError struct {
	*TokenInfoError
}

// NewInvalidL1TokenInfo returns a new instance of InvalidL1TokenInfo.
func NewInvalidL1TokenInfo() *InvalidL1TokenInfoError {
	return &InvalidL1TokenInfoError{
		TokenInfoError: &TokenInfoError{isNested: true},
	}
}

// String is the implementation of the Error interface
func (e *InvalidL1TokenInfoError) String() string {
	return fmt.Sprintf("%s: The L1 token info is invalid. %s",
		InvalidL1TokenInfoErrorType, e.TokenInfo.String())
}

// MissingTokenBalanceProofError is a struct that represents an error that occurs when
// the token balance proof is missing.
type MissingTokenBalanceProofError struct {
	*TokenInfoError
}

// NewMissingTokenBalanceProof returns a new instance of MissingTokenBalanceProof.
func NewMissingTokenBalanceProof() *MissingTokenBalanceProofError {
	return &MissingTokenBalanceProofError{
		TokenInfoError: &TokenInfoError{isNested: true},
	}
}

// String is the implementation of the Error interface
func (e *MissingTokenBalanceProofError) String() string {
	return fmt.Sprintf("%s: The provided token is missing a balance proof. %s",
		MissingTokenBalanceProofErrorType, e.TokenInfo.String())
}

// DuplicateTokenBalanceProofError is a struct that represents an error that occurs when
// the token balance proof is duplicated.
type DuplicateTokenBalanceProofError struct {
	*TokenInfoError
}

// NewDuplicateTokenBalanceProof returns a new instance of DuplicateTokenBalanceProof.
func NewDuplicateTokenBalanceProof() *DuplicateTokenBalanceProofError {
	return &DuplicateTokenBalanceProofError{
		TokenInfoError: &TokenInfoError{isNested: true},
	}
}

// String is the implementation of the Error interface
func (e *DuplicateTokenBalanceProofError) String() string {
	return fmt.Sprintf("%s: The provided token comes with multiple balance proofs. %s",
		DuplicateTokenBalanceProofErrorType, e.TokenInfo.String())
}

// InvalidSignatureError is a struct that represents an error that occurs when
// the signature is invalid.
type InvalidSignatureError struct{}

// String is the implementation of the Error interface
func (e *InvalidSignatureError) Error() string {
	return fmt.Sprintf("%s: The provided signature is invalid.", InvalidSignatureErrorType)
}

// UnknownError is a struct that represents an error that occurs when
// an unknown error is encountered.
type UnknownError struct{}

// String is the implementation of the Error interface
func (e *UnknownError) Error() string {
	return fmt.Sprintf("%s: An unknown error occurred.", UnknownErrorType)
}

// InvalidImportedBridgeExitError is a struct that represents an error that occurs when
// the imported bridge exit is invalid.
type InvalidImportedBridgeExitError struct {
	GlobalIndex *GlobalIndex `json:"global_index"`
	ErrorType   string       `json:"error_type"`
}

// String is the implementation of the Error interface
func (e *InvalidImportedBridgeExitError) Error() string {
	var errorDescription string
	switch e.ErrorType {
	case "MismatchGlobalIndexInclusionProof":
		errorDescription = "The global index and the inclusion proof do not both correspond " +
			"to the same network type: mainnet or rollup."
	case "MismatchL1Root":
		errorDescription = "The provided L1 info root does not match the one provided in the inclusion proof."
	case "MismatchMER":
		errorDescription = "The provided MER does not match the one provided in the inclusion proof."
	case "MismatchRER":
		errorDescription = "The provided RER does not match the one provided in the inclusion proof."
	case "InvalidMerklePathLeafToLER":
		errorDescription = "The inclusion proof from the leaf to the LER is invalid."
	case "InvalidMerklePathLERToRER":
		errorDescription = "The inclusion proof from the LER to the RER is invalid."
	case "InvalidMerklePathGERToL1Root":
		errorDescription = "The inclusion proof from the GER to the L1 info Root is invalid."
	case "InvalidExitNetwork":
		errorDescription = "The provided imported bridge exit does not target the right destination network."
	default:
		errorDescription = "An unknown error occurred."
	}

	return fmt.Sprintf("%s: Global index: %s. Error type: %s. %s",
		InvalidImportedBridgeExitErrorType, e.GlobalIndex.String(), e.ErrorType, errorDescription)
}

func (e *InvalidImportedBridgeExitError) UnmarshalFromMap(data interface{}) error {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return errNotMap
	}

	sourceErr, err := convertMapValue[string](dataMap, "source")
	if err != nil {
		return err
	}

	e.ErrorType = sourceErr

	globalIndexMap, err := convertMapValue[map[string]interface{}](dataMap, "global_index")
	if err != nil {
		return err
	}

	e.GlobalIndex = &GlobalIndex{}
	return e.GlobalIndex.UnmarshalFromMap(globalIndexMap)
}

// convertMapValue converts the value of a key in a map to a target type.
func convertMapValue[T any](data map[string]interface{}, key string) (T, error) {
	value, ok := data[key]
	if !ok {
		var zero T
		return zero, fmt.Errorf("key %s not found in map", key)
	}

	// Try a direct type assertion
	if convertedValue, ok := value.(T); ok {
		return convertedValue, nil
	}

	// If direct assertion fails, handle numeric type conversions
	var target T
	targetType := reflect.TypeOf(target)

	// Check if value is a float64 (default JSON number type) and target is a numeric type
	if floatValue, ok := value.(float64); ok && targetType.Kind() >= reflect.Int && targetType.Kind() <= reflect.Uint64 {
		convertedValue, err := convertNumeric(floatValue, targetType)
		if err != nil {
			return target, fmt.Errorf("conversion error for key %s: %w", key, err)
		}
		return convertedValue.(T), nil //nolint:forcetypeassert
	}

	return target, fmt.Errorf("value of key %s is not of type %T", key, target)
}

// convertNumeric converts a float64 to the specified numeric type.
func convertNumeric(value float64, targetType reflect.Type) (interface{}, error) {
	switch targetType.Kind() {
	case reflect.Int:
		return int(value), nil
	case reflect.Int8:
		return int8(value), nil
	case reflect.Int16:
		return int16(value), nil
	case reflect.Int32:
		return int32(value), nil
	case reflect.Int64:
		return int64(value), nil
	case reflect.Uint:
		return uint(value), nil
	case reflect.Uint8:
		return uint8(value), nil
	case reflect.Uint16:
		return uint16(value), nil
	case reflect.Uint32:
		return uint32(value), nil
	case reflect.Uint64:
		return uint64(value), nil
	case reflect.Float32:
		return float32(value), nil
	case reflect.Float64:
		return value, nil
	default:
		return nil, errors.New("unsupported target type")
	}
}
