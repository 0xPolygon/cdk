// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package verifybatchesmock

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

// VerifybatchesmockMetaData contains all meta data concerning the Verifybatchesmock contract.
var VerifybatchesmockMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"contractIPolygonZkEVMGlobalExitRootV2\",\"name\":\"_globalExitRootManager\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint32\",\"name\":\"rollupID\",\"type\":\"uint32\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"numBatch\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"stateRoot\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"exitRoot\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"aggregator\",\"type\":\"address\"}],\"name\":\"VerifyBatches\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"getRollupExitRoot\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"globalExitRootManager\",\"outputs\":[{\"internalType\":\"contractIPolygonZkEVMGlobalExitRootV2\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"rollupCount\",\"outputs\":[{\"internalType\":\"uint32\",\"name\":\"\",\"type\":\"uint32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"rollupID\",\"type\":\"uint32\"}],\"name\":\"rollupIDToLastExitRoot\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"rollupID\",\"type\":\"uint32\"},{\"internalType\":\"uint64\",\"name\":\"finalNewBatch\",\"type\":\"uint64\"},{\"internalType\":\"bytes32\",\"name\":\"newLocalExitRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"newStateRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bool\",\"name\":\"updateGER\",\"type\":\"bool\"}],\"name\":\"verifyBatches\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x60a060405234801561001057600080fd5b5060405161082938038061082983398101604081905261002f91610040565b6001600160a01b0316608052610070565b60006020828403121561005257600080fd5b81516001600160a01b038116811461006957600080fd5b9392505050565b60805161079861009160003960008181609c01526104e301526107986000f3fe608060405234801561001057600080fd5b50600436106100575760003560e01c80630680cf5c1461005c578063a2967d991461008f578063d02103ca14610097578063db3abdb9146100d6578063f4e92675146100eb575b600080fd5b61007c61006a3660046105de565b60016020526000908152604090205481565b6040519081526020015b60405180910390f35b61007c610110565b6100be7f000000000000000000000000000000000000000000000000000000000000000081565b6040516001600160a01b039091168152602001610086565b6100e96100e4366004610600565b610499565b005b6000546100fb9063ffffffff1681565b60405163ffffffff9091168152602001610086565b6000805463ffffffff1680820361012957506000919050565b60008167ffffffffffffffff8111156101445761014461066f565b60405190808252806020026020018201604052801561016d578160200160208202803683370190505b50905060005b828110156101d35760016000610189838361069b565b63ffffffff1663ffffffff168152602001908152602001600020548282815181106101b6576101b66106b4565b6020908102919091010152806101cb816106ca565b915050610173565b50600060205b836001146103fd5760006101ee6002866106f9565b6101f960028761070d565b610203919061069b565b905060008167ffffffffffffffff8111156102205761022061066f565b604051908082528060200260200182016040528015610249578160200160208202803683370190505b50905060005b828110156103ad57610262600184610721565b8114801561027a57506102766002886106f9565b6001145b156102f7578561028b826002610734565b8151811061029b5761029b6106b4565b6020026020010151856040516020016102be929190918252602082015260400190565b604051602081830303815290604052805190602001208282815181106102e6576102e66106b4565b60200260200101818152505061039b565b85610303826002610734565b81518110610313576103136106b4565b6020026020010151868260026103299190610734565b61033490600161069b565b81518110610344576103446106b4565b6020026020010151604051602001610366929190918252602082015260400190565b6040516020818303038152906040528051906020012082828151811061038e5761038e6106b4565b6020026020010181815250505b806103a5816106ca565b91505061024f565b5080945081955083846040516020016103d0929190918252602082015260400190565b60405160208183030381529060405280519060200120935082806103f39061074b565b93505050506101d9565b600083600081518110610412576104126106b4565b6020026020010151905060005b8281101561048f57604080516020810184905290810185905260600160408051601f19818403018152828252805160209182012090830187905290820186905292506060016040516020818303038152906040528051906020012093508080610487906106ca565b91505061041f565b5095945050505050565b60005463ffffffff90811690861611156104c3576000805463ffffffff191663ffffffff87161790555b63ffffffff851660009081526001602052604090208390558015610569577f00000000000000000000000000000000000000000000000000000000000000006001600160a01b03166333d6247d610518610110565b6040518263ffffffff1660e01b815260040161053691815260200190565b600060405180830381600087803b15801561055057600080fd5b505af1158015610564573d6000803e3d6000fd5b505050505b6040805167ffffffffffffffff8616815260208101849052908101849052339063ffffffff8716907faac1e7a157b259544ebacd6e8a82ae5d6c8f174e12aa48696277bcc9a661f0b49060600160405180910390a35050505050565b803563ffffffff811681146105d957600080fd5b919050565b6000602082840312156105f057600080fd5b6105f9826105c5565b9392505050565b600080600080600060a0868803121561061857600080fd5b610621866105c5565b9450602086013567ffffffffffffffff8116811461063e57600080fd5b935060408601359250606086013591506080860135801515811461066157600080fd5b809150509295509295909350565b634e487b7160e01b600052604160045260246000fd5b634e487b7160e01b600052601160045260246000fd5b808201808211156106ae576106ae610685565b92915050565b634e487b7160e01b600052603260045260246000fd5b6000600182016106dc576106dc610685565b5060010190565b634e487b7160e01b600052601260045260246000fd5b600082610708576107086106e3565b500690565b60008261071c5761071c6106e3565b500490565b818103818111156106ae576106ae610685565b80820281158282048414176106ae576106ae610685565b60008161075a5761075a610685565b50600019019056fea26469706673582212205adc139a1c2a423d3d8d0db882b69ac1b5cdcb3419bc6315ca33eeac9aa68a7464736f6c63430008120033",
}

// VerifybatchesmockABI is the input ABI used to generate the binding from.
// Deprecated: Use VerifybatchesmockMetaData.ABI instead.
var VerifybatchesmockABI = VerifybatchesmockMetaData.ABI

// VerifybatchesmockBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use VerifybatchesmockMetaData.Bin instead.
var VerifybatchesmockBin = VerifybatchesmockMetaData.Bin

// DeployVerifybatchesmock deploys a new Ethereum contract, binding an instance of Verifybatchesmock to it.
func DeployVerifybatchesmock(auth *bind.TransactOpts, backend bind.ContractBackend, _globalExitRootManager common.Address) (common.Address, *types.Transaction, *Verifybatchesmock, error) {
	parsed, err := VerifybatchesmockMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(VerifybatchesmockBin), backend, _globalExitRootManager)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Verifybatchesmock{VerifybatchesmockCaller: VerifybatchesmockCaller{contract: contract}, VerifybatchesmockTransactor: VerifybatchesmockTransactor{contract: contract}, VerifybatchesmockFilterer: VerifybatchesmockFilterer{contract: contract}}, nil
}

// Verifybatchesmock is an auto generated Go binding around an Ethereum contract.
type Verifybatchesmock struct {
	VerifybatchesmockCaller     // Read-only binding to the contract
	VerifybatchesmockTransactor // Write-only binding to the contract
	VerifybatchesmockFilterer   // Log filterer for contract events
}

// VerifybatchesmockCaller is an auto generated read-only Go binding around an Ethereum contract.
type VerifybatchesmockCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// VerifybatchesmockTransactor is an auto generated write-only Go binding around an Ethereum contract.
type VerifybatchesmockTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// VerifybatchesmockFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type VerifybatchesmockFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// VerifybatchesmockSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type VerifybatchesmockSession struct {
	Contract     *Verifybatchesmock // Generic contract binding to set the session for
	CallOpts     bind.CallOpts      // Call options to use throughout this session
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// VerifybatchesmockCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type VerifybatchesmockCallerSession struct {
	Contract *VerifybatchesmockCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts            // Call options to use throughout this session
}

// VerifybatchesmockTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type VerifybatchesmockTransactorSession struct {
	Contract     *VerifybatchesmockTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts            // Transaction auth options to use throughout this session
}

// VerifybatchesmockRaw is an auto generated low-level Go binding around an Ethereum contract.
type VerifybatchesmockRaw struct {
	Contract *Verifybatchesmock // Generic contract binding to access the raw methods on
}

// VerifybatchesmockCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type VerifybatchesmockCallerRaw struct {
	Contract *VerifybatchesmockCaller // Generic read-only contract binding to access the raw methods on
}

// VerifybatchesmockTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type VerifybatchesmockTransactorRaw struct {
	Contract *VerifybatchesmockTransactor // Generic write-only contract binding to access the raw methods on
}

// NewVerifybatchesmock creates a new instance of Verifybatchesmock, bound to a specific deployed contract.
func NewVerifybatchesmock(address common.Address, backend bind.ContractBackend) (*Verifybatchesmock, error) {
	contract, err := bindVerifybatchesmock(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Verifybatchesmock{VerifybatchesmockCaller: VerifybatchesmockCaller{contract: contract}, VerifybatchesmockTransactor: VerifybatchesmockTransactor{contract: contract}, VerifybatchesmockFilterer: VerifybatchesmockFilterer{contract: contract}}, nil
}

// NewVerifybatchesmockCaller creates a new read-only instance of Verifybatchesmock, bound to a specific deployed contract.
func NewVerifybatchesmockCaller(address common.Address, caller bind.ContractCaller) (*VerifybatchesmockCaller, error) {
	contract, err := bindVerifybatchesmock(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &VerifybatchesmockCaller{contract: contract}, nil
}

// NewVerifybatchesmockTransactor creates a new write-only instance of Verifybatchesmock, bound to a specific deployed contract.
func NewVerifybatchesmockTransactor(address common.Address, transactor bind.ContractTransactor) (*VerifybatchesmockTransactor, error) {
	contract, err := bindVerifybatchesmock(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &VerifybatchesmockTransactor{contract: contract}, nil
}

// NewVerifybatchesmockFilterer creates a new log filterer instance of Verifybatchesmock, bound to a specific deployed contract.
func NewVerifybatchesmockFilterer(address common.Address, filterer bind.ContractFilterer) (*VerifybatchesmockFilterer, error) {
	contract, err := bindVerifybatchesmock(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &VerifybatchesmockFilterer{contract: contract}, nil
}

// bindVerifybatchesmock binds a generic wrapper to an already deployed contract.
func bindVerifybatchesmock(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := VerifybatchesmockMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Verifybatchesmock *VerifybatchesmockRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Verifybatchesmock.Contract.VerifybatchesmockCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Verifybatchesmock *VerifybatchesmockRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Verifybatchesmock.Contract.VerifybatchesmockTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Verifybatchesmock *VerifybatchesmockRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Verifybatchesmock.Contract.VerifybatchesmockTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Verifybatchesmock *VerifybatchesmockCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Verifybatchesmock.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Verifybatchesmock *VerifybatchesmockTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Verifybatchesmock.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Verifybatchesmock *VerifybatchesmockTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Verifybatchesmock.Contract.contract.Transact(opts, method, params...)
}

// GetRollupExitRoot is a free data retrieval call binding the contract method 0xa2967d99.
//
// Solidity: function getRollupExitRoot() view returns(bytes32)
func (_Verifybatchesmock *VerifybatchesmockCaller) GetRollupExitRoot(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Verifybatchesmock.contract.Call(opts, &out, "getRollupExitRoot")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetRollupExitRoot is a free data retrieval call binding the contract method 0xa2967d99.
//
// Solidity: function getRollupExitRoot() view returns(bytes32)
func (_Verifybatchesmock *VerifybatchesmockSession) GetRollupExitRoot() ([32]byte, error) {
	return _Verifybatchesmock.Contract.GetRollupExitRoot(&_Verifybatchesmock.CallOpts)
}

// GetRollupExitRoot is a free data retrieval call binding the contract method 0xa2967d99.
//
// Solidity: function getRollupExitRoot() view returns(bytes32)
func (_Verifybatchesmock *VerifybatchesmockCallerSession) GetRollupExitRoot() ([32]byte, error) {
	return _Verifybatchesmock.Contract.GetRollupExitRoot(&_Verifybatchesmock.CallOpts)
}

// GlobalExitRootManager is a free data retrieval call binding the contract method 0xd02103ca.
//
// Solidity: function globalExitRootManager() view returns(address)
func (_Verifybatchesmock *VerifybatchesmockCaller) GlobalExitRootManager(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Verifybatchesmock.contract.Call(opts, &out, "globalExitRootManager")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GlobalExitRootManager is a free data retrieval call binding the contract method 0xd02103ca.
//
// Solidity: function globalExitRootManager() view returns(address)
func (_Verifybatchesmock *VerifybatchesmockSession) GlobalExitRootManager() (common.Address, error) {
	return _Verifybatchesmock.Contract.GlobalExitRootManager(&_Verifybatchesmock.CallOpts)
}

// GlobalExitRootManager is a free data retrieval call binding the contract method 0xd02103ca.
//
// Solidity: function globalExitRootManager() view returns(address)
func (_Verifybatchesmock *VerifybatchesmockCallerSession) GlobalExitRootManager() (common.Address, error) {
	return _Verifybatchesmock.Contract.GlobalExitRootManager(&_Verifybatchesmock.CallOpts)
}

// RollupCount is a free data retrieval call binding the contract method 0xf4e92675.
//
// Solidity: function rollupCount() view returns(uint32)
func (_Verifybatchesmock *VerifybatchesmockCaller) RollupCount(opts *bind.CallOpts) (uint32, error) {
	var out []interface{}
	err := _Verifybatchesmock.contract.Call(opts, &out, "rollupCount")

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// RollupCount is a free data retrieval call binding the contract method 0xf4e92675.
//
// Solidity: function rollupCount() view returns(uint32)
func (_Verifybatchesmock *VerifybatchesmockSession) RollupCount() (uint32, error) {
	return _Verifybatchesmock.Contract.RollupCount(&_Verifybatchesmock.CallOpts)
}

// RollupCount is a free data retrieval call binding the contract method 0xf4e92675.
//
// Solidity: function rollupCount() view returns(uint32)
func (_Verifybatchesmock *VerifybatchesmockCallerSession) RollupCount() (uint32, error) {
	return _Verifybatchesmock.Contract.RollupCount(&_Verifybatchesmock.CallOpts)
}

// RollupIDToLastExitRoot is a free data retrieval call binding the contract method 0x0680cf5c.
//
// Solidity: function rollupIDToLastExitRoot(uint32 rollupID) view returns(bytes32)
func (_Verifybatchesmock *VerifybatchesmockCaller) RollupIDToLastExitRoot(opts *bind.CallOpts, rollupID uint32) ([32]byte, error) {
	var out []interface{}
	err := _Verifybatchesmock.contract.Call(opts, &out, "rollupIDToLastExitRoot", rollupID)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// RollupIDToLastExitRoot is a free data retrieval call binding the contract method 0x0680cf5c.
//
// Solidity: function rollupIDToLastExitRoot(uint32 rollupID) view returns(bytes32)
func (_Verifybatchesmock *VerifybatchesmockSession) RollupIDToLastExitRoot(rollupID uint32) ([32]byte, error) {
	return _Verifybatchesmock.Contract.RollupIDToLastExitRoot(&_Verifybatchesmock.CallOpts, rollupID)
}

// RollupIDToLastExitRoot is a free data retrieval call binding the contract method 0x0680cf5c.
//
// Solidity: function rollupIDToLastExitRoot(uint32 rollupID) view returns(bytes32)
func (_Verifybatchesmock *VerifybatchesmockCallerSession) RollupIDToLastExitRoot(rollupID uint32) ([32]byte, error) {
	return _Verifybatchesmock.Contract.RollupIDToLastExitRoot(&_Verifybatchesmock.CallOpts, rollupID)
}

// VerifyBatches is a paid mutator transaction binding the contract method 0xdb3abdb9.
//
// Solidity: function verifyBatches(uint32 rollupID, uint64 finalNewBatch, bytes32 newLocalExitRoot, bytes32 newStateRoot, bool updateGER) returns()
func (_Verifybatchesmock *VerifybatchesmockTransactor) VerifyBatches(opts *bind.TransactOpts, rollupID uint32, finalNewBatch uint64, newLocalExitRoot [32]byte, newStateRoot [32]byte, updateGER bool) (*types.Transaction, error) {
	return _Verifybatchesmock.contract.Transact(opts, "verifyBatches", rollupID, finalNewBatch, newLocalExitRoot, newStateRoot, updateGER)
}

// VerifyBatches is a paid mutator transaction binding the contract method 0xdb3abdb9.
//
// Solidity: function verifyBatches(uint32 rollupID, uint64 finalNewBatch, bytes32 newLocalExitRoot, bytes32 newStateRoot, bool updateGER) returns()
func (_Verifybatchesmock *VerifybatchesmockSession) VerifyBatches(rollupID uint32, finalNewBatch uint64, newLocalExitRoot [32]byte, newStateRoot [32]byte, updateGER bool) (*types.Transaction, error) {
	return _Verifybatchesmock.Contract.VerifyBatches(&_Verifybatchesmock.TransactOpts, rollupID, finalNewBatch, newLocalExitRoot, newStateRoot, updateGER)
}

// VerifyBatches is a paid mutator transaction binding the contract method 0xdb3abdb9.
//
// Solidity: function verifyBatches(uint32 rollupID, uint64 finalNewBatch, bytes32 newLocalExitRoot, bytes32 newStateRoot, bool updateGER) returns()
func (_Verifybatchesmock *VerifybatchesmockTransactorSession) VerifyBatches(rollupID uint32, finalNewBatch uint64, newLocalExitRoot [32]byte, newStateRoot [32]byte, updateGER bool) (*types.Transaction, error) {
	return _Verifybatchesmock.Contract.VerifyBatches(&_Verifybatchesmock.TransactOpts, rollupID, finalNewBatch, newLocalExitRoot, newStateRoot, updateGER)
}

// VerifybatchesmockVerifyBatchesIterator is returned from FilterVerifyBatches and is used to iterate over the raw logs and unpacked data for VerifyBatches events raised by the Verifybatchesmock contract.
type VerifybatchesmockVerifyBatchesIterator struct {
	Event *VerifybatchesmockVerifyBatches // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *VerifybatchesmockVerifyBatchesIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(VerifybatchesmockVerifyBatches)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(VerifybatchesmockVerifyBatches)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *VerifybatchesmockVerifyBatchesIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *VerifybatchesmockVerifyBatchesIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// VerifybatchesmockVerifyBatches represents a VerifyBatches event raised by the Verifybatchesmock contract.
type VerifybatchesmockVerifyBatches struct {
	RollupID   uint32
	NumBatch   uint64
	StateRoot  [32]byte
	ExitRoot   [32]byte
	Aggregator common.Address
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterVerifyBatches is a free log retrieval operation binding the contract event 0xaac1e7a157b259544ebacd6e8a82ae5d6c8f174e12aa48696277bcc9a661f0b4.
//
// Solidity: event VerifyBatches(uint32 indexed rollupID, uint64 numBatch, bytes32 stateRoot, bytes32 exitRoot, address indexed aggregator)
func (_Verifybatchesmock *VerifybatchesmockFilterer) FilterVerifyBatches(opts *bind.FilterOpts, rollupID []uint32, aggregator []common.Address) (*VerifybatchesmockVerifyBatchesIterator, error) {

	var rollupIDRule []interface{}
	for _, rollupIDItem := range rollupID {
		rollupIDRule = append(rollupIDRule, rollupIDItem)
	}

	var aggregatorRule []interface{}
	for _, aggregatorItem := range aggregator {
		aggregatorRule = append(aggregatorRule, aggregatorItem)
	}

	logs, sub, err := _Verifybatchesmock.contract.FilterLogs(opts, "VerifyBatches", rollupIDRule, aggregatorRule)
	if err != nil {
		return nil, err
	}
	return &VerifybatchesmockVerifyBatchesIterator{contract: _Verifybatchesmock.contract, event: "VerifyBatches", logs: logs, sub: sub}, nil
}

// WatchVerifyBatches is a free log subscription operation binding the contract event 0xaac1e7a157b259544ebacd6e8a82ae5d6c8f174e12aa48696277bcc9a661f0b4.
//
// Solidity: event VerifyBatches(uint32 indexed rollupID, uint64 numBatch, bytes32 stateRoot, bytes32 exitRoot, address indexed aggregator)
func (_Verifybatchesmock *VerifybatchesmockFilterer) WatchVerifyBatches(opts *bind.WatchOpts, sink chan<- *VerifybatchesmockVerifyBatches, rollupID []uint32, aggregator []common.Address) (event.Subscription, error) {

	var rollupIDRule []interface{}
	for _, rollupIDItem := range rollupID {
		rollupIDRule = append(rollupIDRule, rollupIDItem)
	}

	var aggregatorRule []interface{}
	for _, aggregatorItem := range aggregator {
		aggregatorRule = append(aggregatorRule, aggregatorItem)
	}

	logs, sub, err := _Verifybatchesmock.contract.WatchLogs(opts, "VerifyBatches", rollupIDRule, aggregatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(VerifybatchesmockVerifyBatches)
				if err := _Verifybatchesmock.contract.UnpackLog(event, "VerifyBatches", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseVerifyBatches is a log parse operation binding the contract event 0xaac1e7a157b259544ebacd6e8a82ae5d6c8f174e12aa48696277bcc9a661f0b4.
//
// Solidity: event VerifyBatches(uint32 indexed rollupID, uint64 numBatch, bytes32 stateRoot, bytes32 exitRoot, address indexed aggregator)
func (_Verifybatchesmock *VerifybatchesmockFilterer) ParseVerifyBatches(log types.Log) (*VerifybatchesmockVerifyBatches, error) {
	event := new(VerifybatchesmockVerifyBatches)
	if err := _Verifybatchesmock.contract.UnpackLog(event, "VerifyBatches", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
