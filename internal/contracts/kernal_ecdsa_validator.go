// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contracts

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
)

// PackedUserOperation is an auto generated low-level Go binding around an user-defined struct.
type PackedUserOperation struct {
	Sender             common.Address
	Nonce              *big.Int
	InitCode           []byte
	CallData           []byte
	AccountGasLimits   [32]byte
	PreVerificationGas *big.Int
	GasFees            [32]byte
	PaymasterAndData   []byte
	Signature          []byte
}

// ECDSAValidatorMetaData contains all meta data concerning the ECDSAValidator contract.
var ECDSAValidatorMetaData = &bind.MetaData{
	ABI: "[{\"constant\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"smartAccount\",\"type\":\"address\"}],\"name\":\"AlreadyInitialized\",\"payable\":false,\"type\":\"error\"},{\"constant\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"target\",\"type\":\"address\"}],\"name\":\"InvalidTargetAddress\",\"payable\":false,\"type\":\"error\"},{\"constant\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"smartAccount\",\"type\":\"address\"}],\"name\":\"NotInitialized\",\"payable\":false,\"type\":\"error\"},{\"constant\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"kernel\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"OwnerRegistered\",\"payable\":false,\"type\":\"event\"},{\"constant\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"ecdsaValidatorStorage\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"smartAccount\",\"type\":\"address\"}],\"name\":\"isInitialized\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"typeID\",\"type\":\"uint256\"}],\"name\":\"isModuleType\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"pure\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"hash\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"sig\",\"type\":\"bytes\"}],\"name\":\"isValidSignatureWithSender\",\"outputs\":[{\"internalType\":\"bytes4\",\"name\":\"\",\"type\":\"bytes4\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"_data\",\"type\":\"bytes\"}],\"name\":\"onInstall\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"name\":\"onUninstall\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"hookData\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"success\",\"type\":\"bool\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"res\",\"type\":\"bytes\"}],\"name\":\"postCheck\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"msgSender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"name\":\"preCheck\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"payable\":false,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"components\":[{\"name\":\"sender\",\"type\":\"address\"},{\"name\":\"nonce\",\"type\":\"uint256\"},{\"name\":\"initCode\",\"type\":\"bytes\"},{\"name\":\"callData\",\"type\":\"bytes\"},{\"name\":\"accountGasLimits\",\"type\":\"bytes32\"},{\"name\":\"preVerificationGas\",\"type\":\"uint256\"},{\"name\":\"gasFees\",\"type\":\"bytes32\"},{\"name\":\"paymasterAndData\",\"type\":\"bytes\"},{\"name\":\"signature\",\"type\":\"bytes\"}],\"indexed\":false,\"internalType\":\"structPackedUserOperation\",\"name\":\"userOp\",\"type\":\"tuple\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"userOpHash\",\"type\":\"bytes32\"}],\"name\":\"validateUserOp\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"payable\",\"type\":\"function\"}]",
}

// ECDSAValidatorABI is the input ABI used to generate the binding from.
// Deprecated: Use ECDSAValidatorMetaData.ABI instead.
var ECDSAValidatorABI = ECDSAValidatorMetaData.ABI

// ECDSAValidator is an auto generated Go binding around an Ethereum contract.
type ECDSAValidator struct {
	ECDSAValidatorCaller     // Read-only binding to the contract
	ECDSAValidatorTransactor // Write-only binding to the contract
	ECDSAValidatorFilterer   // Log filterer for contract events
}

// ECDSAValidatorCaller is an auto generated read-only Go binding around an Ethereum contract.
type ECDSAValidatorCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ECDSAValidatorTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ECDSAValidatorTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ECDSAValidatorFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ECDSAValidatorFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ECDSAValidatorSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ECDSAValidatorSession struct {
	Contract     *ECDSAValidator   // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ECDSAValidatorCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ECDSAValidatorCallerSession struct {
	Contract *ECDSAValidatorCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts         // Call options to use throughout this session
}

// ECDSAValidatorTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ECDSAValidatorTransactorSession struct {
	Contract     *ECDSAValidatorTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// ECDSAValidatorRaw is an auto generated low-level Go binding around an Ethereum contract.
type ECDSAValidatorRaw struct {
	Contract *ECDSAValidator // Generic contract binding to access the raw methods on
}

// ECDSAValidatorCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ECDSAValidatorCallerRaw struct {
	Contract *ECDSAValidatorCaller // Generic read-only contract binding to access the raw methods on
}

// ECDSAValidatorTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ECDSAValidatorTransactorRaw struct {
	Contract *ECDSAValidatorTransactor // Generic write-only contract binding to access the raw methods on
}

// NewECDSAValidator creates a new instance of ECDSAValidator, bound to a specific deployed contract.
func NewECDSAValidator(address common.Address, backend bind.ContractBackend) (*ECDSAValidator, error) {
	contract, err := bindECDSAValidator(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ECDSAValidator{ECDSAValidatorCaller: ECDSAValidatorCaller{contract: contract}, ECDSAValidatorTransactor: ECDSAValidatorTransactor{contract: contract}, ECDSAValidatorFilterer: ECDSAValidatorFilterer{contract: contract}}, nil
}

// NewECDSAValidatorCaller creates a new read-only instance of ECDSAValidator, bound to a specific deployed contract.
func NewECDSAValidatorCaller(address common.Address, caller bind.ContractCaller) (*ECDSAValidatorCaller, error) {
	contract, err := bindECDSAValidator(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ECDSAValidatorCaller{contract: contract}, nil
}

// NewECDSAValidatorTransactor creates a new write-only instance of ECDSAValidator, bound to a specific deployed contract.
func NewECDSAValidatorTransactor(address common.Address, transactor bind.ContractTransactor) (*ECDSAValidatorTransactor, error) {
	contract, err := bindECDSAValidator(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ECDSAValidatorTransactor{contract: contract}, nil
}

// NewECDSAValidatorFilterer creates a new log filterer instance of ECDSAValidator, bound to a specific deployed contract.
func NewECDSAValidatorFilterer(address common.Address, filterer bind.ContractFilterer) (*ECDSAValidatorFilterer, error) {
	contract, err := bindECDSAValidator(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ECDSAValidatorFilterer{contract: contract}, nil
}

// bindECDSAValidator binds a generic wrapper to an already deployed contract.
func bindECDSAValidator(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ECDSAValidatorABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ECDSAValidator *ECDSAValidatorRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ECDSAValidator.Contract.ECDSAValidatorCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ECDSAValidator *ECDSAValidatorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ECDSAValidator.Contract.ECDSAValidatorTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ECDSAValidator *ECDSAValidatorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ECDSAValidator.Contract.ECDSAValidatorTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ECDSAValidator *ECDSAValidatorCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ECDSAValidator.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ECDSAValidator *ECDSAValidatorTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ECDSAValidator.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ECDSAValidator *ECDSAValidatorTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ECDSAValidator.Contract.contract.Transact(opts, method, params...)
}

// EcdsaValidatorStorage is a free data retrieval call binding the contract method 0x20709efc.
//
// Solidity: function ecdsaValidatorStorage(address ) view returns(address owner)
func (_ECDSAValidator *ECDSAValidatorCaller) EcdsaValidatorStorage(opts *bind.CallOpts, arg0 common.Address) (common.Address, error) {
	var out []interface{}
	err := _ECDSAValidator.contract.Call(opts, &out, "ecdsaValidatorStorage", arg0)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// EcdsaValidatorStorage is a free data retrieval call binding the contract method 0x20709efc.
//
// Solidity: function ecdsaValidatorStorage(address ) view returns(address owner)
func (_ECDSAValidator *ECDSAValidatorSession) EcdsaValidatorStorage(arg0 common.Address) (common.Address, error) {
	return _ECDSAValidator.Contract.EcdsaValidatorStorage(&_ECDSAValidator.CallOpts, arg0)
}

// EcdsaValidatorStorage is a free data retrieval call binding the contract method 0x20709efc.
//
// Solidity: function ecdsaValidatorStorage(address ) view returns(address owner)
func (_ECDSAValidator *ECDSAValidatorCallerSession) EcdsaValidatorStorage(arg0 common.Address) (common.Address, error) {
	return _ECDSAValidator.Contract.EcdsaValidatorStorage(&_ECDSAValidator.CallOpts, arg0)
}

// IsInitialized is a free data retrieval call binding the contract method 0xd60b347f.
//
// Solidity: function isInitialized(address smartAccount) view returns(bool)
func (_ECDSAValidator *ECDSAValidatorCaller) IsInitialized(opts *bind.CallOpts, smartAccount common.Address) (bool, error) {
	var out []interface{}
	err := _ECDSAValidator.contract.Call(opts, &out, "isInitialized", smartAccount)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsInitialized is a free data retrieval call binding the contract method 0xd60b347f.
//
// Solidity: function isInitialized(address smartAccount) view returns(bool)
func (_ECDSAValidator *ECDSAValidatorSession) IsInitialized(smartAccount common.Address) (bool, error) {
	return _ECDSAValidator.Contract.IsInitialized(&_ECDSAValidator.CallOpts, smartAccount)
}

// IsInitialized is a free data retrieval call binding the contract method 0xd60b347f.
//
// Solidity: function isInitialized(address smartAccount) view returns(bool)
func (_ECDSAValidator *ECDSAValidatorCallerSession) IsInitialized(smartAccount common.Address) (bool, error) {
	return _ECDSAValidator.Contract.IsInitialized(&_ECDSAValidator.CallOpts, smartAccount)
}

// IsModuleType is a free data retrieval call binding the contract method 0xecd05961.
//
// Solidity: function isModuleType(uint256 typeID) pure returns(bool)
func (_ECDSAValidator *ECDSAValidatorCaller) IsModuleType(opts *bind.CallOpts, typeID *big.Int) (bool, error) {
	var out []interface{}
	err := _ECDSAValidator.contract.Call(opts, &out, "isModuleType", typeID)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsModuleType is a free data retrieval call binding the contract method 0xecd05961.
//
// Solidity: function isModuleType(uint256 typeID) pure returns(bool)
func (_ECDSAValidator *ECDSAValidatorSession) IsModuleType(typeID *big.Int) (bool, error) {
	return _ECDSAValidator.Contract.IsModuleType(&_ECDSAValidator.CallOpts, typeID)
}

// IsModuleType is a free data retrieval call binding the contract method 0xecd05961.
//
// Solidity: function isModuleType(uint256 typeID) pure returns(bool)
func (_ECDSAValidator *ECDSAValidatorCallerSession) IsModuleType(typeID *big.Int) (bool, error) {
	return _ECDSAValidator.Contract.IsModuleType(&_ECDSAValidator.CallOpts, typeID)
}

// IsValidSignatureWithSender is a free data retrieval call binding the contract method 0xf551e2ee.
//
// Solidity: function isValidSignatureWithSender(address , bytes32 hash, bytes sig) view returns(bytes4)
func (_ECDSAValidator *ECDSAValidatorCaller) IsValidSignatureWithSender(opts *bind.CallOpts, arg0 common.Address, hash [32]byte, sig []byte) ([4]byte, error) {
	var out []interface{}
	err := _ECDSAValidator.contract.Call(opts, &out, "isValidSignatureWithSender", arg0, hash, sig)

	if err != nil {
		return *new([4]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([4]byte)).(*[4]byte)

	return out0, err

}

// IsValidSignatureWithSender is a free data retrieval call binding the contract method 0xf551e2ee.
//
// Solidity: function isValidSignatureWithSender(address , bytes32 hash, bytes sig) view returns(bytes4)
func (_ECDSAValidator *ECDSAValidatorSession) IsValidSignatureWithSender(arg0 common.Address, hash [32]byte, sig []byte) ([4]byte, error) {
	return _ECDSAValidator.Contract.IsValidSignatureWithSender(&_ECDSAValidator.CallOpts, arg0, hash, sig)
}

// IsValidSignatureWithSender is a free data retrieval call binding the contract method 0xf551e2ee.
//
// Solidity: function isValidSignatureWithSender(address , bytes32 hash, bytes sig) view returns(bytes4)
func (_ECDSAValidator *ECDSAValidatorCallerSession) IsValidSignatureWithSender(arg0 common.Address, hash [32]byte, sig []byte) ([4]byte, error) {
	return _ECDSAValidator.Contract.IsValidSignatureWithSender(&_ECDSAValidator.CallOpts, arg0, hash, sig)
}

// OnInstall is a paid mutator transaction binding the contract method 0x6d61fe70.
//
// Solidity: function onInstall(bytes _data) payable returns()
func (_ECDSAValidator *ECDSAValidatorTransactor) OnInstall(opts *bind.TransactOpts, _data []byte) (*types.Transaction, error) {
	return _ECDSAValidator.contract.Transact(opts, "onInstall", _data)
}

// OnInstall is a paid mutator transaction binding the contract method 0x6d61fe70.
//
// Solidity: function onInstall(bytes _data) payable returns()
func (_ECDSAValidator *ECDSAValidatorSession) OnInstall(_data []byte) (*types.Transaction, error) {
	return _ECDSAValidator.Contract.OnInstall(&_ECDSAValidator.TransactOpts, _data)
}

// OnInstall is a paid mutator transaction binding the contract method 0x6d61fe70.
//
// Solidity: function onInstall(bytes _data) payable returns()
func (_ECDSAValidator *ECDSAValidatorTransactorSession) OnInstall(_data []byte) (*types.Transaction, error) {
	return _ECDSAValidator.Contract.OnInstall(&_ECDSAValidator.TransactOpts, _data)
}

// OnUninstall is a paid mutator transaction binding the contract method 0x8a91b0e3.
//
// Solidity: function onUninstall(bytes ) payable returns()
func (_ECDSAValidator *ECDSAValidatorTransactor) OnUninstall(opts *bind.TransactOpts, arg0 []byte) (*types.Transaction, error) {
	return _ECDSAValidator.contract.Transact(opts, "onUninstall", arg0)
}

// OnUninstall is a paid mutator transaction binding the contract method 0x8a91b0e3.
//
// Solidity: function onUninstall(bytes ) payable returns()
func (_ECDSAValidator *ECDSAValidatorSession) OnUninstall(arg0 []byte) (*types.Transaction, error) {
	return _ECDSAValidator.Contract.OnUninstall(&_ECDSAValidator.TransactOpts, arg0)
}

// OnUninstall is a paid mutator transaction binding the contract method 0x8a91b0e3.
//
// Solidity: function onUninstall(bytes ) payable returns()
func (_ECDSAValidator *ECDSAValidatorTransactorSession) OnUninstall(arg0 []byte) (*types.Transaction, error) {
	return _ECDSAValidator.Contract.OnUninstall(&_ECDSAValidator.TransactOpts, arg0)
}

// PostCheck is a paid mutator transaction binding the contract method 0xaacbd72a.
//
// Solidity: function postCheck(bytes hookData, bool success, bytes res) payable returns()
func (_ECDSAValidator *ECDSAValidatorTransactor) PostCheck(opts *bind.TransactOpts, hookData []byte, success bool, res []byte) (*types.Transaction, error) {
	return _ECDSAValidator.contract.Transact(opts, "postCheck", hookData, success, res)
}

// PostCheck is a paid mutator transaction binding the contract method 0xaacbd72a.
//
// Solidity: function postCheck(bytes hookData, bool success, bytes res) payable returns()
func (_ECDSAValidator *ECDSAValidatorSession) PostCheck(hookData []byte, success bool, res []byte) (*types.Transaction, error) {
	return _ECDSAValidator.Contract.PostCheck(&_ECDSAValidator.TransactOpts, hookData, success, res)
}

// PostCheck is a paid mutator transaction binding the contract method 0xaacbd72a.
//
// Solidity: function postCheck(bytes hookData, bool success, bytes res) payable returns()
func (_ECDSAValidator *ECDSAValidatorTransactorSession) PostCheck(hookData []byte, success bool, res []byte) (*types.Transaction, error) {
	return _ECDSAValidator.Contract.PostCheck(&_ECDSAValidator.TransactOpts, hookData, success, res)
}

// PreCheck is a paid mutator transaction binding the contract method 0xd68f6025.
//
// Solidity: function preCheck(address msgSender, uint256 value, bytes ) payable returns(bytes)
func (_ECDSAValidator *ECDSAValidatorTransactor) PreCheck(opts *bind.TransactOpts, msgSender common.Address, value *big.Int, arg2 []byte) (*types.Transaction, error) {
	return _ECDSAValidator.contract.Transact(opts, "preCheck", msgSender, value, arg2)
}

// PreCheck is a paid mutator transaction binding the contract method 0xd68f6025.
//
// Solidity: function preCheck(address msgSender, uint256 value, bytes ) payable returns(bytes)
func (_ECDSAValidator *ECDSAValidatorSession) PreCheck(msgSender common.Address, value *big.Int, arg2 []byte) (*types.Transaction, error) {
	return _ECDSAValidator.Contract.PreCheck(&_ECDSAValidator.TransactOpts, msgSender, value, arg2)
}

// PreCheck is a paid mutator transaction binding the contract method 0xd68f6025.
//
// Solidity: function preCheck(address msgSender, uint256 value, bytes ) payable returns(bytes)
func (_ECDSAValidator *ECDSAValidatorTransactorSession) PreCheck(msgSender common.Address, value *big.Int, arg2 []byte) (*types.Transaction, error) {
	return _ECDSAValidator.Contract.PreCheck(&_ECDSAValidator.TransactOpts, msgSender, value, arg2)
}

// ValidateUserOp is a paid mutator transaction binding the contract method 0x97003203.
//
// Solidity: function validateUserOp((address,uint256,bytes,bytes,bytes32,uint256,bytes32,bytes,bytes) userOp, bytes32 userOpHash) payable returns(uint256)
func (_ECDSAValidator *ECDSAValidatorTransactor) ValidateUserOp(opts *bind.TransactOpts, userOp PackedUserOperation, userOpHash [32]byte) (*types.Transaction, error) {
	return _ECDSAValidator.contract.Transact(opts, "validateUserOp", userOp, userOpHash)
}

// ValidateUserOp is a paid mutator transaction binding the contract method 0x97003203.
//
// Solidity: function validateUserOp((address,uint256,bytes,bytes,bytes32,uint256,bytes32,bytes,bytes) userOp, bytes32 userOpHash) payable returns(uint256)
func (_ECDSAValidator *ECDSAValidatorSession) ValidateUserOp(userOp PackedUserOperation, userOpHash [32]byte) (*types.Transaction, error) {
	return _ECDSAValidator.Contract.ValidateUserOp(&_ECDSAValidator.TransactOpts, userOp, userOpHash)
}

// ValidateUserOp is a paid mutator transaction binding the contract method 0x97003203.
//
// Solidity: function validateUserOp((address,uint256,bytes,bytes,bytes32,uint256,bytes32,bytes,bytes) userOp, bytes32 userOpHash) payable returns(uint256)
func (_ECDSAValidator *ECDSAValidatorTransactorSession) ValidateUserOp(userOp PackedUserOperation, userOpHash [32]byte) (*types.Transaction, error) {
	return _ECDSAValidator.Contract.ValidateUserOp(&_ECDSAValidator.TransactOpts, userOp, userOpHash)
}

// ECDSAValidatorOwnerRegisteredIterator is returned from FilterOwnerRegistered and is used to iterate over the raw logs and unpacked data for OwnerRegistered events raised by the ECDSAValidator contract.
type ECDSAValidatorOwnerRegisteredIterator struct {
	Event *ECDSAValidatorOwnerRegistered // Event containing the contract specifics and raw log

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
func (it *ECDSAValidatorOwnerRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ECDSAValidatorOwnerRegistered)
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
		it.Event = new(ECDSAValidatorOwnerRegistered)
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
func (it *ECDSAValidatorOwnerRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ECDSAValidatorOwnerRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ECDSAValidatorOwnerRegistered represents a OwnerRegistered event raised by the ECDSAValidator contract.
type ECDSAValidatorOwnerRegistered struct {
	Kernel common.Address
	Owner  common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterOwnerRegistered is a free log retrieval operation binding the contract event 0xa5e1f8b4009110f5525798d04ae2125421a12d0590aa52c13682ff1bd3c492ca.
//
// Solidity: event OwnerRegistered(address indexed kernel, address indexed owner)
func (_ECDSAValidator *ECDSAValidatorFilterer) FilterOwnerRegistered(opts *bind.FilterOpts, kernel []common.Address, owner []common.Address) (*ECDSAValidatorOwnerRegisteredIterator, error) {

	var kernelRule []interface{}
	for _, kernelItem := range kernel {
		kernelRule = append(kernelRule, kernelItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _ECDSAValidator.contract.FilterLogs(opts, "OwnerRegistered", kernelRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return &ECDSAValidatorOwnerRegisteredIterator{contract: _ECDSAValidator.contract, event: "OwnerRegistered", logs: logs, sub: sub}, nil
}

// WatchOwnerRegistered is a free log subscription operation binding the contract event 0xa5e1f8b4009110f5525798d04ae2125421a12d0590aa52c13682ff1bd3c492ca.
//
// Solidity: event OwnerRegistered(address indexed kernel, address indexed owner)
func (_ECDSAValidator *ECDSAValidatorFilterer) WatchOwnerRegistered(opts *bind.WatchOpts, sink chan<- *ECDSAValidatorOwnerRegistered, kernel []common.Address, owner []common.Address) (event.Subscription, error) {

	var kernelRule []interface{}
	for _, kernelItem := range kernel {
		kernelRule = append(kernelRule, kernelItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _ECDSAValidator.contract.WatchLogs(opts, "OwnerRegistered", kernelRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ECDSAValidatorOwnerRegistered)
				if err := _ECDSAValidator.contract.UnpackLog(event, "OwnerRegistered", log); err != nil {
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

// ParseOwnerRegistered is a log parse operation binding the contract event 0xa5e1f8b4009110f5525798d04ae2125421a12d0590aa52c13682ff1bd3c492ca.
//
// Solidity: event OwnerRegistered(address indexed kernel, address indexed owner)
func (_ECDSAValidator *ECDSAValidatorFilterer) ParseOwnerRegistered(log types.Log) (*ECDSAValidatorOwnerRegistered, error) {
	event := new(ECDSAValidatorOwnerRegistered)
	if err := _ECDSAValidator.contract.UnpackLog(event, "OwnerRegistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
