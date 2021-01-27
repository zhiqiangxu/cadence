/*
 * Cadence - The resource-oriented smart contract programming language
 *
 * Copyright 2019-2020 Dapper Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package runtime

import (
	"time"

	"github.com/hashicorp/go-multierror"

	"github.com/onflow/cadence"
	"github.com/onflow/cadence/runtime/ast"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/cadence/runtime/sema"
)

const BlockHashLength = 32

type BlockHash [BlockHashLength]byte

type Block struct {
	Height    uint64
	View      uint64
	Hash      BlockHash
	Timestamp int64
}

type ResolvedLocation = sema.ResolvedLocation
type Identifier = ast.Identifier
type Location = common.Location
type AddressLocation = common.AddressLocation

// Accounts manages changes to the accounts
//
// Errors returned by this methods of this interface are considered non-fatal,
// i.e. user errors (e.g. adding a key to an account that does not exist).
//
// Fatal errors (e.g. a storage failure) are panics inside the methods.
type Accounts interface {
	// NewAccount creates a new account and returns its address
	NewAccount() (address Address, err error)
	// AccountExists returns true if the account exists
	AccountExists(address Address) (exists bool, err error)
	// NumberOfAccounts returns the number of accounts
	NumberOfAccounts() (count uint64, err error)

	// SuspendAccount suspends an account (set suspend flag to true)
	SuspendAccount(address Address) error
	// UnsuspendAccount unsuspend an account (set suspend flag to false)
	UnsuspendAccount(address Address) error
	// returns true if account is suspended
	IsAccountSuspended(address Address) (isSuspended bool, err error)
}

// AccountContracts manages contracts stored under accounts
type AccountContracts interface {
	// AccountContractCode returns the code associated with an account contract.
	ContractCode(address AddressLocation) (code []byte, err error)
	// UpdateAccountContractCode updates the code associated with an account contract.
	UpdateContractCode(address AddressLocation, code []byte) (err error)
	// RemoveContractCode removes the code associated with an account contract.
	RemoveContractCode(address AddressLocation) (err error)
	// Contracts returns a list of contract names under this account
	Contracts(address AddressLocation) (name []string, err error)
}

// AccountStorage stores and retrives account key/values
type AccountStorage interface {
	// GetValue gets a value for the given key in the storage.
	GetValue(key StorageKey) (value StorageValue, err error)
	// SetValue sets a value for the given key in the storage.
	SetValue(key StorageKey, value StorageValue) (err error)
	// ValueExists returns true if the given key exists in the storage.
	ValueExists(key StorageKey) (exists bool, err error)
	// StoredKeys returns an iterator of all storage keys and their sizes owned by the given account.
	StoredKeys(address Address) (iter StorageKeyIterator, err error)
	// StorageUsed gets storage used in bytes by the address at the moment of the function call.
	// NOTE: Storage capacity functionality is provided through injected functions.
	StorageUsed(address Address) (value uint64, err error)
}

// AccountKeys manages account keys
type AccountKeys interface {
	// AddAccountKey appends a key to an account.
	AddAccountKey(address Address, publicKey []byte) error
	// RemoveAccountKey removes a key from an account by index.
	RevokeAccountKey(address Address, index int) (publicKey []byte, err error)
	// AccountPublicKey returns the account key for the given index.
	AccountPublicKey(address Address, index int) (publicKey []byte, err error)
}

// LocationResolver provides functionality to locate codes
type LocationResolver interface {
	// GetCode returns the code at a given location
	GetCode(location Location) ([]byte, error)
	// ResolveLocation resolves an import location.
	ResolveLocation(identifiers []Identifier, location Location) ([]ResolvedLocation, error)
}

// Results is an interface that is used to capture artifacts generated by executing a Cadence program.
//
// These functions won't be directly callable by users, but through Cadence language constructs,
// such as the `emit` statement to emit an event.
type Results interface {
	// AppendLog appends a log to the log collection.
	AppendLog(string) error
	// Logs returns all logs that have been appended so far.
	Logs() ([]string, error)
	// LogAt returns the log at the given index of the log collection.
	LogAt(index uint) (string, error)
	// LogCount returns number of logs in the log collection.
	LogCount() uint

	// AppendEvent appends an event to the event collection.
	AppendEvent(cadence.Event) error
	// Events returns all the events in the event collection.
	Events() ([]cadence.Event, error)
	// EventAt returns the event at the given index of the event collection
	EventAt(index uint) (cadence.Event, error)
	// EventCount returns number of events in the event collection.
	EventCount() uint

	// AppendError appends a non-fatal error to the error collection.
	AppendError(error) error
	// Errors returns all errors in the error collection.
	Errors() multierror.Error
	// ErrorAt returns the error at the given index of the error collection.
	// The first return value is the actual non-fatal error,
	// the second return value indicates if there was an error while getting the error.
	ErrorAt(index uint) (Error, error)
	// ErrorCount returns number of errors in the error collection.
	ErrorCount() uint

	// AddComputationUsed increases the computation usage accumulator by the given amount.
	AddComputationUsed(uint64) error
	// ComputationSpent returns the total amount of computation spent during the execution.
	ComputationSpent() uint64
	// ComputationLimit returns the computation limit, the maximum computation that may be used during execution.
	// Ramtin: (we might not need this to be passed and just be enforced in the Results)
	ComputationLimit() uint64
}

// ProgramCache provides caching functionality for Cadence programs (ASTs).
//
type ProgramCache interface {
	// GetCachedProgram attempts to get a parsed program from a cache.
	GetCachedProgram(Location) (*ast.Program, error)
	// CacheProgram adds a parsed program to a cache.
	CacheProgram(Location, *ast.Program) error
}

type CryptoProvider interface {
	// VerifySignature returns true if the given signature was produced by signing the given tag + data
	// using the given public key, signature algorithm, and hash algorithm.
	VerifySignature(
		signature []byte,
		tag string,
		signedData []byte,
		publicKey []byte,
		signatureAlgorithm string,
		hashAlgorithm string,
	) (bool, error)
	// Hash returns the digest of hashing the given data with using the given hash algorithm.
	Hash(data []byte, hashAlgorithm string) ([]byte, error)
}

type Metrics interface {
	// ProgramParsed captures the time spent on parsing the program.
	ProgramParsed(location common.Location, duration time.Duration)
	// ProgramChecked captures the time spent on checking the parsed program.
	ProgramChecked(location common.Location, duration time.Duration)
	// ProgramInterpreted captures the time spent on interpreting the parsed and checked program.
	ProgramInterpreted(location common.Location, duration time.Duration)

	// ValueEncoded captures the time spent on encoding a value.
	// TODO: maybe add type
	ValueEncoded(duration time.Duration)
	// ValueDecoded capture the time spent on decoding an encoded value.
	// TODO: maybe add type
	ValueDecoded(duration time.Duration)
}

type HighLevelAccountStorage interface {
	// HighLevelStorageEnabled should return true
	// if the functions of HighLevelStorage should be called,
	// e.g. SetCadenceValue
	HighLevelStorageEnabled() bool

	// SetCadenceValue sets a value for the given key in the storage, owned by the given account.
	SetCadenceValue(owner Address, key string, value cadence.Value) (err error)
}

// Utils provides some utility functionality needed for cadence
type Utils interface {
	// GenerateUUID generates UUIDs
	GenerateUUID() (uint64, error)
}

type EmptyAccounts struct{}

var _ Accounts = &EmptyAccounts{}

func (i *EmptyAccounts) NewAccount() (Address, error) {
	return Address{}, nil
}

func (i *EmptyAccounts) AccountExists(_ Address) (bool, error) {
	return false, nil
}

func (i *EmptyAccounts) NumberOfAccounts() (uint64, error) {
	return 0, nil
}

func (i *EmptyAccounts) SuspendAccount(_ Address) error {
	return nil
}

func (i *EmptyAccounts) UnsuspendAccount(_ Address) error {
	return nil
}

func (i *EmptyAccounts) IsAccountSuspended(_ Address) (bool, error) {
	return false, nil
}

type EmptyAccountContracts struct{}

var _ AccountContracts = &EmptyAccountContracts{}

func (i *EmptyAccountContracts) ContractCode(_ AddressLocation) ([]byte, error) {
	return nil, nil
}

func (i *EmptyAccountContracts) UpdateContractCode(_ AddressLocation, _ []byte) (err error) {
	return nil
}

func (i *EmptyAccountContracts) RemoveContractCode(_ AddressLocation) (err error) {
	return nil
}

func (i *EmptyAccountContracts) Contracts(_ AddressLocation) (name []string, err error) {
	return nil, nil
}

type EmptyAccountStorage struct{}

var _ AccountStorage = &EmptyAccountStorage{}

func (i *EmptyAccountStorage) ValueExists(_ StorageKey) (exists bool, err error) {
	return false, nil
}

func (i *EmptyAccountStorage) GetValue(_ StorageKey) (value StorageValue, err error) {
	return nil, nil
}

func (i *EmptyAccountStorage) SetValue(_ StorageKey, _ StorageValue) error {
	return nil
}

func (i *EmptyAccountStorage) StorageUsed(_ Address) (uint64, error) {
	return 0, nil
}

func (i *EmptyAccountStorage) StorageCapacity(_ Address) (uint64, error) {
	return 0, nil
}

func (i *EmptyAccountStorage) StoredKeys(_ Address) (StorageKeyIterator, error) {
	return nil, nil
}

type EmptyAccountKeys struct{}

var _ AccountKeys = &EmptyAccountKeys{}

func (i *EmptyAccountKeys) AddAccountKey(_ Address, _ []byte) error {
	return nil
}

func (i *EmptyAccountKeys) RevokeAccountKey(_ Address, _ int) ([]byte, error) {
	return nil, nil
}

func (i *EmptyAccountKeys) AccountPublicKey(_ Address, _ int) ([]byte, error) {
	return nil, nil
}

type EmptyCryptoProvider struct{}

var _ CryptoProvider = &EmptyCryptoProvider{}

func (i *EmptyCryptoProvider) VerifySignature(
	_ []byte,
	_ string,
	_ []byte,
	_ []byte,
	_ string,
	_ string,
) (bool, error) {
	return false, nil
}

func (i *EmptyCryptoProvider) Hash(
	_ []byte,
	_ string,
) ([]byte, error) {
	return nil, nil
}

type EmptyProgramCache struct{}

var _ ProgramCache = &EmptyProgramCache{}

func (i *EmptyProgramCache) GetCachedProgram(_ Location) (*ast.Program, error) {
	return nil, nil
}

func (i *EmptyProgramCache) CacheProgram(_ Location, _ *ast.Program) error {
	return nil
}

type EmptyResults struct{}

var _ Results = &EmptyResults{}

func (i *EmptyResults) AppendLog(_ string) error {
	return nil
}

func (i *EmptyResults) Logs() ([]string, error) {
	return nil, nil
}

func (i *EmptyResults) LogAt(_ uint) (string, error) {
	return "", nil
}

func (i *EmptyResults) LogCount() uint {
	return 0
}

func (i *EmptyResults) AppendEvent(_ cadence.Event) error {
	return nil
}

func (i *EmptyResults) Events() ([]cadence.Event, error) {
	return nil, nil
}

func (i *EmptyResults) EventAt(_ uint) (cadence.Event, error) {
	return cadence.Event{}, nil
}

func (i *EmptyResults) EventCount() uint {
	return 0
}

func (i *EmptyResults) AppendError(_ error) error {
	return nil
}

func (i *EmptyResults) Errors() multierror.Error {
	return multierror.Error{}
}

func (i *EmptyResults) ErrorAt(_ uint) (Error, error) {
	return Error{}, nil
}

func (i *EmptyResults) ErrorCount() uint {
	return 0
}

func (i *EmptyResults) AddComputationUsed(_ uint64) error {
	return nil
}

func (i *EmptyResults) ComputationSpent() uint64 {
	return 0
}

func (i *EmptyResults) ComputationLimit() uint64 {
	return 0
}

type EmptyUtils struct{}

var _ Utils = &EmptyUtils{}

func (i *EmptyUtils) GenerateUUID() (uint64, error) {
	return 0, nil
}

// func (i *EmptyRuntimeInterface) GetCurrentBlockHeight() (uint64, error) {
// 	return 0, nil
// }

// func (i *EmptyRuntimeInterface) GetBlockAtHeight(_ uint64) (block Block, exists bool, err error) {
// 	return
// }

// func (i *EmptyRuntimeInterface) UnsafeRandom() (uint64, error) {
// 	return 0, nil
// }

// func (i *EmptyAccounts) ResolveLocation(identifiers []Identifier, location Location) ([]ResolvedLocation, error) {
// 	return []ResolvedLocation{
// 		{
// 			Location:    location,
// 			Identifiers: identifiers,
// 		},
// 	}, nil
// }
