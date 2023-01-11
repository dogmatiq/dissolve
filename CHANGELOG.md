# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog], and this project adheres to
[Semantic Versioning].

<!-- references -->

[keep a changelog]: https://keepachangelog.com/en/1.0.0/
[semantic versioning]: https://semver.org/spec/v2.0.0.html

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

<!-- version template
## [0.0.1] - YYYY-MM-DD

### Added
### Changed
### Deprecated
### Removed
### Fixed
### Security
-->
