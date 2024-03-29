package dnssd

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"strings"
)

// Attributes represents the set of attributes conveyed in a DNS-SD service
// instance's TXT record.
//
// Each attribute may be either a key/value pair, where the value is a byte
// slice, or a flag (called a boolean attribute in RFC 6763).
//
// Pairs and flags occupy the same keyspace, meaning that it is not possible to
// have a flag with the same name as a pair's key.
//
// This is a consequence of how the attributes are represented inside the TXT
// records. A flag is represented as a key without value, which is also distinct
// from a pair with an empty value.
//
// Keys are case-insensitive. They MUST be at least one character long and
// SHOULD NOT be longer than 9 characters. The characters of a key MUST be
// printable US-ASCII values (0x20-0x7E), excluding '=' (0x3D).
//
// See https://www.rfc-editor.org/rfc/rfc6763#section-6.1
//
// Attributes is not safe for concurrent use without synchronization.
type Attributes struct {
	// m is a map of normalized key to value.
	//
	// A value of nil means the attribute is a flag, any non-nil byte slice
	// (including the empty slice) is a regular binary attribute.
	m map[string][]byte
}

// NewAttributes returns a new empty attribute set.
func NewAttributes() Attributes {
	return Attributes{}
}

// Get returns the value that is associated with the key k.
//
// ok is true there is a key/value pair with this key.
func (a Attributes) Get(k string) (v []byte, ok bool) {
	v = a.m[mustNormalizeAttributeKey(k)]
	return v, v != nil
}

// WithPair returns a clone of the attributes with an additional key/value pair.
//
// It replaces any existing key/value pair or flag with this key.
func (a Attributes) WithPair(k string, v []byte) Attributes {
	return a.mutate(func(m map[string][]byte) {
		// If v is nil, replace it with an empty slice instead, otherwise it is
		// considered a flag.
		if v == nil {
			v = []byte{}
		}
		m[mustNormalizeAttributeKey(k)] = v
	})
}

// Pairs returns the key/value pair (i.e. non-flag) attributes.
func (a Attributes) Pairs() map[string][]byte {
	attrs := map[string][]byte{}

	for k, v := range a.m {
		if v != nil {
			attrs[k] = v
		}
	}

	return attrs
}

// WithFlag returns a lcone of the attributes with an additional flag.
//
// It replaces any existing key/value pair with this key.
//
// Use Without() to clear a flag.
func (a Attributes) WithFlag(k string) Attributes {
	return a.mutate(func(m map[string][]byte) {
		m[mustNormalizeAttributeKey(k)] = nil
	})
}

// HasFlags returns true if all of the given flags are present in the
// attributes.
func (a Attributes) HasFlags(keys ...string) bool {
	for _, k := range keys {
		v, ok := a.m[mustNormalizeAttributeKey(k)]
		if !ok || v != nil {
			return false
		}
	}

	return true
}

// Flags returns the flag (i.e. non-pair) attributes that are set.
func (a Attributes) Flags() map[string]struct{} {
	flags := map[string]struct{}{}

	for k, v := range a.m {
		if v == nil {
			flags[k] = struct{}{}
		}
	}

	return flags
}

// Without returns a clone of the attributes without the given keys, regardless
// of whether they are key/value pairs or flags.
func (a Attributes) Without(keys ...string) Attributes {
	return a.mutate(func(m map[string][]byte) {
		for _, k := range keys {
			delete(m, mustNormalizeAttributeKey(k))
		}
	})
}

// IsEmpty returns true if there are no attributes present.
func (a Attributes) IsEmpty() bool {
	return len(a.m) == 0
}

// WithTXT returns a clone of the attributes containing an attribute parsed from
// a single value within in a DNS-SD service instance's TXT record.
//
// As per RFC 6763, TXT record values that begin with an '=' are ignored, in
// which case ok is false. Empty values are also ignored.
func (a Attributes) WithTXT(pair string) (_ Attributes, ok bool, err error) {
	if pair == "" {
		return a, false, nil
	}

	var (
		k string
		v []byte
	)

	switch n := strings.IndexByte(pair, '='); n {
	case 0:
		// DNS-SD TXT record strings beginning with an '=' character
		// (i.e., the key is missing) MUST be silently ignored.
		return a, false, nil
	case -1:
		// No equals sign, attribute is a flag.
		k = pair
	default:
		v = []byte(pair[n+1:])
		k = pair[:n]
	}

	k, err = normalizeAttributeKey(k)
	if err != nil {
		return Attributes{}, false, err
	}

	return a.mutate(func(m map[string][]byte) {
		m[k] = v
	}), true, nil
}

// ToTXT returns the string representation of each key/value pair, as they
// appear in the TXT record.
//
// The result is deterministic (keys are sorted) to avoid unnecessary DNS churn
// when the attributes are used to construct DNS records.
func (a Attributes) ToTXT() []string {
	type pair struct {
		key   string
		value []byte
	}

	pairs := make([]pair, 0, len(a.m))
	for k, v := range a.m {
		pairs = append(pairs, pair{k, v})
	}

	// https://www.rfc-editor.org/rfc/rfc6763#section-6.7
	//
	// Always place the 'version tag' attribute ("txtvers") in the first
	// entry of the TXT record.
	const versionKey = "txtvers"

	sort.Slice(
		pairs,
		func(i, j int) bool {
			a := pairs[i]
			b := pairs[j]

			if a.key == versionKey {
				return true
			}

			if b.key == versionKey {
				return false
			}

			return a.key < b.key
		},
	)

	var result []string
	for _, p := range pairs {
		if p.value == nil {
			// https://www.rfc-editor.org/rfc/rfc6763#section-6.4
			//
			// If there is no '=' in a DNS-SD TXT record string, then it is a
			// boolean attribute, simply identified as being present, with no
			// value.
			result = append(result, p.key)
		} else {
			result = append(result, p.key+"="+string(p.value))
		}
	}

	return result
}

// Equal returns true if the attributes are equal.
func (a Attributes) Equal(attr Attributes) bool {
	if len(a.m) != len(attr.m) {
		return false
	}

	for k, v1 := range a.m {
		v2, ok := attr.m[k]
		if !ok {
			return false
		}

		isFlag1 := v1 == nil
		isFlag2 := v2 == nil

		if isFlag1 != isFlag2 {
			return false
		}

		if !bytes.Equal(v1, v2) {
			return false
		}
	}

	return true
}

func (a Attributes) mutate(
	fn func(map[string][]byte),
) Attributes {
	m := make(map[string][]byte, len(a.m))

	for k, v := range a.m {
		m[k] = v
	}

	fn(m)

	return Attributes{m}
}

// mustNormalizeAttributeKey normalizes the DNS-SD TXT key, k, or panics if it
// can not be normalized.
//
// See https://www.rfc-editor.org/rfc/rfc6763#section-6.4
func mustNormalizeAttributeKey(k string) string {
	k, err := normalizeAttributeKey(k)
	if err != nil {
		panic(err)
	}

	return k
}

// normalizeAttributeKey normalizes the DNS-SD TXT key, k, or returns an error
// if it can not be normalized.
//
// See https://www.rfc-editor.org/rfc/rfc6763#section-6.4
func normalizeAttributeKey(k string) (string, error) {
	if k == "" {
		return "", errors.New("key must not be empty")
	}

	var w strings.Builder

	for i := 0; i < len(k); i++ {
		ch := k[i]

		// https://www.rfc-editor.org/rfc/rfc6763#section-6.4
		//
		// The characters of a key MUST be printable US-ASCII values (0x20-0x7E)
		// [RFC20], excluding '=' (0x3D).

		if ch == '=' {
			return "", fmt.Errorf("invalid key '%s', key must not contain '=' character", k)
		}

		if ch < 0x20 || ch > 0x7E {
			return "", fmt.Errorf("invalid key '%s', key must contain only printable ASCII characters", k)
		}

		// Convert ASCII letters to lowercase.
		if 'A' <= ch && ch <= 'Z' {
			ch += 'a' - 'A'
		}

		w.WriteByte(ch)
	}

	return w.String(), nil
}

// AttributeCollection is a collection of [Attributes]. Each entry in the slice
// contains the attributes conveyed in a separate TXT record.
type AttributeCollection []Attributes

// Get returns the last value that is associated with the key k.
//
// ok is true there is a key/value pair with this key.
func (c AttributeCollection) Get(k string) (v []byte, ok bool) {
	k = mustNormalizeAttributeKey(k)

	for i := len(c) - 1; i >= 0; i-- {
		a := c[i]
		v = a.m[k]
		if v != nil {
			return v, true
		}
	}

	return nil, false
}

// Pairs returns the key/value pair (i.e. non-flag) attributes.
func (c AttributeCollection) Pairs() map[string][]byte {
	attrs := map[string][]byte{}

	for _, a := range c {
		for k, v := range a.m {
			if v != nil {
				attrs[k] = v
			}
		}
	}

	return attrs
}

// HasFlags returns true if all of the given flags are present in the
// attributes.
func (c AttributeCollection) HasFlags(keys ...string) bool {
key:
	for _, k := range keys {
		k = mustNormalizeAttributeKey(k)

		for _, a := range c {
			v, ok := a.m[k]
			if ok && v == nil {
				continue key
			}
		}

		return false
	}

	return true
}

// Flags returns the flag (i.e. non-pair) attributes that are set.
func (c AttributeCollection) Flags() map[string]struct{} {
	flags := map[string]struct{}{}

	for _, a := range c {
		for k, v := range a.m {
			if v == nil {
				flags[k] = struct{}{}
			}
		}
	}

	return flags
}

// Equal returns true if c and x contain the same sets of attributes, in any
// order.
func (c AttributeCollection) Equal(x AttributeCollection) bool {
	if len(c) != len(x) {
		return false
	}

	visited := make([]bool, len(x))

left:
	for _, l := range c {
		for i, r := range x {
			if visited[i] {
				continue
			}

			if l.Equal(r) {
				visited[i] = true
				continue left
			}
		}

		return false
	}

	return true
}
