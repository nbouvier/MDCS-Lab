package main

import (
	"testing"
)

func TestBucket(t *testing.T) {
	bucketCases := []struct {
		bucket        *bucket
		expectedValue int
	}{
		{newBucket(), 0},
		{newBucket(), 2},
		{newBucket(), 5},
	}

	bucketCases[1].bucket.AddContact(NewContact(NewKademliaID("172.19.0.2:80"), "172.19.0.2:80"))
	bucketCases[1].bucket.AddContact(NewContact(NewKademliaID("172.19.0.3:80"), "172.19.0.3:80"))

	bucketCases[2].bucket.AddContact(NewContact(NewKademliaID("172.19.0.2:80"), "172.19.0.2:80"))
	bucketCases[2].bucket.AddContact(NewContact(NewKademliaID("172.19.0.3:80"), "172.19.0.3:80"))
	bucketCases[2].bucket.AddContact(NewContact(NewKademliaID("172.19.0.4:80"), "172.19.0.4:80"))
	bucketCases[2].bucket.AddContact(NewContact(NewKademliaID("172.19.0.5:80"), "172.19.0.5:80"))
	bucketCases[2].bucket.AddContact(NewContact(NewKademliaID("172.19.0.6:80"), "172.19.0.6:80"))
	// Putting back "172.19.0.4:80" to the front of the bucket
	bucketCases[2].bucket.AddContact(NewContact(NewKademliaID("172.19.0.4:80"), "172.19.0.5:80"))

	for i, c := range bucketCases {
		if c.bucket.Len() != c.expectedValue {
			t.Logf("Error in Bucket testing for test Len%d. Got %d instead of %d.", i, c.bucket.Len(), c.expectedValue)
			t.Fail()
		}
	}

	expectedValues := []string{
		"fd65368e7a6600fdd68383392df67c85742bfb7a", // 172.19.0.4:80
		"0000000000000000000000000000000000000000", // 172.19.0.6:80
		"6113158fd48cbdccd6a6e9ef8cc972f216b8c4d8", // 172.19.0.5:80
		"aa944349861c5d1ce48c49cdd6e0a5bdf8190059", // 172.19.0.3:80
		"00f04ad38162a30a33abf0b5fd3cbe2ba02b0645", // 172.19.0.2:80
	}

	for i, c := range bucketCases[2].bucket.GetContactAndCalcDistance(NewKademliaID("172.19.0.6:80")) {
		if c.distance.String() != expectedValues[i] {
			t.Logf("Error in Bucket testing for test Len%d. Got %s instead of %s.", i, c.distance.String(), expectedValues[i])
			t.Fail()
		}
	}

	t.Log(bucketCases[2].bucket.String())

}
