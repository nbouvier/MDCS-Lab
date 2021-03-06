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
	contact      []Contact
	refreshTime  int
	test         bool
}

// NewKademlia returns a new instance of Kademlia
func NewKademlia(ip net.IP, port int) *Kademlia {
	var kademlia Kademlia

	address := ip.String() + ":" + strconv.Itoa(port)
	kademlia.routingTable = NewRoutingTable(NewContact(NewKademliaID(address), address))
	kademlia.storage = NewStorage()
	kademlia.contact = []Contact{}
	kademlia.refreshTime = 20
	kademlia.test = false

	return &kademlia
}

func NewTestKademlia(kademliaID string) *Kademlia {
	var kademlia Kademlia

	kademlia.routingTable = NewRoutingTable(NewContact(HexToKademliaID(kademliaID), "172.19.0.2:80"))
	kademlia.storage = NewStorage()
	kademlia.contact = []Contact{}
	kademlia.refreshTime = 20
	kademlia.test = true

	return &kademlia
}

func (kademlia *Kademlia) Ping(contact Contact) {

	var network Network
	channel := make(chan bool)
	defer close(channel)

	if !kademlia.test {
		go network.SendPingMessage(kademlia, &contact, channel)
	} else {
		testNetwork := NewTestNetwork()
		go testNetwork.SendPingMessage(kademlia, &contact, channel)
	}

	select {

	case result := <-channel:
		if result {
			if !kademlia.test {
				fmt.Printf("Ping to %s (%s) succeed.\n", contact.Address, contact.ID)
			}
			kademlia.routingTable.AddContact(contact)
		} else {
			fmt.Printf("Failed to ping %s (%s).\n", contact.Address, contact.ID)
		}
		break

	case <-time.After(delayBeforeTimeOut * time.Second):
		fmt.Printf("Ping to %s (%s) timed out.\n", contact.Address, contact.ID)
	}

}

func (kademlia *Kademlia) Refresh() {

	var network Network
	channel := make(chan bool)
	defer close(channel)

	for {

		if len(kademlia.contact) > 0 && !kademlia.test {
			fmt.Printf("Sending refresh to %d nodes.\n", len(kademlia.contact))
		}

		for i := range kademlia.contact {

			contact := kademlia.contact[i]

			if !kademlia.test {
				go network.SendRefreshMessage(kademlia, &contact, channel)
			} else {
				testNetwork := NewTestNetwork()
				go testNetwork.SendRefreshMessage(kademlia, &contact, channel)
			}

			select {

			case result := <-channel:
				if result {
					if debug {
						fmt.Printf("Refresh to %s (%s) succeed.\n", contact.Address, contact.ID)
					}
					kademlia.routingTable.AddContact(contact)
				} else {
					if debug {
						fmt.Printf("Failed to Refresh %s (%s).\n", contact.Address, contact.ID)
					}
				}
				break

			case <-time.After(delayBeforeTimeOut * time.Second):
				if debug {
					fmt.Printf("Refresh to %s (%s) timed out.\n", contact.Address, contact.ID)
				}
			}
		}

		if kademlia.test {
			break
		}

		time.Sleep(time.Duration(kademlia.refreshTime) * time.Second)

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
				if !kademlia.test {
					go network.SendFindContactMessage(kademlia, &searchedContact, &contactsToContact[i], &closestContacts, channel)
				} else {
					testNetwork := NewTestNetwork()
					go testNetwork.SendFindContactMessage(kademlia, &searchedContact, &contactsToContact[i], &closestContacts, channel)
				}
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
				fmt.Println("Timeout.")
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

	data := ""
	dataFound := false
	for notContactedContacts.Len() != 0 {
		contactsToContact := notContactedContacts.GetContacts(alpha)
		responseWaitingNumber := len(contactsToContact)
		for i := range contactsToContact {
			if !kademlia.test {
				go network.SendFindDataMessage(kademlia, dataKademliaID, &contactsToContact[i], &closestContacts, channel)
			} else {
				testNetwork := NewTestNetwork()
				go testNetwork.SendFindDataMessage(kademlia, dataKademliaID, &contactsToContact[i], &closestContacts, channel)
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
					// Still waiting for other contacted nodes to respond
					if len(responseArgs) == 2 /*&& NewKademliaID(responseArgs[1]).Equals(dataKademliaID)*/ {
						data = responseArgs[1]
						dataFound = true
					}
				}
				break

			case <-time.After(delayBeforeTimeOut * time.Second):
				break

			}
		}

		if dataFound {
			return data, nil
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
	channel := make(chan *Contact)
	defer close(channel)

	dataKademliaID := NewKademliaID(data)
	// contacts := kademlia.routingTable.FindClosestContacts(dataKademliaID, bucketSize)
	contacts := kademlia.LookupContact(NewContact(NewKademliaID(data), data))
	responseWaitingNumber := len(contacts)

	for i := range contacts {

		if !kademlia.test {
			go network.SendStoreMessage(kademlia, data, &contacts[i], channel)
		} else {
			testNetwork := NewTestNetwork()
			go testNetwork.SendStoreMessage(kademlia, data, &contacts[i], channel)
		}

	}

	for i := 0; i < responseWaitingNumber; i++ {
		select {

		case contacted := <-channel:
			if contacted != nil {
				if !kademlia.test {
					fmt.Printf("Stored \"%s\" (%s) to %s (%s) successfully.\n", data, dataKademliaID, contacted.Address, contacted.ID)
				}
				kademlia.routingTable.AddContact(*contacted)
			} else {
				fmt.Printf("Failed to store \"%s\" (%s) to %s (%s).\n", data, dataKademliaID, contacted.Address, contacted.ID)
			}
			break

		case <-time.After(delayBeforeTimeOut * time.Second):
			fmt.Println("Timeout.")
			break

		}
	}
}
