package monime

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

// validate is the shared validator instance used across all resources.
var validate = validator.New(validator.WithRequiredStructEnabled())

// validateStruct runs struct-tag validation on s and converts any failure into
// a *ValidationError.
func validateStruct(s any) error {
	if err := validate.Struct(s); err != nil {
		return newValidationError(err.Error())
	}
	return nil
}

// Metadata limits from https://docs.monime.io/docs/developer-resources/api-basics/object-metadata.
const (
	maxMetadataPairs       = 64
	maxMetadataValueLength = 100
)

var metadataKeyPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,64}$`)

// validateMetadata checks metadata against the API's rules before it is sent:
// at most 64 pairs; keys of 1 to 64 characters from [a-zA-Z0-9_-] that do not
// begin with "_", "-" or "monime" (any case); values that are strings of at most
// 100 characters. On updates a nil or "" value deletes the key, so nil is
// accepted there (allowDelete).
func validateMetadata(m map[string]any, allowDelete bool) error {
	if len(m) > maxMetadataPairs {
		return newValidationError(fmt.Sprintf("metadata has %d pairs; the limit is %d", len(m), maxMetadataPairs))
	}
	for k, v := range m {
		if !metadataKeyPattern.MatchString(k) {
			return newValidationError(fmt.Sprintf("metadata key %q must be 1 to 64 characters of letters, digits, '_' or '-'", k))
		}
		if strings.HasPrefix(k, "_") || strings.HasPrefix(k, "-") || strings.HasPrefix(strings.ToLower(k), "monime") {
			return newValidationError(fmt.Sprintf("metadata key %q must not begin with '_', '-' or 'monime'", k))
		}
		switch val := v.(type) {
		case nil:
			if !allowDelete {
				return newValidationError(fmt.Sprintf("metadata value for %q must be a string", k))
			}
		case string:
			if len(val) > maxMetadataValueLength {
				return newValidationError(fmt.Sprintf("metadata value for %q is longer than %d characters", k, maxMetadataValueLength))
			}
		default:
			return newValidationError(fmt.Sprintf("metadata value for %q must be a string", k))
		}
	}
	return nil
}

// currencyOrDefault returns c, or SLE when c is empty, and refuses anything the
// API does not support (the docs list SLE and USD).
func currencyOrDefault(c Currency) (Currency, error) {
	switch c {
	case "":
		return CurrencySLE, nil
	case CurrencySLE, CurrencyUSD:
		return c, nil
	}
	return "", newValidationError(fmt.Sprintf("currency %q must be SLE or USD", c))
}

// updateField describes one field an update (PATCH) body may carry. maxLen, when
// set, limits a string value; a nil value is always allowed, because the API
// treats null as "clear this field".
type updateField struct {
	maxLen int
}

// validateUpdate checks an untyped update body against the fields the API
// accepts for it. Every PATCH body in the spec rejects unknown properties, so an
// unknown key fails here instead of at the server.
func validateUpdate(params map[string]any, allowed map[string]updateField) error {
	if len(params) == 0 {
		return newValidationError("at least one field is required to update")
	}
	for k, v := range params {
		field, ok := allowed[k]
		if !ok {
			return newValidationError(fmt.Sprintf("%q cannot be updated on this resource", k))
		}
		if k == "metadata" {
			if v == nil {
				continue
			}
			var m map[string]any
			switch typed := v.(type) {
			case map[string]any:
				m = typed
			case map[string]string:
				m = make(map[string]any, len(typed))
				for mk, mv := range typed {
					m[mk] = mv
				}
			default:
				return newValidationError("metadata must be a map of strings")
			}
			if err := validateMetadata(m, true); err != nil {
				return err
			}
			continue
		}
		if s, isString := v.(string); isString && field.maxLen > 0 && len(s) > field.maxLen {
			return newValidationError(fmt.Sprintf("%s is longer than %d characters", k, field.maxLen))
		}
	}
	return nil
}

// putIfSet adds value to body under key unless it is the zero string, so
// optional fields the caller left empty are omitted rather than sent as "".
func putIfSet(body map[string]any, key, value string) {
	if value != "" {
		body[key] = value
	}
}
