// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package claimmockcaller

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// ClaimmockcallerMetaData contains all meta data concerning the Claimmockcaller contract.
var ClaimmockcallerMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"contractIClaimMock\",\"name\":\"_claimMock\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"bytes32[32]\",\"name\":\"smtProofLocalExitRoot\",\"type\":\"bytes32[32]\"},{\"internalType\":\"bytes32[32]\",\"name\":\"smtProofRollupExitRoot\",\"type\":\"bytes32[32]\"},{\"internalType\":\"uint256\",\"name\":\"globalIndex\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"mainnetExitRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"rollupExitRoot\",\"type\":\"bytes32\"},{\"internalType\":\"uint32\",\"name\":\"originNetwork\",\"type\":\"uint32\"},{\"internalType\":\"address\",\"name\":\"originTokenAddress\",\"type\":\"address\"},{\"internalType\":\"uint32\",\"name\":\"destinationNetwork\",\"type\":\"uint32\"},{\"internalType\":\"address\",\"name\":\"destinationAddress\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"metadata\",\"type\":\"bytes\"}],\"name\":\"claimAsset\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32[32]\",\"name\":\"smtProofLocalExitRoot\",\"type\":\"bytes32[32]\"},{\"internalType\":\"bytes32[32]\",\"name\":\"smtProofRollupExitRoot\",\"type\":\"bytes32[32]\"},{\"internalType\":\"uint256\",\"name\":\"globalIndex\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"mainnetExitRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"rollupExitRoot\",\"type\":\"bytes32\"},{\"internalType\":\"uint32\",\"name\":\"originNetwork\",\"type\":\"uint32\"},{\"internalType\":\"address\",\"name\":\"originAddress\",\"type\":\"address\"},{\"internalType\":\"uint32\",\"name\":\"destinationNetwork\",\"type\":\"uint32\"},{\"internalType\":\"address\",\"name\":\"destinationAddress\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"metadata\",\"type\":\"bytes\"}],\"name\":\"claimMessage\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"claimMock\",\"outputs\":[{\"internalType\":\"contractIClaimMock\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
	Bin: "0x60a060405234801561001057600080fd5b5060405161050238038061050283398101604081905261002f91610040565b6001600160a01b0316608052610070565b60006020828403121561005257600080fd5b81516001600160a01b038116811461006957600080fd5b9392505050565b60805161046b61009760003960008181604b0152818160fb01526101c3015261046b6000f3fe608060405234801561001057600080fd5b50600436106100415760003560e01c806383f5b00614610046578063ccaa2d1114610096578063f5efcd79146100ab575b600080fd5b61006d7f000000000000000000000000000000000000000000000000000000000000000081565b60405173ffffffffffffffffffffffffffffffffffffffff909116815260200160405180910390f35b6100a96100a4366004610263565b6100be565b005b6100a96100b9366004610263565b610186565b6040517fccaa2d1100000000000000000000000000000000000000000000000000000000815273ffffffffffffffffffffffffffffffffffffffff7f0000000000000000000000000000000000000000000000000000000000000000169063ccaa2d1190610146908f908f908f908f908f908f908f908f908f908f908f908f9060040161036b565b600060405180830381600087803b15801561016057600080fd5b505af1158015610174573d6000803e3d6000fd5b50505050505050505050505050505050565b6040517ff5efcd7900000000000000000000000000000000000000000000000000000000815273ffffffffffffffffffffffffffffffffffffffff7f0000000000000000000000000000000000000000000000000000000000000000169063f5efcd7990610146908f908f908f908f908f908f908f908f908f908f908f908f9060040161036b565b80610400810183101561022057600080fd5b92915050565b803563ffffffff8116811461023a57600080fd5b919050565b803573ffffffffffffffffffffffffffffffffffffffff8116811461023a57600080fd5b6000806000806000806000806000806000806109208d8f03121561028657600080fd5b6102908e8e61020e565b9b506102a08e6104008f0161020e565b9a506108008d013599506108208d013598506108408d013597506102c76108608e01610226565b96506102d66108808e0161023f565b95506102e56108a08e01610226565b94506102f46108c08e0161023f565b93506108e08d013592506109008d013567ffffffffffffffff8082111561031a57600080fd5b818f0191508f601f83011261032e57600080fd5b808235111561033c57600080fd5b508e60208235830101111561035057600080fd5b60208101925080359150509295989b509295989b509295989b565b6000610400808f8437808e82850137508b6108008301528a6108208301528961084083015263ffffffff808a1661086084015273ffffffffffffffffffffffffffffffffffffffff808a166108808501528189166108a08501528088166108c08501525050846108e0830152610920610900830152826109208301526109408385828501376000838501820152601f9093017fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0169091019091019c9b50505050505050505050505056fea26469706673582212206727c3a78a95fb986d1a4568ebb299bd4398120ac4cd67a415f2c5c02e5ee78f64736f6c63430008140033",
}

// ClaimmockcallerABI is the input ABI used to generate the binding from.
// Deprecated: Use ClaimmockcallerMetaData.ABI instead.
var ClaimmockcallerABI = ClaimmockcallerMetaData.ABI

// ClaimmockcallerBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use ClaimmockcallerMetaData.Bin instead.
var ClaimmockcallerBin = ClaimmockcallerMetaData.Bin

// DeployClaimmockcaller deploys a new Ethereum contract, binding an instance of Claimmockcaller to it.
func DeployClaimmockcaller(auth *bind.TransactOpts, backend bind.ContractBackend, _claimMock common.Address) (common.Address, *types.Transaction, *Claimmockcaller, error) {
	parsed, err := ClaimmockcallerMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(ClaimmockcallerBin), backend, _claimMock)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Claimmockcaller{ClaimmockcallerCaller: ClaimmockcallerCaller{contract: contract}, ClaimmockcallerTransactor: ClaimmockcallerTransactor{contract: contract}, ClaimmockcallerFilterer: ClaimmockcallerFilterer{contract: contract}}, nil
}

// Claimmockcaller is an auto generated Go binding around an Ethereum contract.
type Claimmockcaller struct {
	ClaimmockcallerCaller     // Read-only binding to the contract
	ClaimmockcallerTransactor // Write-only binding to the contract
	ClaimmockcallerFilterer   // Log filterer for contract events
}

// ClaimmockcallerCaller is an auto generated read-only Go binding around an Ethereum contract.
type ClaimmockcallerCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ClaimmockcallerTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ClaimmockcallerTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ClaimmockcallerFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ClaimmockcallerFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ClaimmockcallerSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ClaimmockcallerSession struct {
	Contract     *Claimmockcaller  // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ClaimmockcallerCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ClaimmockcallerCallerSession struct {
	Contract *ClaimmockcallerCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts          // Call options to use throughout this session
}

// ClaimmockcallerTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ClaimmockcallerTransactorSession struct {
	Contract     *ClaimmockcallerTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts          // Transaction auth options to use throughout this session
}

// ClaimmockcallerRaw is an auto generated low-level Go binding around an Ethereum contract.
type ClaimmockcallerRaw struct {
	Contract *Claimmockcaller // Generic contract binding to access the raw methods on
}

// ClaimmockcallerCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ClaimmockcallerCallerRaw struct {
	Contract *ClaimmockcallerCaller // Generic read-only contract binding to access the raw methods on
}

// ClaimmockcallerTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ClaimmockcallerTransactorRaw struct {
	Contract *ClaimmockcallerTransactor // Generic write-only contract binding to access the raw methods on
}

// NewClaimmockcaller creates a new instance of Claimmockcaller, bound to a specific deployed contract.
func NewClaimmockcaller(address common.Address, backend bind.ContractBackend) (*Claimmockcaller, error) {
	contract, err := bindClaimmockcaller(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Claimmockcaller{ClaimmockcallerCaller: ClaimmockcallerCaller{contract: contract}, ClaimmockcallerTransactor: ClaimmockcallerTransactor{contract: contract}, ClaimmockcallerFilterer: ClaimmockcallerFilterer{contract: contract}}, nil
}

// NewClaimmockcallerCaller creates a new read-only instance of Claimmockcaller, bound to a specific deployed contract.
func NewClaimmockcallerCaller(address common.Address, caller bind.ContractCaller) (*ClaimmockcallerCaller, error) {
	contract, err := bindClaimmockcaller(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ClaimmockcallerCaller{contract: contract}, nil
}

// NewClaimmockcallerTransactor creates a new write-only instance of Claimmockcaller, bound to a specific deployed contract.
func NewClaimmockcallerTransactor(address common.Address, transactor bind.ContractTransactor) (*ClaimmockcallerTransactor, error) {
	contract, err := bindClaimmockcaller(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ClaimmockcallerTransactor{contract: contract}, nil
}

// NewClaimmockcallerFilterer creates a new log filterer instance of Claimmockcaller, bound to a specific deployed contract.
func NewClaimmockcallerFilterer(address common.Address, filterer bind.ContractFilterer) (*ClaimmockcallerFilterer, error) {
	contract, err := bindClaimmockcaller(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ClaimmockcallerFilterer{contract: contract}, nil
}

// bindClaimmockcaller binds a generic wrapper to an already deployed contract.
func bindClaimmockcaller(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := ClaimmockcallerMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Claimmockcaller *ClaimmockcallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Claimmockcaller.Contract.ClaimmockcallerCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Claimmockcaller *ClaimmockcallerRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Claimmockcaller.Contract.ClaimmockcallerTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Claimmockcaller *ClaimmockcallerRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Claimmockcaller.Contract.ClaimmockcallerTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Claimmockcaller *ClaimmockcallerCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Claimmockcaller.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Claimmockcaller *ClaimmockcallerTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Claimmockcaller.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Claimmockcaller *ClaimmockcallerTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Claimmockcaller.Contract.contract.Transact(opts, method, params...)
}

// ClaimMock is a free data retrieval call binding the contract method 0x83f5b006.
//
// Solidity: function claimMock() view returns(address)
func (_Claimmockcaller *ClaimmockcallerCaller) ClaimMock(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Claimmockcaller.contract.Call(opts, &out, "claimMock")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// ClaimMock is a free data retrieval call binding the contract method 0x83f5b006.
//
// Solidity: function claimMock() view returns(address)
func (_Claimmockcaller *ClaimmockcallerSession) ClaimMock() (common.Address, error) {
	return _Claimmockcaller.Contract.ClaimMock(&_Claimmockcaller.CallOpts)
}

// ClaimMock is a free data retrieval call binding the contract method 0x83f5b006.
//
// Solidity: function claimMock() view returns(address)
func (_Claimmockcaller *ClaimmockcallerCallerSession) ClaimMock() (common.Address, error) {
	return _Claimmockcaller.Contract.ClaimMock(&_Claimmockcaller.CallOpts)
}

// ClaimAsset is a paid mutator transaction binding the contract method 0xccaa2d11.
//
// Solidity: function claimAsset(bytes32[32] smtProofLocalExitRoot, bytes32[32] smtProofRollupExitRoot, uint256 globalIndex, bytes32 mainnetExitRoot, bytes32 rollupExitRoot, uint32 originNetwork, address originTokenAddress, uint32 destinationNetwork, address destinationAddress, uint256 amount, bytes metadata) returns()
func (_Claimmockcaller *ClaimmockcallerTransactor) ClaimAsset(opts *bind.TransactOpts, smtProofLocalExitRoot [32][32]byte, smtProofRollupExitRoot [32][32]byte, globalIndex *big.Int, mainnetExitRoot [32]byte, rollupExitRoot [32]byte, originNetwork uint32, originTokenAddress common.Address, destinationNetwork uint32, destinationAddress common.Address, amount *big.Int, metadata []byte) (*types.Transaction, error) {
	return _Claimmockcaller.contract.Transact(opts, "claimAsset", smtProofLocalExitRoot, smtProofRollupExitRoot, globalIndex, mainnetExitRoot, rollupExitRoot, originNetwork, originTokenAddress, destinationNetwork, destinationAddress, amount, metadata)
}

// ClaimAsset is a paid mutator transaction binding the contract method 0xccaa2d11.
//
// Solidity: function claimAsset(bytes32[32] smtProofLocalExitRoot, bytes32[32] smtProofRollupExitRoot, uint256 globalIndex, bytes32 mainnetExitRoot, bytes32 rollupExitRoot, uint32 originNetwork, address originTokenAddress, uint32 destinationNetwork, address destinationAddress, uint256 amount, bytes metadata) returns()
func (_Claimmockcaller *ClaimmockcallerSession) ClaimAsset(smtProofLocalExitRoot [32][32]byte, smtProofRollupExitRoot [32][32]byte, globalIndex *big.Int, mainnetExitRoot [32]byte, rollupExitRoot [32]byte, originNetwork uint32, originTokenAddress common.Address, destinationNetwork uint32, destinationAddress common.Address, amount *big.Int, metadata []byte) (*types.Transaction, error) {
	return _Claimmockcaller.Contract.ClaimAsset(&_Claimmockcaller.TransactOpts, smtProofLocalExitRoot, smtProofRollupExitRoot, globalIndex, mainnetExitRoot, rollupExitRoot, originNetwork, originTokenAddress, destinationNetwork, destinationAddress, amount, metadata)
}

// ClaimAsset is a paid mutator transaction binding the contract method 0xccaa2d11.
//
// Solidity: function claimAsset(bytes32[32] smtProofLocalExitRoot, bytes32[32] smtProofRollupExitRoot, uint256 globalIndex, bytes32 mainnetExitRoot, bytes32 rollupExitRoot, uint32 originNetwork, address originTokenAddress, uint32 destinationNetwork, address destinationAddress, uint256 amount, bytes metadata) returns()
func (_Claimmockcaller *ClaimmockcallerTransactorSession) ClaimAsset(smtProofLocalExitRoot [32][32]byte, smtProofRollupExitRoot [32][32]byte, globalIndex *big.Int, mainnetExitRoot [32]byte, rollupExitRoot [32]byte, originNetwork uint32, originTokenAddress common.Address, destinationNetwork uint32, destinationAddress common.Address, amount *big.Int, metadata []byte) (*types.Transaction, error) {
	return _Claimmockcaller.Contract.ClaimAsset(&_Claimmockcaller.TransactOpts, smtProofLocalExitRoot, smtProofRollupExitRoot, globalIndex, mainnetExitRoot, rollupExitRoot, originNetwork, originTokenAddress, destinationNetwork, destinationAddress, amount, metadata)
}

// ClaimMessage is a paid mutator transaction binding the contract method 0xf5efcd79.
//
// Solidity: function claimMessage(bytes32[32] smtProofLocalExitRoot, bytes32[32] smtProofRollupExitRoot, uint256 globalIndex, bytes32 mainnetExitRoot, bytes32 rollupExitRoot, uint32 originNetwork, address originAddress, uint32 destinationNetwork, address destinationAddress, uint256 amount, bytes metadata) returns()
func (_Claimmockcaller *ClaimmockcallerTransactor) ClaimMessage(opts *bind.TransactOpts, smtProofLocalExitRoot [32][32]byte, smtProofRollupExitRoot [32][32]byte, globalIndex *big.Int, mainnetExitRoot [32]byte, rollupExitRoot [32]byte, originNetwork uint32, originAddress common.Address, destinationNetwork uint32, destinationAddress common.Address, amount *big.Int, metadata []byte) (*types.Transaction, error) {
	return _Claimmockcaller.contract.Transact(opts, "claimMessage", smtProofLocalExitRoot, smtProofRollupExitRoot, globalIndex, mainnetExitRoot, rollupExitRoot, originNetwork, originAddress, destinationNetwork, destinationAddress, amount, metadata)
}

// ClaimMessage is a paid mutator transaction binding the contract method 0xf5efcd79.
//
// Solidity: function claimMessage(bytes32[32] smtProofLocalExitRoot, bytes32[32] smtProofRollupExitRoot, uint256 globalIndex, bytes32 mainnetExitRoot, bytes32 rollupExitRoot, uint32 originNetwork, address originAddress, uint32 destinationNetwork, address destinationAddress, uint256 amount, bytes metadata) returns()
func (_Claimmockcaller *ClaimmockcallerSession) ClaimMessage(smtProofLocalExitRoot [32][32]byte, smtProofRollupExitRoot [32][32]byte, globalIndex *big.Int, mainnetExitRoot [32]byte, rollupExitRoot [32]byte, originNetwork uint32, originAddress common.Address, destinationNetwork uint32, destinationAddress common.Address, amount *big.Int, metadata []byte) (*types.Transaction, error) {
	return _Claimmockcaller.Contract.ClaimMessage(&_Claimmockcaller.TransactOpts, smtProofLocalExitRoot, smtProofRollupExitRoot, globalIndex, mainnetExitRoot, rollupExitRoot, originNetwork, originAddress, destinationNetwork, destinationAddress, amount, metadata)
}

// ClaimMessage is a paid mutator transaction binding the contract method 0xf5efcd79.
//
// Solidity: function claimMessage(bytes32[32] smtProofLocalExitRoot, bytes32[32] smtProofRollupExitRoot, uint256 globalIndex, bytes32 mainnetExitRoot, bytes32 rollupExitRoot, uint32 originNetwork, address originAddress, uint32 destinationNetwork, address destinationAddress, uint256 amount, bytes metadata) returns()
func (_Claimmockcaller *ClaimmockcallerTransactorSession) ClaimMessage(smtProofLocalExitRoot [32][32]byte, smtProofRollupExitRoot [32][32]byte, globalIndex *big.Int, mainnetExitRoot [32]byte, rollupExitRoot [32]byte, originNetwork uint32, originAddress common.Address, destinationNetwork uint32, destinationAddress common.Address, amount *big.Int, metadata []byte) (*types.Transaction, error) {
	return _Claimmockcaller.Contract.ClaimMessage(&_Claimmockcaller.TransactOpts, smtProofLocalExitRoot, smtProofRollupExitRoot, globalIndex, mainnetExitRoot, rollupExitRoot, originNetwork, originAddress, destinationNetwork, destinationAddress, amount, metadata)
}
