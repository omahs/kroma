// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package bindings

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

// KromaTokenMinterMetaData contains all meta data concerning the KromaTokenMinter contract.
var KromaTokenMinterMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"contractKromaToken\",\"name\":\"_kromaToken\",\"type\":\"address\"},{\"internalType\":\"address[]\",\"name\":\"_recipients\",\"type\":\"address[]\"},{\"internalType\":\"uint64[]\",\"name\":\"_shares\",\"type\":\"uint64[]\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[],\"name\":\"DEPOSITOR_ACCOUNT\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"MINT_DECREASE_CAP\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"MINT_DENOMINATOR\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"MINT_INCREASE_CAP\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"MINT_MAX_INCREASE\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"MINT_MIN_DECREASE\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"SHARE_DENOMINATOR\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"kromaToken\",\"outputs\":[{\"internalType\":\"contractKromaToken\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"totalAmount\",\"type\":\"uint256\"}],\"name\":\"mint\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"recipients\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"shareOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
	Bin: "0x60a06040523480156200001157600080fd5b50604051620009fa380380620009fa833981016040819052620000349162000379565b80518251146200008b5760405162461bcd60e51b815260206004820152601760248201527f696e76616c6964206c656e677468206f6620617272617900000000000000000060448201526064015b60405180910390fd5b6001600160a01b0383166080526000805b835181101562000219576000848281518110620000bd57620000bd62000461565b6020026020010151905060006001600160a01b0316816001600160a01b0316036200012b5760405162461bcd60e51b815260206004820152601d60248201527f726563697069656e7420616464726573732063616e6e6f742062652030000000604482015260640162000082565b600084838151811062000142576200014262000461565b60200260200101516001600160401b03169050806000036200019b5760405162461bcd60e51b8152602060048201526011602482015270073686172652063616e6e6f74206265203607c1b604482015260640162000082565b620001a781856200048d565b60008054600181810183557f290decd9548b62a8d60345a988386fc84ba6bc95484008f6362f93160ef3e56390910180546001600160a01b039096166001600160a01b0319909616861790559381526020939093526040909220559150806200021081620004a9565b9150506200009c565b50606481146200025d5760405162461bcd60e51b815260206004820152600e60248201526d696e76616c69642073686172657360901b604482015260640162000082565b50505050620004c5565b6001600160a01b03811681146200027d57600080fd5b50565b634e487b7160e01b600052604160045260246000fd5b604051601f8201601f191681016001600160401b0381118282101715620002c157620002c162000280565b604052919050565b60006001600160401b03821115620002e557620002e562000280565b5060051b60200190565b600082601f8301126200030157600080fd5b815160206200031a6200031483620002c9565b62000296565b82815260059290921b840181019181810190868411156200033a57600080fd5b8286015b848110156200036e5780516001600160401b0381168114620003605760008081fd5b83529183019183016200033e565b509695505050505050565b6000806000606084860312156200038f57600080fd5b83516200039c8162000267565b602085810151919450906001600160401b0380821115620003bc57600080fd5b818701915087601f830112620003d157600080fd5b8151620003e26200031482620002c9565b81815260059190911b8301840190848101908a8311156200040257600080fd5b938501935b828510156200042d5784516200041d8162000267565b8252938501939085019062000407565b60408a015190975094505050808311156200044757600080fd5b50506200045786828701620002ef565b9150509250925092565b634e487b7160e01b600052603260045260246000fd5b634e487b7160e01b600052601160045260246000fd5b80820180821115620004a357620004a362000477565b92915050565b600060018201620004be57620004be62000477565b5060010190565b608051610512620004e860003960008181610133015261031201526105126000f3fe608060405234801561001057600080fd5b50600436106100c95760003560e01c8063a0712d6811610081578063d1bc76a11161005b578063d1bc76a11461017a578063e06939a31461018d578063e591b2821461019557600080fd5b8063a0712d6814610119578063c27ca149146100ce578063c9d2b4961461012e57600080fd5b80632c7dc242116100b25780632c7dc242146101095780636745d032146101115780637eb118451461010957600080fd5b80630ccfab45146100ce57806321e5e2c4146100e9575b600080fd5b6100d6600a81565b6040519081526020015b60405180910390f35b6100d66100f73660046103c1565b60016020526000908152604090205481565b6100d6606481565b6100d6600781565b61012c6101273660046103fe565b6101b0565b005b6101557f000000000000000000000000000000000000000000000000000000000000000081565b60405173ffffffffffffffffffffffffffffffffffffffff90911681526020016100e0565b6101556101883660046103fe565b61038a565b6100d6600381565b61015573deaddeaddeaddeaddeaddeaddeaddeaddead000281565b3373deaddeaddeaddeaddeaddeaddeaddeaddead000214610257576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602560248201527f6f6e6c79206465706f7369746f722063616e2063616c6c20746869732066756e60448201527f6374696f6e000000000000000000000000000000000000000000000000000000606482015260840160405180910390fd5b60005b60005481101561038657600080828154811061027857610278610417565b600091825260208083209091015473ffffffffffffffffffffffffffffffffffffffff16808352600190915260408220549092509060646102b98387610475565b6102c39190610492565b6040517f40c10f1900000000000000000000000000000000000000000000000000000000815273ffffffffffffffffffffffffffffffffffffffff8581166004830152602482018390529192507f0000000000000000000000000000000000000000000000000000000000000000909116906340c10f1990604401600060405180830381600087803b15801561035857600080fd5b505af115801561036c573d6000803e3d6000fd5b50505050505050808061037e906104cd565b91505061025a565b5050565b6000818154811061039a57600080fd5b60009182526020909120015473ffffffffffffffffffffffffffffffffffffffff16905081565b6000602082840312156103d357600080fd5b813573ffffffffffffffffffffffffffffffffffffffff811681146103f757600080fd5b9392505050565b60006020828403121561041057600080fd5b5035919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052603260045260246000fd5b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b808202811582820484141761048c5761048c610446565b92915050565b6000826104c8577f4e487b7100000000000000000000000000000000000000000000000000000000600052601260045260246000fd5b500490565b60007fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff82036104fe576104fe610446565b506001019056fea164736f6c6343000815000a",
}

// KromaTokenMinterABI is the input ABI used to generate the binding from.
// Deprecated: Use KromaTokenMinterMetaData.ABI instead.
var KromaTokenMinterABI = KromaTokenMinterMetaData.ABI

// KromaTokenMinterBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use KromaTokenMinterMetaData.Bin instead.
var KromaTokenMinterBin = KromaTokenMinterMetaData.Bin

// DeployKromaTokenMinter deploys a new Ethereum contract, binding an instance of KromaTokenMinter to it.
func DeployKromaTokenMinter(auth *bind.TransactOpts, backend bind.ContractBackend, _kromaToken common.Address, _recipients []common.Address, _shares []uint64) (common.Address, *types.Transaction, *KromaTokenMinter, error) {
	parsed, err := KromaTokenMinterMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(KromaTokenMinterBin), backend, _kromaToken, _recipients, _shares)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &KromaTokenMinter{KromaTokenMinterCaller: KromaTokenMinterCaller{contract: contract}, KromaTokenMinterTransactor: KromaTokenMinterTransactor{contract: contract}, KromaTokenMinterFilterer: KromaTokenMinterFilterer{contract: contract}}, nil
}

// KromaTokenMinter is an auto generated Go binding around an Ethereum contract.
type KromaTokenMinter struct {
	KromaTokenMinterCaller     // Read-only binding to the contract
	KromaTokenMinterTransactor // Write-only binding to the contract
	KromaTokenMinterFilterer   // Log filterer for contract events
}

// KromaTokenMinterCaller is an auto generated read-only Go binding around an Ethereum contract.
type KromaTokenMinterCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// KromaTokenMinterTransactor is an auto generated write-only Go binding around an Ethereum contract.
type KromaTokenMinterTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// KromaTokenMinterFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type KromaTokenMinterFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// KromaTokenMinterSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type KromaTokenMinterSession struct {
	Contract     *KromaTokenMinter // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// KromaTokenMinterCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type KromaTokenMinterCallerSession struct {
	Contract *KromaTokenMinterCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts           // Call options to use throughout this session
}

// KromaTokenMinterTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type KromaTokenMinterTransactorSession struct {
	Contract     *KromaTokenMinterTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts           // Transaction auth options to use throughout this session
}

// KromaTokenMinterRaw is an auto generated low-level Go binding around an Ethereum contract.
type KromaTokenMinterRaw struct {
	Contract *KromaTokenMinter // Generic contract binding to access the raw methods on
}

// KromaTokenMinterCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type KromaTokenMinterCallerRaw struct {
	Contract *KromaTokenMinterCaller // Generic read-only contract binding to access the raw methods on
}

// KromaTokenMinterTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type KromaTokenMinterTransactorRaw struct {
	Contract *KromaTokenMinterTransactor // Generic write-only contract binding to access the raw methods on
}

// NewKromaTokenMinter creates a new instance of KromaTokenMinter, bound to a specific deployed contract.
func NewKromaTokenMinter(address common.Address, backend bind.ContractBackend) (*KromaTokenMinter, error) {
	contract, err := bindKromaTokenMinter(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &KromaTokenMinter{KromaTokenMinterCaller: KromaTokenMinterCaller{contract: contract}, KromaTokenMinterTransactor: KromaTokenMinterTransactor{contract: contract}, KromaTokenMinterFilterer: KromaTokenMinterFilterer{contract: contract}}, nil
}

// NewKromaTokenMinterCaller creates a new read-only instance of KromaTokenMinter, bound to a specific deployed contract.
func NewKromaTokenMinterCaller(address common.Address, caller bind.ContractCaller) (*KromaTokenMinterCaller, error) {
	contract, err := bindKromaTokenMinter(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &KromaTokenMinterCaller{contract: contract}, nil
}

// NewKromaTokenMinterTransactor creates a new write-only instance of KromaTokenMinter, bound to a specific deployed contract.
func NewKromaTokenMinterTransactor(address common.Address, transactor bind.ContractTransactor) (*KromaTokenMinterTransactor, error) {
	contract, err := bindKromaTokenMinter(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &KromaTokenMinterTransactor{contract: contract}, nil
}

// NewKromaTokenMinterFilterer creates a new log filterer instance of KromaTokenMinter, bound to a specific deployed contract.
func NewKromaTokenMinterFilterer(address common.Address, filterer bind.ContractFilterer) (*KromaTokenMinterFilterer, error) {
	contract, err := bindKromaTokenMinter(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &KromaTokenMinterFilterer{contract: contract}, nil
}

// bindKromaTokenMinter binds a generic wrapper to an already deployed contract.
func bindKromaTokenMinter(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := KromaTokenMinterMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_KromaTokenMinter *KromaTokenMinterRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _KromaTokenMinter.Contract.KromaTokenMinterCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_KromaTokenMinter *KromaTokenMinterRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _KromaTokenMinter.Contract.KromaTokenMinterTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_KromaTokenMinter *KromaTokenMinterRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _KromaTokenMinter.Contract.KromaTokenMinterTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_KromaTokenMinter *KromaTokenMinterCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _KromaTokenMinter.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_KromaTokenMinter *KromaTokenMinterTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _KromaTokenMinter.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_KromaTokenMinter *KromaTokenMinterTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _KromaTokenMinter.Contract.contract.Transact(opts, method, params...)
}

// DEPOSITORACCOUNT is a free data retrieval call binding the contract method 0xe591b282.
//
// Solidity: function DEPOSITOR_ACCOUNT() view returns(address)
func (_KromaTokenMinter *KromaTokenMinterCaller) DEPOSITORACCOUNT(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _KromaTokenMinter.contract.Call(opts, &out, "DEPOSITOR_ACCOUNT")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// DEPOSITORACCOUNT is a free data retrieval call binding the contract method 0xe591b282.
//
// Solidity: function DEPOSITOR_ACCOUNT() view returns(address)
func (_KromaTokenMinter *KromaTokenMinterSession) DEPOSITORACCOUNT() (common.Address, error) {
	return _KromaTokenMinter.Contract.DEPOSITORACCOUNT(&_KromaTokenMinter.CallOpts)
}

// DEPOSITORACCOUNT is a free data retrieval call binding the contract method 0xe591b282.
//
// Solidity: function DEPOSITOR_ACCOUNT() view returns(address)
func (_KromaTokenMinter *KromaTokenMinterCallerSession) DEPOSITORACCOUNT() (common.Address, error) {
	return _KromaTokenMinter.Contract.DEPOSITORACCOUNT(&_KromaTokenMinter.CallOpts)
}

// MINTDECREASECAP is a free data retrieval call binding the contract method 0xc27ca149.
//
// Solidity: function MINT_DECREASE_CAP() view returns(uint256)
func (_KromaTokenMinter *KromaTokenMinterCaller) MINTDECREASECAP(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _KromaTokenMinter.contract.Call(opts, &out, "MINT_DECREASE_CAP")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MINTDECREASECAP is a free data retrieval call binding the contract method 0xc27ca149.
//
// Solidity: function MINT_DECREASE_CAP() view returns(uint256)
func (_KromaTokenMinter *KromaTokenMinterSession) MINTDECREASECAP() (*big.Int, error) {
	return _KromaTokenMinter.Contract.MINTDECREASECAP(&_KromaTokenMinter.CallOpts)
}

// MINTDECREASECAP is a free data retrieval call binding the contract method 0xc27ca149.
//
// Solidity: function MINT_DECREASE_CAP() view returns(uint256)
func (_KromaTokenMinter *KromaTokenMinterCallerSession) MINTDECREASECAP() (*big.Int, error) {
	return _KromaTokenMinter.Contract.MINTDECREASECAP(&_KromaTokenMinter.CallOpts)
}

// MINTDENOMINATOR is a free data retrieval call binding the contract method 0x2c7dc242.
//
// Solidity: function MINT_DENOMINATOR() view returns(uint256)
func (_KromaTokenMinter *KromaTokenMinterCaller) MINTDENOMINATOR(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _KromaTokenMinter.contract.Call(opts, &out, "MINT_DENOMINATOR")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MINTDENOMINATOR is a free data retrieval call binding the contract method 0x2c7dc242.
//
// Solidity: function MINT_DENOMINATOR() view returns(uint256)
func (_KromaTokenMinter *KromaTokenMinterSession) MINTDENOMINATOR() (*big.Int, error) {
	return _KromaTokenMinter.Contract.MINTDENOMINATOR(&_KromaTokenMinter.CallOpts)
}

// MINTDENOMINATOR is a free data retrieval call binding the contract method 0x2c7dc242.
//
// Solidity: function MINT_DENOMINATOR() view returns(uint256)
func (_KromaTokenMinter *KromaTokenMinterCallerSession) MINTDENOMINATOR() (*big.Int, error) {
	return _KromaTokenMinter.Contract.MINTDENOMINATOR(&_KromaTokenMinter.CallOpts)
}

// MINTINCREASECAP is a free data retrieval call binding the contract method 0x0ccfab45.
//
// Solidity: function MINT_INCREASE_CAP() view returns(uint256)
func (_KromaTokenMinter *KromaTokenMinterCaller) MINTINCREASECAP(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _KromaTokenMinter.contract.Call(opts, &out, "MINT_INCREASE_CAP")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MINTINCREASECAP is a free data retrieval call binding the contract method 0x0ccfab45.
//
// Solidity: function MINT_INCREASE_CAP() view returns(uint256)
func (_KromaTokenMinter *KromaTokenMinterSession) MINTINCREASECAP() (*big.Int, error) {
	return _KromaTokenMinter.Contract.MINTINCREASECAP(&_KromaTokenMinter.CallOpts)
}

// MINTINCREASECAP is a free data retrieval call binding the contract method 0x0ccfab45.
//
// Solidity: function MINT_INCREASE_CAP() view returns(uint256)
func (_KromaTokenMinter *KromaTokenMinterCallerSession) MINTINCREASECAP() (*big.Int, error) {
	return _KromaTokenMinter.Contract.MINTINCREASECAP(&_KromaTokenMinter.CallOpts)
}

// MINTMAXINCREASE is a free data retrieval call binding the contract method 0xe06939a3.
//
// Solidity: function MINT_MAX_INCREASE() view returns(uint256)
func (_KromaTokenMinter *KromaTokenMinterCaller) MINTMAXINCREASE(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _KromaTokenMinter.contract.Call(opts, &out, "MINT_MAX_INCREASE")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MINTMAXINCREASE is a free data retrieval call binding the contract method 0xe06939a3.
//
// Solidity: function MINT_MAX_INCREASE() view returns(uint256)
func (_KromaTokenMinter *KromaTokenMinterSession) MINTMAXINCREASE() (*big.Int, error) {
	return _KromaTokenMinter.Contract.MINTMAXINCREASE(&_KromaTokenMinter.CallOpts)
}

// MINTMAXINCREASE is a free data retrieval call binding the contract method 0xe06939a3.
//
// Solidity: function MINT_MAX_INCREASE() view returns(uint256)
func (_KromaTokenMinter *KromaTokenMinterCallerSession) MINTMAXINCREASE() (*big.Int, error) {
	return _KromaTokenMinter.Contract.MINTMAXINCREASE(&_KromaTokenMinter.CallOpts)
}

// MINTMINDECREASE is a free data retrieval call binding the contract method 0x6745d032.
//
// Solidity: function MINT_MIN_DECREASE() view returns(uint256)
func (_KromaTokenMinter *KromaTokenMinterCaller) MINTMINDECREASE(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _KromaTokenMinter.contract.Call(opts, &out, "MINT_MIN_DECREASE")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MINTMINDECREASE is a free data retrieval call binding the contract method 0x6745d032.
//
// Solidity: function MINT_MIN_DECREASE() view returns(uint256)
func (_KromaTokenMinter *KromaTokenMinterSession) MINTMINDECREASE() (*big.Int, error) {
	return _KromaTokenMinter.Contract.MINTMINDECREASE(&_KromaTokenMinter.CallOpts)
}

// MINTMINDECREASE is a free data retrieval call binding the contract method 0x6745d032.
//
// Solidity: function MINT_MIN_DECREASE() view returns(uint256)
func (_KromaTokenMinter *KromaTokenMinterCallerSession) MINTMINDECREASE() (*big.Int, error) {
	return _KromaTokenMinter.Contract.MINTMINDECREASE(&_KromaTokenMinter.CallOpts)
}

// SHAREDENOMINATOR is a free data retrieval call binding the contract method 0x7eb11845.
//
// Solidity: function SHARE_DENOMINATOR() view returns(uint256)
func (_KromaTokenMinter *KromaTokenMinterCaller) SHAREDENOMINATOR(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _KromaTokenMinter.contract.Call(opts, &out, "SHARE_DENOMINATOR")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// SHAREDENOMINATOR is a free data retrieval call binding the contract method 0x7eb11845.
//
// Solidity: function SHARE_DENOMINATOR() view returns(uint256)
func (_KromaTokenMinter *KromaTokenMinterSession) SHAREDENOMINATOR() (*big.Int, error) {
	return _KromaTokenMinter.Contract.SHAREDENOMINATOR(&_KromaTokenMinter.CallOpts)
}

// SHAREDENOMINATOR is a free data retrieval call binding the contract method 0x7eb11845.
//
// Solidity: function SHARE_DENOMINATOR() view returns(uint256)
func (_KromaTokenMinter *KromaTokenMinterCallerSession) SHAREDENOMINATOR() (*big.Int, error) {
	return _KromaTokenMinter.Contract.SHAREDENOMINATOR(&_KromaTokenMinter.CallOpts)
}

// KromaToken is a free data retrieval call binding the contract method 0xc9d2b496.
//
// Solidity: function kromaToken() view returns(address)
func (_KromaTokenMinter *KromaTokenMinterCaller) KromaToken(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _KromaTokenMinter.contract.Call(opts, &out, "kromaToken")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// KromaToken is a free data retrieval call binding the contract method 0xc9d2b496.
//
// Solidity: function kromaToken() view returns(address)
func (_KromaTokenMinter *KromaTokenMinterSession) KromaToken() (common.Address, error) {
	return _KromaTokenMinter.Contract.KromaToken(&_KromaTokenMinter.CallOpts)
}

// KromaToken is a free data retrieval call binding the contract method 0xc9d2b496.
//
// Solidity: function kromaToken() view returns(address)
func (_KromaTokenMinter *KromaTokenMinterCallerSession) KromaToken() (common.Address, error) {
	return _KromaTokenMinter.Contract.KromaToken(&_KromaTokenMinter.CallOpts)
}

// Recipients is a free data retrieval call binding the contract method 0xd1bc76a1.
//
// Solidity: function recipients(uint256 ) view returns(address)
func (_KromaTokenMinter *KromaTokenMinterCaller) Recipients(opts *bind.CallOpts, arg0 *big.Int) (common.Address, error) {
	var out []interface{}
	err := _KromaTokenMinter.contract.Call(opts, &out, "recipients", arg0)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Recipients is a free data retrieval call binding the contract method 0xd1bc76a1.
//
// Solidity: function recipients(uint256 ) view returns(address)
func (_KromaTokenMinter *KromaTokenMinterSession) Recipients(arg0 *big.Int) (common.Address, error) {
	return _KromaTokenMinter.Contract.Recipients(&_KromaTokenMinter.CallOpts, arg0)
}

// Recipients is a free data retrieval call binding the contract method 0xd1bc76a1.
//
// Solidity: function recipients(uint256 ) view returns(address)
func (_KromaTokenMinter *KromaTokenMinterCallerSession) Recipients(arg0 *big.Int) (common.Address, error) {
	return _KromaTokenMinter.Contract.Recipients(&_KromaTokenMinter.CallOpts, arg0)
}

// ShareOf is a free data retrieval call binding the contract method 0x21e5e2c4.
//
// Solidity: function shareOf(address ) view returns(uint256)
func (_KromaTokenMinter *KromaTokenMinterCaller) ShareOf(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _KromaTokenMinter.contract.Call(opts, &out, "shareOf", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ShareOf is a free data retrieval call binding the contract method 0x21e5e2c4.
//
// Solidity: function shareOf(address ) view returns(uint256)
func (_KromaTokenMinter *KromaTokenMinterSession) ShareOf(arg0 common.Address) (*big.Int, error) {
	return _KromaTokenMinter.Contract.ShareOf(&_KromaTokenMinter.CallOpts, arg0)
}

// ShareOf is a free data retrieval call binding the contract method 0x21e5e2c4.
//
// Solidity: function shareOf(address ) view returns(uint256)
func (_KromaTokenMinter *KromaTokenMinterCallerSession) ShareOf(arg0 common.Address) (*big.Int, error) {
	return _KromaTokenMinter.Contract.ShareOf(&_KromaTokenMinter.CallOpts, arg0)
}

// Mint is a paid mutator transaction binding the contract method 0xa0712d68.
//
// Solidity: function mint(uint256 totalAmount) returns()
func (_KromaTokenMinter *KromaTokenMinterTransactor) Mint(opts *bind.TransactOpts, totalAmount *big.Int) (*types.Transaction, error) {
	return _KromaTokenMinter.contract.Transact(opts, "mint", totalAmount)
}

// Mint is a paid mutator transaction binding the contract method 0xa0712d68.
//
// Solidity: function mint(uint256 totalAmount) returns()
func (_KromaTokenMinter *KromaTokenMinterSession) Mint(totalAmount *big.Int) (*types.Transaction, error) {
	return _KromaTokenMinter.Contract.Mint(&_KromaTokenMinter.TransactOpts, totalAmount)
}

// Mint is a paid mutator transaction binding the contract method 0xa0712d68.
//
// Solidity: function mint(uint256 totalAmount) returns()
func (_KromaTokenMinter *KromaTokenMinterTransactorSession) Mint(totalAmount *big.Int) (*types.Transaction, error) {
	return _KromaTokenMinter.Contract.Mint(&_KromaTokenMinter.TransactOpts, totalAmount)
}
