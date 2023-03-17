package dnssd_test

import (
	. "github.com/dogmatiq/dissolve/dnssd"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("type Attributes", func() {
	var attrs *Attributes

	BeforeEach(func() {
		attrs = &Attributes{}
	})

	Describe("func NewAttributes()", func() {
		It("allows for a 'fluent' interface", func() {
			attrs := NewAttributes().
				Set("<key>", []byte("<value>")).
				SetFlag("<flag>")

			v, ok := attrs.Get("<key>")
			Expect(v).To(Equal([]byte("<value>")))
			Expect(ok).To(BeTrue())

			Expect(attrs.HasFlags("<flag>")).To(BeTrue())
		})
	})

	Context("binary attributes", func() {
		Describe("func Set()", func() {
			It("sets the attribute", func() {
				attrs.Set("<key>", []byte("<value>"))

				v, ok := attrs.Get("<key>")
				Expect(v).To(Equal([]byte("<value>")))
				Expect(ok).To(BeTrue())
			})

			It("replaces flags with the same key", func() {
				attrs.SetFlag("<key>")
				attrs.Set("<key>", []byte("<value>"))

				_, ok := attrs.Get("<key>")
				Expect(ok).To(BeTrue())
				Expect(attrs.HasFlags("<key>")).To(BeFalse())
			})

			It("replaces other binary attributes with the same key", func() {
				attrs.Set("<key>", []byte("<value-1>"))
				attrs.Set("<key>", []byte("<value-2>"))

				v, ok := attrs.Get("<key>")
				Expect(v).To(Equal([]byte("<value-2>")))
				Expect(ok).To(BeTrue())
			})

			It("allows association of a nil slice", func() {
				attrs.Set("<key>", nil)

				v, ok := attrs.Get("<key>")
				Expect(v).To(BeEmpty())
				Expect(ok).To(BeTrue())
			})

			It("allows association of an empty slice", func() {
				attrs.Set("<key>", []byte{})

				v, ok := attrs.Get("<key>")
				Expect(v).To(BeEmpty())
				Expect(ok).To(BeTrue())
			})

			DescribeTable(
				"it panics if the key is invalid",
				func(k, expect string) {
					Expect(func() {
						attrs.Set(k, nil)
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
				attrs.Set("<key>", []byte("<value>"))

				v, ok := attrs.Get("<key>")
				Expect(v).To(Equal([]byte("<value>")))
				Expect(ok).To(BeTrue())
			})

			It("returns false if there is no such key", func() {
				_, ok := attrs.Get("<key>")
				Expect(ok).To(BeFalse())
			})

			It("returns false the key is a flag", func() {
				attrs.SetFlag("<key>")

				_, ok := attrs.Get("<key>")
				Expect(ok).To(BeFalse())
			})

			It("is case insensitive", func() {
				attrs.Set("<KEY>", []byte("<value>"))

				v, ok := attrs.Get("<key>")
				Expect(v).To(Equal([]byte("<value>")))
				Expect(ok).To(BeTrue())
			})
		})

		Describe("func Pairs()", func() {
			It("returns the binary attributes", func() {
				attrs.Set("<key-1>", []byte("<value-1>"))
				attrs.Set("<key-2>", nil)
				attrs.SetFlag("<key-3>")

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
		Describe("func SetFlag()", func() {
			It("sets the flag", func() {
				attrs.SetFlag("<key>")
				Expect(attrs.HasFlags("<key>")).To(BeTrue())
			})

			It("replaces binary attributes with the same key", func() {
				attrs.Set("<key>", []byte("<value>"))
				attrs.SetFlag("<key>")

				_, ok := attrs.Get("<key>")
				Expect(ok).To(BeFalse())
				Expect(attrs.HasFlags("<key>")).To(BeTrue())
			})

			It("has no effect if the flag is already set", func() {
				attrs.SetFlag("<key>")
				attrs.SetFlag("<key>")
				Expect(attrs.HasFlags("<key>")).To(BeTrue())
			})

			DescribeTable(
				"it panics if the key is invalid",
				func(k, expect string) {
					Expect(func() {
						attrs.SetFlag(k)
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
					attrs.SetFlag("<key-1>")
					attrs.SetFlag("<key-2>")

					Expect(attrs.HasFlags(keys...)).To(Equal(expect))
				},
				Entry("no flags", true),
				Entry("equivalent set", true, "<key-1>", "<key-2>"),
				Entry("superset", true, "<key-1>"),
				Entry("subset", false, "<key-1>", "<key-2>", "<key-3>"),
				Entry("disjoint set", false, "<key-4>", "<key-5>"),
			)

			It("returns false if the key has a binary value associated with it", func() {
				attrs.Set("<key>", []byte("<value>"))

				Expect(attrs.HasFlags("<key>")).To(BeFalse())
			})

			It("is case insensitive", func() {
				attrs.SetFlag("<KEY>")

				Expect(attrs.HasFlags("<key>")).To(BeTrue())
			})
		})

		Describe("func Flags()", func() {
			It("returns the flags that are set", func() {
				attrs.SetFlag("<key-1>")
				attrs.SetFlag("<key-2>")
				attrs.Set("<key-3>", []byte("<value>"))

				Expect(attrs.Flags()).To(Equal(
					map[string]struct{}{
						"<key-1>": {},
						"<key-2>": {},
					},
				))
			})
		})

		Describe("func Delete()", func() {
			It("clears flags", func() {
				attrs.SetFlag("<key>")
				attrs.Delete("<key>")
				Expect(attrs.HasFlags("<key>")).To(BeFalse())
			})
		})
	})

	Describe("func IsEmpty()", func() {
		It("returns true if there are no attributes", func() {
			Expect(attrs.IsEmpty()).To(BeTrue())

			attrs.Set("<key>", []byte("<value>"))
			attrs.SetFlag("<flag>")

			attrs.Delete("<key>")
			attrs.Delete("<flag>")

			Expect(attrs.IsEmpty()).To(BeTrue())
		})

		It("returns false if there are key/value pairs", func() {
			attrs.Set("<key>", []byte("<value>"))
			Expect(attrs.IsEmpty()).To(BeFalse())
		})

		It("returns false if there are flags", func() {
			attrs.SetFlag("<key>")
			Expect(attrs.IsEmpty()).To(BeFalse())
		})
	})

	Context("TXT records", func() {
		Describe("func FromTXT()", func() {
			It("parses flags", func() {
				ok, err := attrs.FromTXT("<KEY>")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(ok).To(BeTrue())
				Expect(attrs.HasFlags("<key>")).To(BeTrue())
			})

			It("parses binary attributes", func() {
				ok, err := attrs.FromTXT("<KEY>=<value>")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(ok).To(BeTrue())

				v, ok := attrs.Get("<key>")
				Expect(v).To(Equal([]byte("<value>")))
				Expect(ok).To(BeTrue())
			})

			It("parses binary attributes with empty values", func() {
				ok, err := attrs.FromTXT("<KEY>=")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(ok).To(BeTrue())

				v, ok := attrs.Get("<key>")
				Expect(v).To(BeEmpty())
				Expect(ok).To(BeTrue())
			})

			It("ignores values that start with an equals character", func() {
				ok, err := attrs.FromTXT("=<value>")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(ok).To(BeFalse())

				Expect(attrs.Pairs()).To(BeEmpty())
				Expect(attrs.Flags()).To(BeEmpty())
			})

			It("ignores empty strings", func() {
				ok, err := attrs.FromTXT("")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(ok).To(BeFalse())

				Expect(attrs.Pairs()).To(BeEmpty())
				Expect(attrs.Flags()).To(BeEmpty())
			})

			DescribeTable(
				"it returns an error if the key is invalid",
				func(pair, expect string) {
					_, err := attrs.FromTXT(pair)
					Expect(err).To(MatchError(expect))
				},
				Entry("contains UTF-8", "<ключ>", "invalid key '<ключ>', key must contain only printable ASCII characters"),
				Entry("contains non-printable ASCII", "<\x00>", "invalid key '<\x00>', key must contain only printable ASCII characters"),
			)
		})

		Describe("func ToTXT()", func() {
			It("returns TXT record values containing key/value pairs as per RFC 6763", func() {
				attrs.SetFlag("<key-1>")
				attrs.Set("<key-2>", []byte("<value>"))
				attrs.Set("<key-3>", nil)

				Expect(attrs.ToTXT()).To(ConsistOf(
					"<key-1>",
					"<key-2>=<value>",
					"<key-3>=",
				))
			})

			It("always places the 'version tag' attribute at the beginning, and sorts the other keys lexically", func() {
				attrs.SetFlag("<key-3>")
				attrs.Set("<key-2>", []byte("<value>"))
				attrs.Set("<key-1>", nil)
				attrs.Set("txtvers", []byte("1"))

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
