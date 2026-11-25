package entity

import (
	"regexp"
	"strings"
)

var countryCodePattern = regexp.MustCompile(`^[A-Z]{2}$`)

type Address struct {
	CountryCode string
	PostalCode  string
	AddressLine string
}

func NewAddress(countryCode, postalCode, addressLine string) (Address, error) {
	address := Address{
		CountryCode: strings.ToUpper(strings.TrimSpace(countryCode)),
		PostalCode:  strings.TrimSpace(postalCode),
		AddressLine: strings.TrimSpace(addressLine),
	}
	if !address.Valid() {
		return Address{}, ErrInput
	}
	return address, nil
}

func (a Address) Valid() bool {
	return countryCodePattern.MatchString(a.CountryCode) && len([]rune(a.PostalCode)) <= 32 && len([]rune(a.AddressLine)) >= 1 && len([]rune(a.AddressLine)) <= 500
}
