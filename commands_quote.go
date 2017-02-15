package main

import "fmt"

type quoteCmd struct {
	id     int
	userID string
	stock  string
}

func (q quoteCmd) Name() string {
	return fmt.Sprintf("[%d] QUOTE", q.id)
}

func (q quoteCmd) ToAuditEntry() string {
	return "<butts></lol>"
}

func (q quoteCmd) Execute() {
	if q.stock == "MHC" {
		abortTx("MHC get out!!")
	}
}
