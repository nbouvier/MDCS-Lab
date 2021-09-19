package main

type Storage struct {
	data map[string]string
}

func (storage *Storage) Put(key string, data string) {
	storage.data[key] = data
}

func (storage *Storage) Get(key string) string {
	return storage.data[key]
}
