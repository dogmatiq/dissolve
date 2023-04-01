# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog], and this project adheres to
[Semantic Versioning].

<!-- references -->

[keep a changelog]: https://keepachangelog.com/en/1.0.0/
[semantic versioning]: https://semver.org/spec/v2.0.0.html

## [Unreleased]

### Added

- Added `dnssd.RelativeServiceInstanceName()`
- Added `dnssd.ServiceInstanceName` struct (this name was previously used for a function)

### Changed

- **[BC]** Renamed `dnssd.ServiceInstanceName()` to `AbsoluteServiceInstanceName()`
- **[BC]** `Name`, `ServiceType` and `Domain` fields in `dnssd.ServiceInstance` are now provided by embedding `ServiceInstanceName`

## [0.2.0] - 2023-03-17

### Added

- Added `dnssd.ServiceInstance.Equal()` method
- Added `dnssd.Attributes.Equal()` method
- Added `dnssd.AttributeCollectionsEqual()` function

### Changed

- **[BC]** Renamed `dnssd.ServiceInstance.Instance` to `Name`
- **[BC]** `dnssd.Attributes` now presents an immutable interface
- **[BC]** Renamed `dnssd.Attributes.Set()` to `WithPair()`
- **[BC]** Renamed `dnssd.Attributes.SetFlag()` to `WithFlag()`
- **[BC]** Renamed `dnssd.Attributes.Delete()` to `Without()`

## [0.1.3] - 2023-03-17

### Added

- Added `dnssd.NewAttributes()`

### Changed

- `dnssd.Attributes.Set()`, `SetFlag()` and `Delete()` now return the attribute set, allowing for a "fluent" interface

## [0.1.2] - 2023-01-11

### Changed

- Change `dnssd.Attributes.ToTXT()` to return values in a deterministic order

## [0.1.1] - 2023-01-03

### Fixed

- Add an empty string value to empty `TXT` records, as per https://www.rfc-editor.org/rfc/rfc6763#section-6.1

## [0.1.0] - 2022-08-18

- Initial release

<!-- references -->

[unreleased]: https://github.com/dogmatiq/dissolve
[0.1.0]: https://github.com/dogmatiq/dissolve/releases/tag/v0.1.0
[0.1.1]: https://github.com/dogmatiq/dissolve/releases/tag/v0.1.1
[0.1.2]: https://github.com/dogmatiq/dissolve/releases/tag/v0.1.2
[0.1.3]: https://github.com/dogmatiq/dissolve/releases/tag/v0.1.3
[0.2.0]: https://github.com/dogmatiq/dissolve/releases/tag/v0.2.0

<!-- version template
## [0.0.1] - YYYY-MM-DD

### Added
### Changed
### Deprecated
### Removed
### Fixed
### Security
-->
