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

		case "ping":
			if len(inputs) < 2 {
				fmt.Println("Error: You need to provide a KademliaID.\n     $ ping <kademlia_id>")
				continue
			}
			fmt.Printf("Pinging %s ...\n", inputs[1])
			// kademlia.network.SendPingMessage(inputs[1])
			break

		case "store":
			fmt.Println("Sending to the network ...")
			break

		case "findValue":
			break

		case "join":
			if len(inputs) < 3 {
				fmt.Println("Error: You need to provide an ip address and a KademliaID.\n     $ join <ip_address> <kademlia_id>")
				continue
			}
			fmt.Printf("Joining the network via %s (%s) ...\n", inputs[1], inputs[2])
			kademlia.network.routingTable.AddContact(NewContact(NewKademliaID(inputs[2]), inputs[1]))
			kademlia.LookupContact(&kademlia.network.routingTable.me)
			fmt.Print("Network joined.\n\n")
			break

		case "exit":
			fmt.Println("Leaving the network ...")
			return

		default:
			fmt.Println("Invalid command, please try again.")
			break

		}

	}

}
