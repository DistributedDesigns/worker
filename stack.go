package main

import (
	"errors"
)

type txStack []pendingTx

func (s *txStack) IsEmpty() bool {
	return len(*s) == 0
}

func (s *txStack) Push(ptx pendingTx) {
	(*s) = append(*s, ptx)
}

func (s *txStack) Pop() (pendingTx, error) {
	var ptx pendingTx
	if s.IsEmpty() {
		return ptx, errors.New("Empty txStack")
	}

	ptx = (*s)[len(*s)-1]
	*s = (*s)[:len(*s)-1]

	return ptx, nil
}

func (s *txStack) SplitExpired() *txStack {
	var expiredTxs txStack
	if s.IsEmpty() {
		return &expiredTxs
	}

	// Determine position of first expired item, then split stack
	var i int
	for i = 0; i < len(*s); i++ {
		if !(*s)[i].IsExpired() {
			break
		}
	}

	expiredTxs = (*s)[:i]
	*s = (*s)[i:]

	return &expiredTxs
}
