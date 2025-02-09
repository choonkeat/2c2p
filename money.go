package api2c2p

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
)

// Cents represents monetary value in cents
type Cents int64

// ZeroPrefixed12DCents returns a string representation of the Cents value with leading zeros
func (c Cents) ZeroPrefixed12DCents() string {
	return fmt.Sprintf("%012d", c)
}

// Format: 12 digits with 5 decimal places (e.g., 000000002500.90000)
func (c Cents) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%012d.%02d000\"", c/100, c%100)), nil
}

// UnmarshalJSON decodes "000000000012.34000" into 1234
func (c *Cents) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	// Split by decimal point
	split := strings.Split(s, ".")
	if len(split) != 2 {
		return fmt.Errorf("invalid format")
	}

	// Parse first part
	whole, err := strconv.ParseInt(split[0], 10, 64)
	if err != nil {
		return fmt.Errorf("strconv.ParseInt: %v", err)
	}

	// Parse second part
	decimal, err := strconv.ParseInt(split[1][:5], 10, 64)
	if err != nil {
		return fmt.Errorf("strconv.ParseInt: %v", err)
	}

	// Combine whole and decimal parts
	*c = Cents(whole*100 + decimal/1000)
	return nil
}

// ToDollars converts Cents to Dollars
func (c Cents) ToDollars() Dollars {
	return Dollars{cents: c}
}

// Dollars represents monetary value in dollars, can only be created from Cents
type Dollars struct {
	cents Cents
}

// String implements fmt.Stringer
func (d Dollars) String() string {
	return fmt.Sprintf("%.2f", float64(d.cents)/100)
}

// MarshalXML implements xml.Marshaler
func (d Dollars) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(d.String(), start)
}

// UnmarshalXML implements xml.Unmarshaler
func (d *Dollars) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	var s string
	if err := dec.DecodeElement(&s, &start); err != nil {
		return err
	}
	var f float64
	if _, err := fmt.Sscanf(s, "%f", &f); err != nil {
		return err
	}
	d.cents = Cents(f * 100)
	return nil
}
