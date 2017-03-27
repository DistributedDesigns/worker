package main

import (
	"errors"
)

type txStack []pendingTx

func (s txStack) isEmpty() bool {
	return len(s) == 0
}

func (s *txStack) push(ptx pendingTx) {
	(*s) = append(*s, ptx)
}

func (s *txStack) pop() (pendingTx, error) {
	var ptx pendingTx
	if (*s).isEmpty() {
		return ptx, errors.New("Empty txStack")
	}

	ptx = (*s)[len(*s)-1]
	*s = (*s)[:len(*s)-1]

	return ptx, nil
}

func (s txStack) peek() (pendingTx, error) {
	var ptx pendingTx
	if s.isEmpty() {
		return ptx, errors.New("Empty txStack")
	}

	return s[len(s)-1], nil
}
