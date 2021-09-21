package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"
)

// the static number of simultaneous asynchronous sends
const alpha = 3

// the static number of second before a RPC times out
const delayBeforeTimeOut = 5

// the static port number for UDP requests
// should not stay static, but it's easier actually
const listeningPort = 80

type Network struct {
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
	rand.Seed(time.Now().Unix())

	for {
		var data string

		n, addr, err := connection.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println(err)
		}

		senderAddress := addr.IP.String() + ":" + strconv.Itoa(listeningPort)
		message := strings.Split(string(buffer), " ")
		senderKademliaID := HexToKademliaID(message[1])
		unknownRPC := false
		fmt.Printf("\n-> Received from %s (%s): %s\n", senderAddress, senderKademliaID, buffer[0:n])
		switch message[0] {

		case "PING":
			data = kademlia.routingTable.me.ID.String()
			break

		case "STORE":
			kademlia.storage.Put(HexToKademliaID(message[2]), message[3])
			data = kademlia.routingTable.me.ID.String()
			break

		case "FIND_NODE":
			lookedKademliaID := HexToKademliaID(message[2])
			contacts := kademlia.routingTable.FindClosestContacts(lookedKademliaID, bucketSize)
			data = kademlia.routingTable.me.ID.String()
			for _, contact := range contacts {
				data += " " + contact.ID.String() + " " + contact.Address
			}
			break

		case "FIND_VALUE":
			dataKademliaID := HexToKademliaID(message[2])
			searchedData, exists := kademlia.storage.Get(dataKademliaID)
			data = kademlia.routingTable.me.ID.String()
			if !exists {
				contacts := kademlia.routingTable.FindClosestContacts(dataKademliaID, bucketSize)
				for _, contact := range contacts {
					data += " " + contact.ID.String() + " " + contact.Address
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
			SendUDPResponse(data, addr, connection)
			kademlia.routingTable.AddContact(NewContact(senderKademliaID, senderAddress))
		}
	}

}

// Send stringData as an UDP response to addr
func SendUDPResponse(stringData string, addr *net.UDPAddr, connection *net.UDPConn) {
	data := []byte(stringData)
	fmt.Printf("-> Respond to %s: %s\n\n", addr, stringData)
	_, err := connection.WriteToUDP(data, addr)
	if err != nil {
		fmt.Println(err)
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

	message := "PING " + kademlia.routingTable.me.ID.String()

	fmt.Printf("Sending to %s: %s\n", target.ID.String(), message)
	_, err := network.SendUDPMessage(target, message)

	if err == nil {
		channel <- true
	} else {
		channel <- false
		fmt.Println(err)
	}

}

func (network *Network) SendFindContactMessage(kademlia *Kademlia, searchedKademliaID *KademliaID, target *Contact, closestContacts *ContactCandidates, channel chan *Contact) {

	message := "FIND_NODE " + kademlia.routingTable.me.ID.String() + " " + searchedKademliaID.String()

	fmt.Printf("Sending to %s: %s\n", target.ID.String(), message)
	reply, err := network.SendUDPMessage(target, message)

	if err != nil {
		fmt.Println(err)
		channel <- nil
		return
	}

	replyArgs := strings.Split(strings.TrimSpace(reply), " ")[1:]
	for i := 0; i < len(replyArgs); i += 2 {
		newKademliaID := HexToKademliaID(replyArgs[i])
		existingContact := closestContacts.Find(newKademliaID)
		if existingContact == nil {
			newContact := NewContact(newKademliaID, replyArgs[i+1])
			newContact.CalcDistance(searchedKademliaID)
			closestContacts.AppendOne(newContact)
		} else {
			existingContact.CalcDistance(searchedKademliaID)
		}
	}

	channel <- target

}

func (network *Network) SendFindDataMessage(kademlia *Kademlia, dataKademliaID *KademliaID, target *Contact, closestContacts *ContactCandidates, channel chan string) {

	message := "FIND_VALUE " + kademlia.routingTable.me.ID.String() + " " + dataKademliaID.String()

	fmt.Printf("Sending to %s: %s\n", target.ID.String(), message)
	reply, err := network.SendUDPMessage(target, message)

	if err != nil {
		fmt.Println(err)
		channel <- ""
		return
	}

	replyArgs := strings.Split(strings.TrimSpace(reply), " ")[1:]
	// Why does NewKademliaID(replyArgs[0]).Equals(dataKademliaID) = false ?!
	// fmt.Printf("%s, %s, %s", replyArgs[0], NewKademliaID(replyArgs[0]), dataKademliaID)
	if len(replyArgs) == 1 /*&& NewKademliaID(replyArgs[0]).Equals(dataKademliaID)*/ {
		channel <- target.Address + " " + replyArgs[0]
		return
	}

	for i := 0; i < len(replyArgs); i += 2 {
		newKademliaID := HexToKademliaID(replyArgs[i])
		existingContact := closestContacts.Find(newKademliaID)
		if existingContact == nil {
			newContact := NewContact(newKademliaID, replyArgs[i+1])
			newContact.CalcDistance(dataKademliaID)
			closestContacts.AppendOne(newContact)
		} else {
			existingContact.CalcDistance(dataKademliaID)
		}
	}

	channel <- target.Address

}

func (network *Network) SendStoreMessage(kademlia *Kademlia, data string, target *Contact, channel chan bool) {

	message := "STORE " + kademlia.routingTable.me.ID.String() + " " + NewKademliaID(data).String() + " " + data

	fmt.Printf("Sending to %s: %s\n", target.ID.String(), message)
	_, err := network.SendUDPMessage(target, message)

	if err == nil {
		channel <- true
	} else {
		channel <- false
		fmt.Println(err)
	}

}

func (network *Network) SendUDPMessage(contact *Contact, message string) (string, error) {

	s, err := net.ResolveUDPAddr("udp4", contact.Address)
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
