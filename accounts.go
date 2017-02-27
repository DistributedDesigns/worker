package main

import (
	// "errors"
	// "time"

	"errors"

	"github.com/distributeddesigns/currency"
	// "github.com/op/go-logging"
)

// account : State of a particular account
type account struct {
	Balance currency.Currency
}

// accounts : Maps name -> account
type accounts map[string]*account

// accountStore : A collection of accouunts
type accountStore struct {
	accounts map[string]*account
}

// NewAccountStore : A constructor that returns an initialized accountStore
func NewAccountStore() *accountStore {
	var as accountStore
	as.accounts = make(accounts)
	return &as
}

// HasAccount : Checks if there's an existing account for the user
func (as accountStore) HasAccount(name string) bool {
	_, ok := as.accounts[name]
	return ok
}

// GetAccount ; Grab an account if it exists for the user
func (as accountStore) GetAccount(name string) *account {
	account, ok := as.accounts[name]
	if !ok {
		return nil
	}
	return account
}

// CreateAccount : Initialize a new account. Fail if one already exists
func (as accountStore) CreateAccount(name string) error {
	// Check for pre-existing accounts
	if as.HasAccount(name) {
		return errors.New("Account already exists")
	}

	// Add account with initial values
	as.accounts[name] = &account{}

	return nil
}

// AddFunds : Increases the balance of the account
func (ac *account) AddFunds(amount currency.Currency) {
	// Only allow > $0.00 to be added
	ac.Balance.Add(amount)
}

// RemoveFunds : Decrease balance of the account
func (ac *account) RemoveFunds(amount currency.Currency) error {
	err := ac.Balance.Sub(amount)
	if err != nil {
		return errors.New("Insufficient Funds")
	}
	return nil
}
