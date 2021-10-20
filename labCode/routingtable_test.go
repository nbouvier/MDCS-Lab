package main

import (
	"testing"
)

func TestRoutingTable(t *testing.T) {

	routingTable := NewRoutingTable(NewContact(NewKademliaID("172.19.0.2:80"), "172.19.0.2:80"))
	routingTable.AddContact(NewContact(NewKademliaID("172.19.0.2:80"), "172.19.0.3:80"))
	routingTable.AddContact(NewContact(NewKademliaID("172.19.0.4:80"), "172.19.0.4:80"))
	routingTable.AddContact(NewContact(NewKademliaID("172.19.0.5:80"), "172.19.0.5:80"))

	bucketsCases := []struct {
		bucketID      int
		expectedValue int
	}{
		{routingTable.getBucketIndex(HexToKademliaID("d0205ebb2d57baa254b82b7e96b45776c14ac13a")), 1},
		{routingTable.getBucketIndex(NewKademliaID("172.19.0.85:55789")), 0},
		{routingTable.getBucketIndex(NewKademliaID("some data to retrieve")), 1},
		{routingTable.getBucketIndex(HexToKademliaID("baf34551fecb48acc3da868eb85e1b6dac9de356")), 2},
	}

	for i, c := range bucketsCases {
		if c.bucketID != c.expectedValue {
			t.Logf("Error in RoutingTable testing for test Bucket%d. Got %d instead of %d.", i, c.bucketID, c.expectedValue)
			t.Fail()
		}
	}

	closestCases := []struct {
		closestContacts []Contact
		index           int
		expectedValue   string
	}{
		{routingTable.FindClosestContacts(NewKademliaID("172.19.0.2:80"), bucketSize), 0, "contact{id=97046d62a163244c54791a35b1932698484a59c4,address=172.19.0.2:80}"},
		{routingTable.FindClosestContacts(NewKademliaID("172.19.0.85:55789"), bucketSize), 1, "contact{id=97046d62a163244c54791a35b1932698484a59c4,address=172.19.0.2:80}"},
		{routingTable.FindClosestContacts(NewKademliaID("some data to retrieve"), bucketSize), 0, "contact{id=f6e7323ef48d3a8ab174036fc066ea41fed99b59,address=172.19.0.5:80}"},
		{routingTable.FindClosestContacts(HexToKademliaID("baf34551fecb48acc3da868eb85e1b6dac9de356"), bucketSize), 2, "contact{id=6a91113f5a6787bbb15169b96159e4369c4aa4fb,address=172.19.0.4:80}"},
	}

	for i, c := range closestCases {
		if c.closestContacts[c.index].String() != c.expectedValue {
			t.Logf("Error in RoutingTable testing for test Closest%d. Got %s instead of %s.", i, c.closestContacts[c.index].String(), c.expectedValue)
			t.Fail()
		}
	}

	t.Log(routingTable.String(false))

}
