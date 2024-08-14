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
	Bin: "0x60a060405234801561001057600080fd5b50604051610d2e380380610d2e833981810160405281019061003291906100e1565b8073ffffffffffffffffffffffffffffffffffffffff1660808173ffffffffffffffffffffffffffffffffffffffff16815250505061010e565b600080fd5b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b600061009c82610071565b9050919050565b60006100ae82610091565b9050919050565b6100be816100a3565b81146100c957600080fd5b50565b6000815190506100db816100b5565b92915050565b6000602082840312156100f7576100f661006c565b5b6000610105848285016100cc565b91505092915050565b608051610bfe610130600039600081816104d501526105680152610bfe6000f3fe608060405234801561001057600080fd5b50600436106100575760003560e01c80630680cf5c1461005c578063a2967d991461008c578063d02103ca146100aa578063db3abdb9146100c8578063f4e92675146100e4575b600080fd5b610076600480360381019061007191906106ae565b610102565b60405161008391906106f4565b60405180910390f35b61009461011a565b6040516100a191906106f4565b60405180910390f35b6100b26104d3565b6040516100bf919061078e565b60405180910390f35b6100e260048036038101906100dd919061084d565b6104f7565b005b6100ec610659565b6040516100f991906108d7565b60405180910390f35b60016020528060005260406000206000915090505481565b60008060008054906101000a900463ffffffff1663ffffffff1690506000810361014a576000801b9150506104d0565b60008167ffffffffffffffff811115610166576101656108f2565b5b6040519080825280602002602001820160405280156101945781602001602082028036833780820191505090505b50905060005b8281101561020057600160006001836101b3919061095a565b63ffffffff1663ffffffff168152602001908152602001600020548282815181106101e1576101e061098e565b5b60200260200101818152505080806101f8906109bd565b91505061019a565b5060008060001b90506000602090505b600184146104325760006002856102279190610a34565b6002866102349190610a65565b61023e919061095a565b905060008167ffffffffffffffff81111561025c5761025b6108f2565b5b60405190808252806020026020018201604052801561028a5781602001602082028036833780820191505090505b50905060005b828110156103eb576001836102a59190610a96565b811480156102bf575060016002886102bd9190610a34565b145b1561033757856002826102d29190610aca565b815181106102e3576102e261098e565b5b6020026020010151856040516020016102fd929190610b2d565b604051602081830303815290604052805190602001208282815181106103265761032561098e565b5b6020026020010181815250506103d8565b856002826103459190610aca565b815181106103565761035561098e565b5b602002602001015186600160028461036e9190610aca565b610378919061095a565b815181106103895761038861098e565b5b60200260200101516040516020016103a2929190610b2d565b604051602081830303815290604052805190602001208282815181106103cb576103ca61098e565b5b6020026020010181815250505b80806103e3906109bd565b915050610290565b508094508195508384604051602001610405929190610b2d565b604051602081830303815290604052805190602001209350828061042890610b59565b9350505050610210565b6000836000815181106104485761044761098e565b5b6020026020010151905060005b828110156104c6578184604051602001610470929190610b2d565b604051602081830303815290604052805190602001209150838460405160200161049b929190610b2d565b60405160208183030381529060405280519060200120935080806104be906109bd565b915050610455565b5080955050505050505b90565b7f000000000000000000000000000000000000000000000000000000000000000081565b60008054906101000a900463ffffffff1663ffffffff168563ffffffff16111561053c57846000806101000a81548163ffffffff021916908363ffffffff1602179055505b82600160008763ffffffff1663ffffffff1681526020019081526020016000208190555080156105f9577f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff166333d6247d6105aa61011a565b6040518263ffffffff1660e01b81526004016105c691906106f4565b600060405180830381600087803b1580156105e057600080fd5b505af11580156105f4573d6000803e3d6000fd5b505050505b3373ffffffffffffffffffffffffffffffffffffffff168563ffffffff167faac1e7a157b259544ebacd6e8a82ae5d6c8f174e12aa48696277bcc9a661f0b486858760405161064a93929190610b91565b60405180910390a35050505050565b60008054906101000a900463ffffffff1681565b600080fd5b600063ffffffff82169050919050565b61068b81610672565b811461069657600080fd5b50565b6000813590506106a881610682565b92915050565b6000602082840312156106c4576106c361066d565b5b60006106d284828501610699565b91505092915050565b6000819050919050565b6106ee816106db565b82525050565b600060208201905061070960008301846106e5565b92915050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b6000819050919050565b600061075461074f61074a8461070f565b61072f565b61070f565b9050919050565b600061076682610739565b9050919050565b60006107788261075b565b9050919050565b6107888161076d565b82525050565b60006020820190506107a3600083018461077f565b92915050565b600067ffffffffffffffff82169050919050565b6107c6816107a9565b81146107d157600080fd5b50565b6000813590506107e3816107bd565b92915050565b6107f2816106db565b81146107fd57600080fd5b50565b60008135905061080f816107e9565b92915050565b60008115159050919050565b61082a81610815565b811461083557600080fd5b50565b60008135905061084781610821565b92915050565b600080600080600060a086880312156108695761086861066d565b5b600061087788828901610699565b9550506020610888888289016107d4565b945050604061089988828901610800565b93505060606108aa88828901610800565b92505060806108bb88828901610838565b9150509295509295909350565b6108d181610672565b82525050565b60006020820190506108ec60008301846108c8565b92915050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b6000819050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b600061096582610921565b915061097083610921565b92508282019050808211156109885761098761092b565b5b92915050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052603260045260246000fd5b60006109c882610921565b91507fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff82036109fa576109f961092b565b5b600182019050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601260045260246000fd5b6000610a3f82610921565b9150610a4a83610921565b925082610a5a57610a59610a05565b5b828206905092915050565b6000610a7082610921565b9150610a7b83610921565b925082610a8b57610a8a610a05565b5b828204905092915050565b6000610aa182610921565b9150610aac83610921565b9250828203905081811115610ac457610ac361092b565b5b92915050565b6000610ad582610921565b9150610ae083610921565b9250828202610aee81610921565b91508282048414831517610b0557610b0461092b565b5b5092915050565b6000819050919050565b610b27610b22826106db565b610b0c565b82525050565b6000610b398285610b16565b602082019150610b498284610b16565b6020820191508190509392505050565b6000610b6482610921565b915060008203610b7757610b7661092b565b5b600182039050919050565b610b8b816107a9565b82525050565b6000606082019050610ba66000830186610b82565b610bb360208301856106e5565b610bc060408301846106e5565b94935050505056fea2646970667358221220638bb76fb15c02bf12ba8ed137eb377ccf7928acb4bfb7cf7d3aeb9c24f9163564736f6c63430008120033",
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
