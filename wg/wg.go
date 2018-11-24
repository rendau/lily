package wg

import "sync"

var _wg sync.WaitGroup

func Add() {
	_wg.Add(1)
}

func Done() {
	_wg.Done()
}

func Wait() {
	_wg.Wait()
}
