package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {

	var network Network

	ip, port := GetOutboundIP()
	if len(os.Args) > 2 && os.Args[2] == "entry" {
		port = kademliaEntryListeningPort
	}
	kademlia := NewKademlia(ip, port)
	fmt.Printf("IP address is %s\nKademliaID is %s\n", kademlia.routingTable.me.Address, kademlia.routingTable.me.ID)

	go network.Listen(kademlia, port)
	go kademlia.storage.timeToLive()
	if len(os.Args) > 2 && os.Args[1] == "auto" {
		idle()
	} else if len(os.Args) > 1 && os.Args[1] == "auto" {
		autoConnect(kademlia)
		idle()
	} else {
		handleCommandLine(kademlia)
	}

}

func autoConnect(kademlia *Kademlia) {
	var addr *net.IPAddr

	err := errors.New("Looking for kademliaEntry")
	fmt.Println("Looking for kademliaEntry...")
	for err != nil {
		addr, err = net.ResolveIPAddr("ip4:icmp", "kademliaEntry")
	}
	fmt.Printf("KademliaEntry found\n\n.")

	address := addr.IP.String() + ":" + strconv.Itoa(kademliaEntryListeningPort)
	kademliaID := NewKademliaID(address)
	fmt.Printf("Joining the network via %s (%s) ...\n", address, kademliaID)
	kademlia.routingTable.AddContact(NewContact(kademliaID, address))
	kademlia.LookupContact(kademlia.routingTable.me.ID)
	fmt.Printf("Network joined.\n")
}

func idle() {
	// Just looping around to keep the program alive
	for {
		time.Sleep(60 * time.Second)
	}
}

func handleCommandLine(kademlia *Kademlia) {

	for {

		fmt.Print("\n$ ")

		reader := bufio.NewReader(os.Stdin)
		// ReadString will block until the delimiter is entered
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("An error occured while reading input. Please try again", err)
			continue
		}

		// Compute entered command
		inputs := strings.Split(strings.TrimSpace(input), " ")
		switch inputs[0] {

		case "join":
			if len(inputs) < 3 {
				fmt.Println("Error: You need to provide an ip address and a KademliaID.\n     $ join <ip_address> <port>")
				continue
			}
			address := fmt.Sprintf("%s:%s", inputs[1], inputs[2])
			kademliaID := NewKademliaID(address)
			fmt.Printf("Joining the network via %s (%s) ...\n", address, kademliaID)
			kademlia.routingTable.AddContact(NewContact(kademliaID, address))
			kademlia.LookupContact(kademlia.routingTable.me.ID)
			fmt.Printf("Network joined.\n")
			break

		case "lookup":
			if len(inputs) < 2 {
				fmt.Println("Error: You need to provide a KademliaID. \n     $ lookup <ip> <port>")
			}
			address := fmt.Sprintf("%s:%s", inputs[1], inputs[2])
			kademliaID := NewKademliaID(address)
			fmt.Printf("Looking for %s (%s) ...\n", address, kademliaID)
			kademlia.LookupContact(kademliaID)
			break

		case "ping":
			if len(inputs) < 2 {
				fmt.Println("Error: You need to provide an IP address and a KademliaID.\n     $ ping <ip> <port>")
			}
			address := fmt.Sprintf("%s:%s", inputs[1], inputs[2])
			kademliaID := NewKademliaID(address)
			fmt.Printf("Pinging %s (%s) ...\n", address, kademliaID)
			kademlia.Ping(NewContact(kademliaID, address))
			break

		case "put":
			if len(inputs) < 1 {
				fmt.Println("Error: You need to provide some data.\n     $ put <data>")
			}
			data := inputs[1]
			fmt.Printf("Storing \"%s\" (%s) ...\n", data, NewKademliaID(data))
			kademlia.Store(data)
			break

		case "get":
			if len(inputs) < 1 {
				fmt.Println("Error: You need to provide a KademliaID.\n     $ get <kademlia_id>")
			}
			kademliaID := HexToKademliaID(inputs[1])
			fmt.Printf("Looking for \"%s\" ...\n", kademliaID)
			kademlia.LookupData(kademliaID)
			break

		case "show-storage":
			fmt.Println(kademlia.storage)
			break

		case "exit":
			return

		default:
			fmt.Println("Invalid command, please try again.\nValids commands are:\n     $ join <ip> <port>\n     $ lookup <ip> <port>\n     $ ping <ip> <port>\n     $ put <data>\n     $ get <kademlia_id>\n     $ exit")
			break

		}

	}

}
