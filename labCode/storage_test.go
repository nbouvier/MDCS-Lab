package main

import (
	"testing"
)

func TestStorage(t *testing.T) {
	storage := NewStorage()
	kademliaID := NewKademliaID("data")
	storage.Put(kademliaID, "data")

	data, exists := storage.Get(kademliaID)
	if !exists {
		t.Logf("Error in Storage testing. \"%s\" doesn't exists in storage.", data)
		t.Fail()
	} else {
		if data != "data" {
			t.Logf("Error in Storage testing. Got \"%s\" instead of \"%s\".", data, "data")
			t.Fail()
		}
	}
}
