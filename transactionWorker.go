package main

func transactionWorker(unprocessedTxs <-chan string) {
	for {
		select {
		case <-done:
			consoleLog.Notice(" [x] Finished processing transactions")
			return
		default:
			consoleLog.Info("Processing:", <-unprocessedTxs)
		}
	}
}
