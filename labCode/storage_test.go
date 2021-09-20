package main

import (
	"testing"
)

func TestStorage(t *testing.T) {
	storage := NewStorage()
	storage.Put("a17c9aaa61e80a1bf71d0d850af4e5baa9800bbd", "data")

	if storage.Get("a17c9aaa61e80a1bf71d0d850af4e5baa9800bbd") != "data" {
		t.Logf("Error in Storage testing. Got %s instead of %s.", storage.Get("a17c9aaa61e80a1bf71d0d850af4e5baa9800bbd"), "data")
		t.Fail()
	}
}
