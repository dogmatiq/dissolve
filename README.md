<div align="center">

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="https://dogmatiq.io/res/img/dynamic/repo/dissolve/banner/dark.svg">
  <img alt="Dogmatiq dissolve" src="https://dogmatiq.io/res/img/dynamic/repo/dissolve/banner/light.svg" height="44">
</picture>

Dissolve is a [DNS-SD], [Multicast DNS] and Zeroconf/Bonjour toolkit for Go.

[![API reference](https://dogmatiq.io/res/img/dynamic/repo/dissolve/badge/docs.svg)](https://pkg.go.dev/github.com/dogmatiq/dissolve)
[![Latest version](https://dogmatiq.io/res/img/dynamic/repo/dissolve/badge/semver.svg)](https://github.com/dogmatiq/dissolve/releases)
[![Build status](https://dogmatiq.io/res/img/dynamic/repo/dissolve/badge/build.svg)](https://github.com/dogmatiq/dissolve/actions/workflows/ci.yml)
[![Code coverage](https://dogmatiq.io/res/img/dynamic/repo/dissolve/badge/coverage.svg)](https://codecov.io/github/dogmatiq/dissolve)

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

[multicast dns]: https://www.rfc-editor.org/rfc/rfc6762
[rfc 6762]: https://www.rfc-editor.org/rfc/rfc6762
[dns-sd]: https://www.rfc-editor.org/rfc/rfc6763
[rfc 6763]: https://www.rfc-editor.org/rfc/rfc6763
