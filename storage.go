package main

// Storage defines the interface for a storage adapter.
type Storage interface {
	Put(apt Apartment) error
}
