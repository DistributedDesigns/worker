package main

import (
	"time"

	"errors"

	"github.com/distributeddesigns/currency"
)

type buyItem struct {
	amount         currency.Currency
	numStocks      uint
	price          currency.Currency
	stock          string
	quoteTimeStamp time.Time
}

type buyStack []buyItem

func (s buyStack) isEmpty() bool {
	return len(s) == 0
}

func (bItem buyItem) isExpired() bool {
	return bItem.quoteTimeStamp.Before(time.Now().Add(time.Second * -60))
}

func (s *buyStack) push(element buyItem) {
	(*s) = append(*s, element)
}

func (s *buyStack) pop() (buyItem, error) {
	var val buyItem
	if (*s).isEmpty() {
		return val, errors.New("Empty Stack")
	}
	val = (*s)[len(*s)-1]
	*s = (*s)[:len(*s)-1]
	return val, nil
}

func (s *buyStack) peek() (buyItem, error) {
	var val buyItem
	if (*s).isEmpty() {
		return val, errors.New("Empty Stack")
	}
	return (*s)[len((*s))-1], nil
}

func (s *buyStack) dequeue() (buyItem, error) {
	var val buyItem
	if (*s).isEmpty() {
		return val, errors.New("Empty Stack")
	}
	val = (*s)[0]
	*s = (*s)[1:]
	return val, nil
}

func (s *buyStack) headPeek() (buyItem, error) {
	var val buyItem
	if (*s).isEmpty() {
		return val, errors.New("Empty Stack")
	}
	return (*s)[0], nil
}
