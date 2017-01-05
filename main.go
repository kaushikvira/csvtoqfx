package main

import (
	"bytes"
	"io/ioutil"
	"os"

	goflag "flag"

	"github.com/golang/glog"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	configfile = flag.String("configfile", "", "path to the config file.")
	// Configuration options.
	format      = flag.String("format", "ofx", "format to convert to, one of (ofx|qfx).")
	currency    = flag.String("currency", "USD", "Currency of amounts.")
	orgName     = flag.String("org_name", "", "Financial Organization name, used for OFX FI>ORG value.")
	orgID       = flag.String("org_id", "", "Financial Organization name, used for OFX FI>FID value.")
	intuitID    = flag.String("intuit_id", "", "Financial Organization name, used for OFX FI>FID value.")
	bankID      = flag.String("bank_id", "", "Bank Routing Number, used for OFX BANKID value.")
	accountID   = flag.String("account_id", "", "Bank Account Number, used for OFX ACCTID value.")
	accountType = flag.String("account_type", "", "Bank Account Type, used for OFX ACCTTYPE value.")
	startDate   = flag.String("start_date", "", "Start date for the statement.")
	endDate     = flag.String("end_date", "", "end date for the statement.")
	balance     = flag.Float64("balance", 0, "balance of the statement.")
	hasHeader   = flag.Bool("has_header", true, "If the input file has a header row.")
	dateLayout  = flag.String("date_layout", "2006/01/02", "Format to parse dates with.")
)

func main() {
	flag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	flag.Parse()
	// To supress errors from glog.
	goflag.CommandLine.Parse([]string{})
	viper.BindPFlags(flag.CommandLine)

	if len(flag.Args()) < 1 {
		glog.Exitf("no files to convert.")
	}

	if err := validateConfig(*configfile); err != nil {
		glog.Exitf("error validating configuration file - %s", err)
	}

	for _, f := range flag.Args() {
		var (
			data   []byte
			output bytes.Buffer
			err    error
		)

		// Read the input file.
		if data, err = ioutil.ReadFile(f); err != nil {
			glog.Errorf("unable to read file %s - %s", f, err)
			continue
		}
		// Parse to a Document.
		document := NewDocument(viper.GetViper())
		document.Parse(data)

		// Serialize with the OFX template.
		if err = OFX102.Execute(&output, document); err != nil {
			glog.Errorf("unable to serialize file %s - %s", f, err)
			continue
		}
		filename := outfileName(f, viper.GetString("format"))
		glog.Infof("writing to %s", filename)
		// Write the output file.
		if err = ioutil.WriteFile(filename, output.Bytes(), os.ModePerm); err != nil {
			glog.Errorf("unable to write file %s - %s", f, err)
			continue
		}
	}
}
