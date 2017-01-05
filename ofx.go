package main

import (
	"bytes"
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"io"
	"time"

	"github.com/golang/glog"
	"github.com/spf13/viper"
)

type TransactionType string

const (
	// Common Transaction Types
	DEBIT  TransactionType = "DEBIT"
	CREDIT TransactionType = "CREDIT"
	// Uncommon Transaction Types
	INTEREST      TransactionType = "INT"
	DIVIDENT      TransactionType = "DIV"
	FEE           TransactionType = "FEE"
	SERVICECHARGE TransactionType = "SRVCHG"
	DEPOSIT       TransactionType = "DEP"
	ATM           TransactionType = "ATM"
	POS           TransactionType = "POS"
	TRANSFER      TransactionType = "XFER"
	CHECK         TransactionType = "CHECK"
	PAYMENT       TransactionType = "PAYMENT"
	CASH          TransactionType = "CASH"
	DIRECTDEPOSIT TransactionType = "DIRECTDEP"
	DIRECTDEBIT   TransactionType = "DIRECTDEBIT"
	REPEATPAYMENT TransactionType = "REPEATPMT"
	OTHER         TransactionType = "OTHER"
)

type Amount float64

func (a Amount) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(fmt.Sprintf("%0.2f", a), start)
}

type Transaction struct {
	Type   TransactionType `xml:"TRNTYPE"`
	Posted string          `xml:"DTPOSTED"`
	Amount Amount          `xml:"TRNAMT"`
	ID     string          `xml:"FITID"`
	Date   string          `xml:"DTUSER,omitempty"`
	Name   string          `xml:"NAME,omitempty"`
	Payee  string          `xml:"PAYEE,omitempty"`
	Memo   string          `xml:"MEMO,omitempty"`
}

type SignOnResponse struct {
	Code           int    `xml:"STATUS>CODE"`
	Severity       string `xml:"STATUS>SEVERITY"`
	Date           string `xml:"DTSERVER"`
	Language       string `xml:"LANGUAGE"`
	Organization   string `xml:"FI>ORG"`
	OrganizationID string `xml:"FI>FID"`
	IntuitID       string `xml:"INTU.BID,omitempty"`
}

type StatementTransactionResponseSet struct {
	ID       int                  `xml:"TRNUID"`
	Code     int                  `xml:"STATUS>CODE"`
	Severity string               `xml:"STATUS>SEVERITY"`
	RS       StatementResponseSet `xml:"STMTRS"`
}

type Balance struct {
	Amount Amount `xml:"BALAMT"`
	Date   string `xml:"DTASOF"`
}

type StatementResponseSet struct {
	Currency         string        `xml:"CURDEF"`
	BankID           string        `xml:"BANKACCTFROM>BANKID"`
	AccountID        string        `xml:"BANKACCTFROM>ACCTID"`
	AccountType      string        `xml:"BANKACCTFROM>ACCTTYPE"`
	StartDate        string        `xml:"BANKTRANLIST>DTSTART"`
	EndDate          string        `xml:"BANKTRANLIST>DTEND"`
	Transactions     []Transaction `xml:"BANKTRANLIST>STMTTRN"`
	LedgerBalance    Balance       `xml:"LEDGERBAL"`
	AvailableBalance Balance       `xml:"AVAILBAL"`
}

type Document struct {
	config   *viper.Viper
	XMLName  xml.Name                        `xml:"OFX"`
	Response SignOnResponse                  `xml:"SIGNONMSGSRSV1>SONRS"`
	TRS      StatementTransactionResponseSet `xml:"BANKMSGSRSV1>STMTTRNRS"`
}

func NewDocument(c *viper.Viper) *Document {
	d := &Document{config: c}
	d.Response = SignOnResponse{
		Severity:       "INFO",
		Date:           time.Now().Format("20060102150405.000[-7]"),
		Language:       "ENG",
		Organization:   c.GetString("org_name"),
		OrganizationID: c.GetString("org_id"),
	}
	d.TRS = StatementTransactionResponseSet{
		Severity: "INFO",
		RS: StatementResponseSet{
			Currency:    c.GetString("currency"),
			BankID:      c.GetString("bank_id"),
			AccountID:   c.GetString("account_id"),
			AccountType: c.GetString("account_type"),
		},
	}
	if c.GetString("format") == "qfx" {
		d.Response.IntuitID = c.GetString("org_id")
		if c.IsSet("intuit_id") && c.GetString("intuit_id") != "" {
			d.Response.IntuitID = c.GetString("intuit_id")
		}
	}

	return d
}

func (d *Document) parseTransaction(row []string) (*Transaction, error) {
	var (
		t   = &Transaction{}
		err error
	)

	glog.Infof("parsing row - %#v", row)

	dateLayout := d.config.GetString("date_layout")

	// Posted Date
	if idx := d.config.GetInt("indices.posted"); d.config.IsSet("indices.posted") && len(row) > idx {
		if postedDate, err := time.Parse(dateLayout, row[idx-1]); err != nil {
			glog.Errorf("error parsing date %s with format %s", row[idx-1], dateLayout)
			return nil, err
		} else {
			t.Posted = postedDate.Format("20060102150405.000[0:GMT]")
		}
	}

	// Transaction Date
	if idx := d.config.GetInt("indices.date"); d.config.IsSet("indices.date") && len(row) > idx {
		if date, err := time.Parse(dateLayout, row[idx-1]); err != nil {
			glog.Errorf("error parsing date %s with format %s", row[idx-1], dateLayout)
		} else {
			t.Date = date.Format("20060102150405.000[0:GMT]")
		}
	}

	// Amount
	if t.Amount, err = toAmount(row[d.config.GetInt("indices.amount")-1]); err != nil {
		return nil, err
	}

	// Type
	if idx := d.config.GetInt("indices.type"); d.config.IsSet("indices.type") && len(row) > idx {
		t.Type = TransactionType(row[idx-1])
	} else {
		// Fallback to assuming that positive entries are debits and
		// negative entries are credits. This is reverse of OFX/QFX where
		// Debits are negative and credits are positive.
		t.Amount = t.Amount * -1
		t.Type = DEBIT
		if t.Amount > 0 {
			t.Type = CREDIT
		}
	}

	if idx := d.config.GetInt("indices.id"); d.config.IsSet("indices.id") && len(row) > idx {
		t.ID = row[idx-1]
	}
	if idx := d.config.GetInt("indices.name"); d.config.IsSet("indices.name") && len(row) > idx {
		t.Name = row[idx-1]
	}
	if idx := d.config.GetInt("indices.memo"); d.config.IsSet("indices.memo") && len(row) > idx {
		t.Memo = row[idx-1]
	}
	if idx := d.config.GetInt("indices.payee"); d.config.IsSet("indices.payee") && len(row) > idx {
		t.Payee = row[idx-1]
	}

	return t, nil
}

func (d *Document) Parse(data []byte) {
	reader := csv.NewReader(bytes.NewReader(data))
	// Discard the header row.
	if d.config.GetBool("has_header") {
		reader.Read()
	}
	for {
		r, err := reader.Read()
		if err != nil {
			if err != io.EOF {
				glog.Errorf("error parsing data - %s", err)
			}
			break
		}

		if txn, err := d.parseTransaction(r); err != nil {
			glog.Errorf("error parsing transaction row - %s")
		} else {
			d.TRS.RS.Transactions = append(d.TRS.RS.Transactions, *txn)
		}
	}
}

func (d *Document) ToXML() (string, error) {
	data, err := xml.Marshal(d)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
