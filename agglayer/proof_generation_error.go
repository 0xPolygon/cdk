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

type PPError interface {
	String() string
}

// ProofGenerationError is a struct that represents an error that occurs when generating a proof.
type ProofGenerationError struct {
	GenerationType string
	InnerErrors    []PPError
}

// String is the implementation of the Error interface
func (p *ProofGenerationError) String() string {
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

	getPPErrFn := func(key string, value interface{}) (PPError, error) {
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
			p.InnerErrors = append(p.InnerErrors, &MismatchImportedExitsRoot{})
		case InvalidNullifierPathErrorType:
			p.InnerErrors = append(p.InnerErrors, &InvalidNullifierPath{})
		case InvalidBalancePathErrorType:
			p.InnerErrors = append(p.InnerErrors, &InvalidBalancePath{})
		case BalanceOverflowInBridgeExitErrorType:
			p.InnerErrors = append(p.InnerErrors, &BalanceOverflowInBridgeExit{})
		case BalanceUnderflowInBridgeExitErrorType:
			p.InnerErrors = append(p.InnerErrors, &BalanceUnderflowInBridgeExit{})
		case CannotExitToSameNetworkErrorType:
			p.InnerErrors = append(p.InnerErrors, &CannotExitToSameNetwork{})
		case InvalidMessageOriginNetworkErrorType:
			p.InnerErrors = append(p.InnerErrors, &InvalidMessageOriginNetwork{})
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
			p.InnerErrors = append(p.InnerErrors, &InvalidSignature{})
		case InvalidImportedBridgeExitErrorType:
			invalidImportedBridgeExit := &InvalidImportedBridgeExit{}
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
func (e *DeclaredComputedError) String() string {
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

// InvalidPreviousLocalExitRoot is a struct that represents an error that occurs when
// the previous local exit root is invalid.
type InvalidPreviousLocalExitRoot struct {
	*DeclaredComputedError
}

func NewInvalidPreviousLocalExitRoot() *InvalidPreviousLocalExitRoot {
	return &InvalidPreviousLocalExitRoot{
		DeclaredComputedError: &DeclaredComputedError{ErrType: InvalidPreviousLERErrorType},
	}
}

// InvalidPreviousBalanceRoot is a struct that represents an error that occurs when
// the previous balance root is invalid.
type InvalidPreviousBalanceRoot struct {
	*DeclaredComputedError
}

func NewInvalidPreviousBalanceRoot() *InvalidPreviousBalanceRoot {
	return &InvalidPreviousBalanceRoot{
		DeclaredComputedError: &DeclaredComputedError{ErrType: InvalidPreviousBalanceRootErrorType},
	}
}

// InvalidPreviousNullifierRoot is a struct that represents an error that occurs when
// the previous nullifier root is invalid.
type InvalidPreviousNullifierRoot struct {
	*DeclaredComputedError
}

func NewInvalidPreviousNullifierRoot() *InvalidPreviousNullifierRoot {
	return &InvalidPreviousNullifierRoot{
		DeclaredComputedError: &DeclaredComputedError{ErrType: InvalidPreviousNullifierRootErrorType},
	}
}

// InvalidNewLocalExitRoot is a struct that represents an error that occurs when
// the new local exit root is invalid.
type InvalidNewLocalExitRoot struct {
	*DeclaredComputedError
}

func NewInvalidNewLocalExitRoot() *InvalidNewLocalExitRoot {
	return &InvalidNewLocalExitRoot{
		DeclaredComputedError: &DeclaredComputedError{ErrType: InvalidNewLocalExitRootErrorType},
	}
}

// InvalidNewBalanceRoot is a struct that represents an error that occurs when
// the new balance root is invalid.
type InvalidNewBalanceRoot struct {
	*DeclaredComputedError
}

func NewInvalidNewBalanceRoot() *InvalidNewBalanceRoot {
	return &InvalidNewBalanceRoot{
		DeclaredComputedError: &DeclaredComputedError{ErrType: InvalidNewBalanceRootErrorType},
	}
}

// InvalidNewNullifierRoot is a struct that represents an error that occurs when
// the new nullifier root is invalid.
type InvalidNewNullifierRoot struct {
	*DeclaredComputedError
}

func NewInvalidNewNullifierRoot() *InvalidNewNullifierRoot {
	return &InvalidNewNullifierRoot{
		DeclaredComputedError: &DeclaredComputedError{ErrType: InvalidNewNullifierRootErrorType},
	}
}

// InvalidImportedExitsRoot is a struct that represents an error that occurs when
// the imported exits root is invalid.
type InvalidImportedExitsRoot struct {
	*DeclaredComputedError
}

func NewInvalidImportedExitsRoot() *InvalidImportedExitsRoot {
	return &InvalidImportedExitsRoot{
		DeclaredComputedError: &DeclaredComputedError{ErrType: InvalidImportedExitsRootErrorType},
	}
}

// MismatchImportedExitsRoot is a struct that represents an error that occurs when
// the commitment to the list of imported bridge exits but the list of imported bridge exits is empty.
type MismatchImportedExitsRoot struct{}

// String is the implementation of the Error interface
func (e *MismatchImportedExitsRoot) String() string {
	return fmt.Sprintf(`%s: The commitment to the list of imported bridge exits 
	should be Some if and only if this list is non-empty, should be None otherwise.`,
		MismatchImportedExitsRootErrorType)
}

// InvalidNullifierPath is a struct that represents an error that occurs when
// the provided nullifier path is invalid.
type InvalidNullifierPath struct{}

// String is the implementation of the Error interface
func (e *InvalidNullifierPath) String() string {
	return fmt.Sprintf("%s: The provided nullifier path is invalid", InvalidNullifierPathErrorType)
}

// InvalidBalancePath is a struct that represents an error that occurs when
// the provided balance path is invalid.
type InvalidBalancePath struct{}

// String is the implementation of the Error interface
func (e *InvalidBalancePath) String() string {
	return fmt.Sprintf("%s: The provided balance path is invalid", InvalidBalancePathErrorType)
}

// BalanceOverflowInBridgeExit is a struct that represents an error that occurs when
// bridge exit led to balance overflow.
type BalanceOverflowInBridgeExit struct{}

// String is the implementation of the Error interface
func (e *BalanceOverflowInBridgeExit) String() string {
	return fmt.Sprintf("%s: The imported bridge exit led to balance overflow.", BalanceOverflowInBridgeExitErrorType)
}

// BalanceUnderflowInBridgeExit is a struct that represents an error that occurs when
// bridge exit led to balance underflow.
type BalanceUnderflowInBridgeExit struct{}

// String is the implementation of the Error interface
func (e *BalanceUnderflowInBridgeExit) String() string {
	return fmt.Sprintf("%s: The imported bridge exit led to balance underflow.", BalanceUnderflowInBridgeExitErrorType)
}

// CannotExitToSameNetwork is a struct that represents an error that occurs when
// the user tries to exit to the same network.
type CannotExitToSameNetwork struct{}

// String is the implementation of the Error interface
func (e *CannotExitToSameNetwork) String() string {
	return fmt.Sprintf("%s: The provided bridge exit goes to the sender’s own network which is not permitted.",
		CannotExitToSameNetworkErrorType)
}

// InvalidMessageOriginNetwork is a struct that represents an error that occurs when
// the origin network of the message is invalid.
type InvalidMessageOriginNetwork struct{}

// String is the implementation of the Error interface
func (e *InvalidMessageOriginNetwork) String() string {
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

// InvalidL1TokenInfo is a struct that represents an error that occurs when
// the L1 token info is invalid.
type InvalidL1TokenInfo struct {
	*TokenInfoError
}

// NewInvalidL1TokenInfo returns a new instance of InvalidL1TokenInfo.
func NewInvalidL1TokenInfo() *InvalidL1TokenInfo {
	return &InvalidL1TokenInfo{
		TokenInfoError: &TokenInfoError{isNested: true},
	}
}

// String is the implementation of the Error interface
func (e *InvalidL1TokenInfo) String() string {
	return fmt.Sprintf("%s: The L1 token info is invalid. %s",
		InvalidL1TokenInfoErrorType, e.TokenInfo.String())
}

// MissingTokenBalanceProof is a struct that represents an error that occurs when
// the token balance proof is missing.
type MissingTokenBalanceProof struct {
	*TokenInfoError
}

// NewMissingTokenBalanceProof returns a new instance of MissingTokenBalanceProof.
func NewMissingTokenBalanceProof() *MissingTokenBalanceProof {
	return &MissingTokenBalanceProof{
		TokenInfoError: &TokenInfoError{isNested: true},
	}
}

// String is the implementation of the Error interface
func (e *MissingTokenBalanceProof) String() string {
	return fmt.Sprintf("%s: The provided token is missing a balance proof. %s",
		MissingTokenBalanceProofErrorType, e.TokenInfo.String())
}

// DuplicateTokenBalanceProof is a struct that represents an error that occurs when
// the token balance proof is duplicated.
type DuplicateTokenBalanceProof struct {
	*TokenInfoError
}

// NewDuplicateTokenBalanceProof returns a new instance of DuplicateTokenBalanceProof.
func NewDuplicateTokenBalanceProof() *DuplicateTokenBalanceProof {
	return &DuplicateTokenBalanceProof{
		TokenInfoError: &TokenInfoError{isNested: true},
	}
}

// String is the implementation of the Error interface
func (e *DuplicateTokenBalanceProof) String() string {
	return fmt.Sprintf("%s: The provided token comes with multiple balance proofs. %s",
		DuplicateTokenBalanceProofErrorType, e.TokenInfo.String())
}

// InvalidSignature is a struct that represents an error that occurs when
// the signature is invalid.
type InvalidSignature struct{}

// String is the implementation of the Error interface
func (e *InvalidSignature) String() string {
	return fmt.Sprintf("%s: The provided signature is invalid.", InvalidSignatureErrorType)
}

// UnknownError is a struct that represents an error that occurs when
// an unknown error is encountered.
type UnknownError struct{}

// String is the implementation of the Error interface
func (e *UnknownError) String() string {
	return fmt.Sprintf("%s: An unknown error occurred.", UnknownErrorType)
}

// InvalidImportedBridgeExit is a struct that represents an error that occurs when
// the imported bridge exit is invalid.
type InvalidImportedBridgeExit struct {
	GlobalIndex *GlobalIndex `json:"global_index"`
	ErrorType   string       `json:"error_type"`
}

// String is the implementation of the Error interface
func (e *InvalidImportedBridgeExit) String() string {
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

func (e *InvalidImportedBridgeExit) UnmarshalFromMap(data interface{}) error {
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
