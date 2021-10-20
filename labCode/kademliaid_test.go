package main

import (
	"testing"
)

func TestKademliaID(t *testing.T) {
	idCases := []struct {
		kademliaID    *KademliaID
		expectedValue string
	}{
		{NewKademliaID("172.19.0.2:50"), "d0205ebb2d57baa254b82b7e96b45776c14ac13a"},
		{NewKademliaID("172.19.0.2:50"), "d0205ebb2d57baa254b82b7e96b45776c14ac13a"},
		{NewKademliaID("172.19.0.2:51"), "31f7b844fb8d47b3f463ef9dfa470260ede744cc"},
		{NewKademliaID("some data"), "baf34551fecb48acc3da868eb85e1b6dac9de356"},
		{HexToKademliaID("5ea22da908f2395a39e8b01dbdd7d1a67cb4e556"), "5ea22da908f2395a39e8b01dbdd7d1a67cb4e556"},
	}

	for i, c := range idCases {
		if c.kademliaID.String() != c.expectedValue {
			t.Logf("Error in KademliaID testing for test ID%d. Got %s instead of %s.", i, c.kademliaID.String(), c.expectedValue)
			t.Fail()
		}
	}

	for i := 0; i < 10; i++ {
		if NewRandomKademliaID() == NewRandomKademliaID() {
			t.Logf("Error in KademliaID testing for test Random%d. Got twice the same ID in a raw.", i)
			t.Fail()
		}
	}

	logicalCases := []struct {
		logicalResult bool
		expectedValue bool
	}{
		{NewKademliaID("172.19.0.2:50").Equals(NewKademliaID("172.19.0.2:50")), true},
		{NewKademliaID("172.19.0.18:45933").Equals(NewKademliaID("some data")), false},
		{NewKademliaID("172.19.0.2:51").Less(NewKademliaID("172.19.0.2:50")), true},
		{NewKademliaID("172.19.0.3:69523").Less(NewKademliaID("another data")), true},
		{NewKademliaID("another data").Less(NewKademliaID("172.19.0.3:69523")), false},
	}

	for i, c := range logicalCases {
		if c.logicalResult != c.expectedValue {
			t.Logf("Error in KademliaID testing for test Logical%d.", i)
			t.Fail()
		}
	}

	distanceCases := []struct {
		distance      *KademliaID
		expectedValue string
	}{
		{NewKademliaID("172.19.0.2:50").CalcDistance(NewKademliaID("172.19.0.2:51")), "e1d7e6ffd6dafd11a0dbc4e36cf355162cad85f6"},
		{NewKademliaID("172.19.0.2:45302").CalcDistance(NewKademliaID("172.19.0.2:51")), "1ea5b37ec1f4647fa92628e574a62e04525b7463"},
		{NewKademliaID("172.19.0.33:89654").CalcDistance(NewKademliaID("172.19.0.2:51")), "9dd0c952437fead2d1ca0956d27c94f6b4d267ec"},
	}

	for i, c := range distanceCases {
		if c.distance.String() != c.expectedValue {
			t.Logf("Error in KademliaID testing for test Distance%d. Got %s instead of %s.", i, c.distance.String(), c.expectedValue)
			t.Fail()
		}
	}

}
