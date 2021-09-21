package main

import (
	"testing"
)

type KademliaIDCase struct {
	kademliaID    *KademliaID
	expectedValue string
}

func TestKademliaID(t *testing.T) {
	cases := []KademliaIDCase{
		{NewKademliaID("172.19.0.2:50"), "d0205ebb2d57baa254b82b7e96b45776c14ac13a"},
		{NewKademliaID("172.19.0.2:50"), "d0205ebb2d57baa254b82b7e96b45776c14ac13a"},
		{NewKademliaID("172.19.0.2:51"), "31f7b844fb8d47b3f463ef9dfa470260ede744cc"},
		{NewKademliaID("some data"), "baf34551fecb48acc3da868eb85e1b6dac9de356"},
		{HexToKademliaID("5ea22da908f2395a39e8b01dbdd7d1a67cb4e556"), "5ea22da908f2395a39e8b01dbdd7d1a67cb4e556"},
	}

	for i, c := range cases {
		if c.kademliaID.String() != c.expectedValue {
			t.Logf("Error in KademliaID testing for test %d. Got %s instead of %s.", i, c.kademliaID.String(), c.expectedValue)
			t.Fail()
		}
	}
}
