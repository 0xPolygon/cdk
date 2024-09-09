// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package claimmock

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

// ClaimmockMetaData contains all meta data concerning the Claimmock contract.
var ClaimmockMetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"globalIndex\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint32\",\"name\":\"originNetwork\",\"type\":\"uint32\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"originAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"destinationAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"ClaimEvent\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"bytes32[32]\",\"name\":\"smtProofLocalExitRoot\",\"type\":\"bytes32[32]\"},{\"internalType\":\"bytes32[32]\",\"name\":\"smtProofRollupExitRoot\",\"type\":\"bytes32[32]\"},{\"internalType\":\"uint256\",\"name\":\"globalIndex\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"mainnetExitRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"rollupExitRoot\",\"type\":\"bytes32\"},{\"internalType\":\"uint32\",\"name\":\"originNetwork\",\"type\":\"uint32\"},{\"internalType\":\"address\",\"name\":\"originTokenAddress\",\"type\":\"address\"},{\"internalType\":\"uint32\",\"name\":\"destinationNetwork\",\"type\":\"uint32\"},{\"internalType\":\"address\",\"name\":\"destinationAddress\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"metadata\",\"type\":\"bytes\"}],\"name\":\"claimAsset\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32[32]\",\"name\":\"smtProofLocalExitRoot\",\"type\":\"bytes32[32]\"},{\"internalType\":\"bytes32[32]\",\"name\":\"smtProofRollupExitRoot\",\"type\":\"bytes32[32]\"},{\"internalType\":\"uint256\",\"name\":\"globalIndex\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"mainnetExitRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"rollupExitRoot\",\"type\":\"bytes32\"},{\"internalType\":\"uint32\",\"name\":\"originNetwork\",\"type\":\"uint32\"},{\"internalType\":\"address\",\"name\":\"originAddress\",\"type\":\"address\"},{\"internalType\":\"uint32\",\"name\":\"destinationNetwork\",\"type\":\"uint32\"},{\"internalType\":\"address\",\"name\":\"destinationAddress\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"metadata\",\"type\":\"bytes\"}],\"name\":\"claimMessage\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"}]",
	Bin: "0x6080806040523461001657610227908161001c8239f35b600080fdfe608080604052600436101561001357600080fd5b600090813560e01c908163ccaa2d11146100b5575063f5efcd791461003757600080fd5b6100403661012f565b5050945097505092509350600134146100b1576040805193845263ffffffff9490941660208401526001600160a01b039182169383019390935292909216606083015260808201527f1df3f2a973a00d6635911755c260704e95e8a5876997546798770f76396fda4d9060a090a180f35b8580fd5b7f1df3f2a973a00d6635911755c260704e95e8a5876997546798770f76396fda4d91508061011e6100e53661012f565b5050968b5263ffffffff90931660208b0152506001600160a01b0390811660408a015216606088015250506080850152505060a0820190565b0390a16001341461012c5780f35b80fd5b906109206003198301126101ec57610404908282116101ec57600492610804928184116101ec579235916108243591610844359163ffffffff906108643582811681036101ec57926001600160a01b03916108843583811681036101ec57936108a43590811681036101ec57926108c43590811681036101ec57916108e435916109043567ffffffffffffffff928382116101ec57806023830112156101ec57818e01359384116101ec57602484830101116101ec576024019190565b600080fdfea2646970667358221220360ea7019315ab59618e13d469f48b1816436744772ab76ff89153af49fb746a64736f6c63430008120033",
}

// ClaimmockABI is the input ABI used to generate the binding from.
// Deprecated: Use ClaimmockMetaData.ABI instead.
var ClaimmockABI = ClaimmockMetaData.ABI

// ClaimmockBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use ClaimmockMetaData.Bin instead.
var ClaimmockBin = ClaimmockMetaData.Bin

// DeployClaimmock deploys a new Ethereum contract, binding an instance of Claimmock to it.
func DeployClaimmock(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Claimmock, error) {
	parsed, err := ClaimmockMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(ClaimmockBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Claimmock{ClaimmockCaller: ClaimmockCaller{contract: contract}, ClaimmockTransactor: ClaimmockTransactor{contract: contract}, ClaimmockFilterer: ClaimmockFilterer{contract: contract}}, nil
}

// Claimmock is an auto generated Go binding around an Ethereum contract.
type Claimmock struct {
	ClaimmockCaller     // Read-only binding to the contract
	ClaimmockTransactor // Write-only binding to the contract
	ClaimmockFilterer   // Log filterer for contract events
}

// ClaimmockCaller is an auto generated read-only Go binding around an Ethereum contract.
type ClaimmockCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ClaimmockTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ClaimmockTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ClaimmockFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ClaimmockFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ClaimmockSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ClaimmockSession struct {
	Contract     *Claimmock        // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ClaimmockCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ClaimmockCallerSession struct {
	Contract *ClaimmockCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts    // Call options to use throughout this session
}

// ClaimmockTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ClaimmockTransactorSession struct {
	Contract     *ClaimmockTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// ClaimmockRaw is an auto generated low-level Go binding around an Ethereum contract.
type ClaimmockRaw struct {
	Contract *Claimmock // Generic contract binding to access the raw methods on
}

// ClaimmockCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ClaimmockCallerRaw struct {
	Contract *ClaimmockCaller // Generic read-only contract binding to access the raw methods on
}

// ClaimmockTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ClaimmockTransactorRaw struct {
	Contract *ClaimmockTransactor // Generic write-only contract binding to access the raw methods on
}

// NewClaimmock creates a new instance of Claimmock, bound to a specific deployed contract.
func NewClaimmock(address common.Address, backend bind.ContractBackend) (*Claimmock, error) {
	contract, err := bindClaimmock(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Claimmock{ClaimmockCaller: ClaimmockCaller{contract: contract}, ClaimmockTransactor: ClaimmockTransactor{contract: contract}, ClaimmockFilterer: ClaimmockFilterer{contract: contract}}, nil
}

// NewClaimmockCaller creates a new read-only instance of Claimmock, bound to a specific deployed contract.
func NewClaimmockCaller(address common.Address, caller bind.ContractCaller) (*ClaimmockCaller, error) {
	contract, err := bindClaimmock(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ClaimmockCaller{contract: contract}, nil
}

// NewClaimmockTransactor creates a new write-only instance of Claimmock, bound to a specific deployed contract.
func NewClaimmockTransactor(address common.Address, transactor bind.ContractTransactor) (*ClaimmockTransactor, error) {
	contract, err := bindClaimmock(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ClaimmockTransactor{contract: contract}, nil
}

// NewClaimmockFilterer creates a new log filterer instance of Claimmock, bound to a specific deployed contract.
func NewClaimmockFilterer(address common.Address, filterer bind.ContractFilterer) (*ClaimmockFilterer, error) {
	contract, err := bindClaimmock(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ClaimmockFilterer{contract: contract}, nil
}

// bindClaimmock binds a generic wrapper to an already deployed contract.
func bindClaimmock(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := ClaimmockMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Claimmock *ClaimmockRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Claimmock.Contract.ClaimmockCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Claimmock *ClaimmockRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Claimmock.Contract.ClaimmockTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Claimmock *ClaimmockRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Claimmock.Contract.ClaimmockTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Claimmock *ClaimmockCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Claimmock.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Claimmock *ClaimmockTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Claimmock.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Claimmock *ClaimmockTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Claimmock.Contract.contract.Transact(opts, method, params...)
}

// ClaimAsset is a paid mutator transaction binding the contract method 0xccaa2d11.
//
// Solidity: function claimAsset(bytes32[32] smtProofLocalExitRoot, bytes32[32] smtProofRollupExitRoot, uint256 globalIndex, bytes32 mainnetExitRoot, bytes32 rollupExitRoot, uint32 originNetwork, address originTokenAddress, uint32 destinationNetwork, address destinationAddress, uint256 amount, bytes metadata) payable returns()
func (_Claimmock *ClaimmockTransactor) ClaimAsset(opts *bind.TransactOpts, smtProofLocalExitRoot [32][32]byte, smtProofRollupExitRoot [32][32]byte, globalIndex *big.Int, mainnetExitRoot [32]byte, rollupExitRoot [32]byte, originNetwork uint32, originTokenAddress common.Address, destinationNetwork uint32, destinationAddress common.Address, amount *big.Int, metadata []byte) (*types.Transaction, error) {
	return _Claimmock.contract.Transact(opts, "claimAsset", smtProofLocalExitRoot, smtProofRollupExitRoot, globalIndex, mainnetExitRoot, rollupExitRoot, originNetwork, originTokenAddress, destinationNetwork, destinationAddress, amount, metadata)
}

// ClaimAsset is a paid mutator transaction binding the contract method 0xccaa2d11.
//
// Solidity: function claimAsset(bytes32[32] smtProofLocalExitRoot, bytes32[32] smtProofRollupExitRoot, uint256 globalIndex, bytes32 mainnetExitRoot, bytes32 rollupExitRoot, uint32 originNetwork, address originTokenAddress, uint32 destinationNetwork, address destinationAddress, uint256 amount, bytes metadata) payable returns()
func (_Claimmock *ClaimmockSession) ClaimAsset(smtProofLocalExitRoot [32][32]byte, smtProofRollupExitRoot [32][32]byte, globalIndex *big.Int, mainnetExitRoot [32]byte, rollupExitRoot [32]byte, originNetwork uint32, originTokenAddress common.Address, destinationNetwork uint32, destinationAddress common.Address, amount *big.Int, metadata []byte) (*types.Transaction, error) {
	return _Claimmock.Contract.ClaimAsset(&_Claimmock.TransactOpts, smtProofLocalExitRoot, smtProofRollupExitRoot, globalIndex, mainnetExitRoot, rollupExitRoot, originNetwork, originTokenAddress, destinationNetwork, destinationAddress, amount, metadata)
}

// ClaimAsset is a paid mutator transaction binding the contract method 0xccaa2d11.
//
// Solidity: function claimAsset(bytes32[32] smtProofLocalExitRoot, bytes32[32] smtProofRollupExitRoot, uint256 globalIndex, bytes32 mainnetExitRoot, bytes32 rollupExitRoot, uint32 originNetwork, address originTokenAddress, uint32 destinationNetwork, address destinationAddress, uint256 amount, bytes metadata) payable returns()
func (_Claimmock *ClaimmockTransactorSession) ClaimAsset(smtProofLocalExitRoot [32][32]byte, smtProofRollupExitRoot [32][32]byte, globalIndex *big.Int, mainnetExitRoot [32]byte, rollupExitRoot [32]byte, originNetwork uint32, originTokenAddress common.Address, destinationNetwork uint32, destinationAddress common.Address, amount *big.Int, metadata []byte) (*types.Transaction, error) {
	return _Claimmock.Contract.ClaimAsset(&_Claimmock.TransactOpts, smtProofLocalExitRoot, smtProofRollupExitRoot, globalIndex, mainnetExitRoot, rollupExitRoot, originNetwork, originTokenAddress, destinationNetwork, destinationAddress, amount, metadata)
}

// ClaimMessage is a paid mutator transaction binding the contract method 0xf5efcd79.
//
// Solidity: function claimMessage(bytes32[32] smtProofLocalExitRoot, bytes32[32] smtProofRollupExitRoot, uint256 globalIndex, bytes32 mainnetExitRoot, bytes32 rollupExitRoot, uint32 originNetwork, address originAddress, uint32 destinationNetwork, address destinationAddress, uint256 amount, bytes metadata) payable returns()
func (_Claimmock *ClaimmockTransactor) ClaimMessage(opts *bind.TransactOpts, smtProofLocalExitRoot [32][32]byte, smtProofRollupExitRoot [32][32]byte, globalIndex *big.Int, mainnetExitRoot [32]byte, rollupExitRoot [32]byte, originNetwork uint32, originAddress common.Address, destinationNetwork uint32, destinationAddress common.Address, amount *big.Int, metadata []byte) (*types.Transaction, error) {
	return _Claimmock.contract.Transact(opts, "claimMessage", smtProofLocalExitRoot, smtProofRollupExitRoot, globalIndex, mainnetExitRoot, rollupExitRoot, originNetwork, originAddress, destinationNetwork, destinationAddress, amount, metadata)
}

// ClaimMessage is a paid mutator transaction binding the contract method 0xf5efcd79.
//
// Solidity: function claimMessage(bytes32[32] smtProofLocalExitRoot, bytes32[32] smtProofRollupExitRoot, uint256 globalIndex, bytes32 mainnetExitRoot, bytes32 rollupExitRoot, uint32 originNetwork, address originAddress, uint32 destinationNetwork, address destinationAddress, uint256 amount, bytes metadata) payable returns()
func (_Claimmock *ClaimmockSession) ClaimMessage(smtProofLocalExitRoot [32][32]byte, smtProofRollupExitRoot [32][32]byte, globalIndex *big.Int, mainnetExitRoot [32]byte, rollupExitRoot [32]byte, originNetwork uint32, originAddress common.Address, destinationNetwork uint32, destinationAddress common.Address, amount *big.Int, metadata []byte) (*types.Transaction, error) {
	return _Claimmock.Contract.ClaimMessage(&_Claimmock.TransactOpts, smtProofLocalExitRoot, smtProofRollupExitRoot, globalIndex, mainnetExitRoot, rollupExitRoot, originNetwork, originAddress, destinationNetwork, destinationAddress, amount, metadata)
}

// ClaimMessage is a paid mutator transaction binding the contract method 0xf5efcd79.
//
// Solidity: function claimMessage(bytes32[32] smtProofLocalExitRoot, bytes32[32] smtProofRollupExitRoot, uint256 globalIndex, bytes32 mainnetExitRoot, bytes32 rollupExitRoot, uint32 originNetwork, address originAddress, uint32 destinationNetwork, address destinationAddress, uint256 amount, bytes metadata) payable returns()
func (_Claimmock *ClaimmockTransactorSession) ClaimMessage(smtProofLocalExitRoot [32][32]byte, smtProofRollupExitRoot [32][32]byte, globalIndex *big.Int, mainnetExitRoot [32]byte, rollupExitRoot [32]byte, originNetwork uint32, originAddress common.Address, destinationNetwork uint32, destinationAddress common.Address, amount *big.Int, metadata []byte) (*types.Transaction, error) {
	return _Claimmock.Contract.ClaimMessage(&_Claimmock.TransactOpts, smtProofLocalExitRoot, smtProofRollupExitRoot, globalIndex, mainnetExitRoot, rollupExitRoot, originNetwork, originAddress, destinationNetwork, destinationAddress, amount, metadata)
}

// ClaimmockClaimEventIterator is returned from FilterClaimEvent and is used to iterate over the raw logs and unpacked data for ClaimEvent events raised by the Claimmock contract.
type ClaimmockClaimEventIterator struct {
	Event *ClaimmockClaimEvent // Event containing the contract specifics and raw log

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
func (it *ClaimmockClaimEventIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ClaimmockClaimEvent)
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
		it.Event = new(ClaimmockClaimEvent)
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
func (it *ClaimmockClaimEventIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ClaimmockClaimEventIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ClaimmockClaimEvent represents a ClaimEvent event raised by the Claimmock contract.
type ClaimmockClaimEvent struct {
	GlobalIndex        *big.Int
	OriginNetwork      uint32
	OriginAddress      common.Address
	DestinationAddress common.Address
	Amount             *big.Int
	Raw                types.Log // Blockchain specific contextual infos
}

// FilterClaimEvent is a free log retrieval operation binding the contract event 0x1df3f2a973a00d6635911755c260704e95e8a5876997546798770f76396fda4d.
//
// Solidity: event ClaimEvent(uint256 globalIndex, uint32 originNetwork, address originAddress, address destinationAddress, uint256 amount)
func (_Claimmock *ClaimmockFilterer) FilterClaimEvent(opts *bind.FilterOpts) (*ClaimmockClaimEventIterator, error) {

	logs, sub, err := _Claimmock.contract.FilterLogs(opts, "ClaimEvent")
	if err != nil {
		return nil, err
	}
	return &ClaimmockClaimEventIterator{contract: _Claimmock.contract, event: "ClaimEvent", logs: logs, sub: sub}, nil
}

// WatchClaimEvent is a free log subscription operation binding the contract event 0x1df3f2a973a00d6635911755c260704e95e8a5876997546798770f76396fda4d.
//
// Solidity: event ClaimEvent(uint256 globalIndex, uint32 originNetwork, address originAddress, address destinationAddress, uint256 amount)
func (_Claimmock *ClaimmockFilterer) WatchClaimEvent(opts *bind.WatchOpts, sink chan<- *ClaimmockClaimEvent) (event.Subscription, error) {

	logs, sub, err := _Claimmock.contract.WatchLogs(opts, "ClaimEvent")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ClaimmockClaimEvent)
				if err := _Claimmock.contract.UnpackLog(event, "ClaimEvent", log); err != nil {
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

// ParseClaimEvent is a log parse operation binding the contract event 0x1df3f2a973a00d6635911755c260704e95e8a5876997546798770f76396fda4d.
//
// Solidity: event ClaimEvent(uint256 globalIndex, uint32 originNetwork, address originAddress, address destinationAddress, uint256 amount)
func (_Claimmock *ClaimmockFilterer) ParseClaimEvent(log types.Log) (*ClaimmockClaimEvent, error) {
	event := new(ClaimmockClaimEvent)
	if err := _Claimmock.contract.UnpackLog(event, "ClaimEvent", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
