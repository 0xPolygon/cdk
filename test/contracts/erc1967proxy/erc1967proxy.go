// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package erc1967proxy

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

// Erc1967proxyMetaData contains all meta data concerning the Erc1967proxy contract.
var Erc1967proxyMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_logic\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"_data\",\"type\":\"bytes\"}],\"stateMutability\":\"payable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"previousAdmin\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"newAdmin\",\"type\":\"address\"}],\"name\":\"AdminChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"beacon\",\"type\":\"address\"}],\"name\":\"BeaconUpgraded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"implementation\",\"type\":\"address\"}],\"name\":\"Upgraded\",\"type\":\"event\"},{\"stateMutability\":\"payable\",\"type\":\"fallback\"},{\"stateMutability\":\"payable\",\"type\":\"receive\"}]",
	Bin: "0x60806040526040516104ee3803806104ee833981016040819052610022916102de565b61002e82826000610035565b50506103fb565b61003e83610061565b60008251118061004b5750805b1561005c5761005a83836100a1565b505b505050565b61006a816100cd565b6040516001600160a01b038216907fbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b90600090a250565b60606100c683836040518060600160405280602781526020016104c760279139610180565b9392505050565b6001600160a01b0381163b61013f5760405162461bcd60e51b815260206004820152602d60248201527f455243313936373a206e657720696d706c656d656e746174696f6e206973206e60448201526c1bdd08184818dbdb9d1c9858dd609a1b60648201526084015b60405180910390fd5b7f360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbc80546001600160a01b0319166001600160a01b0392909216919091179055565b6060600080856001600160a01b03168560405161019d91906103ac565b600060405180830381855af49150503d80600081146101d8576040519150601f19603f3d011682016040523d82523d6000602084013e6101dd565b606091505b5090925090506101ef868383876101f9565b9695505050505050565b60608315610268578251600003610261576001600160a01b0385163b6102615760405162461bcd60e51b815260206004820152601d60248201527f416464726573733a2063616c6c20746f206e6f6e2d636f6e74726163740000006044820152606401610136565b5081610272565b610272838361027a565b949350505050565b81511561028a5781518083602001fd5b8060405162461bcd60e51b815260040161013691906103c8565b634e487b7160e01b600052604160045260246000fd5b60005b838110156102d55781810151838201526020016102bd565b50506000910152565b600080604083850312156102f157600080fd5b82516001600160a01b038116811461030857600080fd5b60208401519092506001600160401b038082111561032557600080fd5b818501915085601f83011261033957600080fd5b81518181111561034b5761034b6102a4565b604051601f8201601f19908116603f01168101908382118183101715610373576103736102a4565b8160405282815288602084870101111561038c57600080fd5b61039d8360208301602088016102ba565b80955050505050509250929050565b600082516103be8184602087016102ba565b9190910192915050565b60208152600082518060208401526103e78160408501602087016102ba565b601f01601f19169190910160400192915050565b60be806104096000396000f3fe608060405236601057600e6013565b005b600e5b601f601b6021565b6065565b565b600060607f360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbc5473ffffffffffffffffffffffffffffffffffffffff1690565b905090565b3660008037600080366000845af43d6000803e8080156083573d6000f35b3d6000fdfea2646970667358221220ffbfbaa210c1b5f5ca62a5eba67b7d993e0bdf919f51500f790fb7acf2fd784c64736f6c63430008140033416464726573733a206c6f772d6c6576656c2064656c65676174652063616c6c206661696c6564",
}

// Erc1967proxyABI is the input ABI used to generate the binding from.
// Deprecated: Use Erc1967proxyMetaData.ABI instead.
var Erc1967proxyABI = Erc1967proxyMetaData.ABI

// Erc1967proxyBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use Erc1967proxyMetaData.Bin instead.
var Erc1967proxyBin = Erc1967proxyMetaData.Bin

// DeployErc1967proxy deploys a new Ethereum contract, binding an instance of Erc1967proxy to it.
func DeployErc1967proxy(auth *bind.TransactOpts, backend bind.ContractBackend, _logic common.Address, _data []byte) (common.Address, *types.Transaction, *Erc1967proxy, error) {
	parsed, err := Erc1967proxyMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(Erc1967proxyBin), backend, _logic, _data)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Erc1967proxy{Erc1967proxyCaller: Erc1967proxyCaller{contract: contract}, Erc1967proxyTransactor: Erc1967proxyTransactor{contract: contract}, Erc1967proxyFilterer: Erc1967proxyFilterer{contract: contract}}, nil
}

// Erc1967proxy is an auto generated Go binding around an Ethereum contract.
type Erc1967proxy struct {
	Erc1967proxyCaller     // Read-only binding to the contract
	Erc1967proxyTransactor // Write-only binding to the contract
	Erc1967proxyFilterer   // Log filterer for contract events
}

// Erc1967proxyCaller is an auto generated read-only Go binding around an Ethereum contract.
type Erc1967proxyCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// Erc1967proxyTransactor is an auto generated write-only Go binding around an Ethereum contract.
type Erc1967proxyTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// Erc1967proxyFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type Erc1967proxyFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// Erc1967proxySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type Erc1967proxySession struct {
	Contract     *Erc1967proxy     // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// Erc1967proxyCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type Erc1967proxyCallerSession struct {
	Contract *Erc1967proxyCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts       // Call options to use throughout this session
}

// Erc1967proxyTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type Erc1967proxyTransactorSession struct {
	Contract     *Erc1967proxyTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts       // Transaction auth options to use throughout this session
}

// Erc1967proxyRaw is an auto generated low-level Go binding around an Ethereum contract.
type Erc1967proxyRaw struct {
	Contract *Erc1967proxy // Generic contract binding to access the raw methods on
}

// Erc1967proxyCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type Erc1967proxyCallerRaw struct {
	Contract *Erc1967proxyCaller // Generic read-only contract binding to access the raw methods on
}

// Erc1967proxyTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type Erc1967proxyTransactorRaw struct {
	Contract *Erc1967proxyTransactor // Generic write-only contract binding to access the raw methods on
}

// NewErc1967proxy creates a new instance of Erc1967proxy, bound to a specific deployed contract.
func NewErc1967proxy(address common.Address, backend bind.ContractBackend) (*Erc1967proxy, error) {
	contract, err := bindErc1967proxy(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Erc1967proxy{Erc1967proxyCaller: Erc1967proxyCaller{contract: contract}, Erc1967proxyTransactor: Erc1967proxyTransactor{contract: contract}, Erc1967proxyFilterer: Erc1967proxyFilterer{contract: contract}}, nil
}

// NewErc1967proxyCaller creates a new read-only instance of Erc1967proxy, bound to a specific deployed contract.
func NewErc1967proxyCaller(address common.Address, caller bind.ContractCaller) (*Erc1967proxyCaller, error) {
	contract, err := bindErc1967proxy(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &Erc1967proxyCaller{contract: contract}, nil
}

// NewErc1967proxyTransactor creates a new write-only instance of Erc1967proxy, bound to a specific deployed contract.
func NewErc1967proxyTransactor(address common.Address, transactor bind.ContractTransactor) (*Erc1967proxyTransactor, error) {
	contract, err := bindErc1967proxy(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &Erc1967proxyTransactor{contract: contract}, nil
}

// NewErc1967proxyFilterer creates a new log filterer instance of Erc1967proxy, bound to a specific deployed contract.
func NewErc1967proxyFilterer(address common.Address, filterer bind.ContractFilterer) (*Erc1967proxyFilterer, error) {
	contract, err := bindErc1967proxy(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &Erc1967proxyFilterer{contract: contract}, nil
}

// bindErc1967proxy binds a generic wrapper to an already deployed contract.
func bindErc1967proxy(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := Erc1967proxyMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Erc1967proxy *Erc1967proxyRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Erc1967proxy.Contract.Erc1967proxyCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Erc1967proxy *Erc1967proxyRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Erc1967proxy.Contract.Erc1967proxyTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Erc1967proxy *Erc1967proxyRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Erc1967proxy.Contract.Erc1967proxyTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Erc1967proxy *Erc1967proxyCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Erc1967proxy.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Erc1967proxy *Erc1967proxyTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Erc1967proxy.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Erc1967proxy *Erc1967proxyTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Erc1967proxy.Contract.contract.Transact(opts, method, params...)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() payable returns()
func (_Erc1967proxy *Erc1967proxyTransactor) Fallback(opts *bind.TransactOpts, calldata []byte) (*types.Transaction, error) {
	return _Erc1967proxy.contract.RawTransact(opts, calldata)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() payable returns()
func (_Erc1967proxy *Erc1967proxySession) Fallback(calldata []byte) (*types.Transaction, error) {
	return _Erc1967proxy.Contract.Fallback(&_Erc1967proxy.TransactOpts, calldata)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() payable returns()
func (_Erc1967proxy *Erc1967proxyTransactorSession) Fallback(calldata []byte) (*types.Transaction, error) {
	return _Erc1967proxy.Contract.Fallback(&_Erc1967proxy.TransactOpts, calldata)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_Erc1967proxy *Erc1967proxyTransactor) Receive(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Erc1967proxy.contract.RawTransact(opts, nil) // calldata is disallowed for receive function
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_Erc1967proxy *Erc1967proxySession) Receive() (*types.Transaction, error) {
	return _Erc1967proxy.Contract.Receive(&_Erc1967proxy.TransactOpts)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_Erc1967proxy *Erc1967proxyTransactorSession) Receive() (*types.Transaction, error) {
	return _Erc1967proxy.Contract.Receive(&_Erc1967proxy.TransactOpts)
}

// Erc1967proxyAdminChangedIterator is returned from FilterAdminChanged and is used to iterate over the raw logs and unpacked data for AdminChanged events raised by the Erc1967proxy contract.
type Erc1967proxyAdminChangedIterator struct {
	Event *Erc1967proxyAdminChanged // Event containing the contract specifics and raw log

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
func (it *Erc1967proxyAdminChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(Erc1967proxyAdminChanged)
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
		it.Event = new(Erc1967proxyAdminChanged)
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
func (it *Erc1967proxyAdminChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *Erc1967proxyAdminChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// Erc1967proxyAdminChanged represents a AdminChanged event raised by the Erc1967proxy contract.
type Erc1967proxyAdminChanged struct {
	PreviousAdmin common.Address
	NewAdmin      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterAdminChanged is a free log retrieval operation binding the contract event 0x7e644d79422f17c01e4894b5f4f588d331ebfa28653d42ae832dc59e38c9798f.
//
// Solidity: event AdminChanged(address previousAdmin, address newAdmin)
func (_Erc1967proxy *Erc1967proxyFilterer) FilterAdminChanged(opts *bind.FilterOpts) (*Erc1967proxyAdminChangedIterator, error) {

	logs, sub, err := _Erc1967proxy.contract.FilterLogs(opts, "AdminChanged")
	if err != nil {
		return nil, err
	}
	return &Erc1967proxyAdminChangedIterator{contract: _Erc1967proxy.contract, event: "AdminChanged", logs: logs, sub: sub}, nil
}

// WatchAdminChanged is a free log subscription operation binding the contract event 0x7e644d79422f17c01e4894b5f4f588d331ebfa28653d42ae832dc59e38c9798f.
//
// Solidity: event AdminChanged(address previousAdmin, address newAdmin)
func (_Erc1967proxy *Erc1967proxyFilterer) WatchAdminChanged(opts *bind.WatchOpts, sink chan<- *Erc1967proxyAdminChanged) (event.Subscription, error) {

	logs, sub, err := _Erc1967proxy.contract.WatchLogs(opts, "AdminChanged")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(Erc1967proxyAdminChanged)
				if err := _Erc1967proxy.contract.UnpackLog(event, "AdminChanged", log); err != nil {
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

// ParseAdminChanged is a log parse operation binding the contract event 0x7e644d79422f17c01e4894b5f4f588d331ebfa28653d42ae832dc59e38c9798f.
//
// Solidity: event AdminChanged(address previousAdmin, address newAdmin)
func (_Erc1967proxy *Erc1967proxyFilterer) ParseAdminChanged(log types.Log) (*Erc1967proxyAdminChanged, error) {
	event := new(Erc1967proxyAdminChanged)
	if err := _Erc1967proxy.contract.UnpackLog(event, "AdminChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// Erc1967proxyBeaconUpgradedIterator is returned from FilterBeaconUpgraded and is used to iterate over the raw logs and unpacked data for BeaconUpgraded events raised by the Erc1967proxy contract.
type Erc1967proxyBeaconUpgradedIterator struct {
	Event *Erc1967proxyBeaconUpgraded // Event containing the contract specifics and raw log

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
func (it *Erc1967proxyBeaconUpgradedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(Erc1967proxyBeaconUpgraded)
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
		it.Event = new(Erc1967proxyBeaconUpgraded)
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
func (it *Erc1967proxyBeaconUpgradedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *Erc1967proxyBeaconUpgradedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// Erc1967proxyBeaconUpgraded represents a BeaconUpgraded event raised by the Erc1967proxy contract.
type Erc1967proxyBeaconUpgraded struct {
	Beacon common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterBeaconUpgraded is a free log retrieval operation binding the contract event 0x1cf3b03a6cf19fa2baba4df148e9dcabedea7f8a5c07840e207e5c089be95d3e.
//
// Solidity: event BeaconUpgraded(address indexed beacon)
func (_Erc1967proxy *Erc1967proxyFilterer) FilterBeaconUpgraded(opts *bind.FilterOpts, beacon []common.Address) (*Erc1967proxyBeaconUpgradedIterator, error) {

	var beaconRule []interface{}
	for _, beaconItem := range beacon {
		beaconRule = append(beaconRule, beaconItem)
	}

	logs, sub, err := _Erc1967proxy.contract.FilterLogs(opts, "BeaconUpgraded", beaconRule)
	if err != nil {
		return nil, err
	}
	return &Erc1967proxyBeaconUpgradedIterator{contract: _Erc1967proxy.contract, event: "BeaconUpgraded", logs: logs, sub: sub}, nil
}

// WatchBeaconUpgraded is a free log subscription operation binding the contract event 0x1cf3b03a6cf19fa2baba4df148e9dcabedea7f8a5c07840e207e5c089be95d3e.
//
// Solidity: event BeaconUpgraded(address indexed beacon)
func (_Erc1967proxy *Erc1967proxyFilterer) WatchBeaconUpgraded(opts *bind.WatchOpts, sink chan<- *Erc1967proxyBeaconUpgraded, beacon []common.Address) (event.Subscription, error) {

	var beaconRule []interface{}
	for _, beaconItem := range beacon {
		beaconRule = append(beaconRule, beaconItem)
	}

	logs, sub, err := _Erc1967proxy.contract.WatchLogs(opts, "BeaconUpgraded", beaconRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(Erc1967proxyBeaconUpgraded)
				if err := _Erc1967proxy.contract.UnpackLog(event, "BeaconUpgraded", log); err != nil {
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

// ParseBeaconUpgraded is a log parse operation binding the contract event 0x1cf3b03a6cf19fa2baba4df148e9dcabedea7f8a5c07840e207e5c089be95d3e.
//
// Solidity: event BeaconUpgraded(address indexed beacon)
func (_Erc1967proxy *Erc1967proxyFilterer) ParseBeaconUpgraded(log types.Log) (*Erc1967proxyBeaconUpgraded, error) {
	event := new(Erc1967proxyBeaconUpgraded)
	if err := _Erc1967proxy.contract.UnpackLog(event, "BeaconUpgraded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// Erc1967proxyUpgradedIterator is returned from FilterUpgraded and is used to iterate over the raw logs and unpacked data for Upgraded events raised by the Erc1967proxy contract.
type Erc1967proxyUpgradedIterator struct {
	Event *Erc1967proxyUpgraded // Event containing the contract specifics and raw log

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
func (it *Erc1967proxyUpgradedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(Erc1967proxyUpgraded)
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
		it.Event = new(Erc1967proxyUpgraded)
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
func (it *Erc1967proxyUpgradedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *Erc1967proxyUpgradedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// Erc1967proxyUpgraded represents a Upgraded event raised by the Erc1967proxy contract.
type Erc1967proxyUpgraded struct {
	Implementation common.Address
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterUpgraded is a free log retrieval operation binding the contract event 0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b.
//
// Solidity: event Upgraded(address indexed implementation)
func (_Erc1967proxy *Erc1967proxyFilterer) FilterUpgraded(opts *bind.FilterOpts, implementation []common.Address) (*Erc1967proxyUpgradedIterator, error) {

	var implementationRule []interface{}
	for _, implementationItem := range implementation {
		implementationRule = append(implementationRule, implementationItem)
	}

	logs, sub, err := _Erc1967proxy.contract.FilterLogs(opts, "Upgraded", implementationRule)
	if err != nil {
		return nil, err
	}
	return &Erc1967proxyUpgradedIterator{contract: _Erc1967proxy.contract, event: "Upgraded", logs: logs, sub: sub}, nil
}

// WatchUpgraded is a free log subscription operation binding the contract event 0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b.
//
// Solidity: event Upgraded(address indexed implementation)
func (_Erc1967proxy *Erc1967proxyFilterer) WatchUpgraded(opts *bind.WatchOpts, sink chan<- *Erc1967proxyUpgraded, implementation []common.Address) (event.Subscription, error) {

	var implementationRule []interface{}
	for _, implementationItem := range implementation {
		implementationRule = append(implementationRule, implementationItem)
	}

	logs, sub, err := _Erc1967proxy.contract.WatchLogs(opts, "Upgraded", implementationRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(Erc1967proxyUpgraded)
				if err := _Erc1967proxy.contract.UnpackLog(event, "Upgraded", log); err != nil {
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

// ParseUpgraded is a log parse operation binding the contract event 0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b.
//
// Solidity: event Upgraded(address indexed implementation)
func (_Erc1967proxy *Erc1967proxyFilterer) ParseUpgraded(log types.Log) (*Erc1967proxyUpgraded, error) {
	event := new(Erc1967proxyUpgraded)
	if err := _Erc1967proxy.contract.UnpackLog(event, "Upgraded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
