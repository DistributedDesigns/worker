package main

import (
	"container/ring"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/distributeddesigns/currency"
	types "github.com/distributeddesigns/shared_types"
	"github.com/gorilla/websocket"
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

func (ac *account) toCSV() string {
	return fmt.Sprintf("%s,%.2f", ac.userID, ac.balance.ToFloat())
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
	expiredBuys := ac.pendingBuys.SplitExpired()
	expiredSells := ac.pendingSells.SplitExpired()
	ac.Unlock()

	hasExpiredTxs := (!expiredBuys.IsEmpty() || !expiredSells.IsEmpty())

	if hasExpiredTxs {
		ac.AddSummaryItem("Starting expired TX cleanup")

		for _, buy := range *expiredBuys {
			buy.RollBack()
		}

		for _, sell := range *expiredSells {
			sell.RollBack()
		}

		ac.AddSummaryItem("Finished expired TX cleanup")
	}
}

type pendingATXState struct {
	Stock   string `json:"stock"`
	Amount  string `json:"amount"`
	Trigger string `json:"trigger"`
	Action  string `json:"action"`
}

func serializeATX(autoTx types.AutoTxInit) pendingATXState {
	return pendingATXState{
		Stock:   autoTx.AutoTxKey.Stock,
		Action:  autoTx.AutoTxKey.Action,
		Amount:  autoTx.Amount.String(),
		Trigger: autoTx.Trigger.String(),
	}
}

type accountState struct {
	Balance      string            `json:"balance"`
	Portfolio    map[string]uint64 `json:"portfolio"`
	PendingBuys  []pendingTxState  `json:"pendingBuys"`
	PendingSells []pendingTxState  `json:"pendingSells"`
	AutoTx       []pendingATXState `json:"pendingATX"`
}

func (ac *account) GetState() accountState {
	pendingBuys := make([]pendingTxState, len(ac.pendingBuys))
	for i, pb := range ac.pendingBuys {
		pendingBuys[i] = pb.GetState()
	}

	pendingSells := make([]pendingTxState, len(ac.pendingSells))
	for i, ps := range ac.pendingSells {
		pendingSells[i] = ps.GetState()
	}

	pendingATX := make([]pendingATXState, 0)
	for k, v := range workATXStore {
		if k.UserID == ac.userID {
			serTx := serializeATX(v)
			pendingATX = append(pendingATX, serTx)
		}
	}

	return accountState{
		Balance:      ac.balance.String(),
		Portfolio:    ac.portfolio,
		PendingBuys:  pendingBuys,
		PendingSells: pendingSells,
		AutoTx:       pendingATX,
	}
}

type eventMessage struct {
	Account accountState `json:"account"`
	Message string       `json:"message"`
}

func (ac *account) PushEvent(message string) {
	socket, found := userSocketmap[ac.userID]
	if !found {
		consoleLog.Errorf("User %s is not subscribed to a socket connection", ac.userID)
		return
	}

	payload, err := json.Marshal(&eventMessage{ac.GetState(), message})
	if err != nil {
		consoleLog.Errorf("Failed to json-ify %+v, %s", ac, message)
		return
	}

	socket.WriteMessage(websocket.TextMessage, payload)
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

	ac.PushEvent(s)
}

// GetSummary returns a list of the user's most recent account activities,
// sorted newest to oldest
func (ac *account) GetSummary() []summaryItem {
	s := make([]summaryItem, summarySize)

	// Track non-nil items added
	var i int
	ac.summary.Do(func(node interface{}) {
		if node != nil {
			s[i] = node.(summaryItem)
			i++
		}
	})

	// If we didn't fill the return slice only send back the non-nil items.
	// Complaint: shrinking slices in Go is D:
	if i < summarySize {
		return s[:i]
	}

	return s
}
