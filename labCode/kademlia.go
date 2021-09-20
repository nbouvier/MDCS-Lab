package main

import (
	"fmt"
	"net"
	"strconv"
	"time"
)

type Kademlia struct {
	routingTable *RoutingTable
	storage      *Storage
}

// NewKademlia returns a new instance of Kademlia
func NewKademlia(ip net.IP, port int) *Kademlia {
	var kademlia Kademlia

	address := ip.String() + ":" + strconv.Itoa(listeningPort)
	kademlia.routingTable = NewRoutingTable(NewContact(NewKademliaID(address), address))
	kademlia.storage = NewStorage()

	return &kademlia
}

func (kademlia *Kademlia) Ping(contact Contact) {

	var network Network
	channel := make(chan bool)
	defer close(channel)

	go network.SendPingMessage(kademlia, &contact, channel)

	select {

	case <-channel:
		fmt.Printf("Ping to %s (%s)succeed.\n", contact.Address, contact.ID.String())
		kademlia.routingTable.AddContact(contact)

	case <-time.After(delayBeforeTimeOut * time.Second):
		fmt.Printf("Ping to %s (%s) timed out.\n", contact.Address, contact.ID.String())

	}

}

func (kademlia *Kademlia) LookupContact(searchedKademliaID *KademliaID) []Contact {

	var network Network
	var closestContacts, contactedContacts, notContactedContacts ContactCandidates
	channel := make(chan *Contact)
	defer close(channel)

	contacts := kademlia.routingTable.FindClosestContacts(searchedKademliaID, bucketSize)
	closestContacts.Append(contacts)
	closestContacts.Sort()
	notContactedContacts.Append(contacts)

	for notContactedContacts.Len() != 0 {
		fmt.Println("-----------------")
		contactsToContact := notContactedContacts.GetContacts(alpha)
		responseWaitingNumber := len(contactsToContact)
		for i := range contactsToContact {
			if !contactsToContact[i].Equals(&kademlia.routingTable.me) {
				go network.SendFindContactMessage(kademlia, searchedKademliaID, &contactsToContact[i], &closestContacts, channel)
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
				kademlia.routingTable.AddContact(*contactedContact)

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

func (kademlia *Kademlia) Store(data string) {

	var network Network
	channel := make(chan bool)
	defer close(channel)

	dataKademliaID := NewKademliaID(data)
	// Only sending to closest contact
	// Can be improved by sending to multiple contacts
	// See Lookup to do so
	contact := kademlia.LookupContact(dataKademliaID)[0]

	go network.SendStoreMessage(kademlia, data, &contact, channel)

	select {

	case <-channel:
		fmt.Printf("Stored \"%s\" (%s) to %s (%s) successfully.\n", data, dataKademliaID, contact.Address, contact.ID.String())
		kademlia.routingTable.AddContact(contact)

	case <-time.After(delayBeforeTimeOut * time.Second):
		fmt.Printf("Failed to store \"%s\" (%s) to %s (%s) (Timeout).\n", data, dataKademliaID, contact.Address, contact.ID.String())

	}

}
