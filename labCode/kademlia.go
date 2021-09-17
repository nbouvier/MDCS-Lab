package main

import (
	"fmt"
	"time"
)

type Kademlia struct {
	network *Network
}

// NewKademlia returns a new instance of Kademlia
func NewKademlia(address string) *Kademlia {
	kademlia := &Kademlia{}
	kademlia.network = NewNetwork(NewContact(NewRandomKademliaID(), address))
	return kademlia
}

func (kademlia *Kademlia) Ping(contact Contact) {

	channel := make(chan bool)
	defer close(channel)

	go kademlia.network.SendPingMessage(&contact, channel)

	select {

	case <-channel:
		fmt.Printf("Ping to %s (%s)succeed.\n", contact.Address, contact.ID.String())
		kademlia.network.routingTable.AddContact(contact)

	case <-time.After(delayBeforeTimeOut * time.Second):
		fmt.Printf("Ping to %s (%s) timed out.\n", contact.Address, contact.ID.String())

	}

}

func (kademlia *Kademlia) LookupContact(searchedKademliaID *KademliaID) []Contact {

	var closestContacts, contactedContacts, notContactedContacts ContactCandidates
	channel := make(chan *Contact)
	defer close(channel)

	contacts := kademlia.network.routingTable.FindClosestContacts(searchedKademliaID, bucketSize)
	closestContacts.Append(contacts)
	closestContacts.Sort()
	notContactedContacts.Append(contacts)

	for notContactedContacts.Len() != 0 {
		fmt.Println("-----------------")
		contactsToContact := notContactedContacts.GetContacts(alpha)
		responseWaitingNumber := len(contactsToContact)
		for i := range contactsToContact {
			if !contactsToContact[i].Equals(&kademlia.network.routingTable.me) {
				go kademlia.network.SendFindContactMessage(searchedKademliaID, &contactsToContact[i], &closestContacts, channel)
			} else {
				responseWaitingNumber -= 1
			}
			contactedContacts.AppendOne(contactsToContact[i])
		}

		// This is not totally reliable as if the first "SendFindContactMessage()"
		// times out and finally finished before the n = alpha one, then we will
		// count it twice : 1 for the time out and 1 for the no time out.
		// Doing so, the n = alpha one won't be waited.
		for i := 0; i < responseWaitingNumber; i++ {
			select {

			case contactedContact := <-channel:
				fmt.Printf("Response from %s\n", contactedContact.ID.String())
				kademlia.network.routingTable.AddContact(*contactedContact)

			case <-time.After(delayBeforeTimeOut * time.Second):
				fmt.Println("Timeout.")

			}
		}

		closestContacts.Sort()
		fmt.Printf("\nClosestContacts:")
		notContactedContacts.Empty()
		for _, contact := range closestContacts.GetContacts(bucketSize) {
			fmt.Printf("\nContact: %s", contact.String())
			if contactedContacts.Find(contact.ID) == nil {
				fmt.Printf(" -> Not contacted")
				notContactedContacts.AppendOne(contact)
			}
		}
		fmt.Println()
	}

	fmt.Printf("-----------------\nClosestContacts (%d/%d): %s\n", closestContacts.Len(), bucketSize, closestContacts.String())
	return closestContacts.GetContacts(bucketSize)

}

func (kademlia *Kademlia) LookupData(hash string) {
	// TODO
}

func (kademlia *Kademlia) Store(data []byte) {
	// TODO
}
