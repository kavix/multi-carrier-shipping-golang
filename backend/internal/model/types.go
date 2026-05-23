package model

type QuoteRequest struct {
	Origin      string `json:"origin"`
	Destination string `json:"destination"`
	Weight      string `json:"weight"`
}

type QuoteResponse struct {
	Origin      string  `json:"origin"`
	Destination string  `json:"destination"`
	Weight      string  `json:"weight"`
	Carrier     string  `json:"carrier"`
	Price       float64 `json:"price"`
	TransitDays int     `json:"transitDays"`
}

type Carrier struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	TransitDays int     `json:"transitDays"`
	PriceFactor float64 `json:"priceFactor"`
}

type TrackingResponse struct {
	TrackingNumber string   `json:"trackingNumber"`
	Status         string   `json:"status"`
	LastLocation   string   `json:"lastLocation"`
	History        []string `json:"history"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
