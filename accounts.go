package main

import (
	"sync"

	"time"

	"github.com/distributeddesigns/currency"
)

type portfolio map[string]uint

// Account : State of a particular Account
type account struct {
	userID  string
	balance currency.Currency
	sync.Mutex
	pendingBuys    buyStack
	stockPortfolio portfolio
}

func cleanBuyStacks() {
	consoleLog.Debug("Buy Cleanup is Running")
	for 0 < 1 {
		for _, v := range accountStore {
			v.Lock()
			oldestBuy, err := v.pendingBuys.headPeek()
			v.Unlock()
			if err != nil {
				continue
			}
			if oldestBuy.isExpired() {
				consoleLog.Debugf("Pending buy for %s expired, auto cancelling", v.userID)
				v.Lock()
				v.pendingBuys.dequeue()
				v.Unlock()
				v.AddFunds(oldestBuy.amount)
			}
		}
		time.Sleep(time.Second * 30)
		consoleLog.Debug("Buy cleanup completed loop")
	}
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
