package main

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
)

// the static number of simultaneous asynchronous sends
const alpha = 3

// the static number of second before a RPC times out
const delayBeforeTimeOut = 5

// the static number of tries to send an UDP message
// const udpTryNumber = 3

// the static number of millisecond before each UDP tries
// const delayBetweenUDPTries = 500

// the static listening port number of the kademliaEntry node
const kademliaEntryListeningPort = 80

const testPingResponse = "172.18.0.2:80"
const testFindContactResponse = "172.18.0.2:80 172.18.0.3:80"
const testFindDataResponse = "172.18.0.2:80 172.18.0.3:80 172.18.0.4:80"
const testStoreResponse = "172.18.0.2:80"
const testRefreshResponse = ""

type Network struct {
	test bool
}

func NewTestNetwork() *Network {
	var network Network

	network.test = true

	return &network
}

func (network *Network) Listen(kademlia *Kademlia, port int) {

	portStr := ":" + strconv.Itoa(port)

	s, err := net.ResolveUDPAddr("udp4", portStr)
	if err != nil {
		fmt.Println(err)
		return
	}

	connection, err := net.ListenUDP("udp4", s)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer connection.Close()

	buffer := make([]byte, 1024)
	for {
		n, addr, err := connection.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println(err)
		}

		message := strings.Split(string(buffer), " ")
		senderAddress := message[1]
		senderKademliaID := NewKademliaID(senderAddress)
		data := kademlia.routingTable.me.Address
		unknownRPC := false

		if debug {
			fmt.Printf("\nReceived from %s (%s): %s\n", senderAddress, senderKademliaID, buffer[0:n])
		}
		switch message[0] {

		case "PING":
			break

		case "REFRESH":
			kademlia.storage.RefreshData(message[1], message[2])

		case "STORE":
			kademlia.storage.Put(message[1], HexToKademliaID(message[2]), message[3])
			//kademlia.storage.Put(HexToKademliaID(message[2]),message[3])
			break

		case "FIND_NODE":
			lookedKademliaID := NewKademliaID(message[2])
			contacts := kademlia.routingTable.FindClosestContacts(lookedKademliaID, bucketSize)
			for _, contact := range contacts {
				data += " " + contact.Address
			}
			break

		case "FIND_VALUE":
			dataKademliaID := HexToKademliaID(message[2])
			searchedData, exists := kademlia.storage.Get(dataKademliaID)
			if !exists {
				contacts := kademlia.routingTable.FindClosestContacts(dataKademliaID, bucketSize)
				for _, contact := range contacts {
					data += " " + contact.Address
				}
			} else {
				data += " " + searchedData
			}
			break

		default:
			unknownRPC = true
			break

		}

		if !unknownRPC {
			err = SendUDPResponse(data, addr, connection)
			if err == nil {
				kademlia.routingTable.AddContact(NewContact(senderKademliaID, senderAddress))
			} else if debug {
				fmt.Printf("Error while responding to %s (%s).\n%s\n", senderAddress, senderKademliaID, err)
			}
		}
	}

}

// Get preferred outbound ip of this machine
func GetOutboundIP() (net.IP, int) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP, localAddr.Port
}

func (network *Network) SendPingMessage(kademlia *Kademlia, target *Contact, channel chan bool) {

	var err error

	message := "PING " + kademlia.routingTable.me.Address

	if debug {
		fmt.Printf("Sending to %s (%s): %s\n", target.Address, target.ID, message)
	}
	if !network.test {
		_, err = network.SendUDPMessage(target, message)
	} else {
		_, err = network.SimulateUDPMessage(testPingResponse)
	}

	if err != nil {
		if debug {
			fmt.Printf("Error while sending message to %s (%s).\n%s\n", target.Address, target.ID, err)
		}
		channel <- false
		return
	}

	channel <- true

}

func (network *Network) SendFindContactMessage(kademlia *Kademlia, searchContact *Contact, target *Contact, closestContacts *ContactCandidates, channel chan *Contact) {

	var err error
	var reply string

	message := "FIND_NODE " + kademlia.routingTable.me.Address + " " + searchContact.Address

	if debug {
		fmt.Printf("Sending to %s (%s): %s\n", target.Address, target.ID, message)
	}
	if !network.test {
		reply, err = network.SendUDPMessage(target, message)
	} else {
		reply, err = network.SimulateUDPMessage(testFindContactResponse)
	}

	if err != nil {
		if debug {
			fmt.Printf("Error while sending message to %s (%s).\n%s\n", target.Address, target.ID, err)
		}
		channel <- nil
		return
	}

	replyArgs := strings.Split(strings.TrimSpace(reply), " ")[1:]
	for i := 0; i < len(replyArgs); i++ {
		newAddress := replyArgs[i]
		newKademliaID := NewKademliaID(newAddress)
		existingContact := closestContacts.Find(newKademliaID)
		if existingContact == nil {
			newContact := NewContact(newKademliaID, newAddress)
			newContact.CalcDistance(searchContact.ID)
			closestContacts.AppendOne(newContact)
		} else {
			existingContact.CalcDistance(searchContact.ID)
		}
	}

	channel <- target

}

func (network *Network) SendFindDataMessage(kademlia *Kademlia, dataKademliaID *KademliaID, target *Contact, closestContacts *ContactCandidates, channel chan string) {

	var err error
	var reply string

	message := "FIND_VALUE " + kademlia.routingTable.me.Address + " " + dataKademliaID.String()

	if debug {
		fmt.Printf("Sending to %s (%s): %s\n", target.Address, target.ID, message)
	}
	if !network.test {
		reply, err = network.SendUDPMessage(target, message)
	} else {
		reply, err = network.SimulateUDPMessage(testFindDataResponse)
	}

	if err != nil {
		if debug {
			fmt.Printf("Error while sending message to %s (%s).\n%s\n", target.Address, target.ID, err)
		}
		channel <- ""
		return
	}

	replyArgs := strings.Split(strings.TrimSpace(reply), " ")[1:]
	if len(replyArgs) == 1 {
		channel <- target.Address + " " + replyArgs[0]
		return
	}

	for i := 0; i < len(replyArgs); i++ {
		newAddress := replyArgs[i]
		newKademliaID := NewKademliaID(newAddress)
		existingContact := closestContacts.Find(newKademliaID)
		if existingContact == nil {
			newContact := NewContact(newKademliaID, newAddress)
			newContact.CalcDistance(dataKademliaID)
			closestContacts.AppendOne(newContact)
		} else {
			existingContact.CalcDistance(dataKademliaID)
		}
	}

	channel <- target.Address

}

func (network *Network) SendStoreMessage(kademlia *Kademlia, data string, target *Contact, channel chan *Contact) {

	var err error

	contact := *target
	message := "STORE " + kademlia.routingTable.me.Address + " " + NewKademliaID(data).String() + " " + data

	if debug {
		fmt.Printf("Sending to %s (%s): %s\n", target.Address, target.ID, message)
	}
	if !network.test {
		_, err = network.SendUDPMessage(target, message)
	} else {
		_, err = network.SimulateUDPMessage(testStoreResponse)
	}

	if err != nil {
		if debug {
			fmt.Printf("Error while sending message to %s (%s).\n%s\n", target.Address, target.ID, err)
		}
		channel <- nil
		return
	}

	var flag = 0
	for i := 0; i < len(kademlia.contact); i++ {
		if kademlia.contact[i].Equals(&contact) {
			flag = 1
			break
		}
	}
	if flag == 0 {
		kademlia.contact = append(kademlia.contact, contact)
	}
	channel <- &contact

}

func (network *Network) SendRefreshMessage(kademlia *Kademlia, target *Contact, channel chan bool) {

	var err error

	message := "REFRESH " + kademlia.routingTable.me.Address + " " + target.ID.String()

	if debug {
		fmt.Printf("Sending to %s (%s): %s\n", target.Address, target.ID, message)
	}
	if !network.test {
		_, err = network.SendUDPMessage(target, message)
	} else {
		_, err = network.SimulateUDPMessage(testRefreshResponse)
	}

	if err != nil {
		if debug {
			fmt.Printf("Error while sending REFRESH message to %s (%s).\n%s\n", target.Address, target.ID, err)
		}
		channel <- false
		return
	}

	channel <- true

}

func (network *Network) SimulateUDPMessage(response string) (string, error) {

	return response, nil

}

func (network *Network) SendUDPMessage(contact *Contact, message string) (string, error) {

	s, err := net.ResolveUDPAddr("udp4", contact.Address)
	if err != nil {
		return "", err
	}

	c, err := net.DialUDP("udp4", nil, s)
	if err != nil {
		return "", err
	}
	defer c.Close()

	_, err = c.Write([]byte(message))
	if err != nil {
		return "", err
	}

	buffer := make([]byte, 1024)
	n, _, err := c.ReadFromUDP(buffer)
	if err != nil {
		return "", err
	}

	return string(buffer[0:n]), nil

}

// Send stringData as an UDP response to addr
func SendUDPResponse(stringData string, addr *net.UDPAddr, connection *net.UDPConn) error {

	if debug {
		fmt.Printf("Respond to %s: %s\n\n", addr, stringData)
	}
	data := []byte(stringData)
	_, err := connection.WriteToUDP(data, addr)

	return err
}
