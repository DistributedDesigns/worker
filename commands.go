package main

import (
	"fmt"
	"strconv"
	"strings"
)

type command interface {
	Execute()
	Name() string
	ToAuditEntry() string
}

func parseCommand(s string) command {
	// Convert to a proper .csv, then parse
	// change `[100] STUFF,...` -> `100,STUFF,...`
	csv := strings.Replace(s, "[", "", 1)
	csv = strings.Replace(csv, "] ", ",", 1)
	csv = strings.TrimSpace(csv)

	parts := strings.Split(csv, ",")

	if len(parts) < 3 {
		abortTx("Insufficient command arguments")
	}

	cmdType := parts[1]
	var parsedCommand command

	switch cmdType {
	case "ADD":
	case "QUOTE":
		parsedCommand = parseQuoteCmd(parts)
	case "BUY":
	case "COMMIT_BUY":
	case "CANCEL_BUY":
	case "SELL":
	case "COMMIT_SELL":
	case "CANCEL_SELL":
	case "SET_BUY_AMOUNT":
	case "SET_BUY_TRIGGER":
	case "CANCEL_SET_BUY":
	case "SET_SELL_AMOUNT":
	case "SET_SELL_TRIGGER":
	case "CANCEL_SET_SELL":
	case "DISPLAY_SUMMARY":
	case "DUMPLOG":
	default:
		fmt.Printf("%+v", parts)
		abortTx(fmt.Sprint("Unrecognized command:", cmdType))
	}

	return parsedCommand
}

func parseQuoteCmd(parts []string) quoteCmd {
	if len(parts) != 4 {
		abortTx("QUOTE needs 4 parts")
	}

	id, err := strconv.ParseUint(parts[0], 10, 64)
	abortTxOnError(err, "Could not parse ID")

	return quoteCmd{
		id:     id,
		userID: parts[2],
		stock:  parts[3],
	}
}
