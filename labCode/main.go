package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {

	ip := GetOutboundIP()
	kademlia := NewKademlia(ip.String())
	fmt.Printf("IP address is %s\nKademliaID is %s\n\n", ip.String(), kademlia.network.routingTable.me.ID)

	go Listen(kademlia, 80)

	handleCommandLine(kademlia)

}

func handleCommandLine(kademlia *Kademlia) {

	for {

		fmt.Print("$ ")

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
				fmt.Println("Error: You need to provide an ip address and a KademliaID.\n     $ join <ip_address> <kademlia_id>")
				continue
			}
			ip, kademliaID := inputs[1], NewKademliaID(inputs[2])
			fmt.Printf("Joining the network via %s (%s) ...\n", ip, kademliaID.String())
			kademlia.network.routingTable.AddContact(NewContact(kademliaID, ip))
			// No go routine because you don't necessarily want the node to respond
			// other requests before he completes the join processus
			kademlia.LookupContact(kademlia.network.routingTable.me.ID)
			fmt.Print("Network joined.\n\n")
			break

		case "lookup":
			if len(inputs) < 2 {
				fmt.Println("Error: You need to provide a KademliaID. \n     $ lookup <kademlia_id>")
			}
			kademliaID := NewKademliaID(inputs[1])
			fmt.Printf("Looking for %s ...\n", kademliaID.String())
			go kademlia.LookupContact(kademliaID)
			break

		case "ping":
			if len(inputs) < 2 {
				fmt.Println("Error: You need to provide an IP address and a KademliaID.\n     $ ping <ip> <kademlia_id>")
			}
			ip, kademliaID := inputs[1], NewKademliaID(inputs[2])
			fmt.Printf("Pinging %s (%s) ...\n", ip, kademliaID)
			go kademlia.Ping(NewContact(kademliaID, ip))
			break

		case "exit":
			return

		default:
			fmt.Println("Invalid command, please try again.\nValids commands are:\n     $ join <ip> <kademlia_id>\n     $ lookup <kademlia_id>\n     $ ping <ip> <kademlia_id>")
			break

		}

	}

}
