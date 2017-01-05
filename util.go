package main

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var (
	amountPattern = regexp.MustCompile("[()$,]")
)

func toAmount(value string) (Amount, error) {
	multiplier := 1.0
	if strings.HasPrefix(value, "(") {
		multiplier = -1.0
	}
	amount, err := strconv.ParseFloat(amountPattern.ReplaceAllString(value, ""), 64)
	if err != nil {
		return 0, err
	}
	return Amount(multiplier * amount), nil
}

func outfileName(infile string, format string) string {
	var (
		name = filepath.Base(infile)
		ext  = filepath.Ext(infile)
	)
	outfile := fmt.Sprintf("%s.%s", filepath.Base(infile)[:len(name)-len(ext)], format)
	return filepath.Join(filepath.Dir(infile), outfile)
}
