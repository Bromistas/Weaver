package common

type URLType int

const (
	AmazonProduct URLType = iota
	NeweggProduct
)

type URLMessage struct {
	URL     string
	URLType URLType
}
