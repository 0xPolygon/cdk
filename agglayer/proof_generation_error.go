package agglayer

import (
	"errors"
	"fmt"

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

type Error interface {
	String() string
}

// ProofGenerationError is a struct that represents an error that occurs when generating a proof.
type ProofGenerationError struct {
	GenerationType string
	Errors         []Error
}

// String is the implementation of the Error interface
func (p *ProofGenerationError) String() string {
	return fmt.Sprintf("Proof generation error: %s. %s", p.GenerationType, p.Errors)
}

// UnmarshalFromMap unmarshals the data from a map into a ProofGenerationError struct.
func (p *ProofGenerationError) UnmarshalFromMap(data map[string]interface{}) error {
	inErr, ok := data["InError"]
	if !ok {
		return errors.New("InError field not found")
	}

	inErrMap, ok := inErr.(map[string]interface{})
	if !ok {
		return errNotMap
	}

	inErrData, ok := inErrMap["error"]
	if !ok {
		return errors.New("error field not found")
	}

	inErrDataMap, ok := inErrData.(map[string]interface{})
	if !ok {
		return errNotMap
	}

	proofGenerationError, ok := inErrDataMap["ProofGenerationError"]
	if !ok {
		return errors.New("ProofGenerationError field not found")
	}

	proofGenerationErrorMap, ok := proofGenerationError.(map[string]interface{})
	if !ok {
		return errNotMap
	}

	p.GenerationType = proofGenerationErrorMap["generation_type"].(string)

	sourceMap, ok := proofGenerationErrorMap["source"].(map[string]interface{})
	if !ok {
		return errNotMap
	}

	// there will always be only one key in the source map
	for key, value := range sourceMap {
		data, ok := value.(map[string]interface{})
		if !ok {
			return errNotMap
		}

		switch key {
		case InvalidSignerErrorType:
			invalidSigner := &InvalidSignerError{}
			if err := invalidSigner.UnmarshalFromMap(data); err != nil {
				return err
			}
			p.Errors = append(p.Errors, invalidSigner)
		case InvalidPreviousLERErrorType:
			invalidPreviousLER := NewInvalidPreviousLocalExitRoot()
			if err := invalidPreviousLER.UnmarshalFromMap(data); err != nil {
				return err
			}
			p.Errors = append(p.Errors, invalidPreviousLER)
		case InvalidPreviousBalanceRootErrorType:
			invalidPreviousBalanceRoot := NewInvalidPreviousBalanceRoot()
			if err := invalidPreviousBalanceRoot.UnmarshalFromMap(data); err != nil {
				return err
			}
			p.Errors = append(p.Errors, invalidPreviousBalanceRoot)
		case InvalidPreviousNullifierRootErrorType:
			invalidPreviousNullifierRoot := NewInvalidPreviousNullifierRoot()
			if err := invalidPreviousNullifierRoot.UnmarshalFromMap(data); err != nil {
				return err
			}
			p.Errors = append(p.Errors, invalidPreviousNullifierRoot)
		case InvalidNewLocalExitRootErrorType:
			invalidNewLocalExitRoot := NewInvalidNewLocalExitRoot()
			if err := invalidNewLocalExitRoot.UnmarshalFromMap(data); err != nil {
				return err
			}
			p.Errors = append(p.Errors, invalidNewLocalExitRoot)
		case InvalidNewBalanceRootErrorType:
			invalidNewBalanceRoot := NewInvalidNewBalanceRoot()
			if err := invalidNewBalanceRoot.UnmarshalFromMap(data); err != nil {
				return err
			}
			p.Errors = append(p.Errors, invalidNewBalanceRoot)
		case InvalidNewNullifierRootErrorType:
			invalidNewNullifierRoot := NewInvalidNewNullifierRoot()
			if err := invalidNewNullifierRoot.UnmarshalFromMap(data); err != nil {
				return err
			}
			p.Errors = append(p.Errors, invalidNewNullifierRoot)
		case InvalidImportedExitsRootErrorType:
			invalidImportedExitsRoot := NewInvalidImportedExitsRoot()
			if err := invalidImportedExitsRoot.UnmarshalFromMap(data); err != nil {
				return err
			}
			p.Errors = append(p.Errors, invalidImportedExitsRoot)
		case MismatchImportedExitsRootErrorType:
			p.Errors = append(p.Errors, &MismatchImportedExitsRoot{})
		case InvalidNullifierPathErrorType:
			p.Errors = append(p.Errors, &InvalidNullifierPath{})
		case InvalidBalancePathErrorType:
			p.Errors = append(p.Errors, &InvalidBalancePath{})
		case BalanceOverflowInBridgeExitErrorType:
			p.Errors = append(p.Errors, &BalanceOverflowInBridgeExit{})
		case BalanceUnderflowInBridgeExitErrorType:
			p.Errors = append(p.Errors, &BalanceUnderflowInBridgeExit{})
		case CannotExitToSameNetworkErrorType:
			p.Errors = append(p.Errors, &CannotExitToSameNetwork{})
		case InvalidMessageOriginNetworkErrorType:
			p.Errors = append(p.Errors, &InvalidMessageOriginNetwork{})
		case InvalidL1TokenInfoErrorType:
			invalidL1TokenInfo := &InvalidL1TokenInfo{}
			if err := invalidL1TokenInfo.UnmarshalFromMap(data); err != nil {
				return err
			}
			p.Errors = append(p.Errors, invalidL1TokenInfo)
		case MissingTokenBalanceProofErrorType:
			missingTokenBalanceProof := &MissingTokenBalanceProof{}
			if err := missingTokenBalanceProof.UnmarshalFromMap(data); err != nil {
				return err
			}
			p.Errors = append(p.Errors, missingTokenBalanceProof)
		case DuplicateTokenBalanceProofErrorType:
			duplicateTokenBalanceProof := &DuplicateTokenBalanceProof{}
			if err := duplicateTokenBalanceProof.UnmarshalFromMap(data); err != nil {
				return err
			}
			p.Errors = append(p.Errors, duplicateTokenBalanceProof)
		case InvalidSignatureErrorType:
			p.Errors = append(p.Errors, &InvalidSignature{})
		case InvalidImportedBridgeExitErrorType:
			invalidImportedBridgeExit := &InvalidImportedBridgeExit{}
			if err := invalidImportedBridgeExit.UnmarshalFromMap(data); err != nil {
				return err
			}
			p.Errors = append(p.Errors, invalidImportedBridgeExit)
		case UnknownErrorType:
			p.Errors = append(p.Errors, &UnknownError{})
		default:
			return fmt.Errorf("unknown error type: %s", key)
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

func (e *InvalidSignerError) UnmarshalFromMap(data map[string]interface{}) error {
	e.Declared = common.HexToAddress(data["declared"].(string))
	e.Recovered = common.HexToAddress(data["recovered"].(string))

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
func (e *DeclaredComputedError) UnmarshalFromMap(data map[string]interface{}) error {
	e.Declared = common.HexToHash(data["declared"].(string))
	e.Computed = common.HexToHash(data["computed"].(string))

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

// InvalidL1TokenInfo is a struct that represents an error that occurs when
// the L1 token info is invalid.
type InvalidL1TokenInfo struct {
	TokenInfo *TokenInfo `json:"token_info"`
}

// String is the implementation of the Error interface
func (e *InvalidL1TokenInfo) String() string {
	return fmt.Sprintf("%s: The L1 token info is invalid. %s",
		InvalidL1TokenInfoErrorType, e.TokenInfo.String())
}

func (e *InvalidL1TokenInfo) UnmarshalFromMap(data map[string]interface{}) error {
	ti, err := unmarshalTokenInfo(data)
	if err != nil {
		return err
	}

	e.TokenInfo = ti

	return nil
}

// MissingTokenBalanceProof is a struct that represents an error that occurs when
// the token balance proof is missing.
type MissingTokenBalanceProof struct {
	TokenInfo *TokenInfo `json:"token_info"`
}

// String is the implementation of the Error interface
func (e *MissingTokenBalanceProof) String() string {
	return fmt.Sprintf("%s: The provided token is missing a balance proof. %s",
		MissingTokenBalanceProofErrorType, e.TokenInfo.String())
}

func (e *MissingTokenBalanceProof) UnmarshalFromMap(data map[string]interface{}) error {
	ti, err := unmarshalTokenInfo(data)
	if err != nil {
		return err
	}

	e.TokenInfo = ti

	return nil
}

// DuplicateTokenBalanceProof is a struct that represents an error that occurs when
// the token balance proof is duplicated.
type DuplicateTokenBalanceProof struct {
	TokenInfo *TokenInfo `json:"token_info"`
}

// String is the implementation of the Error interface
func (e *DuplicateTokenBalanceProof) String() string {
	return fmt.Sprintf("%s: The provided token comes with multiple balance proofs. %s",
		DuplicateTokenBalanceProofErrorType, e.TokenInfo.String())
}

func (e *DuplicateTokenBalanceProof) UnmarshalFromMap(data map[string]interface{}) error {
	ti, err := unmarshalTokenInfo(data)
	if err != nil {
		return err
	}

	e.TokenInfo = ti

	return nil
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
		errorDescription = "The global index and the inclusion proof do not both correspond to the same network type: mainnet or rollup."
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

func (e *InvalidImportedBridgeExit) UnmarshalFromMap(data map[string]interface{}) error {
	sourceError, ok := data["source"]
	if !ok {
		return errors.New("source field not found")
	}

	e.ErrorType = sourceError.(string)

	globalIndex, ok := data["global_index"]
	if !ok {
		return errors.New("global_index field not found")
	}

	globalIndexMap, ok := globalIndex.(map[string]interface{})
	if !ok {
		return errNotMap
	}

	e.GlobalIndex = &GlobalIndex{
		MainnetFlag: globalIndexMap["mainnet_flag"].(bool),
		RollupIndex: globalIndexMap["rollup_index"].(uint32),
		LeafIndex:   globalIndexMap["leaf_index"].(uint32),
	}

	return nil
}

// unmashalTokenInfo unmarshals the data from a map into a TokenInfo struct.
func unmarshalTokenInfo(data map[string]interface{}) (*TokenInfo, error) {
	raw, ok := data["TokenInfo"]
	if !ok {
		return nil, errors.New("TokenInfo field not found")
	}

	tokenInfoMap, ok := raw.(map[string]interface{})
	if !ok {
		return nil, errNotMap
	}

	return &TokenInfo{
		OriginNetwork:      tokenInfoMap["origin_network"].(uint32),
		OriginTokenAddress: common.HexToAddress(tokenInfoMap["origin_address"].(string)),
	}, nil
}
