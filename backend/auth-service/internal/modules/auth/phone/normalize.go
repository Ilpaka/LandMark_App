package phone

import (
	"errors"

	"github.com/nyaruka/phonenumbers"
)

// ErrInvalid is returned when the number is not a valid E.164 candidate.
var ErrInvalid = errors.New("invalid phone number")

// ToE164 parses and formats a number to E.164 (e.g. +79991234567). defaultRegion is used when
// the input is a local number (e.g. "9123456789" for RU use "RU").
func ToE164(raw string, defaultRegion string) (string, error) {
	if defaultRegion == "" {
		defaultRegion = "RU"
	}
	p, err := phonenumbers.Parse(raw, defaultRegion)
	if err != nil {
		return "", err
	}
	if !phonenumbers.IsValidNumber(p) {
		return "", ErrInvalid
	}
	return phonenumbers.Format(p, phonenumbers.E164), nil
}
