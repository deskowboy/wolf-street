package model

type Stock struct {
	Name        string
	Code        string
	Number      string
	Description string
}

func GetAllStock() ([]Stock, error) {
	var Stocks = []Stock{
		{
			Name:        "AuMas Resources Bhd",
			Code:        "AUMAS",
			Number:      "0098",
			Description: "Investment holding company & segments include Aquaculture operations",
		},
		{
			Name:        "MN Holdings Bhd",
			Code:        "MNHLDG",
			Number:      "0245",
			Description: "Infrastructure utilities construction industries",
		},
		{
			Name:        "Pharmaniaga Bhd",
			Code:        "PHARMA",
			Number:      "7081",
			Description: "R&D, manufacturing of generic pharmaceutical products",
		},
		{
			Name:        "Zetrix AI Bhd",
			Code:        "ZETRIX",
			Number:      "0138",
			Description: "myeg",
		},
	}

	return Stocks, nil
}
