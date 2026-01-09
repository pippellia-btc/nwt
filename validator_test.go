package nwt

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestStrictValidator(t *testing.T) {
	tests := []struct {
		token Token
		err   error
	}{
		{
			token: Token{},
			err:   ErrEmptyID,
		},
		{
			token: Token{ID: "id"},
			err:   ErrInvalidIssuedAt,
		},
		{
			token: Token{
				ID:         "id",
				IssuedAt:   time.Unix(420, 0).UTC(),
				Expiration: time.Unix(69, 0).UTC(),
				NotBefore:  time.Unix(420, 0).UTC(),
			},
			err: ErrInvalidTimeWindow,
		},
		{
			token: Token{
				ID:         "id",
				IssuedAt:   time.Now().UTC(),
				Expiration: MaxTime,
				NotBefore:  time.Now().Add(time.Hour).UTC(),
			},
			err: ErrNotYetValid,
		},
		{
			token: Token{
				ID:         "id",
				IssuedAt:   time.Now().UTC(),
				Expiration: time.Now().Add(-time.Hour).UTC(),
				NotBefore:  MinTime,
			},
			err: ErrExpired,
		},
		{
			token: Token{
				ID:         "id",
				IssuedAt:   time.Now().UTC(),
				Audience:   []string{"other-identifier"},
				Expiration: time.Now().Add(time.Hour).UTC(),
				NotBefore:  time.Now().Add(-time.Hour).UTC(),
			},
			err: ErrInvalidAudience,
		},
		{
			token: Token{
				ID:         "id",
				IssuedAt:   time.Now().UTC(),
				Audience:   []string{"identifier"},
				Expiration: time.Now().Add(time.Hour).UTC(),
				NotBefore:  time.Now().Add(-time.Hour).UTC(),
			},
			err: nil,
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			validator := StrictValidator{
				Identifier: "identifier",
				ClockSkew:  time.Minute,
			}

			err := validator.Validate(test.token)
			if !errors.Is(err, test.err) {
				t.Fatalf("expected error %v, got %v", test.err, err)
			}
		})
	}

}
