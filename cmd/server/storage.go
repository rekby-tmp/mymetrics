package main

type Storage interface {
	StoreGauge(name string, value float64) error
	StoreCounter(name string, value int64) error
}
