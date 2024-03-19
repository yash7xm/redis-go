package main

type Storage struct {
	data map[string]Data_Value
}

type Data_Value struct {
	value string
}

func NewStorage() *Storage {
	return &Storage{
		data: make(map[string]Data_Value),
	}
}

func (kv *Storage) Get(Key string) (string, bool) {
	return kv.data[Key].value, true
}

func (kv *Storage) Set(key string, value string) {
	kv.data[key] = Data_Value{value: value}
}
