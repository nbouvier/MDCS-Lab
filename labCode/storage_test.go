package main

import (
	"testing"
)

func TestStorage(t *testing.T) {
	storage := NewStorage()
	dataKademliaID := NewKademliaID("data")
	storage.Put("172.18.0.1:80", dataKademliaID, "data")

	data, exists := storage.Get(dataKademliaID)
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
