# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- App-related metrics (installed, available updates) are opt-in now, mirroring the change in Nextcloud 28

### Fixed

- Error parsing "free space" reported by Nextcloud when number is very large

## [0.6.2] - 2023-10-15

### Changed

- Maintenance release, updates Go and dependencies

## [0.6.1] - 2023-05-29

### Changed

- `latest` Docker tag now points to most recent release and `master` points to the build from the default branch

## [0.6.0] - 2022-10-19

### Added

- New database info metric (`nextcloud_database_info`) containing type and version
- New metrics for hourly and daily active users
- Additional label on `nextcloud_scrape_errors_total` for errors due to rate-limiting
- Additional labels for email and room shares on `nextcloud_shares_total`

## [0.5.1] - 2022-04-02

### Fixed

- Updated Prometheus client library for CVE-2022-21698

## [0.5.0] - 2022-01-15

### Added

- Flag for showing version information
- Option to disable TLS validation
- Token authentication for Nextcloud 22 and newer

### Changed

- Switched to JSON from XML for getting information from server
- Use different metric for authentication errors

## [0.4.0] - 2021-01-21

### Added

- Metrics for installed apps and available updates

## [0.3.0] - 2020-06-01

### Added

- Makefile target for building deb
- Login flow for app password

### Changed

- Simpler configuration of server URL

### Fixed

- Error in version information

## [0.2.0] - 2020-05-20

### Added

- Version information in binary
- Custom User-Agent header
- systemd service unit

### Changed

- No timestamp in log output

## [0.1.0] - 2019-10-12

- Initial release

[0.6.2]: https://github.com/xperimental/nextcloud-exporter/releases/tag/v0.6.2
[0.6.1]: https://github.com/xperimental/nextcloud-exporter/releases/tag/v0.6.1
[0.6.0]: https://github.com/xperimental/nextcloud-exporter/releases/tag/v0.6.0
[0.5.1]: https://github.com/xperimental/nextcloud-exporter/releases/tag/v0.5.1
[0.5.0]: https://github.com/xperimental/nextcloud-exporter/releases/tag/v0.5.0
[0.4.0]: https://github.com/xperimental/nextcloud-exporter/releases/tag/v0.4.0
[0.3.0]: https://github.com/xperimental/nextcloud-exporter/releases/tag/v0.3.0
[0.2.0]: https://github.com/xperimental/nextcloud-exporter/releases/tag/v0.2.0
[0.1.0]: https://github.com/xperimental/nextcloud-exporter/releases/tag/v0.1.0
