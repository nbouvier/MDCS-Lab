package main

import (
	"testing"
)

func TestKademlia(t *testing.T) {

	ip, port := GetOutboundIP()
	kademlia := NewKademlia(ip, port)
	t.Log(kademlia)

}
