package nwt

import (
	"errors"
	"fmt"
	"slices"
	"time"
)

var (
	// MinTime represents the minimum valid time for NWT claims, corresponding to the 0 Unix epoch.
	MinTime = time.Unix(0, 0).UTC()

	// MaxTime represents the maximum valid time for NWT claims, set to December 31, 9999.
	MaxTime = time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)
)

// Token validation errors
var (
	ErrEmptyID           = errors.New("token ID is empty")
	ErrInvalidIssuedAt   = errors.New("issued at claim is invalid")
	ErrInvalidExpiration = errors.New("expiration claim is invalid")
	ErrInvalidNotBefore  = errors.New("not before claim is invalid")
	ErrInvalidTimeWindow = errors.New("not before is after expiration")
	ErrInvalidAudience   = errors.New("audience claim is invalid")
	ErrNotYetValid       = errors.New("token not yet valid (before NotBefore)")
	ErrExpired           = errors.New("token expired (after Expiration)")
)

// Validator wraps the Validate method for validating Tokens.
// The token is considered valid iff Validate returns nil.
//
// Implementations may enforce different policies for what constitutes a valid token,
// but are generally expected to at least validate the time-based claims with [ValidateTimeBounds].
//
// As an example, check out [StrictValidator].
type Validator interface {
	Validate(Token) error
}

// StrictValidator performs validation on the Token claims.
// It checks time-based claims with a configurable clock skew tolerance
// and verifies that the Audience claim contains an exact match of the specified identifier.
type StrictValidator struct {
	Identifier string
	ClockSkew  time.Duration
}

func (v StrictValidator) Validate(t Token) error {
	if t.ID == "" {
		return ErrEmptyID
	}

	if err := ValidateTimeBounds(t); err != nil {
		return err
	}

	now := time.Now()
	if now.Before(t.NotBefore.Add(-v.ClockSkew)) {
		return ErrNotYetValid
	}
	if now.After(t.Expiration.Add(v.ClockSkew)) {
		return ErrExpired
	}

	if len(t.Audience) > 0 {
		if !slices.Contains(t.Audience, v.Identifier) {
			return fmt.Errorf("%w: it doesn't contain an exact match of %q", ErrInvalidAudience, v.Identifier)
		}
	}
	return nil
}

// ValidateTimeBounds checks that the Token's time-based claims are within valid bounds.
func ValidateTimeBounds(t Token) error {
	if t.IssuedAt.Before(MinTime) || t.IssuedAt.After(MaxTime) {
		return ErrInvalidIssuedAt
	}

	if t.Expiration.Before(MinTime) || t.Expiration.After(MaxTime) {
		return ErrInvalidExpiration
	}

	if t.NotBefore.Before(MinTime) || t.NotBefore.After(MaxTime) {
		return ErrInvalidNotBefore
	}

	if t.NotBefore.After(t.Expiration) {
		return ErrInvalidTimeWindow
	}
	return nil
}
