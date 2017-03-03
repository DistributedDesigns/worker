package main

import (
	"errors"
	"sync"

	"github.com/distributeddesigns/currency"
)

// Account : State of a particular Account
type Account struct {
	Balance      currency.Currency
	BalanceMutex sync.Mutex
}

// Accounts : Maps name -> account
type Accounts map[string]*Account

// AddFunds : Increases the balance of the account
func (ac *Account) AddFunds(amount currency.Currency) {
	// Only allow > $0.00 to be added
	ac.BalanceMutex.Lock()
	ac.Balance.Add(amount)
	ac.BalanceMutex.Unlock()
}

// RemoveFunds : Decrease balance of the account
func (ac *Account) RemoveFunds(amount currency.Currency) error {
	ac.BalanceMutex.Lock()
	err := ac.Balance.Sub(amount)
	ac.BalanceMutex.Unlock()
	if err != nil {
		return errors.New("Insufficient Funds")
	}
	return nil
}
