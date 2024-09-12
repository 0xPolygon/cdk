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
	ABI: "[{\"inputs\":[{\"internalType\":\"contractIClaimMock\",\"name\":\"_claimMock\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"claim1\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"claim2\",\"type\":\"bytes\"},{\"internalType\":\"bool[2]\",\"name\":\"reverted\",\"type\":\"bool[2]\"}],\"name\":\"claim2Bytes\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32[32]\",\"name\":\"smtProofLocalExitRoot\",\"type\":\"bytes32[32]\"},{\"internalType\":\"bytes32[32]\",\"name\":\"smtProofRollupExitRoot\",\"type\":\"bytes32[32]\"},{\"internalType\":\"uint256\",\"name\":\"globalIndex\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"mainnetExitRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"rollupExitRoot\",\"type\":\"bytes32\"},{\"internalType\":\"uint32\",\"name\":\"originNetwork\",\"type\":\"uint32\"},{\"internalType\":\"address\",\"name\":\"originTokenAddress\",\"type\":\"address\"},{\"internalType\":\"uint32\",\"name\":\"destinationNetwork\",\"type\":\"uint32\"},{\"internalType\":\"address\",\"name\":\"destinationAddress\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"metadata\",\"type\":\"bytes\"},{\"internalType\":\"bool\",\"name\":\"reverted\",\"type\":\"bool\"}],\"name\":\"claimAsset\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"claim\",\"type\":\"bytes\"},{\"internalType\":\"bool\",\"name\":\"reverted\",\"type\":\"bool\"}],\"name\":\"claimBytes\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32[32]\",\"name\":\"smtProofLocalExitRoot\",\"type\":\"bytes32[32]\"},{\"internalType\":\"bytes32[32]\",\"name\":\"smtProofRollupExitRoot\",\"type\":\"bytes32[32]\"},{\"internalType\":\"uint256\",\"name\":\"globalIndex\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"mainnetExitRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"rollupExitRoot\",\"type\":\"bytes32\"},{\"internalType\":\"uint32\",\"name\":\"originNetwork\",\"type\":\"uint32\"},{\"internalType\":\"address\",\"name\":\"originAddress\",\"type\":\"address\"},{\"internalType\":\"uint32\",\"name\":\"destinationNetwork\",\"type\":\"uint32\"},{\"internalType\":\"address\",\"name\":\"destinationAddress\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"metadata\",\"type\":\"bytes\"},{\"internalType\":\"bool\",\"name\":\"reverted\",\"type\":\"bool\"}],\"name\":\"claimMessage\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"claimMock\",\"outputs\":[{\"internalType\":\"contractIClaimMock\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
	Bin: "0x60a03461008557601f61063738819003918201601f19168301916001600160401b0383118484101761008a5780849260209460405283398101031261008557516001600160a01b03811681036100855760805260405161059690816100a1823960805181818160d4015281816103bc01528181610407015281816104b701526104f80152f35b600080fd5b634e487b7160e01b600052604160045260246000fdfe6080604052600436101561001257600080fd5b6000803560e01c90816301beea651461006a575080631cf865cf1461006557806327e358431461006057806383f5b0061461005b5763a51061701461005657600080fd5b610436565b6103f1565b61036a565b6102d0565b3461010b57806020610aac608083610081366101a0565b929b939a949995989697969594939291506101029050575b63f5efcd7960e01b8c5260a00135610124528a013561050452610884526108a4526108c4526108e452610904526109245261094452610964527f00000000000000000000000000000000000000000000000000000000000000006001600160a01b03165af15080f35b60019a50610099565b80fd5b6108a4359063ffffffff8216820361012257565b600080fd5b61088435906001600160a01b038216820361012257565b6108c435906001600160a01b038216820361012257565b9181601f840112156101225782359167ffffffffffffffff8311610122576020838186019501011161012257565b6109243590811515820361012257565b3590811515820361012257565b906109406003198301126101225761040490828211610122576004926108049281841161012257923591610824359161084435916108643563ffffffff8116810361012257916101ee610127565b916101f761010e565b9161020061013e565b916108e43591610904359067ffffffffffffffff821161012257610225918d01610155565b909161022f610183565b90565b634e487b7160e01b600052604160045260246000fd5b604051906040820182811067ffffffffffffffff82111761026857604052565b610232565b81601f820112156101225780359067ffffffffffffffff928383116102685760405193601f8401601f19908116603f0116850190811185821017610268576040528284526020838301011161012257816000926020809301838601378301015290565b346101225760803660031901126101225767ffffffffffffffff6004358181116101225761030290369060040161026d565b906024359081116101225761031b90369060040161026d565b9036606312156101225761032d610248565b9182916084368111610122576044945b81861061035257505061035093506104ec565b005b6020809161035f88610193565b81520195019461033d565b346101225760403660031901126101225760043567ffffffffffffffff81116101225761039b90369060040161026d565b602435801515810361012257610aac60209160009384916103e8575b8301907f00000000000000000000000000000000000000000000000000000000000000006001600160a01b03165af1005b600191506103b7565b34610122576000366003190112610122576040517f00000000000000000000000000000000000000000000000000000000000000006001600160a01b03168152602090f35b346101225760006020610aac61044b366101a0565b9a9150508d989198979297969396959495996104e3575b60405163ccaa2d1160e01b815260a09b909b013560a48c0152608001356104848b01526108048a01526108248901526108448801526108648701526108848601526108a48501526108c48401526108e48301527f00000000000000000000000000000000000000000000000000000000000000006001600160a01b03165af1005b60019950610462565b825160009360209384937f00000000000000000000000000000000000000000000000000000000000000006001600160a01b0316938793929190610557575b858893015161054e575b858891819495610aac9889920190885af15001915af150565b60019250610535565b6001935061052b56fea2646970667358221220bbde05c8a8245c4319ff8aa0ce8d95e6c5dd5c5828fe085ba1491ea451b390ba64736f6c63430008120033",
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

// Claim2Bytes is a paid mutator transaction binding the contract method 0x1cf865cf.
//
// Solidity: function claim2Bytes(bytes claim1, bytes claim2, bool[2] reverted) returns()
func (_Claimmockcaller *ClaimmockcallerTransactor) Claim2Bytes(opts *bind.TransactOpts, claim1 []byte, claim2 []byte, reverted [2]bool) (*types.Transaction, error) {
	return _Claimmockcaller.contract.Transact(opts, "claim2Bytes", claim1, claim2, reverted)
}

// Claim2Bytes is a paid mutator transaction binding the contract method 0x1cf865cf.
//
// Solidity: function claim2Bytes(bytes claim1, bytes claim2, bool[2] reverted) returns()
func (_Claimmockcaller *ClaimmockcallerSession) Claim2Bytes(claim1 []byte, claim2 []byte, reverted [2]bool) (*types.Transaction, error) {
	return _Claimmockcaller.Contract.Claim2Bytes(&_Claimmockcaller.TransactOpts, claim1, claim2, reverted)
}

// Claim2Bytes is a paid mutator transaction binding the contract method 0x1cf865cf.
//
// Solidity: function claim2Bytes(bytes claim1, bytes claim2, bool[2] reverted) returns()
func (_Claimmockcaller *ClaimmockcallerTransactorSession) Claim2Bytes(claim1 []byte, claim2 []byte, reverted [2]bool) (*types.Transaction, error) {
	return _Claimmockcaller.Contract.Claim2Bytes(&_Claimmockcaller.TransactOpts, claim1, claim2, reverted)
}

// ClaimAsset is a paid mutator transaction binding the contract method 0xa5106170.
//
// Solidity: function claimAsset(bytes32[32] smtProofLocalExitRoot, bytes32[32] smtProofRollupExitRoot, uint256 globalIndex, bytes32 mainnetExitRoot, bytes32 rollupExitRoot, uint32 originNetwork, address originTokenAddress, uint32 destinationNetwork, address destinationAddress, uint256 amount, bytes metadata, bool reverted) returns()
func (_Claimmockcaller *ClaimmockcallerTransactor) ClaimAsset(opts *bind.TransactOpts, smtProofLocalExitRoot [32][32]byte, smtProofRollupExitRoot [32][32]byte, globalIndex *big.Int, mainnetExitRoot [32]byte, rollupExitRoot [32]byte, originNetwork uint32, originTokenAddress common.Address, destinationNetwork uint32, destinationAddress common.Address, amount *big.Int, metadata []byte, reverted bool) (*types.Transaction, error) {
	return _Claimmockcaller.contract.Transact(opts, "claimAsset", smtProofLocalExitRoot, smtProofRollupExitRoot, globalIndex, mainnetExitRoot, rollupExitRoot, originNetwork, originTokenAddress, destinationNetwork, destinationAddress, amount, metadata, reverted)
}

// ClaimAsset is a paid mutator transaction binding the contract method 0xa5106170.
//
// Solidity: function claimAsset(bytes32[32] smtProofLocalExitRoot, bytes32[32] smtProofRollupExitRoot, uint256 globalIndex, bytes32 mainnetExitRoot, bytes32 rollupExitRoot, uint32 originNetwork, address originTokenAddress, uint32 destinationNetwork, address destinationAddress, uint256 amount, bytes metadata, bool reverted) returns()
func (_Claimmockcaller *ClaimmockcallerSession) ClaimAsset(smtProofLocalExitRoot [32][32]byte, smtProofRollupExitRoot [32][32]byte, globalIndex *big.Int, mainnetExitRoot [32]byte, rollupExitRoot [32]byte, originNetwork uint32, originTokenAddress common.Address, destinationNetwork uint32, destinationAddress common.Address, amount *big.Int, metadata []byte, reverted bool) (*types.Transaction, error) {
	return _Claimmockcaller.Contract.ClaimAsset(&_Claimmockcaller.TransactOpts, smtProofLocalExitRoot, smtProofRollupExitRoot, globalIndex, mainnetExitRoot, rollupExitRoot, originNetwork, originTokenAddress, destinationNetwork, destinationAddress, amount, metadata, reverted)
}

// ClaimAsset is a paid mutator transaction binding the contract method 0xa5106170.
//
// Solidity: function claimAsset(bytes32[32] smtProofLocalExitRoot, bytes32[32] smtProofRollupExitRoot, uint256 globalIndex, bytes32 mainnetExitRoot, bytes32 rollupExitRoot, uint32 originNetwork, address originTokenAddress, uint32 destinationNetwork, address destinationAddress, uint256 amount, bytes metadata, bool reverted) returns()
func (_Claimmockcaller *ClaimmockcallerTransactorSession) ClaimAsset(smtProofLocalExitRoot [32][32]byte, smtProofRollupExitRoot [32][32]byte, globalIndex *big.Int, mainnetExitRoot [32]byte, rollupExitRoot [32]byte, originNetwork uint32, originTokenAddress common.Address, destinationNetwork uint32, destinationAddress common.Address, amount *big.Int, metadata []byte, reverted bool) (*types.Transaction, error) {
	return _Claimmockcaller.Contract.ClaimAsset(&_Claimmockcaller.TransactOpts, smtProofLocalExitRoot, smtProofRollupExitRoot, globalIndex, mainnetExitRoot, rollupExitRoot, originNetwork, originTokenAddress, destinationNetwork, destinationAddress, amount, metadata, reverted)
}

// ClaimBytes is a paid mutator transaction binding the contract method 0x27e35843.
//
// Solidity: function claimBytes(bytes claim, bool reverted) returns()
func (_Claimmockcaller *ClaimmockcallerTransactor) ClaimBytes(opts *bind.TransactOpts, claim []byte, reverted bool) (*types.Transaction, error) {
	return _Claimmockcaller.contract.Transact(opts, "claimBytes", claim, reverted)
}

// ClaimBytes is a paid mutator transaction binding the contract method 0x27e35843.
//
// Solidity: function claimBytes(bytes claim, bool reverted) returns()
func (_Claimmockcaller *ClaimmockcallerSession) ClaimBytes(claim []byte, reverted bool) (*types.Transaction, error) {
	return _Claimmockcaller.Contract.ClaimBytes(&_Claimmockcaller.TransactOpts, claim, reverted)
}

// ClaimBytes is a paid mutator transaction binding the contract method 0x27e35843.
//
// Solidity: function claimBytes(bytes claim, bool reverted) returns()
func (_Claimmockcaller *ClaimmockcallerTransactorSession) ClaimBytes(claim []byte, reverted bool) (*types.Transaction, error) {
	return _Claimmockcaller.Contract.ClaimBytes(&_Claimmockcaller.TransactOpts, claim, reverted)
}

// ClaimMessage is a paid mutator transaction binding the contract method 0x01beea65.
//
// Solidity: function claimMessage(bytes32[32] smtProofLocalExitRoot, bytes32[32] smtProofRollupExitRoot, uint256 globalIndex, bytes32 mainnetExitRoot, bytes32 rollupExitRoot, uint32 originNetwork, address originAddress, uint32 destinationNetwork, address destinationAddress, uint256 amount, bytes metadata, bool reverted) returns()
func (_Claimmockcaller *ClaimmockcallerTransactor) ClaimMessage(opts *bind.TransactOpts, smtProofLocalExitRoot [32][32]byte, smtProofRollupExitRoot [32][32]byte, globalIndex *big.Int, mainnetExitRoot [32]byte, rollupExitRoot [32]byte, originNetwork uint32, originAddress common.Address, destinationNetwork uint32, destinationAddress common.Address, amount *big.Int, metadata []byte, reverted bool) (*types.Transaction, error) {
	return _Claimmockcaller.contract.Transact(opts, "claimMessage", smtProofLocalExitRoot, smtProofRollupExitRoot, globalIndex, mainnetExitRoot, rollupExitRoot, originNetwork, originAddress, destinationNetwork, destinationAddress, amount, metadata, reverted)
}

// ClaimMessage is a paid mutator transaction binding the contract method 0x01beea65.
//
// Solidity: function claimMessage(bytes32[32] smtProofLocalExitRoot, bytes32[32] smtProofRollupExitRoot, uint256 globalIndex, bytes32 mainnetExitRoot, bytes32 rollupExitRoot, uint32 originNetwork, address originAddress, uint32 destinationNetwork, address destinationAddress, uint256 amount, bytes metadata, bool reverted) returns()
func (_Claimmockcaller *ClaimmockcallerSession) ClaimMessage(smtProofLocalExitRoot [32][32]byte, smtProofRollupExitRoot [32][32]byte, globalIndex *big.Int, mainnetExitRoot [32]byte, rollupExitRoot [32]byte, originNetwork uint32, originAddress common.Address, destinationNetwork uint32, destinationAddress common.Address, amount *big.Int, metadata []byte, reverted bool) (*types.Transaction, error) {
	return _Claimmockcaller.Contract.ClaimMessage(&_Claimmockcaller.TransactOpts, smtProofLocalExitRoot, smtProofRollupExitRoot, globalIndex, mainnetExitRoot, rollupExitRoot, originNetwork, originAddress, destinationNetwork, destinationAddress, amount, metadata, reverted)
}

// ClaimMessage is a paid mutator transaction binding the contract method 0x01beea65.
//
// Solidity: function claimMessage(bytes32[32] smtProofLocalExitRoot, bytes32[32] smtProofRollupExitRoot, uint256 globalIndex, bytes32 mainnetExitRoot, bytes32 rollupExitRoot, uint32 originNetwork, address originAddress, uint32 destinationNetwork, address destinationAddress, uint256 amount, bytes metadata, bool reverted) returns()
func (_Claimmockcaller *ClaimmockcallerTransactorSession) ClaimMessage(smtProofLocalExitRoot [32][32]byte, smtProofRollupExitRoot [32][32]byte, globalIndex *big.Int, mainnetExitRoot [32]byte, rollupExitRoot [32]byte, originNetwork uint32, originAddress common.Address, destinationNetwork uint32, destinationAddress common.Address, amount *big.Int, metadata []byte, reverted bool) (*types.Transaction, error) {
	return _Claimmockcaller.Contract.ClaimMessage(&_Claimmockcaller.TransactOpts, smtProofLocalExitRoot, smtProofRollupExitRoot, globalIndex, mainnetExitRoot, rollupExitRoot, originNetwork, originAddress, destinationNetwork, destinationAddress, amount, metadata, reverted)
}
