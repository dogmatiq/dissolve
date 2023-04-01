package domainname

import "strings"

// Absolute joins the given labels to form an absolute domain name including a
// trailing dot.
func Absolute(labels ...string) string {
	return strings.Join(labels, ".") + "."
}

// Relative joins the given labels to form a relative domain name.
func Relative(labels ...string) string {
	return strings.Join(labels, ".")
}
