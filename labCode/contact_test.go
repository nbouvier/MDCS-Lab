package main

import (
	"testing"
)

func TestContact(t *testing.T) {
	contact := NewContact(HexToKademliaID("0000000000000000000000000000000000000001"), "172.19.0.2:80")

	distanceCases := []struct {
		target        *KademliaID
		expectedValue *KademliaID
	}{
		{HexToKademliaID("1000000000000000000000000000000000000000"), HexToKademliaID("1000000000000000000000000000000000000001  ")},
		{HexToKademliaID("0000000000000000000000000000055555555555"), HexToKademliaID("0000000000000000000000000000055555555554 ")},
		{HexToKademliaID("ffffffffffffffffffffff111111111111111111"), HexToKademliaID("ffffffffffffffffffffff111111111111111110 ")},
	}

	for i, c := range distanceCases {
		contact.CalcDistance(c.target)
		if contact.distance.String() != c.expectedValue.String() {
			t.Logf("Error in Contact testing for test Distance%d. Got %s instead of %s.", i, contact.distance, c.expectedValue)
			t.Fail()
		}
	}

	// Other functions are covered in other testing modules

	/*contactTest := NewContact(HexToKademliaID("0000000000000000000000000000055555555555"), "test")

	logicalCases := []struct {
		logical       bool
		expectedValue bool
	}{
		{contact.Less(&contactTest), true},
		{contactTest.Less(&contact), false},
		{contact.Equals(&contact), true},
		{contact.Equals(&contactTest), false},
	}

	for i, c := range logicalCases {
		if c.logical != c.expectedValue {
			t.Logf("Error in Contact testing for test Logical%d.", i)
			t.Fail()
		}
	}*/

}
