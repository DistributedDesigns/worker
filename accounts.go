package main

import (
	"errors"
	"sync"

	"github.com/distributeddesigns/currency"
)

type portfolio map[string]uint64

// Account : State of a particular Account
type account struct {
	userID       string
	balance      currency.Currency
	portfolio    portfolio
	pendingBuys  txStack
	pendingSells txStack
	sync.Mutex
}

// newAccountForUser creates an empty account for the given user
func newAccountForUser(userID string) *account {
	var ac = account{userID: userID}
	ac.portfolio = make(portfolio)
	return &ac
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
	consoleLog.Debugf("Old balance for %s is %s", ac.userID, ac.balance)

	ac.Lock()
	err := ac.balance.Sub(amount)
	ac.Unlock()

	consoleLog.Debugf("New balance for %s is %s", ac.userID, ac.balance)

	return err
}

// AddStock grants the user the specified amount of stock in their portfolio
func (ac *account) AddStock(stock string, quantity uint64) {
	consoleLog.Debugf("Old portfolio for %s: %d x %s",
		ac.userID, ac.portfolio[stock], stock)

	ac.Lock()
	ac.portfolio[stock] += quantity
	ac.Unlock()

	consoleLog.Debugf("New portfolio for %s: %d x %s",
		ac.userID, ac.portfolio[stock], stock)
}

// RemoveStock surrenders stock from the user's portfolio
func (ac *account) RemoveStock(stock string, quantity uint64) error {
	// Lock the account now so next check can be considered valid
	// for length of function
	ac.Lock()
	defer ac.Unlock()

	// Check to see if user can surrender that much stock
	if ac.portfolio[stock] < quantity {
		return errors.New("User does not have enough stock to surrender")
	}

	consoleLog.Debugf("Old portfolio for %s: %d x %s",
		ac.userID, ac.portfolio[stock], stock)

	ac.portfolio[stock] -= quantity

	consoleLog.Debugf("New portfolio for %s: %d x %s",
		ac.userID, ac.portfolio[stock], stock)

	return nil
}

// PruneExpiredTxs will remove all pendingTxs that are expired
func (ac *account) PruneExpiredTxs() {
	ac.Lock()
	expiredBuys := ac.pendingBuys.SplitExpired()
	expiredSells := ac.pendingSells.SplitExpired()
	ac.Unlock()

	for _, buy := range *expiredBuys {
		buy.RollBack()
	}

	for _, sell := range *expiredSells {
		sell.RollBack()
	}
}
