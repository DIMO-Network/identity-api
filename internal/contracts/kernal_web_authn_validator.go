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

// WebAuthnValidatorMetaData contains all meta data concerning the WebAuthnValidator contract.
var WebAuthnValidatorMetaData = &bind.MetaData{
	ABI: "[{\"constant\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"smartAccount\",\"type\":\"address\"}],\"name\":\"AlreadyInitialized\",\"payable\":false,\"type\":\"error\"},{\"constant\":false,\"inputs\":[],\"name\":\"InvalidPublicKey\",\"payable\":false,\"type\":\"error\"},{\"constant\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"smartAccount\",\"type\":\"address\"}],\"name\":\"NotInitialized\",\"payable\":false,\"type\":\"error\"},{\"constant\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"kernel\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"authenticatorIdHash\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"pubKeyX\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"pubKeyY\",\"type\":\"uint256\"}],\"name\":\"WebAuthnPublicKeyRegistered\",\"payable\":false,\"type\":\"event\"},{\"constant\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"id\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"hash\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"sig\",\"type\":\"bytes\"}],\"name\":\"checkSignature\",\"outputs\":[{\"internalType\":\"bytes4\",\"name\":\"\",\"type\":\"bytes4\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"id\",\"type\":\"bytes32\"},{\"components\":[{\"name\":\"sender\",\"type\":\"address\"},{\"name\":\"nonce\",\"type\":\"uint256\"},{\"name\":\"initCode\",\"type\":\"bytes\"},{\"name\":\"callData\",\"type\":\"bytes\"},{\"name\":\"accountGasLimits\",\"type\":\"bytes32\"},{\"name\":\"preVerificationGas\",\"type\":\"uint256\"},{\"name\":\"gasFees\",\"type\":\"bytes32\"},{\"name\":\"paymasterAndData\",\"type\":\"bytes\"},{\"name\":\"signature\",\"type\":\"bytes\"}],\"indexed\":false,\"internalType\":\"structPackedUserOperation\",\"name\":\"userOp\",\"type\":\"tuple\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"userOpHash\",\"type\":\"bytes32\"}],\"name\":\"checkUserOpSignature\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"kernel\",\"type\":\"address\"}],\"name\":\"isInitialized\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"}],\"name\":\"isModuleType\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"pure\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"onInstall\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"onUninstall\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"usedIds\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"id\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"kernel\",\"type\":\"address\"}],\"name\":\"webAuthnSignerStorage\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"pubKeyX\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pubKeyY\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// WebAuthnValidatorABI is the input ABI used to generate the binding from.
// Deprecated: Use WebAuthnValidatorMetaData.ABI instead.
var WebAuthnValidatorABI = WebAuthnValidatorMetaData.ABI

// WebAuthnValidator is an auto generated Go binding around an Ethereum contract.
type WebAuthnValidator struct {
	WebAuthnValidatorCaller     // Read-only binding to the contract
	WebAuthnValidatorTransactor // Write-only binding to the contract
	WebAuthnValidatorFilterer   // Log filterer for contract events
}

// WebAuthnValidatorCaller is an auto generated read-only Go binding around an Ethereum contract.
type WebAuthnValidatorCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// WebAuthnValidatorTransactor is an auto generated write-only Go binding around an Ethereum contract.
type WebAuthnValidatorTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// WebAuthnValidatorFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type WebAuthnValidatorFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// WebAuthnValidatorSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type WebAuthnValidatorSession struct {
	Contract     *WebAuthnValidator // Generic contract binding to set the session for
	CallOpts     bind.CallOpts      // Call options to use throughout this session
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// WebAuthnValidatorCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type WebAuthnValidatorCallerSession struct {
	Contract *WebAuthnValidatorCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts            // Call options to use throughout this session
}

// WebAuthnValidatorTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type WebAuthnValidatorTransactorSession struct {
	Contract     *WebAuthnValidatorTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts            // Transaction auth options to use throughout this session
}

// WebAuthnValidatorRaw is an auto generated low-level Go binding around an Ethereum contract.
type WebAuthnValidatorRaw struct {
	Contract *WebAuthnValidator // Generic contract binding to access the raw methods on
}

// WebAuthnValidatorCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type WebAuthnValidatorCallerRaw struct {
	Contract *WebAuthnValidatorCaller // Generic read-only contract binding to access the raw methods on
}

// WebAuthnValidatorTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type WebAuthnValidatorTransactorRaw struct {
	Contract *WebAuthnValidatorTransactor // Generic write-only contract binding to access the raw methods on
}

// NewWebAuthnValidator creates a new instance of WebAuthnValidator, bound to a specific deployed contract.
func NewWebAuthnValidator(address common.Address, backend bind.ContractBackend) (*WebAuthnValidator, error) {
	contract, err := bindWebAuthnValidator(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &WebAuthnValidator{WebAuthnValidatorCaller: WebAuthnValidatorCaller{contract: contract}, WebAuthnValidatorTransactor: WebAuthnValidatorTransactor{contract: contract}, WebAuthnValidatorFilterer: WebAuthnValidatorFilterer{contract: contract}}, nil
}

// NewWebAuthnValidatorCaller creates a new read-only instance of WebAuthnValidator, bound to a specific deployed contract.
func NewWebAuthnValidatorCaller(address common.Address, caller bind.ContractCaller) (*WebAuthnValidatorCaller, error) {
	contract, err := bindWebAuthnValidator(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &WebAuthnValidatorCaller{contract: contract}, nil
}

// NewWebAuthnValidatorTransactor creates a new write-only instance of WebAuthnValidator, bound to a specific deployed contract.
func NewWebAuthnValidatorTransactor(address common.Address, transactor bind.ContractTransactor) (*WebAuthnValidatorTransactor, error) {
	contract, err := bindWebAuthnValidator(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &WebAuthnValidatorTransactor{contract: contract}, nil
}

// NewWebAuthnValidatorFilterer creates a new log filterer instance of WebAuthnValidator, bound to a specific deployed contract.
func NewWebAuthnValidatorFilterer(address common.Address, filterer bind.ContractFilterer) (*WebAuthnValidatorFilterer, error) {
	contract, err := bindWebAuthnValidator(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &WebAuthnValidatorFilterer{contract: contract}, nil
}

// bindWebAuthnValidator binds a generic wrapper to an already deployed contract.
func bindWebAuthnValidator(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(WebAuthnValidatorABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_WebAuthnValidator *WebAuthnValidatorRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _WebAuthnValidator.Contract.WebAuthnValidatorCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_WebAuthnValidator *WebAuthnValidatorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _WebAuthnValidator.Contract.WebAuthnValidatorTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_WebAuthnValidator *WebAuthnValidatorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _WebAuthnValidator.Contract.WebAuthnValidatorTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_WebAuthnValidator *WebAuthnValidatorCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _WebAuthnValidator.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_WebAuthnValidator *WebAuthnValidatorTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _WebAuthnValidator.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_WebAuthnValidator *WebAuthnValidatorTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _WebAuthnValidator.Contract.contract.Transact(opts, method, params...)
}

// CheckSignature is a free data retrieval call binding the contract method 0x392dffaf.
//
// Solidity: function checkSignature(bytes32 id, address sender, bytes32 hash, bytes sig) view returns(bytes4)
func (_WebAuthnValidator *WebAuthnValidatorCaller) CheckSignature(opts *bind.CallOpts, id [32]byte, sender common.Address, hash [32]byte, sig []byte) ([4]byte, error) {
	var out []interface{}
	err := _WebAuthnValidator.contract.Call(opts, &out, "checkSignature", id, sender, hash, sig)

	if err != nil {
		return *new([4]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([4]byte)).(*[4]byte)

	return out0, err

}

// CheckSignature is a free data retrieval call binding the contract method 0x392dffaf.
//
// Solidity: function checkSignature(bytes32 id, address sender, bytes32 hash, bytes sig) view returns(bytes4)
func (_WebAuthnValidator *WebAuthnValidatorSession) CheckSignature(id [32]byte, sender common.Address, hash [32]byte, sig []byte) ([4]byte, error) {
	return _WebAuthnValidator.Contract.CheckSignature(&_WebAuthnValidator.CallOpts, id, sender, hash, sig)
}

// CheckSignature is a free data retrieval call binding the contract method 0x392dffaf.
//
// Solidity: function checkSignature(bytes32 id, address sender, bytes32 hash, bytes sig) view returns(bytes4)
func (_WebAuthnValidator *WebAuthnValidatorCallerSession) CheckSignature(id [32]byte, sender common.Address, hash [32]byte, sig []byte) ([4]byte, error) {
	return _WebAuthnValidator.Contract.CheckSignature(&_WebAuthnValidator.CallOpts, id, sender, hash, sig)
}

// IsInitialized is a free data retrieval call binding the contract method 0xd60b347f.
//
// Solidity: function isInitialized(address kernel) view returns(bool)
func (_WebAuthnValidator *WebAuthnValidatorCaller) IsInitialized(opts *bind.CallOpts, kernel common.Address) (bool, error) {
	var out []interface{}
	err := _WebAuthnValidator.contract.Call(opts, &out, "isInitialized", kernel)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsInitialized is a free data retrieval call binding the contract method 0xd60b347f.
//
// Solidity: function isInitialized(address kernel) view returns(bool)
func (_WebAuthnValidator *WebAuthnValidatorSession) IsInitialized(kernel common.Address) (bool, error) {
	return _WebAuthnValidator.Contract.IsInitialized(&_WebAuthnValidator.CallOpts, kernel)
}

// IsInitialized is a free data retrieval call binding the contract method 0xd60b347f.
//
// Solidity: function isInitialized(address kernel) view returns(bool)
func (_WebAuthnValidator *WebAuthnValidatorCallerSession) IsInitialized(kernel common.Address) (bool, error) {
	return _WebAuthnValidator.Contract.IsInitialized(&_WebAuthnValidator.CallOpts, kernel)
}

// IsModuleType is a free data retrieval call binding the contract method 0xecd05961.
//
// Solidity: function isModuleType(uint256 id) pure returns(bool)
func (_WebAuthnValidator *WebAuthnValidatorCaller) IsModuleType(opts *bind.CallOpts, id *big.Int) (bool, error) {
	var out []interface{}
	err := _WebAuthnValidator.contract.Call(opts, &out, "isModuleType", id)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsModuleType is a free data retrieval call binding the contract method 0xecd05961.
//
// Solidity: function isModuleType(uint256 id) pure returns(bool)
func (_WebAuthnValidator *WebAuthnValidatorSession) IsModuleType(id *big.Int) (bool, error) {
	return _WebAuthnValidator.Contract.IsModuleType(&_WebAuthnValidator.CallOpts, id)
}

// IsModuleType is a free data retrieval call binding the contract method 0xecd05961.
//
// Solidity: function isModuleType(uint256 id) pure returns(bool)
func (_WebAuthnValidator *WebAuthnValidatorCallerSession) IsModuleType(id *big.Int) (bool, error) {
	return _WebAuthnValidator.Contract.IsModuleType(&_WebAuthnValidator.CallOpts, id)
}

// UsedIds is a free data retrieval call binding the contract method 0x244d6cb2.
//
// Solidity: function usedIds(address ) view returns(uint256)
func (_WebAuthnValidator *WebAuthnValidatorCaller) UsedIds(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _WebAuthnValidator.contract.Call(opts, &out, "usedIds", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// UsedIds is a free data retrieval call binding the contract method 0x244d6cb2.
//
// Solidity: function usedIds(address ) view returns(uint256)
func (_WebAuthnValidator *WebAuthnValidatorSession) UsedIds(arg0 common.Address) (*big.Int, error) {
	return _WebAuthnValidator.Contract.UsedIds(&_WebAuthnValidator.CallOpts, arg0)
}

// UsedIds is a free data retrieval call binding the contract method 0x244d6cb2.
//
// Solidity: function usedIds(address ) view returns(uint256)
func (_WebAuthnValidator *WebAuthnValidatorCallerSession) UsedIds(arg0 common.Address) (*big.Int, error) {
	return _WebAuthnValidator.Contract.UsedIds(&_WebAuthnValidator.CallOpts, arg0)
}

// WebAuthnSignerStorage is a free data retrieval call binding the contract method 0x1811663f.
//
// Solidity: function webAuthnSignerStorage(bytes32 id, address kernel) view returns(uint256 pubKeyX, uint256 pubKeyY)
func (_WebAuthnValidator *WebAuthnValidatorCaller) WebAuthnSignerStorage(opts *bind.CallOpts, id [32]byte, kernel common.Address) (struct {
	PubKeyX *big.Int
	PubKeyY *big.Int
}, error) {
	var out []interface{}
	err := _WebAuthnValidator.contract.Call(opts, &out, "webAuthnSignerStorage", id, kernel)

	outstruct := new(struct {
		PubKeyX *big.Int
		PubKeyY *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.PubKeyX = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.PubKeyY = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// WebAuthnSignerStorage is a free data retrieval call binding the contract method 0x1811663f.
//
// Solidity: function webAuthnSignerStorage(bytes32 id, address kernel) view returns(uint256 pubKeyX, uint256 pubKeyY)
func (_WebAuthnValidator *WebAuthnValidatorSession) WebAuthnSignerStorage(id [32]byte, kernel common.Address) (struct {
	PubKeyX *big.Int
	PubKeyY *big.Int
}, error) {
	return _WebAuthnValidator.Contract.WebAuthnSignerStorage(&_WebAuthnValidator.CallOpts, id, kernel)
}

// WebAuthnSignerStorage is a free data retrieval call binding the contract method 0x1811663f.
//
// Solidity: function webAuthnSignerStorage(bytes32 id, address kernel) view returns(uint256 pubKeyX, uint256 pubKeyY)
func (_WebAuthnValidator *WebAuthnValidatorCallerSession) WebAuthnSignerStorage(id [32]byte, kernel common.Address) (struct {
	PubKeyX *big.Int
	PubKeyY *big.Int
}, error) {
	return _WebAuthnValidator.Contract.WebAuthnSignerStorage(&_WebAuthnValidator.CallOpts, id, kernel)
}

// CheckUserOpSignature is a paid mutator transaction binding the contract method 0x0ccab7a1.
//
// Solidity: function checkUserOpSignature(bytes32 id, (address,uint256,bytes,bytes,bytes32,uint256,bytes32,bytes,bytes) userOp, bytes32 userOpHash) payable returns(uint256)
func (_WebAuthnValidator *WebAuthnValidatorTransactor) CheckUserOpSignature(opts *bind.TransactOpts, id [32]byte, userOp PackedUserOperation, userOpHash [32]byte) (*types.Transaction, error) {
	return _WebAuthnValidator.contract.Transact(opts, "checkUserOpSignature", id, userOp, userOpHash)
}

// CheckUserOpSignature is a paid mutator transaction binding the contract method 0x0ccab7a1.
//
// Solidity: function checkUserOpSignature(bytes32 id, (address,uint256,bytes,bytes,bytes32,uint256,bytes32,bytes,bytes) userOp, bytes32 userOpHash) payable returns(uint256)
func (_WebAuthnValidator *WebAuthnValidatorSession) CheckUserOpSignature(id [32]byte, userOp PackedUserOperation, userOpHash [32]byte) (*types.Transaction, error) {
	return _WebAuthnValidator.Contract.CheckUserOpSignature(&_WebAuthnValidator.TransactOpts, id, userOp, userOpHash)
}

// CheckUserOpSignature is a paid mutator transaction binding the contract method 0x0ccab7a1.
//
// Solidity: function checkUserOpSignature(bytes32 id, (address,uint256,bytes,bytes,bytes32,uint256,bytes32,bytes,bytes) userOp, bytes32 userOpHash) payable returns(uint256)
func (_WebAuthnValidator *WebAuthnValidatorTransactorSession) CheckUserOpSignature(id [32]byte, userOp PackedUserOperation, userOpHash [32]byte) (*types.Transaction, error) {
	return _WebAuthnValidator.Contract.CheckUserOpSignature(&_WebAuthnValidator.TransactOpts, id, userOp, userOpHash)
}

// OnInstall is a paid mutator transaction binding the contract method 0x6d61fe70.
//
// Solidity: function onInstall(bytes data) payable returns()
func (_WebAuthnValidator *WebAuthnValidatorTransactor) OnInstall(opts *bind.TransactOpts, data []byte) (*types.Transaction, error) {
	return _WebAuthnValidator.contract.Transact(opts, "onInstall", data)
}

// OnInstall is a paid mutator transaction binding the contract method 0x6d61fe70.
//
// Solidity: function onInstall(bytes data) payable returns()
func (_WebAuthnValidator *WebAuthnValidatorSession) OnInstall(data []byte) (*types.Transaction, error) {
	return _WebAuthnValidator.Contract.OnInstall(&_WebAuthnValidator.TransactOpts, data)
}

// OnInstall is a paid mutator transaction binding the contract method 0x6d61fe70.
//
// Solidity: function onInstall(bytes data) payable returns()
func (_WebAuthnValidator *WebAuthnValidatorTransactorSession) OnInstall(data []byte) (*types.Transaction, error) {
	return _WebAuthnValidator.Contract.OnInstall(&_WebAuthnValidator.TransactOpts, data)
}

// OnUninstall is a paid mutator transaction binding the contract method 0x8a91b0e3.
//
// Solidity: function onUninstall(bytes data) payable returns()
func (_WebAuthnValidator *WebAuthnValidatorTransactor) OnUninstall(opts *bind.TransactOpts, data []byte) (*types.Transaction, error) {
	return _WebAuthnValidator.contract.Transact(opts, "onUninstall", data)
}

// OnUninstall is a paid mutator transaction binding the contract method 0x8a91b0e3.
//
// Solidity: function onUninstall(bytes data) payable returns()
func (_WebAuthnValidator *WebAuthnValidatorSession) OnUninstall(data []byte) (*types.Transaction, error) {
	return _WebAuthnValidator.Contract.OnUninstall(&_WebAuthnValidator.TransactOpts, data)
}

// OnUninstall is a paid mutator transaction binding the contract method 0x8a91b0e3.
//
// Solidity: function onUninstall(bytes data) payable returns()
func (_WebAuthnValidator *WebAuthnValidatorTransactorSession) OnUninstall(data []byte) (*types.Transaction, error) {
	return _WebAuthnValidator.Contract.OnUninstall(&_WebAuthnValidator.TransactOpts, data)
}

// WebAuthnValidatorWebAuthnPublicKeyRegisteredIterator is returned from FilterWebAuthnPublicKeyRegistered and is used to iterate over the raw logs and unpacked data for WebAuthnPublicKeyRegistered events raised by the WebAuthnValidator contract.
type WebAuthnValidatorWebAuthnPublicKeyRegisteredIterator struct {
	Event *WebAuthnValidatorWebAuthnPublicKeyRegistered // Event containing the contract specifics and raw log

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
func (it *WebAuthnValidatorWebAuthnPublicKeyRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(WebAuthnValidatorWebAuthnPublicKeyRegistered)
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
		it.Event = new(WebAuthnValidatorWebAuthnPublicKeyRegistered)
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
func (it *WebAuthnValidatorWebAuthnPublicKeyRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *WebAuthnValidatorWebAuthnPublicKeyRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// WebAuthnValidatorWebAuthnPublicKeyRegistered represents a WebAuthnPublicKeyRegistered event raised by the WebAuthnValidator contract.
type WebAuthnValidatorWebAuthnPublicKeyRegistered struct {
	Kernel              common.Address
	AuthenticatorIdHash [32]byte
	PubKeyX             *big.Int
	PubKeyY             *big.Int
	Raw                 types.Log // Blockchain specific contextual infos
}

// FilterWebAuthnPublicKeyRegistered is a free log retrieval operation binding the contract event 0xdaa12c36d531747b295ac442f2dc73409156b4e78117b4b178bc019014b6cf5b.
//
// Solidity: event WebAuthnPublicKeyRegistered(address indexed kernel, bytes32 indexed authenticatorIdHash, uint256 pubKeyX, uint256 pubKeyY)
func (_WebAuthnValidator *WebAuthnValidatorFilterer) FilterWebAuthnPublicKeyRegistered(opts *bind.FilterOpts, kernel []common.Address, authenticatorIdHash [][32]byte) (*WebAuthnValidatorWebAuthnPublicKeyRegisteredIterator, error) {

	var kernelRule []interface{}
	for _, kernelItem := range kernel {
		kernelRule = append(kernelRule, kernelItem)
	}
	var authenticatorIdHashRule []interface{}
	for _, authenticatorIdHashItem := range authenticatorIdHash {
		authenticatorIdHashRule = append(authenticatorIdHashRule, authenticatorIdHashItem)
	}

	logs, sub, err := _WebAuthnValidator.contract.FilterLogs(opts, "WebAuthnPublicKeyRegistered", kernelRule, authenticatorIdHashRule)
	if err != nil {
		return nil, err
	}
	return &WebAuthnValidatorWebAuthnPublicKeyRegisteredIterator{contract: _WebAuthnValidator.contract, event: "WebAuthnPublicKeyRegistered", logs: logs, sub: sub}, nil
}

// WatchWebAuthnPublicKeyRegistered is a free log subscription operation binding the contract event 0xdaa12c36d531747b295ac442f2dc73409156b4e78117b4b178bc019014b6cf5b.
//
// Solidity: event WebAuthnPublicKeyRegistered(address indexed kernel, bytes32 indexed authenticatorIdHash, uint256 pubKeyX, uint256 pubKeyY)
func (_WebAuthnValidator *WebAuthnValidatorFilterer) WatchWebAuthnPublicKeyRegistered(opts *bind.WatchOpts, sink chan<- *WebAuthnValidatorWebAuthnPublicKeyRegistered, kernel []common.Address, authenticatorIdHash [][32]byte) (event.Subscription, error) {

	var kernelRule []interface{}
	for _, kernelItem := range kernel {
		kernelRule = append(kernelRule, kernelItem)
	}
	var authenticatorIdHashRule []interface{}
	for _, authenticatorIdHashItem := range authenticatorIdHash {
		authenticatorIdHashRule = append(authenticatorIdHashRule, authenticatorIdHashItem)
	}

	logs, sub, err := _WebAuthnValidator.contract.WatchLogs(opts, "WebAuthnPublicKeyRegistered", kernelRule, authenticatorIdHashRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(WebAuthnValidatorWebAuthnPublicKeyRegistered)
				if err := _WebAuthnValidator.contract.UnpackLog(event, "WebAuthnPublicKeyRegistered", log); err != nil {
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

// ParseWebAuthnPublicKeyRegistered is a log parse operation binding the contract event 0xdaa12c36d531747b295ac442f2dc73409156b4e78117b4b178bc019014b6cf5b.
//
// Solidity: event WebAuthnPublicKeyRegistered(address indexed kernel, bytes32 indexed authenticatorIdHash, uint256 pubKeyX, uint256 pubKeyY)
func (_WebAuthnValidator *WebAuthnValidatorFilterer) ParseWebAuthnPublicKeyRegistered(log types.Log) (*WebAuthnValidatorWebAuthnPublicKeyRegistered, error) {
	event := new(WebAuthnValidatorWebAuthnPublicKeyRegistered)
	if err := _WebAuthnValidator.contract.UnpackLog(event, "WebAuthnPublicKeyRegistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
