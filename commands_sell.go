package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/distributeddesigns/currency"
	types "github.com/distributeddesigns/shared_types"
)

type sellCmd struct {
	id             uint64
	userID         string
	stock          string
	amount         currency.Currency
	profit         currency.Currency
	quantityToSell uint64
	quoteTimestamp time.Time
}

func parseSellCmd(parts []string) sellCmd {
	if len(parts) != 5 {
		abortTx("SELL needs 5 parts")
	}

	id, err := strconv.ParseUint(parts[0], 10, 64)
	abortTxOnError(err, "Could not parse ID")

	amount, err := currency.NewFromString(parts[4])
	abortTxOnError(err, "Could not parse amount in transaction")

	return sellCmd{
		id:     id,
		userID: parts[2],
		stock:  parts[3],
		amount: amount,
	}
}

func (s sellCmd) Name() string {
	return fmt.Sprintf("[%d] SELL", s.id)
}

func (s sellCmd) ToAuditEvent() types.AuditEvent {
	xmlElement := fmt.Sprintf(`
	<userCommand>
		<timestamp>%d</timestamp>
		<server>%s</server>
		<transactionNum>%d</transactionNum>
		<command>SELL</command>
		<username>%s</username>
		<stockSymbol>%s</stockSymbol>
		<funds>%.02f</funds>
	</userCommand>`,
		time.Now().UnixNano()/1e6, redisBaseKey, s.id, s.userID,
		s.stock, s.amount.ToFloat(),
	)

	return types.AuditEvent{
		UserID:    s.userID,
		ID:        s.id,
		EventType: "command",
		Content:   xmlElement,
	}
}

func (s sellCmd) Execute() {
	abortTxIfNoAccount(s.userID)

	// Check if user has any stock, abort early
	acct := accountStore[s.userID]
	stockHoldings, found := acct.portfolio[s.stock]
	if !found || stockHoldings == 0 {
		abortTx("User does not have any stock to sell")
	}

	// Get a quote for the stock
	qr := types.QuoteRequest{
		Stock:      s.stock,
		UserID:     s.userID,
		AllowCache: true,
		ID:         s.id,
	}

	q := getQuote(qr)

	// Get a fresh one if about to expire
	quoteTTL := q.Timestamp.Add(time.Second*60).Unix() - time.Now().Unix()
	if quoteTTL < config.QuotePolicy.UseInBuySell {
		consoleLog.Info(" [!] Getting a fresh quote for", s.Name())
		qr.AllowCache = false
		q = getQuote(qr)
	}

	// Check if user can sell stock at quote price
	quantityToSell, _ := q.Price.FitsInto(s.amount)
	consoleLog.Debugf("Want to sell %s stock", quantityToSell)
	if quantityToSell < 1 {
		abortTx("Cannot sell less than one stock")
	}

	profit := q.Price
	profit.Mul(float64(quantityToSell))

	// If yes...
	// 1. Populate the profit, quantityToSell, quoteTimestamp fields
	// 2. Remove stock from the user
	// 3. Add the sellCmd to the accounts pendingSells stack

	s.quantityToSell = uint64(quantityToSell)
	s.profit = profit
	s.quoteTimestamp = q.Timestamp

	err := acct.RemoveStock(s.stock, s.quantityToSell)
	abortTxOnError(err, "User does not have enough stock to sell")
	acct.pendingSells.push(s)

	consoleLog.Notice(" [✔] Finished", s.Name())
}

func (s sellCmd) Commit() {
	acct := accountStore[s.userID]
	acct.AddFunds(s.profit)
}

func (s sellCmd) RollBack() {
	acct := accountStore[s.userID]
	acct.AddStock(s.stock, s.quantityToSell)
}

func (s sellCmd) IsExpired() bool {
	expiry := s.quoteTimestamp.Add(time.Second * 60)
	return time.Now().After(expiry)
}
