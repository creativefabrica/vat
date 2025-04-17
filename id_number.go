package vat

import (
	"fmt"
	"regexp"
	"strings"
)

//nolint:gochecknoglobals // This is a constant map of country codes to their VAT ID number regex patterns.

var patterns = map[string]*regexp.Regexp{
	"AU": regexp.MustCompile(`[0-9]{11}`),
	"AT": regexp.MustCompile(`U[A-Z0-9]{8}`),
	"BE": regexp.MustCompile(`(0[0-9]{9}|[0-9]{10})`),
	"BG": regexp.MustCompile(`[0-9]{9,10}`),
	"CH": regexp.MustCompile(
		`(?:E(?:-| )[0-9]{3}(?:\.| )[0-9]{3}(?:\.| )[0-9]{3}( MWST)?|E[0-9]{9}(?:MWST)?)`,
	),
	"CY": regexp.MustCompile(`[0-9]{8}[A-Z]`),
	"CZ": regexp.MustCompile(`[0-9]{8,10}`),
	"DE": regexp.MustCompile(`[0-9]{9}`),
	"DK": regexp.MustCompile(`[0-9]{8}`),
	"EE": regexp.MustCompile(`[0-9]{9}`),
	"EL": regexp.MustCompile(`[0-9]{9}`),
	"ES": regexp.MustCompile(`[A-Z][0-9]{7}[A-Z]|[0-9]{8}[A-Z]|[A-Z][0-9]{8}`),
	"FI": regexp.MustCompile(`[0-9]{8}`),
	"FR": regexp.MustCompile(`([A-Z]{2}|[0-9]{2})[0-9]{9}`),
	// Supposedly the regex for GB numbers is `[0-9]{9}|[0-9]{12}|(GD|HA)[0-9]{3}`,
	// but our validator service only accepts numbers with 9 or 12 digits following the country code.
	// Seems like the official site only accepts 9 digits... https://www.gov.uk/check-uk-vat-number
	"GB": regexp.MustCompile(`([0-9]{9}|[0-9]{12})`),
	"HR": regexp.MustCompile(`[0-9]{11}`),
	"HU": regexp.MustCompile(`[0-9]{8}`),
	"IE": regexp.MustCompile(`[A-Z0-9]{7}[A-Z]|[A-Z0-9]{7}[A-W][A-I]`),
	"IT": regexp.MustCompile(`[0-9]{11}`),
	"LT": regexp.MustCompile(`([0-9]{9}|[0-9]{12})`),
	"LU": regexp.MustCompile(`[0-9]{8}`),
	"LV": regexp.MustCompile(`[0-9]{11}`),
	"MT": regexp.MustCompile(`[0-9]{8}`),
	"NL": regexp.MustCompile(`[0-9]{9}B[0-9]{2}`),
	"PL": regexp.MustCompile(`[0-9]{10}`),
	"PT": regexp.MustCompile(`[0-9]{9}`),
	"RO": regexp.MustCompile(`[0-9]{2,10}`),
	"SE": regexp.MustCompile(`[0-9]{12}`),
	"SI": regexp.MustCompile(`[0-9]{8}`),
	"SK": regexp.MustCompile(`[0-9]{10}`),
	"XI": regexp.MustCompile(`([0-9]{9}|[0-9]{12})`), // Northern Ireland, same format as GB
}

const idNumberMinLength = 3

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
	s = strings.ReplaceAll(s, " ", "")

	if len(s) < idNumberMinLength {
		return IDNumber{}, ErrInvalidFormat
	}

	s = strings.ToUpper(s)
	num := IDNumber{
		CountryCode: s[:2],
		Number:      s[2:],
	}

	pattern, ok := patterns[num.CountryCode]
	if !ok {
		return IDNumber{}, ErrInvalidCountryCode
	}

	if !pattern.MatchString(num.Number) {
		return IDNumber{}, ErrInvalidFormat
	}

	if num.CountryCode == "AU" && !validaABN(num.Number) {
		return IDNumber{}, ErrInvalidFormat
	}

	return num, nil
}

// validaABN will check if an ABN is valid.
// For more information on how this works you can
// refer to: https://abr.business.gov.au/Help/AbnFormat
func validaABN(abn string) bool {
	// If the first check digit is a 0 then it's not a valid ABN
	if abn[:1] == "0" {
		return false
	}

	// Subtract 1 from the first check digit of the abn
	abnByte := []byte(abn)
	abnByte[0]--
	abn = string(abnByte)

	if abn == "" || len(abn) != 11 {
		return false
	}

	var abnWeights = []int{10, 1, 3, 5, 7, 9, 11, 13, 15, 17, 19}
	var weightingSum int

	for i := range abnWeights {
		num := int(abn[i]) - 48
		weightingSum += num * abnWeights[i]
	}

	// If the weightedSum is a multiple of 89 then it's a valid ABN
	return (weightingSum % 89) == 0
}
