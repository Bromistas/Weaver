package common

type Product struct {
	Name        string  `json:"name"`
	Price       float32 `json:"price"`
	URL         string  `json:"url"`
	Description string  `json:"description"`
	Rating      string  `json:"rating"`
	NodeAuthor  string  `json:"node_author"`
	Replicated  bool    `json:"replicated"`
}

type URLMessage struct {
	URL     string
	URLType URLType
}

type URLType int

const (
	AmazonProduct URLType = iota
	NeweggProduct
	NeweggRoot
	AmazonRoot
)
