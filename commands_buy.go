package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/distributeddesigns/currency"
	types "github.com/distributeddesigns/shared_types"
)

type buyCmd struct {
	id             uint64
	userID         string
	stock          string
	amount         currency.Currency
	purchaseAmount currency.Currency
	quoteTimestamp time.Time
	quantityToBuy  uint64
}

func parseBuyCmd(parts []string) buyCmd {
	if len(parts) != 5 {
		abortTx("BUY needs 5 parts")
	}

	id, err := strconv.ParseUint(parts[0], 10, 64)
	abortTxOnError(err, "Could not parse ID")

	amount, err := currency.NewFromString(parts[4])
	abortTxOnError(err, "Could not parse amount in transaction")

	return buyCmd{
		id:     id,
		userID: parts[2],
		stock:  parts[3],
		amount: amount,
	}
}

func (b buyCmd) Name() string {
	return fmt.Sprintf("[%d] BUY", b.id)
}

func (b buyCmd) ToAuditEvent() types.AuditEvent {
	xmlElement := fmt.Sprintf(`
	<userCommand>
		<timestamp>%d</timestamp>
		<server>%s</server>
		<transactionNum>%d</transactionNum>
		<command>BUY</command>
		<username>%s</username>
		<stockSymbol>%s</stockSymbol>
		<funds>%.02f</funds>
	</userCommand>`,
		time.Now().UnixNano()/1e6, redisBaseKey, b.id, b.userID, b.stock, b.amount.ToFloat(),
	)

	return types.AuditEvent{
		UserID:    b.userID,
		ID:        b.id,
		EventType: "command",
		Content:   xmlElement,
	}
}

func (b buyCmd) Execute() {
	abortTxIfNoAccount(b.userID)

	// Get a quote for the stock
	qr := types.QuoteRequest{
		Stock:      b.stock,
		UserID:     b.userID,
		AllowCache: true,
		ID:         b.id,
	}

	q := getQuote(qr)

	// Get a fresh quote if quote is about to expire
	quoteTTL := q.Timestamp.Add(time.Second*60).Unix() - time.Now().Unix()
	if quoteTTL < config.QuotePolicy.UseInBuySell {
		consoleLog.Info(" [!] Getting a fresh quote for", b.Name())
		qr.AllowCache = false
		q = getQuote(qr)
	}

	// Check if user can buy any stock at quote price
	quantityToBuy, remainder := q.Price.FitsInto(b.amount)
	consoleLog.Debugf("Want to buy %d stock", quantityToBuy)
	if quantityToBuy < 1 {
		abortTx("Cannot buy less than one stock")
	}

	purchaseAmount := b.amount
	err := purchaseAmount.Sub(remainder)
	abortTxOnError(err, "This should be impossible!")

	// If yes...
	// 1. Populate the quantityToBuy, purchaseAmount and quoteTimestamp fields
	// 2. Remove the funds from the user
	// 3. Add the buyCmd to the account's pendingBuys stack

	b.quantityToBuy = uint64(quantityToBuy)
	b.purchaseAmount = purchaseAmount
	b.quoteTimestamp = q.Timestamp

	acct := accountStore[b.userID]
	err = acct.RemoveFunds(purchaseAmount)
	abortTxOnError(err, "User does not have enough funds to purchase stock")
	acct.pendingBuys.push(b)

	consoleLog.Notice(" [✔] Finished", b.Name())
}

func (b buyCmd) Commit() {
	acct := accountStore[b.userID]
	acct.AddStock(b.stock, b.quantityToBuy)
}

func (b buyCmd) RollBack() {
	acct := accountStore[b.userID]
	acct.AddFunds(b.purchaseAmount)
}

func (b buyCmd) IsExpired() bool {
	expiry := b.quoteTimestamp.Add(time.Second * 60)
	return time.Now().After(expiry)
}
