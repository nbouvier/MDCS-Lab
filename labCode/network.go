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
		n, addr, err := connection.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Printf("\n-> Received from %s: %s\n", addr, string(buffer[0:n]))
		message := strings.Split(string(buffer), " ")

		switch message[0] {

		case "PING":
			kademliaID := NewKademliaID(message[1])

			data := kademlia.routingTable.me.ID.String()
			SendUDPResponse(data, addr, connection)

			kademlia.routingTable.AddContact(NewContact(kademliaID, addr.IP.String()))

			break

		case "STORE":
			break

		case "FIND_NODE":
			kademliaID := NewKademliaID(message[1])
			contacts := kademlia.routingTable.FindClosestContacts(kademliaID, bucketSize)

			data := kademlia.routingTable.me.ID.String()
			for _, contact := range contacts {
				data += " " + contact.ID.String() + " " + contact.Address
			}

			SendUDPResponse(data, addr, connection)

			kademlia.routingTable.AddContact(NewContact(kademliaID, addr.IP.String()))

			break

		case "FIND_VALUE":
			break

		default:
			break

		}
	}

}

// Send stringData as an UDP response to addr
func SendUDPResponse(stringData string, addr *net.UDPAddr, connection *net.UDPConn) {
	data := []byte(stringData)
	fmt.Printf("-> Response to %s: %s\n\n", addr, stringData)
	_, err := connection.WriteToUDP(data, addr)
	if err != nil {
		fmt.Println(err)
	}
}

// Get preferred outbound ip of this machine
func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
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

	replyArgs := strings.Split(reply, " ")[1:]
	for i := 0; i < len(replyArgs); i += 2 {
		newKademliaID := NewKademliaID(replyArgs[i])
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

func (network *Network) SendFindDataMessage(hash string) {
	// TODO
}

func (network *Network) SendStoreMessage(kademlia *Kademlia, data []byte, target *Contact, channel chan bool) {

	message := "STORE " + kademlia.routingTable.me.ID.String() + " " + NewKademliaID(string(data)).String()

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

	ip := contact.Address + ":80"

	s, err := net.ResolveUDPAddr("udp4", ip)
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
