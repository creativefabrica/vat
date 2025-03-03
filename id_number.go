package vat

import (
	"fmt"
	"regexp"
	"strings"
)

var patterns = map[string]string{
	"AT": `U[A-Z0-9]{8}`,
	"BE": `(0[0-9]{9}|[0-9]{10})`,
	"BG": `[0-9]{9,10}`,
	"CH": `(?:E(?:-| )[0-9]{3}(?:\.| )[0-9]{3}(?:\.| )[0-9]{3}( MWST)?|E[0-9]{9}(?:MWST)?)`,
	"CY": `[0-9]{8}[A-Z]`,
	"CZ": `[0-9]{8,10}`,
	"DE": `[0-9]{9}`,
	"DK": `[0-9]{8}`,
	"EE": `[0-9]{9}`,
	"EL": `[0-9]{9}`,
	"ES": `[A-Z][0-9]{7}[A-Z]|[0-9]{8}[A-Z]|[A-Z][0-9]{8}`,
	"FI": `[0-9]{8}`,
	"FR": `([A-Z]{2}|[0-9]{2})[0-9]{9}`,
	// Supposedly the regex for GB numbers is `[0-9]{9}|[0-9]{12}|(GD|HA)[0-9]{3}`,
	// but our validator service only accepts numbers with 9 or 12 digits following the country code.
	// Seems like the official site only accepts 9 digits... https://www.gov.uk/check-uk-vat-number
	"GB": `([0-9]{9}|[0-9]{12})`,
	"HR": `[0-9]{11}`,
	"HU": `[0-9]{8}`,
	"IE": `[A-Z0-9]{7}[A-Z]|[A-Z0-9]{7}[A-W][A-I]`,
	"IT": `[0-9]{11}`,
	"LT": `([0-9]{9}|[0-9]{12})`,
	"LU": `[0-9]{8}`,
	"LV": `[0-9]{11}`,
	"MT": `[0-9]{8}`,
	"NL": `[0-9]{9}B[0-9]{2}`,
	"PL": `[0-9]{10}`,
	"PT": `[0-9]{9}`,
	"RO": `[0-9]{2,10}`,
	"SE": `[0-9]{12}`,
	"SI": `[0-9]{8}`,
	"SK": `[0-9]{10}`,
	"XI": `([0-9]{9}|[0-9]{12})`, // Northern Ireland, same format as GB
}

type IDNumber struct {
	CountryCode string
	Number      string
}

func (id IDNumber) String() string {
	return fmt.Sprintf("%s%s", id.CountryCode, id.Number)
}

func MustParse(s string) IDNumber {
	id, err := Parse(s)
	if err != nil {
		panic(err)
	}
	return id
}

func Parse(s string) (IDNumber, error) {
	if len(s) < 3 {
		return IDNumber{}, ErrInvalidFormat
	}

	s = strings.ToUpper(strings.ReplaceAll(s, " ", ""))
	num := IDNumber{
		CountryCode: s[:2],
		Number:      s[2:],
	}

	pattern, ok := patterns[num.CountryCode]
	if !ok {
		return IDNumber{}, ErrInvalidCountryCode
	}

	matched, err := regexp.MatchString(fmt.Sprintf("^%s$", pattern), num.Number)
	if err != nil {
		return IDNumber{}, err
	}
	if !matched {
		return IDNumber{}, ErrInvalidFormat
	}

	return num, nil
}
