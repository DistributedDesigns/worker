package main

func txWorker(unprocessedTxs <-chan string) {
	for {
		select {
		case <-done:
			consoleLog.Notice(" [x] Finished processing transactions")
			cleanUpTransactions(unprocessedTxs)
			return
		default:
			// Parse command from tx
			// execute command
			// (go) log command
		}
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
