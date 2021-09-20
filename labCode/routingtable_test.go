package main

import (
	"testing"
)

type RoutingTableCase struct {
	routingTable  *RoutingTable
	expectedValue string
}

func TestRoutingTable(t *testing.T) {
	cases := []RoutingTableCase{
		{NewRoutingTable(NewContact(NewKademliaID("172.19.0.3:80"), "172.19.0.3:80")), "routingTable{me=contact{id=3d6064f8a61dda5a835ea34d9a4f3d0e10785fd8,address=172.19.0.3:80},buckets=(160)[hidden]}"},
	}

	for i, c := range cases {
		if c.routingTable.String(false) != c.expectedValue {
			t.Logf("Error in RoutingTable testing for test %d. Got %s instead of %s.", i, c.routingTable.String(false), c.expectedValue)
			t.Fail()
		}
	}
}
