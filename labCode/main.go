package main

import (
	"bufio"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

const debug = false

func main() {

	var network Network

	ip, port := GetOutboundIP()
	if len(os.Args) > 1 && os.Args[1] == "entry" {
		port = kademliaEntryListeningPort
	}
	kademlia := NewKademlia(ip, port)
	fmt.Printf("IP address is %s\nKademliaID is %s\n", kademlia.routingTable.me.Address, kademlia.routingTable.me.ID)

	go network.Listen(kademlia, port)
	go kademlia.storage.TimeToLive()
	go kademlia.Refresh()

	if len(os.Args) > 1 && os.Args[1] == "entry" {
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
	for err != nil {
		addr, err = net.ResolveIPAddr("ip4:icmp", "kademliaEntry")
	}

	rand.Seed(time.Now().UnixNano())
	time.Sleep(time.Second*time.Duration(rand.Intn(60)) + time.Second*time.Duration(10))

	address := addr.IP.String() + ":" + strconv.Itoa(kademliaEntryListeningPort)
	kademliaID := NewKademliaID(address)
	fmt.Printf("Joining the network via %s (%s) ...\n", address, kademliaID)
	kademlia.routingTable.AddContact(NewContact(kademliaID, address))
	kademlia.LookupContact(kademlia.routingTable.me)
	fmt.Printf("Network joined.\n")
}

func idle() {
	// Just looping around to keep the program alive
	for {
		time.Sleep(30 * time.Second)
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
			kademlia.LookupContact(kademlia.routingTable.me)
			fmt.Printf("Network joined.\n")
			break

		case "lookup":
			if len(inputs) < 3 {
				fmt.Println("Error: You need to provide a KademliaID. \n     $ lookup <ip> <port>")
			}
			address := fmt.Sprintf("%s:%s", inputs[1], inputs[2])
			kademliaID := NewKademliaID(address)
			fmt.Printf("Looking for %s (%s) ...\n", address, kademliaID)
			contacts := kademlia.LookupContact(NewContact(kademliaID, address))
			stringifiedContacts := ""
			for _, contact := range contacts {
				stringifiedContacts += contact.String() + "\n"
			}
			fmt.Printf("Closest contacts for %s (%s): (%d)[\n%s]\n", address, kademliaID, len(contacts), stringifiedContacts)
			break

		case "ping":
			if len(inputs) < 3 {
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
			data, contacts := kademlia.LookupData(kademliaID)
			if contacts != nil {
				stringifiedContacts := ""
				for _, contact := range contacts {
					stringifiedContacts += contact.String() + "\n"
				}
				fmt.Printf("Data not found.\nClosest contacts for %s: (%d)[\n%s]\n", kademliaID, len(contacts), stringifiedContacts)
			} else {
				fmt.Printf("Data found: %s.\n", data)
			}
			break

		case "show-routing-table":
			if len(inputs) < 1 {
				fmt.Println("Error: You need to provide a boolean for hidding buckets.\n     $ show-routing-table <buckets_hidding>")
			}
			bucketsHidding, err := strconv.ParseBool(string(input[1]))
			if err == nil {
				fmt.Println(kademlia.routingTable.String(bucketsHidding))
			} else {
				fmt.Printf("%s found, boolean expected.\n", string(input[1]))
			}
			break

		case "show-storage":
			fmt.Println(kademlia.storage)
			break

		case "set-refresh-time":
			if len(inputs) < 1 {
				fmt.Println("Error: You need to provide a number of seconds.\n     $ set-refresh-time <seconds>")
			}
			kademlia.refreshTime, _ = strconv.Atoi(inputs[1])
			break

		case "exit":
			return

		case "help":
			fmt.Println("     $ join <ip> <port>\n" +
				"	  $ lookup <ip> <port>\n" +
				"     $ ping <ip> <port>\n" +
				"     $ put <data>\n" +
				"     $ get <data_kademlia_id>\n" +
				"     $ show-storage\n" +
				"     $ show-routing-table <buckets_hidding>\n" +
				"     $ exit")
			break

		default:
			fmt.Println("Invalid command, please try again or print \"help\".")
			break

		}

	}

}
