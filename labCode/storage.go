package main

import "fmt"

type Storage struct {
	data map[string]string
}

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
