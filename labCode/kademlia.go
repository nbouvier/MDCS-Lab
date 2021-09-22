package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

type Kademlia struct {
	routingTable *RoutingTable
	storage      *Storage
}

// NewKademlia returns a new instance of Kademlia
func NewKademlia(ip net.IP, port int) *Kademlia {
	var kademlia Kademlia

	address := ip.String() + ":" + strconv.Itoa(port)
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

	case result := <-channel:
		if result {
			fmt.Printf("Ping to %s (%s) succeed.\n", contact.Address, contact.ID)
			kademlia.routingTable.AddContact(contact)
		} else {
			fmt.Printf("Failed to ping %s (%s).\n", contact.Address, contact.ID)
		}
		break

	case <-time.After(delayBeforeTimeOut * time.Second):
		fmt.Printf("Ping to %s (%s) timed out.\n", contact.Address, contact.ID)
	}

}

func (kademlia *Kademlia) LookupContact(searchedContact Contact) []Contact {

	var network Network
	var closestContacts, contactedContacts, notContactedContacts ContactCandidates
	channel := make(chan *Contact)
	defer close(channel)

	contacts := kademlia.routingTable.FindClosestContacts(searchedContact.ID, bucketSize)
	closestContacts.Append(contacts)
	closestContacts.Sort()
	notContactedContacts.Append(contacts)

	for notContactedContacts.Len() != 0 {
		contactsToContact := notContactedContacts.GetContacts(alpha)
		responseWaitingNumber := len(contactsToContact)
		for i := range contactsToContact {
			if !contactsToContact[i].Equals(&kademlia.routingTable.me) {
				go network.SendFindContactMessage(kademlia, &searchedContact, &contactsToContact[i], &closestContacts, channel)
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
				if contactedContact != nil {
					kademlia.routingTable.AddContact(*contactedContact)
				}
				break

			case <-time.After(delayBeforeTimeOut * time.Second):
				break

			}
		}

		closestContacts.Sort()

		notContactedContacts.Empty()
		for _, contact := range closestContacts.GetContacts(bucketSize) {
			if contactedContacts.Find(contact.ID) == nil {
				notContactedContacts.AppendOne(contact)
			}
		}
	}

	return closestContacts.GetContacts(bucketSize)

}

func (kademlia *Kademlia) LookupData(dataKademliaID *KademliaID) (string, []Contact) {

	var network Network
	var closestContacts, contactedContacts, notContactedContacts ContactCandidates
	channel := make(chan string)
	defer close(channel)

	contacts := kademlia.routingTable.FindClosestContacts(dataKademliaID, bucketSize)
	closestContacts.Append(contacts)
	closestContacts.Sort()
	notContactedContacts.Append(contacts)

	for notContactedContacts.Len() != 0 {
		contactsToContact := notContactedContacts.GetContacts(alpha)
		responseWaitingNumber := len(contactsToContact)
		for i := range contactsToContact {
			if !contactsToContact[i].Equals(&kademlia.routingTable.me) {
				go network.SendFindDataMessage(kademlia, dataKademliaID, &contactsToContact[i], &closestContacts, channel)
			} else {
				responseWaitingNumber -= 1
			}
			contactedContacts.AppendOne(contactsToContact[i])
		}

		// This is not totally reliable as if the first "SendFindDataMessage()"
		// times out and finally finished before the n = alpha one, then we will
		// count it twice : 1 for the time out and 1 for the no time out.
		// Doing so, the n = alpha one won't be waited.
		for i := 0; i < responseWaitingNumber; i++ {
			select {

			case response := <-channel:
				if response != "" {
					responseArgs := strings.Split(strings.TrimSpace(response), " ")
					responseContact := NewContact(NewKademliaID(responseArgs[0]), responseArgs[0])
					kademlia.routingTable.AddContact(responseContact)
					// Check if data was found
					// /!\ Not waiting for other nodes to responde. Should we ?
					if len(responseArgs) == 2 /*&& NewKademliaID(responseArgs[1]).Equals(dataKademliaID)*/ {
						return responseArgs[1], nil
					}
				}
				break

			case <-time.After(delayBeforeTimeOut * time.Second):
				break

			}
		}

		closestContacts.Sort()

		notContactedContacts.Empty()
		for _, contact := range closestContacts.GetContacts(bucketSize) {
			if contactedContacts.Find(contact.ID) == nil {
				notContactedContacts.AppendOne(contact)
			}
		}
	}

	return "", closestContacts.GetContacts(bucketSize)

}

func (kademlia *Kademlia) Store(data string) {

	var network Network
	channel := make(chan bool)
	defer close(channel)

	dataKademliaID := NewKademliaID(data)
	// Only sending to closest contact
	// Can be improved by sending to multiple contacts
	// See Lookup to do so
	_, contacts := kademlia.LookupData(dataKademliaID)
	contact := contacts[0]

	go network.SendStoreMessage(kademlia, data, &contact, channel)

	select {

	case result := <-channel:
		if result {
			fmt.Printf("Stored \"%s\" (%s) to %s (%s) successfully.\n", data, dataKademliaID, contact.Address, contact.ID)
			kademlia.routingTable.AddContact(contact)
		} else {
			fmt.Printf("Failed to store \"%s\" (%s) to %s (%s).\n", data, dataKademliaID, contact.Address, contact.ID)
		}
		break

	case <-time.After(delayBeforeTimeOut * time.Second):
		fmt.Printf("Failed to store \"%s\" (%s) to %s (%s) (Timeout).\n", data, dataKademliaID, contact.Address, contact.ID)
		break

	}

}
