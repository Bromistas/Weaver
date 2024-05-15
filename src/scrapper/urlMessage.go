package main

type URLType int

const (
	AmazonProduct URLType = iota
	NeweggProduct
	NeweggRoot
)

type URLMessage struct {
	URL     string
	URLType URLType
}
