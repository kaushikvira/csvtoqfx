package main

import (
	"fmt"

	"github.com/spf13/viper"
)

func validateConfig(configfile string) error {
	viper.SetDefault("indices", map[string]interface{}{"date": 1, "posted": 2, "name": 3, "id": 4, "amount": 5})

	if configfile != "" {
		viper.SetConfigFile(configfile)
		if err := viper.ReadInConfig(); err != nil {
			return err
		}
	}

	// Required string arguments
	required := []string{
		"org_name",
		"org_id",
		"bank_name",
		"bank_id",
		"account_id",
		"account_type",
		"start_date",
		"end_date",
		"balance",
		"available_balance",
	}

	for _, attr := range required {
		if viper.GetString(attr) == "" {
			return fmt.Errorf("required argument %s not specified via configfile or flags.", attr)
		}
	}

	return nil
}
