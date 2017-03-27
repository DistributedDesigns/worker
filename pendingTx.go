package main

type pendingTx interface {
	Commit()
	RollBack()
	IsValid() bool
}
