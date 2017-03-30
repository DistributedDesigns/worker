package main

import (
	"fmt"
	"strconv"
	"time"

	types "github.com/distributeddesigns/shared_types"
)

type commitBuyCmd struct {
	id     uint64
	userID string
}

func parseCommitBuyCmd(parts []string) commitBuyCmd {
	if len(parts) != 3 {
		abortTx("COMMIT_BUY needs 3 parts")
	}

	id, err := strconv.ParseUint(parts[0], 10, 64)
	abortTxOnError(err, "Could not parse ID")

	return commitBuyCmd{
		id:     id,
		userID: parts[2],
	}
}

func (cb commitBuyCmd) Name() string {
	return fmt.Sprintf("[%d] COMMIT_BUY", cb.id)
}

func (cb commitBuyCmd) ToAuditEvent() types.AuditEvent {
	xmlElement := fmt.Sprintf(`
	<userCommand>
		<timestamp>%d</timestamp>
		<server>%s</server>
		<transactionNum>%d</transactionNum>
		<command>COMMIT_BUY</command>
		<username>%s</username>
	</userCommand>`,
		time.Now().UnixNano()/1e6, redisBaseKey, cb.id, cb.userID,
	)

	return types.AuditEvent{
		UserID:    cb.userID,
		ID:        cb.id,
		EventType: "command",
		Content:   xmlElement,
	}
}

func (cb commitBuyCmd) Execute() {
	userAccount := accountStore[cb.userID]
	userAccount.Lock()
	bItem, err := userAccount.pendingBuys.pop()
	userAccount.Unlock()
	if err != nil {
		consoleLog.Debugf("Cannot confirm buy, no pending buys for %s", cb.userID)
		return
	}

	if bItem.isExpired() {
		consoleLog.Noticef("Stock price has expired, unable to confirm; refunding account")
		userAccount.AddFunds(bItem.amount)
		return
	}
	consoleLog.Debugf("%s purchased %s of %s stock, adding to portfolio", cb.userID, bItem.amount, bItem.stock)
	userAccount.stockPortfolio[bItem.stock] += bItem.numStocks
	consoleLog.Notice(" [✔] Finished", cb.Name())
}
