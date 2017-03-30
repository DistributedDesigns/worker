package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/distributeddesigns/currency"
	types "github.com/distributeddesigns/shared_types"
)

type buyCmd struct {
	id     uint64
	userID string
	stock  string
	amount currency.Currency
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
	consoleLog.Debug("Starting Buy")
	if _, accountExists := accountStore[b.userID]; !accountExists {
		consoleLog.Errorf("No account for %s exists", b.userID)
		return
	}

	userAccount := accountStore[b.userID]

	if userAccount.balance.ToFloat() < b.amount.ToFloat() {
		consoleLog.Errorf("User %s does not have enough available funds for this purchase", b.userID)
		return
	}

	qr := types.QuoteRequest{
		Stock:      b.stock,
		UserID:     b.userID,
		AllowCache: true,
		ID:         b.id,
	}

	var theQuote types.Quote
	theQuote = getQuote(qr)
	numStocks, spent := theQuote.Price.FitsInto(b.amount)
	if numStocks < 1 {
		abortTx("Cannot buy less than one stock")
	} else {
		consoleLog.Infof("Queuing buy %s of %s for %s", spent, b.stock, b.userID)
		consoleLog.Debugf("Removing %s from %s", spent, b.userID)
		err := userAccount.RemoveFunds(spent)
		abortTxOnError(err, "Error removing funds for buy, cancelling transaction")
		bi := buyItem{
			amount:         spent,
			numStocks:      numStocks,
			price:          theQuote.Price,
			stock:          b.stock,
			quoteTimeStamp: theQuote.Timestamp,
		}
		userAccount.Lock()
		userAccount.pendingBuys.push(bi)
		userAccount.Unlock()
	}

	consoleLog.Notice(" [✔] Finished", b.Name())
}
