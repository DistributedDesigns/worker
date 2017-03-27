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
	for {
		for _, account := range accountStore {
			for {
				account.Lock()
				oldestBuy, err := account.pendingBuys.headPeek()
				if err != nil {
					account.Unlock()
					break
				}
				if !oldestBuy.isExpired() {
					account.Unlock()
					break
				}
				consoleLog.Debugf("Pending buy for %s expired, auto cancelling", account.userID)
				account.pendingBuys.dequeue()
				account.Unlock()
				account.AddFunds(oldestBuy.amount)
			}
		}
		time.Sleep(time.Second * time.Duration(config.CleanPolicy.CleanFrequency))
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
