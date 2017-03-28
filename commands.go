package main

import (
	"fmt"
	"strings"

	types "github.com/distributeddesigns/shared_types"
)

type command interface {
	Execute()
	Name() string
	ToAuditEvent() types.AuditEvent
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
		parsedCommand = parseAddCmd(parts)
	case "QUOTE":
		parsedCommand = parseQuoteCmd(parts)
	case "BUY":
		parsedCommand = parseBuyCmd(parts)
	case "COMMIT_BUY":
		parsedCommand = parseCommitBuyCmd(parts)
	case "CANCEL_BUY":
		parsedCommand = parseCancelBuyCmd(parts)
	case "SELL":
		parsedCommand = parseSellCmd(parts)
	case "COMMIT_SELL":
		parsedCommand = parseCommitSellCmd(parts)
	case "CANCEL_SELL":
		parsedCommand = parseCancelSellCmd(parts)
	case "SET_BUY_AMOUNT":
		parsedCommand = parseSetBuyAmountCmd(parts)
	case "SET_BUY_TRIGGER":
		parsedCommand = parseSetBuyTriggerCmd(parts)
	case "CANCEL_SET_BUY":
		parsedCommand = parseCancelSetBuyCmd(parts)
	case "SET_SELL_AMOUNT":
		parsedCommand = parseSetSellAmountCmd(parts)
	case "SET_SELL_TRIGGER":
		parsedCommand = parseSetSellTriggerCmd(parts)
	case "CANCEL_SET_SELL":
		parsedCommand = parseCancelSetSellCmd(parts)
	case "DISPLAY_SUMMARY":
		parsedCommand = parseDisplaySummaryCmd(parts)
	case "DUMPLOG":
		parsedCommand = parseDumplogCmd(parts)
	default:
		abortTx(fmt.Sprintf("Unrecognized command %s: %+v", cmdType, parts))
	}

	return parsedCommand
}
