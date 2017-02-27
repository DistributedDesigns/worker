package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/distributeddesigns/currency"
)

type addCmd struct {
	txID   uint64
	userID string
	amount currency.Currency
}

func parseAddCmd(parts []string) addCmd {
	if len(parts) != 4 {
		abortTx("ADD needs 4 parts")
	}

	txID, err := strconv.ParseUint(parts[0], 10, 64)
	abortTxOnError(err, "Could not parse ID")

	amount, err := currency.NewFromString(parts[3])
	abortTxOnError(err, "Could not parse amount in transaction")

	return addCmd{
		txID:   txID,
		userID: parts[2],
		amount: amount,
	}
}

func (a addCmd) Name() string {
	return fmt.Sprintf("[%d] ADD", a.txID)
}

func (a addCmd) ToAuditEntry() string {
	return fmt.Sprintf(`
	<userCommand>
		<timestamp>%d</timestamp>
		<server>%s</server>
		<transactionNum>%d</transactionNum>
		<command>ADD</command>
		<username>%s</username>
		<funds>%.02f</funds>
	</userCommand>`,
		time.Now().UnixNano()/1e6, redisBaseKey, a.txID, a.userID, a.amount.ToFloat(),
	)
}

func (a addCmd) Execute() {
	// Create an account if one does not exist
	if !accountStore.HasAccount(a.userID) {
		consoleLog.Noticef("Creating account for %s", a.userID)
		if err := accountStore.CreateAccount(a.userID); err != nil {
			consoleLog.Error(err.Error())
			abortTx("Account creation failed")
		}
	}

	// Add the amounts
	consoleLog.Infof("Adding %s to %s", a.amount, a.userID)
	userAccount := accountStore.GetAccount(a.userID)
	if userAccount == nil {
		consoleLog.Errorf("Internal account creation error for %s", a.userID)
		abortTx("Internal account creation error")
	}

	consoleLog.Debugf("Old balance for %s is %s", a.userID, userAccount.Balance)

	userAccount.AddFunds(a.amount)

	consoleLog.Infof("New balance for %s is %s", a.userID, userAccount.Balance)
}
