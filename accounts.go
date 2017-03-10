package main

import (
	"errors"
	"sync"

	"github.com/distributeddesigns/currency"
)

// Account : State of a particular Account
type account struct {
	userID  string
	balance currency.Currency
	sync.Mutex
}

// AddFunds : Increases the balance of the account
func (ac *account) AddFunds(amount currency.Currency) {
	consoleLog.Debugf("Old balance for %s is %s", ac.userID, ac.balance)

	ac.Lock()
	ac.balance.Add(amount)
	ac.Unlock()

	consoleLog.Debugf("New balance for %s is %s", ac.userID, ac.balance)
}

// RemoveFunds : Decrease balance of the account
func (ac *account) RemoveFunds(amount currency.Currency) error {
	ac.Lock()
	err := ac.balance.Sub(amount)
	ac.Unlock()
	if err != nil {
		return errors.New("Insufficient Funds")
	}
	return nil
}
