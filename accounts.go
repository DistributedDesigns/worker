package main

import (
	"container/ring"
	"errors"
	"fmt"
	"sync"
	"time"

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
	summary      *ring.Ring
	sync.Mutex
}

const (
	summarySize = 20
)

// newAccountForUser creates an empty account for the given user
func newAccountForUser(userID string) *account {
	var ac = account{userID: userID}
	ac.portfolio = make(portfolio)
	ac.summary = ring.New(summarySize)

	ac.AddSummaryItem("Created")

	return &ac
}

// AddFunds : Increases the balance of the account
func (ac *account) AddFunds(amount currency.Currency) {
	consoleLog.Debugf("Old balance for %s is %s", ac.userID, ac.balance)

	ac.balance.Add(amount)
	ac.AddSummaryItem("Added " + amount.String())

	consoleLog.Debugf("New balance for %s is %s", ac.userID, ac.balance)
}

// RemoveFunds : Decrease balance of the account
func (ac *account) RemoveFunds(amount currency.Currency) error {
	consoleLog.Debugf("Old balance for %s is %s", ac.userID, ac.balance)

	err := ac.balance.Sub(amount)
	if err != nil {
		ac.AddSummaryItem("Removed " + amount.String())
	}

	consoleLog.Debugf("New balance for %s is %s", ac.userID, ac.balance)

	return err
}

// AddStock grants the user the specified amount of stock in their portfolio
func (ac *account) AddStock(stock string, quantity uint64) {
	consoleLog.Debugf("Old portfolio for %s: %d x %s",
		ac.userID, ac.portfolio[stock], stock)

	ac.portfolio[stock] += quantity
	ac.AddSummaryItem(fmt.Sprintf("Added %dx%s", quantity, stock))

	consoleLog.Debugf("New portfolio for %s: %d x %s",
		ac.userID, ac.portfolio[stock], stock)
}

// RemoveStock surrenders stock from the user's portfolio
func (ac *account) RemoveStock(stock string, quantity uint64) error {
	// Check to see if user can surrender that much stock
	if ac.portfolio[stock] < quantity {
		return errors.New("User does not have enough stock to surrender")
	}

	consoleLog.Debugf("Old portfolio for %s: %d x %s",
		ac.userID, ac.portfolio[stock], stock)

	ac.portfolio[stock] -= quantity
	ac.AddSummaryItem(fmt.Sprintf("Removed %dx%s", quantity, stock))

	consoleLog.Debugf("New portfolio for %s: %d x %s",
		ac.userID, ac.portfolio[stock], stock)

	return nil
}

// PruneExpiredTxs will remove all pendingTxs that are expired
func (ac *account) PruneExpiredTxs() {
	ac.Lock()
	ac.AddSummaryItem("Starting expired TX cleanup")
	expiredBuys := ac.pendingBuys.SplitExpired()
	expiredSells := ac.pendingSells.SplitExpired()
	ac.Unlock()

	for _, buy := range *expiredBuys {
		buy.RollBack()
	}

	for _, sell := range *expiredSells {
		sell.RollBack()
	}

	ac.AddSummaryItem("Finished expired TX cleanup")
}

type summaryItem struct {
	loggedAt time.Time
	message  string
}

func (ac *account) AddSummaryItem(s string) {
	// Since ring.Do() always goes _forward_ we want to make sure the forward
	// order of elements is newest -> oldest. This saves us a reverse after
	// we convert the ring to a slice.
	ac.summary = ac.summary.Prev()
	ac.summary.Value = summaryItem{time.Now(), s}
}

// GetSummary returns a list of the user's most recent account activities,
// sorted newest to oldest
func (ac *account) GetSummary() []summaryItem {
	s := make([]summaryItem, 0)

	ac.summary.Do(func(node interface{}) {
		if node != nil {
			s = append(s, node.(summaryItem))
		}
	})

	return s
}
