package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const InvalidEmailErrorMessage = "invalid email address"

func TestValidator(t *testing.T) {
	testCases := []struct {
		name                 string
		email                string
		isValid              bool
		expectedErrorMessage string
	}{
		{name: "Hyphen in address", email: "some-email@address.com", isValid: true},
		{name: "Dot in address", email: "some.email@address.com", isValid: true},
		{name: "Multi domain", email: "someemail@address.edu.ua", isValid: true},
		{name: "Hyphen in the end", email: "someemail-@address.ua", isValid: true},
		{name: "No domain", email: "some-email@", isValid: false, expectedErrorMessage: InvalidEmailErrorMessage},
		{name: "Hash symbol in email", email: "abc.def@mail#archive.com", isValid: false,
			expectedErrorMessage: InvalidEmailErrorMessage},
		{name: "Without at sign", email: "someemailaddress.edu.ua", isValid: false,
			expectedErrorMessage: InvalidEmailErrorMessage},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			v := New()
			v.Check(IsValidEmail(tc.email), "email", InvalidEmailErrorMessage)

			assert.Equal(t, tc.isValid, v.IsValid())
			assert.Equal(t, tc.expectedErrorMessage, v.Errors["email"])
		})
	}
}
