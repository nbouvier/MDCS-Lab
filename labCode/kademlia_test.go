package main

import (
	"testing"
)

func TestKademlia(t *testing.T) {

	kademlia := NewTestKademlia("ffffffffffffffffffff00000000000000000000")

	kademlia.Ping(NewContact(HexToKademliaID("ffffffffffffffffffff11111111111111111111"), "172.19.0.3:80"))
	kademlia.LookupContact(NewContact(HexToKademliaID("ffffffffffffffffffff11111111111111111111"), "172.19.0.3:80"))
	kademlia.LookupData(NewKademliaID("some data"))
	kademlia.Store("some data")
	kademlia.Refresh()

}
