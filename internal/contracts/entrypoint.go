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

// EntrypointMetaData contains all meta data concerning the Entrypoint contract.
var EntrypointMetaData = &bind.MetaData{
	ABI: "[{\"constant\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"smartAccount\",\"type\":\"address\"}],\"name\":\"AlreadyInitialized\",\"payable\":false,\"type\":\"error\"},{\"constant\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"target\",\"type\":\"address\"}],\"name\":\"InvalidTargetAddress\",\"payable\":false,\"type\":\"error\"},{\"constant\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"smartAccount\",\"type\":\"address\"}],\"name\":\"NotInitialized\",\"payable\":false,\"type\":\"error\"},{\"constant\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"kernel\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"OwnerRegistered\",\"payable\":false,\"type\":\"event\"},{\"constant\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"ecdsaValidatorStorage\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"smartAccount\",\"type\":\"address\"}],\"name\":\"isInitialized\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"typeID\",\"type\":\"uint256\"}],\"name\":\"isModuleType\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"pure\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"hash\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"sig\",\"type\":\"bytes\"}],\"name\":\"isValidSignatureWithSender\",\"outputs\":[{\"internalType\":\"bytes4\",\"name\":\"\",\"type\":\"bytes4\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"_data\",\"type\":\"bytes\"}],\"name\":\"onInstall\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"name\":\"onUninstall\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"hookData\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"success\",\"type\":\"bool\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"res\",\"type\":\"bytes\"}],\"name\":\"postCheck\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"msgSender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"name\":\"preCheck\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"payable\":false,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"components\":[{\"name\":\"sender\",\"type\":\"address\"},{\"name\":\"nonce\",\"type\":\"uint256\"},{\"name\":\"initCode\",\"type\":\"bytes\"},{\"name\":\"callData\",\"type\":\"bytes\"},{\"name\":\"accountGasLimits\",\"type\":\"bytes32\"},{\"name\":\"preVerificationGas\",\"type\":\"uint256\"},{\"name\":\"gasFees\",\"type\":\"bytes32\"},{\"name\":\"paymasterAndData\",\"type\":\"bytes\"},{\"name\":\"signature\",\"type\":\"bytes\"}],\"indexed\":false,\"internalType\":\"structPackedUserOperation\",\"name\":\"userOp\",\"type\":\"tuple\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"userOpHash\",\"type\":\"bytes32\"}],\"name\":\"validateUserOp\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"payable\",\"type\":\"function\"}]",
}

// EntrypointABI is the input ABI used to generate the binding from.
// Deprecated: Use EntrypointMetaData.ABI instead.
var EntrypointABI = EntrypointMetaData.ABI

// Entrypoint is an auto generated Go binding around an Ethereum contract.
type Entrypoint struct {
	EntrypointCaller     // Read-only binding to the contract
	EntrypointTransactor // Write-only binding to the contract
	EntrypointFilterer   // Log filterer for contract events
}

// EntrypointCaller is an auto generated read-only Go binding around an Ethereum contract.
type EntrypointCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EntrypointTransactor is an auto generated write-only Go binding around an Ethereum contract.
type EntrypointTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EntrypointFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type EntrypointFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EntrypointSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type EntrypointSession struct {
	Contract     *Entrypoint       // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// EntrypointCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type EntrypointCallerSession struct {
	Contract *EntrypointCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts     // Call options to use throughout this session
}

// EntrypointTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type EntrypointTransactorSession struct {
	Contract     *EntrypointTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// EntrypointRaw is an auto generated low-level Go binding around an Ethereum contract.
type EntrypointRaw struct {
	Contract *Entrypoint // Generic contract binding to access the raw methods on
}

// EntrypointCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type EntrypointCallerRaw struct {
	Contract *EntrypointCaller // Generic read-only contract binding to access the raw methods on
}

// EntrypointTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type EntrypointTransactorRaw struct {
	Contract *EntrypointTransactor // Generic write-only contract binding to access the raw methods on
}

// NewEntrypoint creates a new instance of Entrypoint, bound to a specific deployed contract.
func NewEntrypoint(address common.Address, backend bind.ContractBackend) (*Entrypoint, error) {
	contract, err := bindEntrypoint(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Entrypoint{EntrypointCaller: EntrypointCaller{contract: contract}, EntrypointTransactor: EntrypointTransactor{contract: contract}, EntrypointFilterer: EntrypointFilterer{contract: contract}}, nil
}

// NewEntrypointCaller creates a new read-only instance of Entrypoint, bound to a specific deployed contract.
func NewEntrypointCaller(address common.Address, caller bind.ContractCaller) (*EntrypointCaller, error) {
	contract, err := bindEntrypoint(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &EntrypointCaller{contract: contract}, nil
}

// NewEntrypointTransactor creates a new write-only instance of Entrypoint, bound to a specific deployed contract.
func NewEntrypointTransactor(address common.Address, transactor bind.ContractTransactor) (*EntrypointTransactor, error) {
	contract, err := bindEntrypoint(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &EntrypointTransactor{contract: contract}, nil
}

// NewEntrypointFilterer creates a new log filterer instance of Entrypoint, bound to a specific deployed contract.
func NewEntrypointFilterer(address common.Address, filterer bind.ContractFilterer) (*EntrypointFilterer, error) {
	contract, err := bindEntrypoint(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &EntrypointFilterer{contract: contract}, nil
}

// bindEntrypoint binds a generic wrapper to an already deployed contract.
func bindEntrypoint(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(EntrypointABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Entrypoint *EntrypointRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Entrypoint.Contract.EntrypointCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Entrypoint *EntrypointRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Entrypoint.Contract.EntrypointTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Entrypoint *EntrypointRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Entrypoint.Contract.EntrypointTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Entrypoint *EntrypointCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Entrypoint.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Entrypoint *EntrypointTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Entrypoint.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Entrypoint *EntrypointTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Entrypoint.Contract.contract.Transact(opts, method, params...)
}

// EcdsaValidatorStorage is a free data retrieval call binding the contract method 0x20709efc.
//
// Solidity: function ecdsaValidatorStorage(address ) view returns(address owner)
func (_Entrypoint *EntrypointCaller) EcdsaValidatorStorage(opts *bind.CallOpts, arg0 common.Address) (common.Address, error) {
	var out []interface{}
	err := _Entrypoint.contract.Call(opts, &out, "ecdsaValidatorStorage", arg0)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// EcdsaValidatorStorage is a free data retrieval call binding the contract method 0x20709efc.
//
// Solidity: function ecdsaValidatorStorage(address ) view returns(address owner)
func (_Entrypoint *EntrypointSession) EcdsaValidatorStorage(arg0 common.Address) (common.Address, error) {
	return _Entrypoint.Contract.EcdsaValidatorStorage(&_Entrypoint.CallOpts, arg0)
}

// EcdsaValidatorStorage is a free data retrieval call binding the contract method 0x20709efc.
//
// Solidity: function ecdsaValidatorStorage(address ) view returns(address owner)
func (_Entrypoint *EntrypointCallerSession) EcdsaValidatorStorage(arg0 common.Address) (common.Address, error) {
	return _Entrypoint.Contract.EcdsaValidatorStorage(&_Entrypoint.CallOpts, arg0)
}

// IsInitialized is a free data retrieval call binding the contract method 0xd60b347f.
//
// Solidity: function isInitialized(address smartAccount) view returns(bool)
func (_Entrypoint *EntrypointCaller) IsInitialized(opts *bind.CallOpts, smartAccount common.Address) (bool, error) {
	var out []interface{}
	err := _Entrypoint.contract.Call(opts, &out, "isInitialized", smartAccount)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsInitialized is a free data retrieval call binding the contract method 0xd60b347f.
//
// Solidity: function isInitialized(address smartAccount) view returns(bool)
func (_Entrypoint *EntrypointSession) IsInitialized(smartAccount common.Address) (bool, error) {
	return _Entrypoint.Contract.IsInitialized(&_Entrypoint.CallOpts, smartAccount)
}

// IsInitialized is a free data retrieval call binding the contract method 0xd60b347f.
//
// Solidity: function isInitialized(address smartAccount) view returns(bool)
func (_Entrypoint *EntrypointCallerSession) IsInitialized(smartAccount common.Address) (bool, error) {
	return _Entrypoint.Contract.IsInitialized(&_Entrypoint.CallOpts, smartAccount)
}

// IsModuleType is a free data retrieval call binding the contract method 0xecd05961.
//
// Solidity: function isModuleType(uint256 typeID) pure returns(bool)
func (_Entrypoint *EntrypointCaller) IsModuleType(opts *bind.CallOpts, typeID *big.Int) (bool, error) {
	var out []interface{}
	err := _Entrypoint.contract.Call(opts, &out, "isModuleType", typeID)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsModuleType is a free data retrieval call binding the contract method 0xecd05961.
//
// Solidity: function isModuleType(uint256 typeID) pure returns(bool)
func (_Entrypoint *EntrypointSession) IsModuleType(typeID *big.Int) (bool, error) {
	return _Entrypoint.Contract.IsModuleType(&_Entrypoint.CallOpts, typeID)
}

// IsModuleType is a free data retrieval call binding the contract method 0xecd05961.
//
// Solidity: function isModuleType(uint256 typeID) pure returns(bool)
func (_Entrypoint *EntrypointCallerSession) IsModuleType(typeID *big.Int) (bool, error) {
	return _Entrypoint.Contract.IsModuleType(&_Entrypoint.CallOpts, typeID)
}

// IsValidSignatureWithSender is a free data retrieval call binding the contract method 0xf551e2ee.
//
// Solidity: function isValidSignatureWithSender(address , bytes32 hash, bytes sig) view returns(bytes4)
func (_Entrypoint *EntrypointCaller) IsValidSignatureWithSender(opts *bind.CallOpts, arg0 common.Address, hash [32]byte, sig []byte) ([4]byte, error) {
	var out []interface{}
	err := _Entrypoint.contract.Call(opts, &out, "isValidSignatureWithSender", arg0, hash, sig)

	if err != nil {
		return *new([4]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([4]byte)).(*[4]byte)

	return out0, err

}

// IsValidSignatureWithSender is a free data retrieval call binding the contract method 0xf551e2ee.
//
// Solidity: function isValidSignatureWithSender(address , bytes32 hash, bytes sig) view returns(bytes4)
func (_Entrypoint *EntrypointSession) IsValidSignatureWithSender(arg0 common.Address, hash [32]byte, sig []byte) ([4]byte, error) {
	return _Entrypoint.Contract.IsValidSignatureWithSender(&_Entrypoint.CallOpts, arg0, hash, sig)
}

// IsValidSignatureWithSender is a free data retrieval call binding the contract method 0xf551e2ee.
//
// Solidity: function isValidSignatureWithSender(address , bytes32 hash, bytes sig) view returns(bytes4)
func (_Entrypoint *EntrypointCallerSession) IsValidSignatureWithSender(arg0 common.Address, hash [32]byte, sig []byte) ([4]byte, error) {
	return _Entrypoint.Contract.IsValidSignatureWithSender(&_Entrypoint.CallOpts, arg0, hash, sig)
}

// OnInstall is a paid mutator transaction binding the contract method 0x6d61fe70.
//
// Solidity: function onInstall(bytes _data) payable returns()
func (_Entrypoint *EntrypointTransactor) OnInstall(opts *bind.TransactOpts, _data []byte) (*types.Transaction, error) {
	return _Entrypoint.contract.Transact(opts, "onInstall", _data)
}

// OnInstall is a paid mutator transaction binding the contract method 0x6d61fe70.
//
// Solidity: function onInstall(bytes _data) payable returns()
func (_Entrypoint *EntrypointSession) OnInstall(_data []byte) (*types.Transaction, error) {
	return _Entrypoint.Contract.OnInstall(&_Entrypoint.TransactOpts, _data)
}

// OnInstall is a paid mutator transaction binding the contract method 0x6d61fe70.
//
// Solidity: function onInstall(bytes _data) payable returns()
func (_Entrypoint *EntrypointTransactorSession) OnInstall(_data []byte) (*types.Transaction, error) {
	return _Entrypoint.Contract.OnInstall(&_Entrypoint.TransactOpts, _data)
}

// OnUninstall is a paid mutator transaction binding the contract method 0x8a91b0e3.
//
// Solidity: function onUninstall(bytes ) payable returns()
func (_Entrypoint *EntrypointTransactor) OnUninstall(opts *bind.TransactOpts, arg0 []byte) (*types.Transaction, error) {
	return _Entrypoint.contract.Transact(opts, "onUninstall", arg0)
}

// OnUninstall is a paid mutator transaction binding the contract method 0x8a91b0e3.
//
// Solidity: function onUninstall(bytes ) payable returns()
func (_Entrypoint *EntrypointSession) OnUninstall(arg0 []byte) (*types.Transaction, error) {
	return _Entrypoint.Contract.OnUninstall(&_Entrypoint.TransactOpts, arg0)
}

// OnUninstall is a paid mutator transaction binding the contract method 0x8a91b0e3.
//
// Solidity: function onUninstall(bytes ) payable returns()
func (_Entrypoint *EntrypointTransactorSession) OnUninstall(arg0 []byte) (*types.Transaction, error) {
	return _Entrypoint.Contract.OnUninstall(&_Entrypoint.TransactOpts, arg0)
}

// PostCheck is a paid mutator transaction binding the contract method 0xaacbd72a.
//
// Solidity: function postCheck(bytes hookData, bool success, bytes res) payable returns()
func (_Entrypoint *EntrypointTransactor) PostCheck(opts *bind.TransactOpts, hookData []byte, success bool, res []byte) (*types.Transaction, error) {
	return _Entrypoint.contract.Transact(opts, "postCheck", hookData, success, res)
}

// PostCheck is a paid mutator transaction binding the contract method 0xaacbd72a.
//
// Solidity: function postCheck(bytes hookData, bool success, bytes res) payable returns()
func (_Entrypoint *EntrypointSession) PostCheck(hookData []byte, success bool, res []byte) (*types.Transaction, error) {
	return _Entrypoint.Contract.PostCheck(&_Entrypoint.TransactOpts, hookData, success, res)
}

// PostCheck is a paid mutator transaction binding the contract method 0xaacbd72a.
//
// Solidity: function postCheck(bytes hookData, bool success, bytes res) payable returns()
func (_Entrypoint *EntrypointTransactorSession) PostCheck(hookData []byte, success bool, res []byte) (*types.Transaction, error) {
	return _Entrypoint.Contract.PostCheck(&_Entrypoint.TransactOpts, hookData, success, res)
}

// PreCheck is a paid mutator transaction binding the contract method 0xd68f6025.
//
// Solidity: function preCheck(address msgSender, uint256 value, bytes ) payable returns(bytes)
func (_Entrypoint *EntrypointTransactor) PreCheck(opts *bind.TransactOpts, msgSender common.Address, value *big.Int, arg2 []byte) (*types.Transaction, error) {
	return _Entrypoint.contract.Transact(opts, "preCheck", msgSender, value, arg2)
}

// PreCheck is a paid mutator transaction binding the contract method 0xd68f6025.
//
// Solidity: function preCheck(address msgSender, uint256 value, bytes ) payable returns(bytes)
func (_Entrypoint *EntrypointSession) PreCheck(msgSender common.Address, value *big.Int, arg2 []byte) (*types.Transaction, error) {
	return _Entrypoint.Contract.PreCheck(&_Entrypoint.TransactOpts, msgSender, value, arg2)
}

// PreCheck is a paid mutator transaction binding the contract method 0xd68f6025.
//
// Solidity: function preCheck(address msgSender, uint256 value, bytes ) payable returns(bytes)
func (_Entrypoint *EntrypointTransactorSession) PreCheck(msgSender common.Address, value *big.Int, arg2 []byte) (*types.Transaction, error) {
	return _Entrypoint.Contract.PreCheck(&_Entrypoint.TransactOpts, msgSender, value, arg2)
}

// ValidateUserOp is a paid mutator transaction binding the contract method 0x97003203.
//
// Solidity: function validateUserOp((address,uint256,bytes,bytes,bytes32,uint256,bytes32,bytes,bytes) userOp, bytes32 userOpHash) payable returns(uint256)
func (_Entrypoint *EntrypointTransactor) ValidateUserOp(opts *bind.TransactOpts, userOp PackedUserOperation, userOpHash [32]byte) (*types.Transaction, error) {
	return _Entrypoint.contract.Transact(opts, "validateUserOp", userOp, userOpHash)
}

// ValidateUserOp is a paid mutator transaction binding the contract method 0x97003203.
//
// Solidity: function validateUserOp((address,uint256,bytes,bytes,bytes32,uint256,bytes32,bytes,bytes) userOp, bytes32 userOpHash) payable returns(uint256)
func (_Entrypoint *EntrypointSession) ValidateUserOp(userOp PackedUserOperation, userOpHash [32]byte) (*types.Transaction, error) {
	return _Entrypoint.Contract.ValidateUserOp(&_Entrypoint.TransactOpts, userOp, userOpHash)
}

// ValidateUserOp is a paid mutator transaction binding the contract method 0x97003203.
//
// Solidity: function validateUserOp((address,uint256,bytes,bytes,bytes32,uint256,bytes32,bytes,bytes) userOp, bytes32 userOpHash) payable returns(uint256)
func (_Entrypoint *EntrypointTransactorSession) ValidateUserOp(userOp PackedUserOperation, userOpHash [32]byte) (*types.Transaction, error) {
	return _Entrypoint.Contract.ValidateUserOp(&_Entrypoint.TransactOpts, userOp, userOpHash)
}

// EntrypointOwnerRegisteredIterator is returned from FilterOwnerRegistered and is used to iterate over the raw logs and unpacked data for OwnerRegistered events raised by the Entrypoint contract.
type EntrypointOwnerRegisteredIterator struct {
	Event *EntrypointOwnerRegistered // Event containing the contract specifics and raw log

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
func (it *EntrypointOwnerRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EntrypointOwnerRegistered)
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
		it.Event = new(EntrypointOwnerRegistered)
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
func (it *EntrypointOwnerRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EntrypointOwnerRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EntrypointOwnerRegistered represents a OwnerRegistered event raised by the Entrypoint contract.
type EntrypointOwnerRegistered struct {
	Kernel common.Address
	Owner  common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterOwnerRegistered is a free log retrieval operation binding the contract event 0xa5e1f8b4009110f5525798d04ae2125421a12d0590aa52c13682ff1bd3c492ca.
//
// Solidity: event OwnerRegistered(address indexed kernel, address indexed owner)
func (_Entrypoint *EntrypointFilterer) FilterOwnerRegistered(opts *bind.FilterOpts, kernel []common.Address, owner []common.Address) (*EntrypointOwnerRegisteredIterator, error) {

	var kernelRule []interface{}
	for _, kernelItem := range kernel {
		kernelRule = append(kernelRule, kernelItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _Entrypoint.contract.FilterLogs(opts, "OwnerRegistered", kernelRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return &EntrypointOwnerRegisteredIterator{contract: _Entrypoint.contract, event: "OwnerRegistered", logs: logs, sub: sub}, nil
}

// WatchOwnerRegistered is a free log subscription operation binding the contract event 0xa5e1f8b4009110f5525798d04ae2125421a12d0590aa52c13682ff1bd3c492ca.
//
// Solidity: event OwnerRegistered(address indexed kernel, address indexed owner)
func (_Entrypoint *EntrypointFilterer) WatchOwnerRegistered(opts *bind.WatchOpts, sink chan<- *EntrypointOwnerRegistered, kernel []common.Address, owner []common.Address) (event.Subscription, error) {

	var kernelRule []interface{}
	for _, kernelItem := range kernel {
		kernelRule = append(kernelRule, kernelItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _Entrypoint.contract.WatchLogs(opts, "OwnerRegistered", kernelRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EntrypointOwnerRegistered)
				if err := _Entrypoint.contract.UnpackLog(event, "OwnerRegistered", log); err != nil {
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
func (_Entrypoint *EntrypointFilterer) ParseOwnerRegistered(log types.Log) (*EntrypointOwnerRegistered, error) {
	event := new(EntrypointOwnerRegistered)
	if err := _Entrypoint.contract.UnpackLog(event, "OwnerRegistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
