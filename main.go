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
	format       = flag.String("format", "ofx", "format to convert to, one of (ofx|qfx).")
	currency     = flag.String("currency", "USD", "Currency of amounts.")
	orgName      = flag.String("org_name", "", "Financial Organization name, used for OFX FI>ORG value.")
	orgID        = flag.String("org_id", "", "Financial Organization name, used for OFX FI>FID value.")
	intuitID     = flag.String("intuit_id", "", "Financial Organization name, used for OFX FI>FID value.")
	bankID       = flag.String("bank_id", "", "Bank Routing Number, used for OFX BANKID value.")
	accountID    = flag.String("account_id", "", "Bank Account Number, used for OFX ACCTID value.")
	accountType  = flag.String("account_type", "", "Bank Account Type, used for OFX ACCTTYPE value.")
	startDate    = flag.String("start_date", "", "Start date for the statement.")
	endDate      = flag.String("end_date", "", "end date for the statement.")
	asOfDate     = flag.String("asof_date", "", "as of date for balances of this statement.")
	balance      = flag.Float64("balance", 0, "balance of the statement.")
	availBalance = flag.Float64("avail_balance", 0, "balance of the statement.")
	hasHeader    = flag.Bool("has_header", true, "If the input file has a header row.")
	dateLayout   = flag.String("date_layout", "2006/01/02", "Format to parse dates with.")
)

func main() {
	var (
		data     []byte
		parsed   *Parsed
		document *Document
		output   bytes.Buffer
		err      error
	)

	flag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	flag.Parse()
	// To supress errors from glog.
	goflag.CommandLine.Parse([]string{})
	viper.BindPFlags(flag.CommandLine)

	if len(flag.Args()) < 1 {
		glog.Exitf("no files to convert.")
	}

	if parsed, err = validateConfig(*configfile); err != nil {
		glog.Exitf("error validating configuration file - %s", err)
	}

	infile := flag.Arg(1)
	// Read the input file.
	if data, err = ioutil.ReadFile(infile); err != nil {
		glog.Exitf("unable to read file %s - %s", infile, err)
	}
	// Parse to a Document.
	if document, err = NewDocument(viper.GetViper(), parsed); err != nil {
		glog.Exitf("unable to create OFX document for %s - %s", infile, err)
	}
	document.Parse(data)
	// Serialize with the OFX template.
	if err = OFX102.Execute(&output, document); err != nil {
		glog.Exitf("unable to serialize file %s - %s", infile, err)
	}
	filename := outfileName(infile, viper.GetString("format"))
	glog.Infof("writing to %s", filename)
	// Write the output file.
	if err = ioutil.WriteFile(filename, output.Bytes(), os.ModePerm); err != nil {
		glog.Exitf("unable to write file %s - %s", infile, err)
	}
}
