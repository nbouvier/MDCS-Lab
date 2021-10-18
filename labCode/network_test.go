package main

import (
	"testing"
)

func TestNetwork(t *testing.T) {

	ip, port := GetOutboundIP()
	t.Logf("Ip: %s:%d", ip, port)

}
