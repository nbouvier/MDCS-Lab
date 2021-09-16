package main

import (
	"fmt"
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

func (kademlia *Kademlia) LookupContact(searchedContact *Contact) {

	var closestContacts, contactedContacts, notContactedContacts ContactCandidates
	contactChannel := make(chan *Contact)
	defer close(contactChannel)

	contacts := kademlia.network.routingTable.FindClosestContacts(searchedContact.ID, bucketSize)
	closestContacts.Append(contacts)
	closestContacts.Sort()
	notContactedContacts.Append(contacts)

	for notContactedContacts.Len() != 0 {
		fmt.Println("-----------------")
		contactsToContact := notContactedContacts.GetContacts(alpha)
		responseWaitingNumber := len(contactsToContact)
		for i := range contactsToContact {
			if !contactsToContact[i].Equals(&kademlia.network.routingTable.me) {
				go kademlia.network.SendFindContactMessage(searchedContact, &contactsToContact[i], &closestContacts, contactChannel)
			} else {
				responseWaitingNumber -= 1
			}
			contactedContacts.AppendOne(contactsToContact[i])
		}

		for i := 0; i < responseWaitingNumber; i++ {
			contactedContact := <-contactChannel
			fmt.Printf("Response from %s\n", contactedContact.ID.String())
			kademlia.network.routingTable.AddContact(*contactedContact)
			// TODO: TTL
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

}

func (kademlia *Kademlia) LookupData(hash string) {
	// TODO
}

func (kademlia *Kademlia) Store(data []byte) {
	// TODO
}
