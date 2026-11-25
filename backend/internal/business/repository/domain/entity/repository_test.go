package entity

import (
	"strings"
	"testing"
)

func TestAddressNormalizesAndValidates(t *testing.T) {
	address, err := NewAddress(" tw ", " 100 ", " 臺北市中正區 ")
	if err != nil || address.CountryCode != "TW" || address.PostalCode != "100" || address.AddressLine != "臺北市中正區" {
		t.Fatalf("address = %+v, err = %v", address, err)
	}
	if _, err := NewAddress("TWN", "", "臺北市"); err != ErrInput {
		t.Fatalf("invalid country = %v", err)
	}
	if _, err := NewAddress("TW", "", " "); err != ErrInput {
		t.Fatalf("empty address = %v", err)
	}
}

func TestRepositoryIdentityIsIndependentFromOwnerReference(t *testing.T) {
	repositoryID := strings.Repeat("a", 32)
	userOwner := "U" + strings.Repeat("b", 32)
	organizationOwner := "11111111-1111-4111-8111-111111111111"
	if !ValidID(repositoryID) || ValidOwner(repositoryID) {
		t.Fatal("repository identity must not be accepted as an owner reference")
	}
	for _, owner := range []string{userOwner, organizationOwner} {
		if !ValidOwner(owner) || ValidID(owner) {
			t.Fatalf("owner reference and repository identity collapsed: %q", owner)
		}
	}
	for _, invalid := range []string{"", "U" + strings.Repeat("B", 32), "R" + strings.Repeat("d", 32), "C" + strings.Repeat("e", 32)} {
		if ValidOwner(invalid) {
			t.Fatalf("invalid owner reference accepted: %q", invalid)
		}
	}
}
