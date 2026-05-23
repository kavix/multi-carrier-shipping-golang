package carrier

import "github.com/example/multi-carrier-shipping-golang/backend/internal/model"

func ListCarriers() []model.Carrier {
	return []model.Carrier{
		{ID: "fedex", Name: "FedEx", TransitDays: 2, PriceFactor: 1.05},
		{ID: "ups", Name: "UPS", TransitDays: 3, PriceFactor: 0.98},
		{ID: "dhl", Name: "DHL", TransitDays: 4, PriceFactor: 0.92},
	}
}

func QuoteForCarriers(origin, destination, weight string) []model.Carrier {
	carriers := ListCarriers()
	for idx := range carriers {
		carriers[idx].PriceFactor += 0.1
	}
	return carriers
}
