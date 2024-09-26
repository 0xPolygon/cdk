// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package claimmocktest

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

// ClaimmocktestMetaData contains all meta data concerning the Claimmocktest contract.
var ClaimmocktestMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"contractIClaimMock\",\"name\":\"_claimMock\",\"type\":\"address\"},{\"internalType\":\"contractIClaimMockCaller\",\"name\":\"_claimMockCaller\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"claim1\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"claim2\",\"type\":\"bytes\"},{\"internalType\":\"bool[2]\",\"name\":\"reverted\",\"type\":\"bool[2]\"}],\"name\":\"claim2TestInternal\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"claim1\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"claim2\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"claim3\",\"type\":\"bytes\"},{\"internalType\":\"bool[3]\",\"name\":\"reverted\",\"type\":\"bool[3]\"}],\"name\":\"claim3TestInternal\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"claimMock\",\"outputs\":[{\"internalType\":\"contractIClaimMock\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"claimMockCaller\",\"outputs\":[{\"internalType\":\"contractIClaimMockCaller\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"claim\",\"type\":\"bytes\"},{\"internalType\":\"bool\",\"name\":\"reverted\",\"type\":\"bool\"}],\"name\":\"claimTestInternal\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x60c0346100a157601f61072e38819003918201601f19168301916001600160401b038311848410176100a65780849260409485528339810103126100a15780516001600160a01b039182821682036100a1576020015191821682036100a15760a05260805260405161067190816100bd82396080518181816102d5015281816103870152818161046a015261054e015260a05181818161031a01526105c20152f35b600080fd5b634e487b7160e01b600052604160045260246000fdfe6080604052600436101561001257600080fd5b6000803560e01c90816348f0c6801461006a575080636e53085414610065578063837a84701461006057806383f5b0061461005b57639bee34681461005657600080fd5b610349565b610304565b6102bf565b610217565b346100f45760c03660031901126100f45767ffffffffffffffff6004358181116100f05761009c903690600401610142565b6024358281116100ec576100b4903690600401610142565b916044359081116100ec576100cd903690600401610142565b36608312156100ec576100e9926100e3366101c6565b92610533565b80f35b8380fd5b8280fd5b80fd5b634e487b7160e01b600052604160045260246000fd5b67ffffffffffffffff811161012157604052565b6100f7565b6040810190811067ffffffffffffffff82111761012157604052565b81601f820112156101a55780359067ffffffffffffffff928383116101215760405193601f8401601f19908116603f011685019081118582101761012157604052828452602083830101116101a557816000926020809301838601378301015290565b600080fd5b6024359081151582036101a557565b359081151582036101a557565b90604051916060830183811067ffffffffffffffff821117610121576040528260c49182116101a5576064905b8282106101ff57505050565b6020809161020c846101b9565b8152019101906101f3565b346101a55760803660031901126101a55767ffffffffffffffff6004358181116101a557610249903690600401610142565b906024359081116101a557610262903690600401610142565b36606312156101a5576040519061027882610126565b819260843681116101a5576044945b81861061029c57505061029a9350610467565b005b602080916102a9886101b9565b815201950194610287565b60009103126101a557565b346101a55760003660031901126101a5576040517f00000000000000000000000000000000000000000000000000000000000000006001600160a01b03168152602090f35b346101a55760003660031901126101a5576040517f00000000000000000000000000000000000000000000000000000000000000006001600160a01b03168152602090f35b346101a557600060403660031901126100f45760043567ffffffffffffffff81116103f75761037c903690600401610142565b816103856101aa565b7f00000000000000000000000000000000000000000000000000000000000000006001600160a01b0316803b156100f0576103d793836040518096819582946327e3584360e01b84526004840161043b565b03925af180156103f2576103e9575080f35b6100e99061010d565b61045b565b5080fd5b919082519283825260005b848110610427575050826000602080949584010152601f8019910116010190565b602081830181015184830182015201610406565b906104536020919493946040845260408401906103fb565b931515910152565b6040513d6000823e3d90fd5b917f00000000000000000000000000000000000000000000000000000000000000006001600160a01b031691823b156101a557604051631cf865cf60e01b815260806004820152938492916104d6916104c49060848601906103fb565b848103600319016024860152906103fb565b90600090604484015b60028310610517575050509181600081819503925af180156103f2576105025750565b8061050f6105159261010d565b806102b4565b565b81511515815286945060019290920191602091820191016104df565b91926000906020810151610632575b80516001600160a01b037f000000000000000000000000000000000000000000000000000000000000000081169490911515853b156101a557600061059d91604051809381926327e3584360e01b9b8c84526004840161043b565b0381838a5af19283156103f25760409560208094610aac9460009761061f575b5001917f0000000000000000000000000000000000000000000000000000000000000000165af1500151151590803b156101a55761060e93600080946040519687958694859384526004840161043b565b03925af180156103f2576105025750565b8061050f61062c9261010d565b386105bd565b6001915061054256fea264697066735822122091357ca0b4807d5579dc633a7d2a9263efbfe31944c644c21b7ccf83594a9e2c64736f6c63430008120033",
}

// ClaimmocktestABI is the input ABI used to generate the binding from.
// Deprecated: Use ClaimmocktestMetaData.ABI instead.
var ClaimmocktestABI = ClaimmocktestMetaData.ABI

// ClaimmocktestBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use ClaimmocktestMetaData.Bin instead.
var ClaimmocktestBin = ClaimmocktestMetaData.Bin

// DeployClaimmocktest deploys a new Ethereum contract, binding an instance of Claimmocktest to it.
func DeployClaimmocktest(auth *bind.TransactOpts, backend bind.ContractBackend, _claimMock common.Address, _claimMockCaller common.Address) (common.Address, *types.Transaction, *Claimmocktest, error) {
	parsed, err := ClaimmocktestMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(ClaimmocktestBin), backend, _claimMock, _claimMockCaller)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Claimmocktest{ClaimmocktestCaller: ClaimmocktestCaller{contract: contract}, ClaimmocktestTransactor: ClaimmocktestTransactor{contract: contract}, ClaimmocktestFilterer: ClaimmocktestFilterer{contract: contract}}, nil
}

// Claimmocktest is an auto generated Go binding around an Ethereum contract.
type Claimmocktest struct {
	ClaimmocktestCaller     // Read-only binding to the contract
	ClaimmocktestTransactor // Write-only binding to the contract
	ClaimmocktestFilterer   // Log filterer for contract events
}

// ClaimmocktestCaller is an auto generated read-only Go binding around an Ethereum contract.
type ClaimmocktestCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ClaimmocktestTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ClaimmocktestTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ClaimmocktestFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ClaimmocktestFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ClaimmocktestSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ClaimmocktestSession struct {
	Contract     *Claimmocktest    // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ClaimmocktestCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ClaimmocktestCallerSession struct {
	Contract *ClaimmocktestCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts        // Call options to use throughout this session
}

// ClaimmocktestTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ClaimmocktestTransactorSession struct {
	Contract     *ClaimmocktestTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// ClaimmocktestRaw is an auto generated low-level Go binding around an Ethereum contract.
type ClaimmocktestRaw struct {
	Contract *Claimmocktest // Generic contract binding to access the raw methods on
}

// ClaimmocktestCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ClaimmocktestCallerRaw struct {
	Contract *ClaimmocktestCaller // Generic read-only contract binding to access the raw methods on
}

// ClaimmocktestTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ClaimmocktestTransactorRaw struct {
	Contract *ClaimmocktestTransactor // Generic write-only contract binding to access the raw methods on
}

// NewClaimmocktest creates a new instance of Claimmocktest, bound to a specific deployed contract.
func NewClaimmocktest(address common.Address, backend bind.ContractBackend) (*Claimmocktest, error) {
	contract, err := bindClaimmocktest(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Claimmocktest{ClaimmocktestCaller: ClaimmocktestCaller{contract: contract}, ClaimmocktestTransactor: ClaimmocktestTransactor{contract: contract}, ClaimmocktestFilterer: ClaimmocktestFilterer{contract: contract}}, nil
}

// NewClaimmocktestCaller creates a new read-only instance of Claimmocktest, bound to a specific deployed contract.
func NewClaimmocktestCaller(address common.Address, caller bind.ContractCaller) (*ClaimmocktestCaller, error) {
	contract, err := bindClaimmocktest(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ClaimmocktestCaller{contract: contract}, nil
}

// NewClaimmocktestTransactor creates a new write-only instance of Claimmocktest, bound to a specific deployed contract.
func NewClaimmocktestTransactor(address common.Address, transactor bind.ContractTransactor) (*ClaimmocktestTransactor, error) {
	contract, err := bindClaimmocktest(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ClaimmocktestTransactor{contract: contract}, nil
}

// NewClaimmocktestFilterer creates a new log filterer instance of Claimmocktest, bound to a specific deployed contract.
func NewClaimmocktestFilterer(address common.Address, filterer bind.ContractFilterer) (*ClaimmocktestFilterer, error) {
	contract, err := bindClaimmocktest(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ClaimmocktestFilterer{contract: contract}, nil
}

// bindClaimmocktest binds a generic wrapper to an already deployed contract.
func bindClaimmocktest(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := ClaimmocktestMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Claimmocktest *ClaimmocktestRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Claimmocktest.Contract.ClaimmocktestCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Claimmocktest *ClaimmocktestRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Claimmocktest.Contract.ClaimmocktestTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Claimmocktest *ClaimmocktestRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Claimmocktest.Contract.ClaimmocktestTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Claimmocktest *ClaimmocktestCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Claimmocktest.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Claimmocktest *ClaimmocktestTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Claimmocktest.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Claimmocktest *ClaimmocktestTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Claimmocktest.Contract.contract.Transact(opts, method, params...)
}

// ClaimMock is a free data retrieval call binding the contract method 0x83f5b006.
//
// Solidity: function claimMock() view returns(address)
func (_Claimmocktest *ClaimmocktestCaller) ClaimMock(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Claimmocktest.contract.Call(opts, &out, "claimMock")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// ClaimMock is a free data retrieval call binding the contract method 0x83f5b006.
//
// Solidity: function claimMock() view returns(address)
func (_Claimmocktest *ClaimmocktestSession) ClaimMock() (common.Address, error) {
	return _Claimmocktest.Contract.ClaimMock(&_Claimmocktest.CallOpts)
}

// ClaimMock is a free data retrieval call binding the contract method 0x83f5b006.
//
// Solidity: function claimMock() view returns(address)
func (_Claimmocktest *ClaimmocktestCallerSession) ClaimMock() (common.Address, error) {
	return _Claimmocktest.Contract.ClaimMock(&_Claimmocktest.CallOpts)
}

// ClaimMockCaller is a free data retrieval call binding the contract method 0x837a8470.
//
// Solidity: function claimMockCaller() view returns(address)
func (_Claimmocktest *ClaimmocktestCaller) ClaimMockCaller(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Claimmocktest.contract.Call(opts, &out, "claimMockCaller")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// ClaimMockCaller is a free data retrieval call binding the contract method 0x837a8470.
//
// Solidity: function claimMockCaller() view returns(address)
func (_Claimmocktest *ClaimmocktestSession) ClaimMockCaller() (common.Address, error) {
	return _Claimmocktest.Contract.ClaimMockCaller(&_Claimmocktest.CallOpts)
}

// ClaimMockCaller is a free data retrieval call binding the contract method 0x837a8470.
//
// Solidity: function claimMockCaller() view returns(address)
func (_Claimmocktest *ClaimmocktestCallerSession) ClaimMockCaller() (common.Address, error) {
	return _Claimmocktest.Contract.ClaimMockCaller(&_Claimmocktest.CallOpts)
}

// Claim2TestInternal is a paid mutator transaction binding the contract method 0x6e530854.
//
// Solidity: function claim2TestInternal(bytes claim1, bytes claim2, bool[2] reverted) returns()
func (_Claimmocktest *ClaimmocktestTransactor) Claim2TestInternal(opts *bind.TransactOpts, claim1 []byte, claim2 []byte, reverted [2]bool) (*types.Transaction, error) {
	return _Claimmocktest.contract.Transact(opts, "claim2TestInternal", claim1, claim2, reverted)
}

// Claim2TestInternal is a paid mutator transaction binding the contract method 0x6e530854.
//
// Solidity: function claim2TestInternal(bytes claim1, bytes claim2, bool[2] reverted) returns()
func (_Claimmocktest *ClaimmocktestSession) Claim2TestInternal(claim1 []byte, claim2 []byte, reverted [2]bool) (*types.Transaction, error) {
	return _Claimmocktest.Contract.Claim2TestInternal(&_Claimmocktest.TransactOpts, claim1, claim2, reverted)
}

// Claim2TestInternal is a paid mutator transaction binding the contract method 0x6e530854.
//
// Solidity: function claim2TestInternal(bytes claim1, bytes claim2, bool[2] reverted) returns()
func (_Claimmocktest *ClaimmocktestTransactorSession) Claim2TestInternal(claim1 []byte, claim2 []byte, reverted [2]bool) (*types.Transaction, error) {
	return _Claimmocktest.Contract.Claim2TestInternal(&_Claimmocktest.TransactOpts, claim1, claim2, reverted)
}

// Claim3TestInternal is a paid mutator transaction binding the contract method 0x48f0c680.
//
// Solidity: function claim3TestInternal(bytes claim1, bytes claim2, bytes claim3, bool[3] reverted) returns()
func (_Claimmocktest *ClaimmocktestTransactor) Claim3TestInternal(opts *bind.TransactOpts, claim1 []byte, claim2 []byte, claim3 []byte, reverted [3]bool) (*types.Transaction, error) {
	return _Claimmocktest.contract.Transact(opts, "claim3TestInternal", claim1, claim2, claim3, reverted)
}

// Claim3TestInternal is a paid mutator transaction binding the contract method 0x48f0c680.
//
// Solidity: function claim3TestInternal(bytes claim1, bytes claim2, bytes claim3, bool[3] reverted) returns()
func (_Claimmocktest *ClaimmocktestSession) Claim3TestInternal(claim1 []byte, claim2 []byte, claim3 []byte, reverted [3]bool) (*types.Transaction, error) {
	return _Claimmocktest.Contract.Claim3TestInternal(&_Claimmocktest.TransactOpts, claim1, claim2, claim3, reverted)
}

// Claim3TestInternal is a paid mutator transaction binding the contract method 0x48f0c680.
//
// Solidity: function claim3TestInternal(bytes claim1, bytes claim2, bytes claim3, bool[3] reverted) returns()
func (_Claimmocktest *ClaimmocktestTransactorSession) Claim3TestInternal(claim1 []byte, claim2 []byte, claim3 []byte, reverted [3]bool) (*types.Transaction, error) {
	return _Claimmocktest.Contract.Claim3TestInternal(&_Claimmocktest.TransactOpts, claim1, claim2, claim3, reverted)
}

// ClaimTestInternal is a paid mutator transaction binding the contract method 0x9bee3468.
//
// Solidity: function claimTestInternal(bytes claim, bool reverted) returns()
func (_Claimmocktest *ClaimmocktestTransactor) ClaimTestInternal(opts *bind.TransactOpts, claim []byte, reverted bool) (*types.Transaction, error) {
	return _Claimmocktest.contract.Transact(opts, "claimTestInternal", claim, reverted)
}

// ClaimTestInternal is a paid mutator transaction binding the contract method 0x9bee3468.
//
// Solidity: function claimTestInternal(bytes claim, bool reverted) returns()
func (_Claimmocktest *ClaimmocktestSession) ClaimTestInternal(claim []byte, reverted bool) (*types.Transaction, error) {
	return _Claimmocktest.Contract.ClaimTestInternal(&_Claimmocktest.TransactOpts, claim, reverted)
}

// ClaimTestInternal is a paid mutator transaction binding the contract method 0x9bee3468.
//
// Solidity: function claimTestInternal(bytes claim, bool reverted) returns()
func (_Claimmocktest *ClaimmocktestTransactorSession) ClaimTestInternal(claim []byte, reverted bool) (*types.Transaction, error) {
	return _Claimmocktest.Contract.ClaimTestInternal(&_Claimmocktest.TransactOpts, claim, reverted)
}
