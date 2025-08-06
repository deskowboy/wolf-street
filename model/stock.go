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
