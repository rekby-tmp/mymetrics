package main

type MemStorage struct {
	gauge   map[string]float64
	counter map[string]int64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauge:   map[string]float64{},
		counter: map[string]int64{},
	}
}

func (m *MemStorage) StoreGauge(name string, value float64) error {
	m.gauge[name] = value
	return nil
}

func (m *MemStorage) StoreCounter(name string, value int64) error {
	m.counter[name] += value
	return nil
}
