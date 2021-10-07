package main

import (
	"fmt"
	"sync"
	"time"
)

const TTL = 40 * time.Duration(1000000000)

type Storage struct {
	storedData map[string]Data
}
type Data struct {
	data       string
	storedTime time.Duration
	storedLock sync.Mutex
	originNode string
	dataNodes  []Contact
}

func NewStorage() *Storage {
	var storage Storage

	storage.storedData = map[string]Data{}
	return &storage
}

func (storage *Storage) Put(nodeAddress string, key *KademliaID, newdata string) {

	kademliaID := NewKademliaID(nodeAddress)
	node := NewContact(kademliaID, nodeAddress)

	NewData := Data{}

	NewData.data = newdata
	NewData.storedTime = TTL
	NewData.storedLock = sync.Mutex{}
	NewData.originNode = node.ID.String()

	storage.storedData[key.String()] = NewData
	fmt.Printf("KADEMLIA  " + node.ID.String())
}

func (storage *Storage) TimeToLive() {

	for {

		for k, t := range storage.storedData {
			f(&t, &k, storage)
		}

		time.Sleep(1000000000)

	}
}

func f(t *Data, k *string, storage *Storage) {

	ttl := t.storedTime
	ttl = ttl - time.Duration(1000000000)

	if ttl != 0 {
		t.storedTime = ttl

		storage.storedData[*k] = *t
	} else {
		delete(storage.storedData, *k)
		delete(storage.storedData, *k)

	}

}

func (storage *Storage) Get(key *KademliaID) (string, bool) {
	data, exists := storage.storedData[key.String()]
	mutex := data.storedLock
	mutex.Lock()
	data.storedTime = TTL
	mutex.Unlock()
	return data.data, exists
}

func (storage *Storage) String() string {
	data := ""
	for k, v := range storage.storedData {
		data += fmt.Sprintf("{key=%s,value=%s , ttl=%s}", k, v.data, v.storedTime.String())
	}

	return fmt.Sprintf("storage{data=[%s]}", data)
}

func (storage *Storage) RefreshData(address string, id string) {
	kademliaID := NewKademliaID(address)

	for k, d := range storage.storedData {

		node := d.originNode

		if node == kademliaID.String() {
			mutex := d.storedLock
			mutex.Lock()
			d.storedTime = TTL
			storage.storedData[k] = d
			mutex.Unlock()
		}
	}
}
