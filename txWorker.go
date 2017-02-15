package main

func txWorker(unprocessedTxs <-chan string) {
	for {
		select {
		case <-done:
			consoleLog.Notice(" [x] Finished processing transactions")
			cleanUpTransactions(unprocessedTxs)
			return
		default:
			processTxs(unprocessedTxs)
		}
	}
}

func processTxs(unprocessedTxs <-chan string) {
	// If a TX is aborted anwhere in its processing it will bubble up
	// to here and catchAbortedTx() will run before processTxs() closes.
	// Control returns to txWorker() where the select{} repeats and the
	// next transaction is grabbed.
	defer catchAbortedTx()
	cmd := parseCommand(<-unprocessedTxs)
	cmd.Execute()
	// (go) log command
}

func abortTx(msg string) {
	panic(msg)
}

func abortTxOnError(err error, msg string) {
	if err != nil {
		panic(msg)
	}
}

func catchAbortedTx() {
	if r := recover(); r != nil {
		consoleLog.Error("Aborted transaction:", r)
	}
}

func cleanUpTransactions(unprocessedTxs <-chan string) {
	// TODO: Put these back in redis? Just warn for now.
	if len(unprocessedTxs) > 0 {
		for i := 0; i < len(unprocessedTxs); i++ {
			consoleLog.Warning("Unprocessed transaction", <-unprocessedTxs)
		}
	}
}
