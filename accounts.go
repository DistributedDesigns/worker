package main

import (
	// "errors"
	// "time"

	"errors"

	"github.com/distributeddesigns/currency"
	// "github.com/op/go-logging"
)

// Account : State of a particular Account
type Account struct {
	Balance currency.Currency
}

// Accounts : Maps name -> account
type Accounts map[string]*Account

// AccountStore : A collection of accouunts
type AccountStore struct {
	accounts map[string]*Account
}

// NewAccountStore : A constructor that returns an initialized accountStore
func NewAccountStore() *AccountStore {
	var as AccountStore
	as.accounts = make(Accounts)
	return &as
}

// HasAccount : Checks if there's an existing account for the user
func (as AccountStore) HasAccount(name string) bool {
	_, ok := as.accounts[name]
	return ok
}

// GetAccount ; Grab an account if it exists for the user
func (as AccountStore) GetAccount(name string) *Account {
	account, ok := as.accounts[name]
	if !ok {
		return nil
	}
	return account
}

// CreateAccount : Initialize a new account. Fail if one already exists
func (as AccountStore) CreateAccount(name string) error {
	// Check for pre-existing accounts
	if as.HasAccount(name) {
		return errors.New("Account already exists")
	}

	// Add account with initial values
	as.accounts[name] = &Account{}

	return nil
}

// AddFunds : Increases the balance of the account
func (ac *Account) AddFunds(amount currency.Currency) {
	// Only allow > $0.00 to be added
	ac.Balance.Add(amount)
}

// RemoveFunds : Decrease balance of the account
func (ac *Account) RemoveFunds(amount currency.Currency) error {
	err := ac.Balance.Sub(amount)
	if err != nil {
		return errors.New("Insufficient Funds")
	}
	return nil
}
