package order

import (
	"math"
	"strconv"

	"github.com/example/multi-carrier-shipping-golang/backend/internal/model"
)

func GenerateQuote(origin, destination, weight string) model.QuoteResponse {
	value := parseWeight(weight)
	price := math.Round((25.0+value*2.5)*100) / 100
	return model.QuoteResponse{
		Origin:      origin,
		Destination: destination,
		Weight:      weight,
		Carrier:     "MultiCarrier Express",
		Price:       price,
		TransitDays: 2 + int(value/10),
	}
}

func parseWeight(weight string) float64 {
	value, err := strconv.ParseFloat(weight, 64)
	if err != nil || value < 0 {
		return 1
	}
	return value
}
