package main

type pendingTx interface {
	Commit()
	RollBack()
	IsExpired() bool
}
