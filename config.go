package main

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Parsed struct {
	Dates   map[string]string
	Amounts map[string]Amount
}

func validateConfig(configfile string) (*Parsed, error) {
	viper.SetDefault("indices", map[string]interface{}{"date": 1, "posted": 2, "name": 3, "id": 4, "amount": 5})

	if configfile != "" {
		viper.SetConfigFile(configfile)
		if err := viper.ReadInConfig(); err != nil {
			return nil, err
		}
	}

	// Required string arguments
	required := []string{
		"org_name",
		"org_id",
		"bank_id",
		"account_id",
		"account_type",
		"start_date",
		"end_date",
		"balance",
		"avail_balance",
	}

	for _, attr := range required {
		if viper.GetString(attr) == "" {
			return nil, fmt.Errorf("required argument %s not specified via configfile or flags.", attr)
		}
	}

	parsed := &Parsed{
		Dates:   make(map[string]string, 3),
		Amounts: make(map[string]Amount, 2),
	}

	// Required date arguments
	for _, attr := range []string{"start_date", "end_date", "asof_date"} {
		date, err := time.Parse(viper.GetString("date_layout"), viper.GetString(attr))
		if err != nil {
			return nil, fmt.Errorf("error parsing %s - %s", attr, err.Error())
		}
		parsed.Dates[attr] = date.Format("20060102150405.000[0:GMT]")
	}

	// Required Amount arguments
	for _, attr := range []string{"balance", "avail_balance"} {
		amount, err := toAmount(viper.GetString(attr))
		if err != nil {
			return nil, fmt.Errorf("error parsing amount %s - %s", attr, err.Error())
		}
		parsed.Amounts[attr] = amount
	}

	return parsed, nil
}
