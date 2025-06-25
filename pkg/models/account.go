package models

type AccountDetails struct {
	AccountNumber        string `json:"account_number"`
	Status              string `json:"status"`
	BuyingPower         string `json:"buying_power"`
	BuyingPowerCurrency string `json:"buying_power_currency"`
}