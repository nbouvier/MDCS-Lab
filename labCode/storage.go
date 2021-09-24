package main

import (
	"fmt"
	"sync"
	"time"
)

const TTL = 40 * time.Duration(1000000000)

type Storage struct {
	storedData map[string]string
	storedTime map[string]time.Duration
	storedLock map[string]sync.Mutex
}
type packet struct {
	ttl time.Duration
}

func NewStorage() *Storage {
	var storage Storage

	storage.storedData = map[string]string{}
	storage.storedTime = map[string]time.Duration{}
	storage.storedLock = map[string]sync.Mutex{}
	return &storage
}

func (storage *Storage) Put(key *KademliaID, data string) {

	storage.storedData[key.String()] = data
	storage.storedTime[key.String()] = TTL
	storage.storedLock[key.String()] = sync.Mutex{}

}

func (storage *Storage) timeToLive() {
	fmt.Print("DEBUG 1")
	var wg sync.WaitGroup
	for {

		for k, t := range storage.storedTime {
			wg.Add(1)
			//m:=storage.storedLock[k]
			f(&t, &k, storage /* &m, &wg*/)
		}
		//wg.Wait()
		time.Sleep(1000000000)

	}
}

func f(t *time.Duration, k *string, storage *Storage /* m *sync.Mutex, wg *sync.WaitGroup*/) {
	//m.Lock()
	//fmt.Print("DEBUG TTL= ", storage.storedData[*k])
	*t = *t - time.Duration(1000000000)

	if *t != 0 {
		storage.storedTime[*k] = *t
	} else {
		delete(storage.storedTime, *k)
		delete(storage.storedData, *k)
	}
	//m.Unlock()
	//wg.Done()
}

func (storage *Storage) Get(key *KademliaID) (string, bool) {
	data, exists := storage.storedData[key.String()]
	mutex := storage.storedLock[key.String()]
	mutex.Lock()
	fmt.Print("DEBUG GET")
	storage.storedTime[key.String()] = TTL
	fmt.Printf("ttl = ", storage.storedTime[key.String()])
	fmt.Print("DEBUG FIN GET")
	mutex.Unlock()
	return data, exists
}

func (storage *Storage) String() string {
	fmt.Sprintf("DEBUG")
	data := ""
	for k, v := range storage.storedData {
		data += fmt.Sprintf("{key=%s,value=%s , ttl=%s}", k, v, storage.storedTime[k].String())
	}

	return fmt.Sprintf("storage{data=[%s]}", data)
}

/*

func NewStorage() *Storage {
	var storage Storage

	storage.data = map[string]string{}

	return &storage
}

func (storage *Storage) Put(key *KademliaID, data string) {
	storage.data[key.String()] = data
}

func (storage *Storage) Get(key *KademliaID) (string, bool) {
	data, exists := storage.data[key.String()]
	return data, exists
}

func (storage *Storage) String() string {
	data := ""
	for k, v := range storage.data {
		data += fmt.Sprintf("{key=%s,value=%s}", k, v)
	}

	return fmt.Sprintf("storage{data=[%s]}", data)
}
*/
