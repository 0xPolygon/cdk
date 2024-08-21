// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package transparentupgradableproxy

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

// TransparentupgradableproxyMetaData contains all meta data concerning the Transparentupgradableproxy contract.
var TransparentupgradableproxyMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_logic\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"admin_\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"_data\",\"type\":\"bytes\"}],\"stateMutability\":\"payable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"previousAdmin\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"newAdmin\",\"type\":\"address\"}],\"name\":\"AdminChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"beacon\",\"type\":\"address\"}],\"name\":\"BeaconUpgraded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"implementation\",\"type\":\"address\"}],\"name\":\"Upgraded\",\"type\":\"event\"},{\"stateMutability\":\"payable\",\"type\":\"fallback\"},{\"inputs\":[],\"name\":\"admin\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"admin_\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newAdmin\",\"type\":\"address\"}],\"name\":\"changeAdmin\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"implementation\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"implementation_\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newImplementation\",\"type\":\"address\"}],\"name\":\"upgradeTo\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newImplementation\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"upgradeToAndCall\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"stateMutability\":\"payable\",\"type\":\"receive\"}]",
	Bin: "0x608060405260405162000f6838038062000f68833981016040819052620000269162000415565b82816200003582825f6200004c565b50620000439050826200007d565b50505062000540565b6200005783620000ee565b5f82511180620000645750805b1562000078576200007683836200012f565b505b505050565b7f7e644d79422f17c01e4894b5f4f588d331ebfa28653d42ae832dc59e38c9798f620000be5f8051602062000f21833981519152546001600160a01b031690565b604080516001600160a01b03928316815291841660208301520160405180910390a1620000eb816200015e565b50565b620000f981620001fb565b6040516001600160a01b038216907fbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b905f90a250565b606062000157838360405180606001604052806027815260200162000f416027913962000292565b9392505050565b6001600160a01b038116620001c95760405162461bcd60e51b815260206004820152602660248201527f455243313936373a206e65772061646d696e20697320746865207a65726f206160448201526564647265737360d01b60648201526084015b60405180910390fd5b805f8051602062000f218339815191525b80546001600160a01b0319166001600160a01b039290921691909117905550565b6001600160a01b0381163b6200026a5760405162461bcd60e51b815260206004820152602d60248201527f455243313936373a206e657720696d706c656d656e746174696f6e206973206e60448201526c1bdd08184818dbdb9d1c9858dd609a1b6064820152608401620001c0565b807f360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbc620001da565b60605f80856001600160a01b031685604051620002b09190620004ef565b5f60405180830381855af49150503d805f8114620002ea576040519150601f19603f3d011682016040523d82523d5f602084013e620002ef565b606091505b50909250905062000303868383876200030d565b9695505050505050565b60608315620003805782515f0362000378576001600160a01b0385163b620003785760405162461bcd60e51b815260206004820152601d60248201527f416464726573733a2063616c6c20746f206e6f6e2d636f6e74726163740000006044820152606401620001c0565b50816200038c565b6200038c838362000394565b949350505050565b815115620003a55781518083602001fd5b8060405162461bcd60e51b8152600401620001c091906200050c565b80516001600160a01b0381168114620003d8575f80fd5b919050565b634e487b7160e01b5f52604160045260245ffd5b5f5b838110156200040d578181015183820152602001620003f3565b50505f910152565b5f805f6060848603121562000428575f80fd5b6200043384620003c1565b92506200044360208501620003c1565b60408501519092506001600160401b038082111562000460575f80fd5b818601915086601f83011262000474575f80fd5b815181811115620004895762000489620003dd565b604051601f8201601f19908116603f01168101908382118183101715620004b457620004b4620003dd565b81604052828152896020848701011115620004cd575f80fd5b620004e0836020830160208801620003f1565b80955050505050509250925092565b5f825162000502818460208701620003f1565b9190910192915050565b602081525f82518060208401526200052c816040850160208701620003f1565b601f01601f19169190910160400192915050565b6109d3806200054e5f395ff3fe60806040526004361061005d575f3560e01c80635c60da1b116100425780635c60da1b146100a65780638f283970146100e3578063f851a440146101025761006c565b80633659cfe6146100745780634f1ef286146100935761006c565b3661006c5761006a610116565b005b61006a610116565b34801561007f575f80fd5b5061006a61008e366004610854565b610130565b61006a6100a136600461086d565b610178565b3480156100b1575f80fd5b506100ba6101eb565b60405173ffffffffffffffffffffffffffffffffffffffff909116815260200160405180910390f35b3480156100ee575f80fd5b5061006a6100fd366004610854565b610228565b34801561010d575f80fd5b506100ba610255565b61011e610282565b61012e610129610359565b610362565b565b610138610380565b73ffffffffffffffffffffffffffffffffffffffff1633036101705761016d8160405180602001604052805f8152505f6103bf565b50565b61016d610116565b610180610380565b73ffffffffffffffffffffffffffffffffffffffff1633036101e3576101de8383838080601f0160208091040260200160405190810160405280939291908181526020018383808284375f92019190915250600192506103bf915050565b505050565b6101de610116565b5f6101f4610380565b73ffffffffffffffffffffffffffffffffffffffff16330361021d57610218610359565b905090565b610225610116565b90565b610230610380565b73ffffffffffffffffffffffffffffffffffffffff1633036101705761016d816103e9565b5f61025e610380565b73ffffffffffffffffffffffffffffffffffffffff16330361021d57610218610380565b61028a610380565b73ffffffffffffffffffffffffffffffffffffffff16330361012e576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152604260248201527f5472616e73706172656e745570677261646561626c6550726f78793a2061646d60448201527f696e2063616e6e6f742066616c6c6261636b20746f2070726f7879207461726760648201527f6574000000000000000000000000000000000000000000000000000000000000608482015260a4015b60405180910390fd5b5f61021861044a565b365f80375f80365f845af43d5f803e80801561037c573d5ff35b3d5ffd5b5f7fb53127684a568b3173ae13b9f8a6016e243e63b6e8ee1178d6a717850b5d61035b5473ffffffffffffffffffffffffffffffffffffffff16919050565b6103c883610471565b5f825111806103d45750805b156101de576103e383836104bd565b50505050565b7f7e644d79422f17c01e4894b5f4f588d331ebfa28653d42ae832dc59e38c9798f610412610380565b6040805173ffffffffffffffffffffffffffffffffffffffff928316815291841660208301520160405180910390a161016d816104e9565b5f7f360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbc6103a3565b61047a816105f5565b60405173ffffffffffffffffffffffffffffffffffffffff8216907fbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b905f90a250565b60606104e28383604051806060016040528060278152602001610977602791396106c0565b9392505050565b73ffffffffffffffffffffffffffffffffffffffff811661058c576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602660248201527f455243313936373a206e65772061646d696e20697320746865207a65726f206160448201527f64647265737300000000000000000000000000000000000000000000000000006064820152608401610350565b807fb53127684a568b3173ae13b9f8a6016e243e63b6e8ee1178d6a717850b5d61035b80547fffffffffffffffffffffffff00000000000000000000000000000000000000001673ffffffffffffffffffffffffffffffffffffffff9290921691909117905550565b73ffffffffffffffffffffffffffffffffffffffff81163b610699576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602d60248201527f455243313936373a206e657720696d706c656d656e746174696f6e206973206e60448201527f6f74206120636f6e7472616374000000000000000000000000000000000000006064820152608401610350565b807f360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbc6105af565b60605f808573ffffffffffffffffffffffffffffffffffffffff16856040516106e9919061090b565b5f60405180830381855af49150503d805f8114610721576040519150601f19603f3d011682016040523d82523d5f602084013e610726565b606091505b509150915061073786838387610741565b9695505050505050565b606083156107d65782515f036107cf5773ffffffffffffffffffffffffffffffffffffffff85163b6107cf576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601d60248201527f416464726573733a2063616c6c20746f206e6f6e2d636f6e74726163740000006044820152606401610350565b50816107e0565b6107e083836107e8565b949350505050565b8151156107f85781518083602001fd5b806040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016103509190610926565b803573ffffffffffffffffffffffffffffffffffffffff8116811461084f575f80fd5b919050565b5f60208284031215610864575f80fd5b6104e28261082c565b5f805f6040848603121561087f575f80fd5b6108888461082c565b9250602084013567ffffffffffffffff808211156108a4575f80fd5b818601915086601f8301126108b7575f80fd5b8135818111156108c5575f80fd5b8760208285010111156108d6575f80fd5b6020830194508093505050509250925092565b5f5b838110156109035781810151838201526020016108eb565b50505f910152565b5f825161091c8184602087016108e9565b9190910192915050565b602081525f82518060208401526109448160408501602087016108e9565b601f017fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe016919091016040019291505056fe416464726573733a206c6f772d6c6576656c2064656c65676174652063616c6c206661696c6564a26469706673582212202ac98acbfbb3d3ac1b74050e18c4e76db25a3ff2801ec69bf85d0c61414d502b64736f6c63430008140033b53127684a568b3173ae13b9f8a6016e243e63b6e8ee1178d6a717850b5d6103416464726573733a206c6f772d6c6576656c2064656c65676174652063616c6c206661696c6564",
}

// TransparentupgradableproxyABI is the input ABI used to generate the binding from.
// Deprecated: Use TransparentupgradableproxyMetaData.ABI instead.
var TransparentupgradableproxyABI = TransparentupgradableproxyMetaData.ABI

// TransparentupgradableproxyBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use TransparentupgradableproxyMetaData.Bin instead.
var TransparentupgradableproxyBin = TransparentupgradableproxyMetaData.Bin

// DeployTransparentupgradableproxy deploys a new Ethereum contract, binding an instance of Transparentupgradableproxy to it.
func DeployTransparentupgradableproxy(auth *bind.TransactOpts, backend bind.ContractBackend, _logic common.Address, admin_ common.Address, _data []byte) (common.Address, *types.Transaction, *Transparentupgradableproxy, error) {
	parsed, err := TransparentupgradableproxyMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(TransparentupgradableproxyBin), backend, _logic, admin_, _data)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Transparentupgradableproxy{TransparentupgradableproxyCaller: TransparentupgradableproxyCaller{contract: contract}, TransparentupgradableproxyTransactor: TransparentupgradableproxyTransactor{contract: contract}, TransparentupgradableproxyFilterer: TransparentupgradableproxyFilterer{contract: contract}}, nil
}

// Transparentupgradableproxy is an auto generated Go binding around an Ethereum contract.
type Transparentupgradableproxy struct {
	TransparentupgradableproxyCaller     // Read-only binding to the contract
	TransparentupgradableproxyTransactor // Write-only binding to the contract
	TransparentupgradableproxyFilterer   // Log filterer for contract events
}

// TransparentupgradableproxyCaller is an auto generated read-only Go binding around an Ethereum contract.
type TransparentupgradableproxyCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TransparentupgradableproxyTransactor is an auto generated write-only Go binding around an Ethereum contract.
type TransparentupgradableproxyTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TransparentupgradableproxyFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type TransparentupgradableproxyFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TransparentupgradableproxySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type TransparentupgradableproxySession struct {
	Contract     *Transparentupgradableproxy // Generic contract binding to set the session for
	CallOpts     bind.CallOpts               // Call options to use throughout this session
	TransactOpts bind.TransactOpts           // Transaction auth options to use throughout this session
}

// TransparentupgradableproxyCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type TransparentupgradableproxyCallerSession struct {
	Contract *TransparentupgradableproxyCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                     // Call options to use throughout this session
}

// TransparentupgradableproxyTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type TransparentupgradableproxyTransactorSession struct {
	Contract     *TransparentupgradableproxyTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                     // Transaction auth options to use throughout this session
}

// TransparentupgradableproxyRaw is an auto generated low-level Go binding around an Ethereum contract.
type TransparentupgradableproxyRaw struct {
	Contract *Transparentupgradableproxy // Generic contract binding to access the raw methods on
}

// TransparentupgradableproxyCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type TransparentupgradableproxyCallerRaw struct {
	Contract *TransparentupgradableproxyCaller // Generic read-only contract binding to access the raw methods on
}

// TransparentupgradableproxyTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type TransparentupgradableproxyTransactorRaw struct {
	Contract *TransparentupgradableproxyTransactor // Generic write-only contract binding to access the raw methods on
}

// NewTransparentupgradableproxy creates a new instance of Transparentupgradableproxy, bound to a specific deployed contract.
func NewTransparentupgradableproxy(address common.Address, backend bind.ContractBackend) (*Transparentupgradableproxy, error) {
	contract, err := bindTransparentupgradableproxy(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Transparentupgradableproxy{TransparentupgradableproxyCaller: TransparentupgradableproxyCaller{contract: contract}, TransparentupgradableproxyTransactor: TransparentupgradableproxyTransactor{contract: contract}, TransparentupgradableproxyFilterer: TransparentupgradableproxyFilterer{contract: contract}}, nil
}

// NewTransparentupgradableproxyCaller creates a new read-only instance of Transparentupgradableproxy, bound to a specific deployed contract.
func NewTransparentupgradableproxyCaller(address common.Address, caller bind.ContractCaller) (*TransparentupgradableproxyCaller, error) {
	contract, err := bindTransparentupgradableproxy(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &TransparentupgradableproxyCaller{contract: contract}, nil
}

// NewTransparentupgradableproxyTransactor creates a new write-only instance of Transparentupgradableproxy, bound to a specific deployed contract.
func NewTransparentupgradableproxyTransactor(address common.Address, transactor bind.ContractTransactor) (*TransparentupgradableproxyTransactor, error) {
	contract, err := bindTransparentupgradableproxy(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &TransparentupgradableproxyTransactor{contract: contract}, nil
}

// NewTransparentupgradableproxyFilterer creates a new log filterer instance of Transparentupgradableproxy, bound to a specific deployed contract.
func NewTransparentupgradableproxyFilterer(address common.Address, filterer bind.ContractFilterer) (*TransparentupgradableproxyFilterer, error) {
	contract, err := bindTransparentupgradableproxy(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &TransparentupgradableproxyFilterer{contract: contract}, nil
}

// bindTransparentupgradableproxy binds a generic wrapper to an already deployed contract.
func bindTransparentupgradableproxy(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := TransparentupgradableproxyMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Transparentupgradableproxy *TransparentupgradableproxyRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Transparentupgradableproxy.Contract.TransparentupgradableproxyCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Transparentupgradableproxy *TransparentupgradableproxyRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Transparentupgradableproxy.Contract.TransparentupgradableproxyTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Transparentupgradableproxy *TransparentupgradableproxyRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Transparentupgradableproxy.Contract.TransparentupgradableproxyTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Transparentupgradableproxy *TransparentupgradableproxyCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Transparentupgradableproxy.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Transparentupgradableproxy *TransparentupgradableproxyTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Transparentupgradableproxy.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Transparentupgradableproxy *TransparentupgradableproxyTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Transparentupgradableproxy.Contract.contract.Transact(opts, method, params...)
}

// Admin is a paid mutator transaction binding the contract method 0xf851a440.
//
// Solidity: function admin() returns(address admin_)
func (_Transparentupgradableproxy *TransparentupgradableproxyTransactor) Admin(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Transparentupgradableproxy.contract.Transact(opts, "admin")
}

// Admin is a paid mutator transaction binding the contract method 0xf851a440.
//
// Solidity: function admin() returns(address admin_)
func (_Transparentupgradableproxy *TransparentupgradableproxySession) Admin() (*types.Transaction, error) {
	return _Transparentupgradableproxy.Contract.Admin(&_Transparentupgradableproxy.TransactOpts)
}

// Admin is a paid mutator transaction binding the contract method 0xf851a440.
//
// Solidity: function admin() returns(address admin_)
func (_Transparentupgradableproxy *TransparentupgradableproxyTransactorSession) Admin() (*types.Transaction, error) {
	return _Transparentupgradableproxy.Contract.Admin(&_Transparentupgradableproxy.TransactOpts)
}

// ChangeAdmin is a paid mutator transaction binding the contract method 0x8f283970.
//
// Solidity: function changeAdmin(address newAdmin) returns()
func (_Transparentupgradableproxy *TransparentupgradableproxyTransactor) ChangeAdmin(opts *bind.TransactOpts, newAdmin common.Address) (*types.Transaction, error) {
	return _Transparentupgradableproxy.contract.Transact(opts, "changeAdmin", newAdmin)
}

// ChangeAdmin is a paid mutator transaction binding the contract method 0x8f283970.
//
// Solidity: function changeAdmin(address newAdmin) returns()
func (_Transparentupgradableproxy *TransparentupgradableproxySession) ChangeAdmin(newAdmin common.Address) (*types.Transaction, error) {
	return _Transparentupgradableproxy.Contract.ChangeAdmin(&_Transparentupgradableproxy.TransactOpts, newAdmin)
}

// ChangeAdmin is a paid mutator transaction binding the contract method 0x8f283970.
//
// Solidity: function changeAdmin(address newAdmin) returns()
func (_Transparentupgradableproxy *TransparentupgradableproxyTransactorSession) ChangeAdmin(newAdmin common.Address) (*types.Transaction, error) {
	return _Transparentupgradableproxy.Contract.ChangeAdmin(&_Transparentupgradableproxy.TransactOpts, newAdmin)
}

// Implementation is a paid mutator transaction binding the contract method 0x5c60da1b.
//
// Solidity: function implementation() returns(address implementation_)
func (_Transparentupgradableproxy *TransparentupgradableproxyTransactor) Implementation(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Transparentupgradableproxy.contract.Transact(opts, "implementation")
}

// Implementation is a paid mutator transaction binding the contract method 0x5c60da1b.
//
// Solidity: function implementation() returns(address implementation_)
func (_Transparentupgradableproxy *TransparentupgradableproxySession) Implementation() (*types.Transaction, error) {
	return _Transparentupgradableproxy.Contract.Implementation(&_Transparentupgradableproxy.TransactOpts)
}

// Implementation is a paid mutator transaction binding the contract method 0x5c60da1b.
//
// Solidity: function implementation() returns(address implementation_)
func (_Transparentupgradableproxy *TransparentupgradableproxyTransactorSession) Implementation() (*types.Transaction, error) {
	return _Transparentupgradableproxy.Contract.Implementation(&_Transparentupgradableproxy.TransactOpts)
}

// UpgradeTo is a paid mutator transaction binding the contract method 0x3659cfe6.
//
// Solidity: function upgradeTo(address newImplementation) returns()
func (_Transparentupgradableproxy *TransparentupgradableproxyTransactor) UpgradeTo(opts *bind.TransactOpts, newImplementation common.Address) (*types.Transaction, error) {
	return _Transparentupgradableproxy.contract.Transact(opts, "upgradeTo", newImplementation)
}

// UpgradeTo is a paid mutator transaction binding the contract method 0x3659cfe6.
//
// Solidity: function upgradeTo(address newImplementation) returns()
func (_Transparentupgradableproxy *TransparentupgradableproxySession) UpgradeTo(newImplementation common.Address) (*types.Transaction, error) {
	return _Transparentupgradableproxy.Contract.UpgradeTo(&_Transparentupgradableproxy.TransactOpts, newImplementation)
}

// UpgradeTo is a paid mutator transaction binding the contract method 0x3659cfe6.
//
// Solidity: function upgradeTo(address newImplementation) returns()
func (_Transparentupgradableproxy *TransparentupgradableproxyTransactorSession) UpgradeTo(newImplementation common.Address) (*types.Transaction, error) {
	return _Transparentupgradableproxy.Contract.UpgradeTo(&_Transparentupgradableproxy.TransactOpts, newImplementation)
}

// UpgradeToAndCall is a paid mutator transaction binding the contract method 0x4f1ef286.
//
// Solidity: function upgradeToAndCall(address newImplementation, bytes data) payable returns()
func (_Transparentupgradableproxy *TransparentupgradableproxyTransactor) UpgradeToAndCall(opts *bind.TransactOpts, newImplementation common.Address, data []byte) (*types.Transaction, error) {
	return _Transparentupgradableproxy.contract.Transact(opts, "upgradeToAndCall", newImplementation, data)
}

// UpgradeToAndCall is a paid mutator transaction binding the contract method 0x4f1ef286.
//
// Solidity: function upgradeToAndCall(address newImplementation, bytes data) payable returns()
func (_Transparentupgradableproxy *TransparentupgradableproxySession) UpgradeToAndCall(newImplementation common.Address, data []byte) (*types.Transaction, error) {
	return _Transparentupgradableproxy.Contract.UpgradeToAndCall(&_Transparentupgradableproxy.TransactOpts, newImplementation, data)
}

// UpgradeToAndCall is a paid mutator transaction binding the contract method 0x4f1ef286.
//
// Solidity: function upgradeToAndCall(address newImplementation, bytes data) payable returns()
func (_Transparentupgradableproxy *TransparentupgradableproxyTransactorSession) UpgradeToAndCall(newImplementation common.Address, data []byte) (*types.Transaction, error) {
	return _Transparentupgradableproxy.Contract.UpgradeToAndCall(&_Transparentupgradableproxy.TransactOpts, newImplementation, data)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() payable returns()
func (_Transparentupgradableproxy *TransparentupgradableproxyTransactor) Fallback(opts *bind.TransactOpts, calldata []byte) (*types.Transaction, error) {
	return _Transparentupgradableproxy.contract.RawTransact(opts, calldata)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() payable returns()
func (_Transparentupgradableproxy *TransparentupgradableproxySession) Fallback(calldata []byte) (*types.Transaction, error) {
	return _Transparentupgradableproxy.Contract.Fallback(&_Transparentupgradableproxy.TransactOpts, calldata)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() payable returns()
func (_Transparentupgradableproxy *TransparentupgradableproxyTransactorSession) Fallback(calldata []byte) (*types.Transaction, error) {
	return _Transparentupgradableproxy.Contract.Fallback(&_Transparentupgradableproxy.TransactOpts, calldata)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_Transparentupgradableproxy *TransparentupgradableproxyTransactor) Receive(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Transparentupgradableproxy.contract.RawTransact(opts, nil) // calldata is disallowed for receive function
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_Transparentupgradableproxy *TransparentupgradableproxySession) Receive() (*types.Transaction, error) {
	return _Transparentupgradableproxy.Contract.Receive(&_Transparentupgradableproxy.TransactOpts)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_Transparentupgradableproxy *TransparentupgradableproxyTransactorSession) Receive() (*types.Transaction, error) {
	return _Transparentupgradableproxy.Contract.Receive(&_Transparentupgradableproxy.TransactOpts)
}

// TransparentupgradableproxyAdminChangedIterator is returned from FilterAdminChanged and is used to iterate over the raw logs and unpacked data for AdminChanged events raised by the Transparentupgradableproxy contract.
type TransparentupgradableproxyAdminChangedIterator struct {
	Event *TransparentupgradableproxyAdminChanged // Event containing the contract specifics and raw log

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
func (it *TransparentupgradableproxyAdminChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TransparentupgradableproxyAdminChanged)
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
		it.Event = new(TransparentupgradableproxyAdminChanged)
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
func (it *TransparentupgradableproxyAdminChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TransparentupgradableproxyAdminChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TransparentupgradableproxyAdminChanged represents a AdminChanged event raised by the Transparentupgradableproxy contract.
type TransparentupgradableproxyAdminChanged struct {
	PreviousAdmin common.Address
	NewAdmin      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterAdminChanged is a free log retrieval operation binding the contract event 0x7e644d79422f17c01e4894b5f4f588d331ebfa28653d42ae832dc59e38c9798f.
//
// Solidity: event AdminChanged(address previousAdmin, address newAdmin)
func (_Transparentupgradableproxy *TransparentupgradableproxyFilterer) FilterAdminChanged(opts *bind.FilterOpts) (*TransparentupgradableproxyAdminChangedIterator, error) {

	logs, sub, err := _Transparentupgradableproxy.contract.FilterLogs(opts, "AdminChanged")
	if err != nil {
		return nil, err
	}
	return &TransparentupgradableproxyAdminChangedIterator{contract: _Transparentupgradableproxy.contract, event: "AdminChanged", logs: logs, sub: sub}, nil
}

// WatchAdminChanged is a free log subscription operation binding the contract event 0x7e644d79422f17c01e4894b5f4f588d331ebfa28653d42ae832dc59e38c9798f.
//
// Solidity: event AdminChanged(address previousAdmin, address newAdmin)
func (_Transparentupgradableproxy *TransparentupgradableproxyFilterer) WatchAdminChanged(opts *bind.WatchOpts, sink chan<- *TransparentupgradableproxyAdminChanged) (event.Subscription, error) {

	logs, sub, err := _Transparentupgradableproxy.contract.WatchLogs(opts, "AdminChanged")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TransparentupgradableproxyAdminChanged)
				if err := _Transparentupgradableproxy.contract.UnpackLog(event, "AdminChanged", log); err != nil {
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
func (_Transparentupgradableproxy *TransparentupgradableproxyFilterer) ParseAdminChanged(log types.Log) (*TransparentupgradableproxyAdminChanged, error) {
	event := new(TransparentupgradableproxyAdminChanged)
	if err := _Transparentupgradableproxy.contract.UnpackLog(event, "AdminChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// TransparentupgradableproxyBeaconUpgradedIterator is returned from FilterBeaconUpgraded and is used to iterate over the raw logs and unpacked data for BeaconUpgraded events raised by the Transparentupgradableproxy contract.
type TransparentupgradableproxyBeaconUpgradedIterator struct {
	Event *TransparentupgradableproxyBeaconUpgraded // Event containing the contract specifics and raw log

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
func (it *TransparentupgradableproxyBeaconUpgradedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TransparentupgradableproxyBeaconUpgraded)
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
		it.Event = new(TransparentupgradableproxyBeaconUpgraded)
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
func (it *TransparentupgradableproxyBeaconUpgradedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TransparentupgradableproxyBeaconUpgradedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TransparentupgradableproxyBeaconUpgraded represents a BeaconUpgraded event raised by the Transparentupgradableproxy contract.
type TransparentupgradableproxyBeaconUpgraded struct {
	Beacon common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterBeaconUpgraded is a free log retrieval operation binding the contract event 0x1cf3b03a6cf19fa2baba4df148e9dcabedea7f8a5c07840e207e5c089be95d3e.
//
// Solidity: event BeaconUpgraded(address indexed beacon)
func (_Transparentupgradableproxy *TransparentupgradableproxyFilterer) FilterBeaconUpgraded(opts *bind.FilterOpts, beacon []common.Address) (*TransparentupgradableproxyBeaconUpgradedIterator, error) {

	var beaconRule []interface{}
	for _, beaconItem := range beacon {
		beaconRule = append(beaconRule, beaconItem)
	}

	logs, sub, err := _Transparentupgradableproxy.contract.FilterLogs(opts, "BeaconUpgraded", beaconRule)
	if err != nil {
		return nil, err
	}
	return &TransparentupgradableproxyBeaconUpgradedIterator{contract: _Transparentupgradableproxy.contract, event: "BeaconUpgraded", logs: logs, sub: sub}, nil
}

// WatchBeaconUpgraded is a free log subscription operation binding the contract event 0x1cf3b03a6cf19fa2baba4df148e9dcabedea7f8a5c07840e207e5c089be95d3e.
//
// Solidity: event BeaconUpgraded(address indexed beacon)
func (_Transparentupgradableproxy *TransparentupgradableproxyFilterer) WatchBeaconUpgraded(opts *bind.WatchOpts, sink chan<- *TransparentupgradableproxyBeaconUpgraded, beacon []common.Address) (event.Subscription, error) {

	var beaconRule []interface{}
	for _, beaconItem := range beacon {
		beaconRule = append(beaconRule, beaconItem)
	}

	logs, sub, err := _Transparentupgradableproxy.contract.WatchLogs(opts, "BeaconUpgraded", beaconRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TransparentupgradableproxyBeaconUpgraded)
				if err := _Transparentupgradableproxy.contract.UnpackLog(event, "BeaconUpgraded", log); err != nil {
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
func (_Transparentupgradableproxy *TransparentupgradableproxyFilterer) ParseBeaconUpgraded(log types.Log) (*TransparentupgradableproxyBeaconUpgraded, error) {
	event := new(TransparentupgradableproxyBeaconUpgraded)
	if err := _Transparentupgradableproxy.contract.UnpackLog(event, "BeaconUpgraded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// TransparentupgradableproxyUpgradedIterator is returned from FilterUpgraded and is used to iterate over the raw logs and unpacked data for Upgraded events raised by the Transparentupgradableproxy contract.
type TransparentupgradableproxyUpgradedIterator struct {
	Event *TransparentupgradableproxyUpgraded // Event containing the contract specifics and raw log

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
func (it *TransparentupgradableproxyUpgradedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TransparentupgradableproxyUpgraded)
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
		it.Event = new(TransparentupgradableproxyUpgraded)
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
func (it *TransparentupgradableproxyUpgradedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TransparentupgradableproxyUpgradedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TransparentupgradableproxyUpgraded represents a Upgraded event raised by the Transparentupgradableproxy contract.
type TransparentupgradableproxyUpgraded struct {
	Implementation common.Address
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterUpgraded is a free log retrieval operation binding the contract event 0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b.
//
// Solidity: event Upgraded(address indexed implementation)
func (_Transparentupgradableproxy *TransparentupgradableproxyFilterer) FilterUpgraded(opts *bind.FilterOpts, implementation []common.Address) (*TransparentupgradableproxyUpgradedIterator, error) {

	var implementationRule []interface{}
	for _, implementationItem := range implementation {
		implementationRule = append(implementationRule, implementationItem)
	}

	logs, sub, err := _Transparentupgradableproxy.contract.FilterLogs(opts, "Upgraded", implementationRule)
	if err != nil {
		return nil, err
	}
	return &TransparentupgradableproxyUpgradedIterator{contract: _Transparentupgradableproxy.contract, event: "Upgraded", logs: logs, sub: sub}, nil
}

// WatchUpgraded is a free log subscription operation binding the contract event 0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b.
//
// Solidity: event Upgraded(address indexed implementation)
func (_Transparentupgradableproxy *TransparentupgradableproxyFilterer) WatchUpgraded(opts *bind.WatchOpts, sink chan<- *TransparentupgradableproxyUpgraded, implementation []common.Address) (event.Subscription, error) {

	var implementationRule []interface{}
	for _, implementationItem := range implementation {
		implementationRule = append(implementationRule, implementationItem)
	}

	logs, sub, err := _Transparentupgradableproxy.contract.WatchLogs(opts, "Upgraded", implementationRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TransparentupgradableproxyUpgraded)
				if err := _Transparentupgradableproxy.contract.UnpackLog(event, "Upgraded", log); err != nil {
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
func (_Transparentupgradableproxy *TransparentupgradableproxyFilterer) ParseUpgraded(log types.Log) (*TransparentupgradableproxyUpgraded, error) {
	event := new(TransparentupgradableproxyUpgraded)
	if err := _Transparentupgradableproxy.contract.UnpackLog(event, "Upgraded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
