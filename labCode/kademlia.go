package main

type Kademlia struct {
	network *Network
}

// NewKademlia returns a new instance of Kademlia
func NewKademlia(address string) *Kademlia {
	kademlia := &Kademlia{}
	kademlia.network = NewNetwork(NewContact(NewRandomKademliaID(), address))
	return kademlia
}

func (kademlia *Kademlia) LookupContact(target *KademliaID /* target *Contact */) []Contact {
	return kademlia.network.routingTable.FindClosestContacts(target, bucketSize)
}

func (kademlia *Kademlia) LookupData(hash string) {
	// TODO
}

func (kademlia *Kademlia) Store(data []byte) {
	// TODO
}
