package dnssd_test

import (
	. "github.com/dogmatiq/dissolve/dnssd"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("type Attributes", func() {
	Context("binary attributes", func() {
		Describe("func WithPair()", func() {
			It("sets the attribute", func() {
				attrs := NewAttributes().
					WithPair("<key>", []byte("<value>"))

				v, ok := attrs.Get("<key>")
				Expect(v).To(Equal([]byte("<value>")))
				Expect(ok).To(BeTrue())
			})

			It("does not modify the original attributes", func() {
				attrs := NewAttributes()
				attrs.WithPair("<key>", []byte("<value>"))

				Expect(attrs.IsEmpty()).To(BeTrue())
			})

			It("replaces flags with the same key", func() {
				attrs := NewAttributes().
					WithFlag("<key>").
					WithPair("<key>", []byte("<value>"))

				_, ok := attrs.Get("<key>")
				Expect(ok).To(BeTrue())
				Expect(attrs.HasFlags("<key>")).To(BeFalse())
			})

			It("replaces other binary attributes with the same key", func() {
				attrs := NewAttributes().
					WithPair("<key>", []byte("<value-1>")).
					WithPair("<key>", []byte("<value-2>"))

				v, ok := attrs.Get("<key>")
				Expect(v).To(Equal([]byte("<value-2>")))
				Expect(ok).To(BeTrue())
			})

			It("allows association of a nil slice", func() {
				attrs := NewAttributes().
					WithPair("<key>", nil)

				v, ok := attrs.Get("<key>")
				Expect(v).To(BeEmpty())
				Expect(ok).To(BeTrue())
			})

			It("allows association of an empty slice", func() {
				attrs := NewAttributes().
					WithPair("<key>", []byte{})

				v, ok := attrs.Get("<key>")
				Expect(v).To(BeEmpty())
				Expect(ok).To(BeTrue())
			})

			DescribeTable(
				"it panics if the key is invalid",
				func(k, expect string) {
					Expect(func() {
						NewAttributes().
							WithPair(k, nil)
					}).To(PanicWith(MatchError(expect)))
				},
				Entry("empty", "", "key must not be empty"),
				Entry("contains equals", "<k=v>", "invalid key '<k=v>', key must not contain '=' character"),
				Entry("contains UTF-8", "<ключ>", "invalid key '<ключ>', key must contain only printable ASCII characters"),
				Entry("contains non-printable ASCII", "<\x00>", "invalid key '<\x00>', key must contain only printable ASCII characters"),
			)
		})

		Describe("func Get()", func() {
			It("returns the associated value", func() {
				attrs := NewAttributes().
					WithPair("<key>", []byte("<value>"))

				v, ok := attrs.Get("<key>")
				Expect(v).To(Equal([]byte("<value>")))
				Expect(ok).To(BeTrue())
			})

			It("returns false if there is no such key", func() {
				attrs := NewAttributes()

				_, ok := attrs.Get("<key>")
				Expect(ok).To(BeFalse())
			})

			It("returns false the key is a flag", func() {
				attrs := NewAttributes().
					WithFlag("<key>")

				_, ok := attrs.Get("<key>")
				Expect(ok).To(BeFalse())
			})

			It("is case insensitive", func() {
				attrs := NewAttributes().
					WithPair("<KEY>", []byte("<value>"))

				v, ok := attrs.Get("<key>")
				Expect(v).To(Equal([]byte("<value>")))
				Expect(ok).To(BeTrue())
			})
		})

		Describe("func Pairs()", func() {
			It("returns the binary attributes", func() {
				attrs := NewAttributes().
					WithPair("<key-1>", []byte("<value-1>")).
					WithPair("<key-2>", nil).
					WithFlag("<key-3>")

				Expect(attrs.Pairs()).To(Equal(
					map[string][]byte{
						"<key-1>": []byte("<value-1>"),
						"<key-2>": {},
					},
				))
			})
		})
	})

	Context("flags", func() {
		Describe("func WithFlag()", func() {
			It("sets the flag", func() {
				attrs := NewAttributes().
					WithFlag("<key>")

				Expect(attrs.HasFlags("<key>")).To(BeTrue())
			})

			It("does not modify the original attributes", func() {
				attrs := NewAttributes()
				attrs.WithFlag("<key>")

				Expect(attrs.IsEmpty()).To(BeTrue())
			})

			It("replaces binary attributes with the same key", func() {
				attrs := NewAttributes().
					WithPair("<key>", []byte("<value>")).
					WithFlag("<key>")

				_, ok := attrs.Get("<key>")
				Expect(ok).To(BeFalse())
				Expect(attrs.HasFlags("<key>")).To(BeTrue())
			})

			It("has no effect if the flag is already set", func() {
				attrs := NewAttributes().
					WithFlag("<key>").
					WithFlag("<key>")

				Expect(attrs.HasFlags("<key>")).To(BeTrue())
			})

			DescribeTable(
				"it panics if the key is invalid",
				func(k, expect string) {
					Expect(func() {
						NewAttributes().
							WithFlag(k)
					}).To(PanicWith(MatchError(expect)))
				},
				Entry("empty", "", "key must not be empty"),
				Entry("contains equals", "<k=v>", "invalid key '<k=v>', key must not contain '=' character"),
				Entry("contains UTF-8", "<ключ>", "invalid key '<ключ>', key must contain only printable ASCII characters"),
				Entry("contains non-printable ASCII", "<\x00>", "invalid key '<\x00>', key must contain only printable ASCII characters"),
			)
		})

		Describe("func HasFlags()", func() {
			DescribeTable(
				"it returns the expected result",
				func(expect bool, keys ...string) {
					attrs := NewAttributes().
						WithFlag("<key-1>").
						WithFlag("<key-2>")

					Expect(attrs.HasFlags(keys...)).To(Equal(expect))
				},
				Entry("no flags", true),
				Entry("equivalent set", true, "<key-1>", "<key-2>"),
				Entry("superset", true, "<key-1>"),
				Entry("subset", false, "<key-1>", "<key-2>", "<key-3>"),
				Entry("disjoint set", false, "<key-4>", "<key-5>"),
			)

			It("returns false if the key has a binary value associated with it", func() {
				attrs := NewAttributes().
					WithPair("<key>", []byte("<value>"))

				Expect(attrs.HasFlags("<key>")).To(BeFalse())
			})

			It("is case insensitive", func() {
				attrs := NewAttributes().
					WithFlag("<KEY>")

				Expect(attrs.HasFlags("<key>")).To(BeTrue())
			})
		})

		Describe("func Flags()", func() {
			It("returns the flags that are set", func() {
				attrs := NewAttributes().
					WithFlag("<key-1>").
					WithFlag("<key-2>").
					WithPair("<key-3>", []byte("<value>"))

				Expect(attrs.Flags()).To(Equal(
					map[string]struct{}{
						"<key-1>": {},
						"<key-2>": {},
					},
				))
			})
		})

		Describe("func Without()", func() {
			It("clears flags", func() {
				attrs := NewAttributes().
					WithFlag("<key>").
					Without("<key>")

				Expect(attrs.HasFlags("<key>")).To(BeFalse())
			})

			It("does not modify the original attributes", func() {
				attrs := NewAttributes().
					WithFlag("<key>")
				attrs.Without("<key>")

				Expect(attrs.HasFlags("<key>")).To(BeTrue())
			})
		})
	})

	Describe("func IsEmpty()", func() {
		It("returns true if there are no attributes", func() {
			attrs := NewAttributes()
			Expect(attrs.IsEmpty()).To(BeTrue())

			attrs = attrs.
				WithPair("<key>", []byte("<value>")).
				WithFlag("<flag>").
				Without("<key>").
				Without("<flag>")

			Expect(attrs.IsEmpty()).To(BeTrue())
		})

		It("returns false if there are key/value pairs", func() {
			attrs := NewAttributes().
				WithPair("<key>", []byte("<value>"))

			Expect(attrs.IsEmpty()).To(BeFalse())
		})

		It("returns false if there are flags", func() {
			attrs := NewAttributes().
				WithFlag("<key>")

			Expect(attrs.IsEmpty()).To(BeFalse())
		})
	})

	Context("TXT records", func() {
		Describe("func WithTXT()", func() {
			It("parses flags", func() {
				attrs, ok, err := NewAttributes().
					WithTXT("<KEY-1>")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(ok).To(BeTrue())

				attrs, ok, err = attrs.
					WithTXT("<KEY-2>")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(ok).To(BeTrue())

				Expect(attrs.HasFlags("<key-1>")).To(BeTrue())
				Expect(attrs.HasFlags("<key-2>")).To(BeTrue())
			})

			It("parses binary attributes", func() {
				attrs, ok, err := NewAttributes().
					WithTXT("<KEY-1>=<value-1>")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(ok).To(BeTrue())

				attrs, ok, err = attrs.
					WithTXT("<KEY-2>=<value-2>")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(ok).To(BeTrue())

				Expect(attrs.Pairs()).To(Equal(
					map[string][]byte{
						"<key-1>": []byte("<value-1>"),
						"<key-2>": []byte("<value-2>"),
					},
				))
			})

			It("parses binary attributes with empty values", func() {
				attrs, ok, err := NewAttributes().
					WithTXT("<KEY>=")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(ok).To(BeTrue())

				v, ok := attrs.Get("<key>")
				Expect(v).To(BeEmpty())
				Expect(ok).To(BeTrue())
			})

			It("ignores values that start with an equals character", func() {
				attrs, ok, err := NewAttributes().
					WithTXT("=<value>")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(ok).To(BeFalse())

				Expect(attrs.Pairs()).To(BeEmpty())
				Expect(attrs.Flags()).To(BeEmpty())
			})

			It("ignores empty strings", func() {
				attrs, ok, err := NewAttributes().
					WithTXT("")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(ok).To(BeFalse())

				Expect(attrs.Pairs()).To(BeEmpty())
				Expect(attrs.Flags()).To(BeEmpty())
			})

			DescribeTable(
				"it returns an error if the key is invalid",
				func(pair, expect string) {
					_, _, err := NewAttributes().
						WithTXT(pair)
					Expect(err).To(MatchError(expect))
				},
				Entry("contains UTF-8", "<ключ>", "invalid key '<ключ>', key must contain only printable ASCII characters"),
				Entry("contains non-printable ASCII", "<\x00>", "invalid key '<\x00>', key must contain only printable ASCII characters"),
			)

			It("does not modify the original attributes", func() {
				attrs := NewAttributes()

				_, ok, err := attrs.
					WithTXT("<key>=<value>")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(ok).To(BeTrue())

				Expect(attrs.IsEmpty()).To(BeTrue())
			})
		})

		Describe("func ToTXT()", func() {
			It("returns TXT record values containing key/value pairs as per RFC 6763", func() {
				attrs := NewAttributes().
					WithFlag("<key-1>").
					WithPair("<key-2>", []byte("<value>")).
					WithPair("<key-3>", nil)

				Expect(attrs.ToTXT()).To(ConsistOf(
					"<key-1>",
					"<key-2>=<value>",
					"<key-3>=",
				))
			})

			It("always places the 'version tag' attribute at the beginning, and sorts the other keys lexically", func() {
				attrs := NewAttributes().
					WithFlag("<key-3>").
					WithPair("<key-2>", []byte("<value>")).
					WithPair("<key-1>", nil).
					WithPair("txtvers", []byte("1"))

				// Repeat test several times to ensure it's not just passing due
				// to Go's psuedo-random map ordering.
				for i := 0; i < 1000; i++ {
					Expect(attrs.ToTXT()).To(Equal(
						[]string{
							"txtvers=1",
							"<key-1>=",
							"<key-2>=<value>",
							"<key-3>",
						},
					))
				}
			})
		})
	})
})
