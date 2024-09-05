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
	ABI: "[{\"inputs\":[{\"internalType\":\"contractIClaimMock\",\"name\":\"_claimMock\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"bytes32[32]\",\"name\":\"smtProofLocalExitRoot\",\"type\":\"bytes32[32]\"},{\"internalType\":\"bytes32[32]\",\"name\":\"smtProofRollupExitRoot\",\"type\":\"bytes32[32]\"},{\"internalType\":\"uint256\",\"name\":\"globalIndex\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"mainnetExitRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"rollupExitRoot\",\"type\":\"bytes32\"},{\"internalType\":\"uint32\",\"name\":\"originNetwork\",\"type\":\"uint32\"},{\"internalType\":\"address\",\"name\":\"originTokenAddress\",\"type\":\"address\"},{\"internalType\":\"uint32\",\"name\":\"destinationNetwork\",\"type\":\"uint32\"},{\"internalType\":\"address\",\"name\":\"destinationAddress\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"metadata\",\"type\":\"bytes\"},{\"internalType\":\"bool\",\"name\":\"reverted\",\"type\":\"bool\"}],\"name\":\"claimAsset\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32[32]\",\"name\":\"smtProofLocalExitRoot\",\"type\":\"bytes32[32]\"},{\"internalType\":\"bytes32[32]\",\"name\":\"smtProofRollupExitRoot\",\"type\":\"bytes32[32]\"},{\"internalType\":\"uint256\",\"name\":\"globalIndex\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"mainnetExitRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"rollupExitRoot\",\"type\":\"bytes32\"},{\"internalType\":\"uint32\",\"name\":\"originNetwork\",\"type\":\"uint32\"},{\"internalType\":\"address\",\"name\":\"originTokenAddress\",\"type\":\"address\"},{\"internalType\":\"uint32\",\"name\":\"destinationNetwork\",\"type\":\"uint32\"},{\"internalType\":\"address\",\"name\":\"destinationAddress\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"metadata\",\"type\":\"bytes\"},{\"internalType\":\"bool[2]\",\"name\":\"reverted\",\"type\":\"bool[2]\"}],\"name\":\"claimAsset2\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32[32]\",\"name\":\"smtProofLocalExitRoot\",\"type\":\"bytes32[32]\"},{\"internalType\":\"bytes32[32]\",\"name\":\"smtProofRollupExitRoot\",\"type\":\"bytes32[32]\"},{\"internalType\":\"uint256\",\"name\":\"globalIndex\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"mainnetExitRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"rollupExitRoot\",\"type\":\"bytes32\"},{\"internalType\":\"uint32\",\"name\":\"originNetwork\",\"type\":\"uint32\"},{\"internalType\":\"address\",\"name\":\"originAddress\",\"type\":\"address\"},{\"internalType\":\"uint32\",\"name\":\"destinationNetwork\",\"type\":\"uint32\"},{\"internalType\":\"address\",\"name\":\"destinationAddress\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"metadata\",\"type\":\"bytes\"},{\"internalType\":\"bool\",\"name\":\"reverted\",\"type\":\"bool\"}],\"name\":\"claimMessage\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32[32]\",\"name\":\"smtProofLocalExitRoot\",\"type\":\"bytes32[32]\"},{\"internalType\":\"bytes32[32]\",\"name\":\"smtProofRollupExitRoot\",\"type\":\"bytes32[32]\"},{\"internalType\":\"uint256\",\"name\":\"globalIndex\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"mainnetExitRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"rollupExitRoot\",\"type\":\"bytes32\"},{\"internalType\":\"uint32\",\"name\":\"originNetwork\",\"type\":\"uint32\"},{\"internalType\":\"address\",\"name\":\"originAddress\",\"type\":\"address\"},{\"internalType\":\"uint32\",\"name\":\"destinationNetwork\",\"type\":\"uint32\"},{\"internalType\":\"address\",\"name\":\"destinationAddress\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"metadata\",\"type\":\"bytes\"},{\"internalType\":\"bool[2]\",\"name\":\"reverted\",\"type\":\"bool[2]\"}],\"name\":\"claimMessage2\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"claimMock\",\"outputs\":[{\"internalType\":\"contractIClaimMock\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
	Bin: "0x60a03461009357601f61079738819003918201601f19168301916001600160401b038311848410176100985780849260209460405283398101031261009357516001600160a01b0381168103610093576080526040516106e890816100af823960805181818160dd01528181610375015281816103ec0152818161043e015281816104ec015281816105d101526106480152f35b600080fd5b634e487b7160e01b600052604160045260246000fdfe6080604052600436101561001257600080fd5b6000803560e01c90816301beea651461006a575080637a6760ca1461006557806383f5b00614610060578063a51061701461005b5763a8bad36e1461005657600080fd5b610520565b61046d565b610428565b6102bd565b3461010b57610078366101a1565b9b9291505098919897929796939695949599610103575b9060a060809263f5efcd7960e01b8452013561012452013561050452610884526108a4526108c4526108e45261090452610924526109445261096452806020610aac60808360018060a01b037f0000000000000000000000000000000000000000000000000000000000000000165af15080f35b8a995061008f565b80fd5b610864359063ffffffff8216820361012257565b600080fd5b6108a4359063ffffffff8216820361012257565b61088435906001600160a01b038216820361012257565b6108c435906001600160a01b038216820361012257565b9181601f840112156101225782359167ffffffffffffffff8311610122576020838186019501011161012257565b8015150361012257565b906109406003198301126101225761040490828211610122576004926108049281841161012257923591610824359161084435916101dd61010e565b916101e661013b565b916101ef610127565b916101f8610152565b916108e43591610904359067ffffffffffffffff82116101225761021d918d01610169565b90916109243561022c81610197565b90565b9061096060031983011261012257610404908282116101225760049261080492818411610122579235916108243591610844359161026b61010e565b9161027461013b565b9161027d610127565b91610286610152565b916108e435916109043567ffffffffffffffff811161012257816102ab918e01610169565b92909291610964116101225761092490565b346101225760006020610aac6102d23661022f565b9a939994989597959493928c92509050806102ec8c610684565b610420575b8e6102fe6103059261068e565b9c01610684565b610418575b8d8f91608094938f9160a06040519263ccaa2d1160e01b8452013560a4830152868601356104848301528761080483015288610824830152896108448301528a6108648301528b6108848301528c6108a48301528d6108c48301526108e48201528360018060a01b037f0000000000000000000000000000000000000000000000000000000000000000165af15060405163ccaa2d1160e01b815260a09b909b013560a48c015201356104848a01526108048901526108248801526108448701526108648601526108848501526108a48401526108c48301526108e4820152837f00000000000000000000000000000000000000000000000000000000000000006001600160a01b03165af1005b8e9a5061030a565b8f91506102f1565b34610122576000366003190112610122576040517f00000000000000000000000000000000000000000000000000000000000000006001600160a01b03168152602090f35b346101225760006020610aac610482366101a1565b9a9291505097919796929695939598610518575b60405163ccaa2d1160e01b815260a09a909a013560a48b0152608001356104848a01526108048901526108248801526108448701526108648601526108848501526108a48401526108c48301526108e4820152837f00000000000000000000000000000000000000000000000000000000000000006001600160a01b03165af1005b8c9850610496565b346101225760006020610aac6105353661022f565b9a939994989597959493928c925090508061054f8c610684565b61067c575b8e6102fe6105619261068e565b610674575b8d8f91608094938f9160a06040519263f5efcd7960e01b8452013560a4830152868601356104848301528761080483015288610824830152896108448301528a6108648301528b6108848301528c6108a48301528d6108c48301526108e48201528360018060a01b037f0000000000000000000000000000000000000000000000000000000000000000165af15060405163f5efcd7960e01b815260a09b909b013560a48c015201356104848a01526108048901526108248801526108448701526108648601526108848501526108a48401526108c48301526108e4820152837f00000000000000000000000000000000000000000000000000000000000000006001600160a01b03165af1005b8e9a50610566565b8f9150610554565b3561022c81610197565b906001820180921161069c57565b634e487b7160e01b600052601160045260246000fdfea2646970667358221220c04c7f39b262e6c0fcd8ddd51e2d50fb9d6b33d0b11bd2015734dffd6d60383364736f6c63430008120033",
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

// ClaimAsset2 is a paid mutator transaction binding the contract method 0x7a6760ca.
//
// Solidity: function claimAsset2(bytes32[32] smtProofLocalExitRoot, bytes32[32] smtProofRollupExitRoot, uint256 globalIndex, bytes32 mainnetExitRoot, bytes32 rollupExitRoot, uint32 originNetwork, address originTokenAddress, uint32 destinationNetwork, address destinationAddress, uint256 amount, bytes metadata, bool[2] reverted) returns()
func (_Claimmockcaller *ClaimmockcallerTransactor) ClaimAsset2(opts *bind.TransactOpts, smtProofLocalExitRoot [32][32]byte, smtProofRollupExitRoot [32][32]byte, globalIndex *big.Int, mainnetExitRoot [32]byte, rollupExitRoot [32]byte, originNetwork uint32, originTokenAddress common.Address, destinationNetwork uint32, destinationAddress common.Address, amount *big.Int, metadata []byte, reverted [2]bool) (*types.Transaction, error) {
	return _Claimmockcaller.contract.Transact(opts, "claimAsset2", smtProofLocalExitRoot, smtProofRollupExitRoot, globalIndex, mainnetExitRoot, rollupExitRoot, originNetwork, originTokenAddress, destinationNetwork, destinationAddress, amount, metadata, reverted)
}

// ClaimAsset2 is a paid mutator transaction binding the contract method 0x7a6760ca.
//
// Solidity: function claimAsset2(bytes32[32] smtProofLocalExitRoot, bytes32[32] smtProofRollupExitRoot, uint256 globalIndex, bytes32 mainnetExitRoot, bytes32 rollupExitRoot, uint32 originNetwork, address originTokenAddress, uint32 destinationNetwork, address destinationAddress, uint256 amount, bytes metadata, bool[2] reverted) returns()
func (_Claimmockcaller *ClaimmockcallerSession) ClaimAsset2(smtProofLocalExitRoot [32][32]byte, smtProofRollupExitRoot [32][32]byte, globalIndex *big.Int, mainnetExitRoot [32]byte, rollupExitRoot [32]byte, originNetwork uint32, originTokenAddress common.Address, destinationNetwork uint32, destinationAddress common.Address, amount *big.Int, metadata []byte, reverted [2]bool) (*types.Transaction, error) {
	return _Claimmockcaller.Contract.ClaimAsset2(&_Claimmockcaller.TransactOpts, smtProofLocalExitRoot, smtProofRollupExitRoot, globalIndex, mainnetExitRoot, rollupExitRoot, originNetwork, originTokenAddress, destinationNetwork, destinationAddress, amount, metadata, reverted)
}

// ClaimAsset2 is a paid mutator transaction binding the contract method 0x7a6760ca.
//
// Solidity: function claimAsset2(bytes32[32] smtProofLocalExitRoot, bytes32[32] smtProofRollupExitRoot, uint256 globalIndex, bytes32 mainnetExitRoot, bytes32 rollupExitRoot, uint32 originNetwork, address originTokenAddress, uint32 destinationNetwork, address destinationAddress, uint256 amount, bytes metadata, bool[2] reverted) returns()
func (_Claimmockcaller *ClaimmockcallerTransactorSession) ClaimAsset2(smtProofLocalExitRoot [32][32]byte, smtProofRollupExitRoot [32][32]byte, globalIndex *big.Int, mainnetExitRoot [32]byte, rollupExitRoot [32]byte, originNetwork uint32, originTokenAddress common.Address, destinationNetwork uint32, destinationAddress common.Address, amount *big.Int, metadata []byte, reverted [2]bool) (*types.Transaction, error) {
	return _Claimmockcaller.Contract.ClaimAsset2(&_Claimmockcaller.TransactOpts, smtProofLocalExitRoot, smtProofRollupExitRoot, globalIndex, mainnetExitRoot, rollupExitRoot, originNetwork, originTokenAddress, destinationNetwork, destinationAddress, amount, metadata, reverted)
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

// ClaimMessage2 is a paid mutator transaction binding the contract method 0xa8bad36e.
//
// Solidity: function claimMessage2(bytes32[32] smtProofLocalExitRoot, bytes32[32] smtProofRollupExitRoot, uint256 globalIndex, bytes32 mainnetExitRoot, bytes32 rollupExitRoot, uint32 originNetwork, address originAddress, uint32 destinationNetwork, address destinationAddress, uint256 amount, bytes metadata, bool[2] reverted) returns()
func (_Claimmockcaller *ClaimmockcallerTransactor) ClaimMessage2(opts *bind.TransactOpts, smtProofLocalExitRoot [32][32]byte, smtProofRollupExitRoot [32][32]byte, globalIndex *big.Int, mainnetExitRoot [32]byte, rollupExitRoot [32]byte, originNetwork uint32, originAddress common.Address, destinationNetwork uint32, destinationAddress common.Address, amount *big.Int, metadata []byte, reverted [2]bool) (*types.Transaction, error) {
	return _Claimmockcaller.contract.Transact(opts, "claimMessage2", smtProofLocalExitRoot, smtProofRollupExitRoot, globalIndex, mainnetExitRoot, rollupExitRoot, originNetwork, originAddress, destinationNetwork, destinationAddress, amount, metadata, reverted)
}

// ClaimMessage2 is a paid mutator transaction binding the contract method 0xa8bad36e.
//
// Solidity: function claimMessage2(bytes32[32] smtProofLocalExitRoot, bytes32[32] smtProofRollupExitRoot, uint256 globalIndex, bytes32 mainnetExitRoot, bytes32 rollupExitRoot, uint32 originNetwork, address originAddress, uint32 destinationNetwork, address destinationAddress, uint256 amount, bytes metadata, bool[2] reverted) returns()
func (_Claimmockcaller *ClaimmockcallerSession) ClaimMessage2(smtProofLocalExitRoot [32][32]byte, smtProofRollupExitRoot [32][32]byte, globalIndex *big.Int, mainnetExitRoot [32]byte, rollupExitRoot [32]byte, originNetwork uint32, originAddress common.Address, destinationNetwork uint32, destinationAddress common.Address, amount *big.Int, metadata []byte, reverted [2]bool) (*types.Transaction, error) {
	return _Claimmockcaller.Contract.ClaimMessage2(&_Claimmockcaller.TransactOpts, smtProofLocalExitRoot, smtProofRollupExitRoot, globalIndex, mainnetExitRoot, rollupExitRoot, originNetwork, originAddress, destinationNetwork, destinationAddress, amount, metadata, reverted)
}

// ClaimMessage2 is a paid mutator transaction binding the contract method 0xa8bad36e.
//
// Solidity: function claimMessage2(bytes32[32] smtProofLocalExitRoot, bytes32[32] smtProofRollupExitRoot, uint256 globalIndex, bytes32 mainnetExitRoot, bytes32 rollupExitRoot, uint32 originNetwork, address originAddress, uint32 destinationNetwork, address destinationAddress, uint256 amount, bytes metadata, bool[2] reverted) returns()
func (_Claimmockcaller *ClaimmockcallerTransactorSession) ClaimMessage2(smtProofLocalExitRoot [32][32]byte, smtProofRollupExitRoot [32][32]byte, globalIndex *big.Int, mainnetExitRoot [32]byte, rollupExitRoot [32]byte, originNetwork uint32, originAddress common.Address, destinationNetwork uint32, destinationAddress common.Address, amount *big.Int, metadata []byte, reverted [2]bool) (*types.Transaction, error) {
	return _Claimmockcaller.Contract.ClaimMessage2(&_Claimmockcaller.TransactOpts, smtProofLocalExitRoot, smtProofRollupExitRoot, globalIndex, mainnetExitRoot, rollupExitRoot, originNetwork, originAddress, destinationNetwork, destinationAddress, amount, metadata, reverted)
}
