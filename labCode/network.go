package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/tatsushid/go-fastping"
)

// the static number of simultaneous asynchronous sends
const alpha = 3

type Network struct {
	routingTable *RoutingTable
}

// NewNetwork returns a new instance of a Network
func NewNetwork(me Contact) *Network {
	network := &Network{}
	network.routingTable = NewRoutingTable(me)
	return network
}

func Listen(kademlia *Kademlia /* ip string, */, port int) {

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
		fmt.Printf("\n-> Received from %s: %s\n", addr, string(buffer[0:n]))
		message := strings.Split(string(buffer), " ")

		switch message[0] {

		case "STORE":
			break

		case "FIND_NODE":
			kademliaID := NewKademliaID(message[1])
			contacts := kademlia.network.routingTable.FindClosestContacts(kademliaID, bucketSize)

			contactsData := kademlia.network.routingTable.me.ID.String()
			for _, contact := range contacts {
				contactsData += " " + contact.ID.String() + " " + contact.Address
			}

			data := []byte(contactsData)
			fmt.Printf("-> Response to %s: %s\n\n", addr, string(data))
			_, err = connection.WriteToUDP(data, addr)
			if err != nil {
				fmt.Println(err)
				continue
			}

			kademlia.network.routingTable.AddContact(NewContact(kademliaID, addr.IP.String()))

			break

		case "FIND_VALUE":
			break

		default:
			break

		}
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

func (network *Network) SendPingMessage(contact *Contact) {

	p := fastping.NewPinger()

	ra, err := net.ResolveIPAddr("ip4:icmp", contact.Address)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	p.AddIPAddr(ra)
	p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		fmt.Printf("IP Address: %s receive, RTT: %v\n", addr.String(), rtt)
	}
	p.OnIdle = func() {
		fmt.Print("Pinging done.\n\n")
	}

	err = p.Run()
	if err != nil {
		fmt.Println(err)
	}

}

func (network *Network) SendFindContactMessage(searchedContact *Contact, target *Contact, closestContacts *ContactCandidates, contactChannel chan *Contact) {

	message := "FIND_NODE " + network.routingTable.me.ID.String() + " " + searchedContact.ID.String()

	fmt.Printf("Sending to %s: %s\n", target.ID.String(), message)
	reply := network.SendUDPMessage(target, message)

	replyArgs := strings.Split(reply, " ")[1:]
	for i := 0; i < len(replyArgs); i += 2 {
		newKademliaID := NewKademliaID(replyArgs[i])
		existingContact := closestContacts.Find(newKademliaID)
		if existingContact == nil {
			newContact := NewContact(newKademliaID, replyArgs[i+1])
			newContact.CalcDistance(searchedContact.ID)
			closestContacts.AppendOne(newContact)
		} else {
			existingContact.CalcDistance(searchedContact.ID)
		}
	}

	contactChannel <- target

}

func (network *Network) SendFindDataMessage(hash string) {
	// TODO
}

func (network *Network) SendStoreMessage(data []byte) {
	// TODO
}

func (network *Network) SendUDPMessage(contact *Contact, message string) string {

	ip := contact.Address + ":80"

	s, err := net.ResolveUDPAddr("udp4", ip)
	c, err := net.DialUDP("udp4", nil, s)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	defer c.Close()

	_, err = c.Write([]byte(message))
	if err != nil {
		fmt.Println(err)
		return ""
	}

	buffer := make([]byte, 1024)
	n, _, err := c.ReadFromUDP(buffer)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	return string(buffer[0:n])

}
