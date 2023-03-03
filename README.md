<div align="center">

# Dissolve

Dissolve is a [DNS-SD](https://datatracker.ietf.org/doc/html/rfc6763),
[Multicast DNS](https://datatracker.ietf.org/doc/html/rfc6762) and
Zeroconf/Bonjour toolkit for Go.

[![Documentation](https://img.shields.io/badge/go.dev-documentation-007d9c?&style=for-the-badge)](https://pkg.go.dev/github.com/dogmatiq/dissolve)
[![Latest Version](https://img.shields.io/github/tag/dogmatiq/dissolve.svg?&style=for-the-badge&label=semver)](https://github.com/dogmatiq/dissolve/releases)
[![Build Status](https://img.shields.io/github/actions/workflow/status/dogmatiq/dissolve/ci.yml?style=for-the-badge&branch=main)](https://github.com/dogmatiq/dissolve/actions/workflows/ci.yml)
[![Code Coverage](https://img.shields.io/codecov/c/github/dogmatiq/dissolve/main.svg?style=for-the-badge)](https://codecov.io/github/dogmatiq/dissolve)

</div>

- DNS-based service discovery (aka DNS-SD) is a method of using a standard set
  of DNS records to describe network services so that they may be discovered by
  clients. DNS-SD is defined in [RFC 6763].

- Multicast DNS (aka mDNS) provides a way to respond to DNS queries without the
  need for a centralised DNS server. mDNS is defined in [RFC 6762].

- The combination of these two technologies, wherein a multicast DNS server
  answers DNS-SD queries gives us Zero Configuration networking, commonly known
  as Zeroconf or Bonjour.

## Goals

- Advertise and discover services via Zeroconf/Bonjour
- Discover DNS-SD services via conventional unicast DNS
- Advertise DNS-SD services via common unicast DNS services (Route53, etc)
- Expose methods for writing custom mDNS responders
- Allow mDNS queries without CGO

<!-- references -->

[rfc 6762]: https://datatracker.ietf.org/doc/html/rfc6762
[rfc 6763]: https://datatracker.ietf.org/doc/html/rfc6763
