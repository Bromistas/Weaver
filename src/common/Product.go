package common

type Product struct {
	Name        string  `json:"name"`
	Price       float32 `json:"price"`
	URL         string  `json:"url"`
	Description string  `json:"description"`
	Rating      string  `json:"rating"`
}
