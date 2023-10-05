# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).



## [Unreleased]

### Added

- Add `global.podSecurityStandards.enforced` value for PSS migration.

## [0.5.0] - 2023-07-13

### Fixed

- Add required values for pss policies

### Modified

- Add missing volumes to PSP
- Bump Golang to 1.19
- Add use of runtime/default seccomp profile.

## [0.4.0] - 2022-07-19

### Added

- Add Management ARN.

## [0.3.0] - 2022-06-30

### Added

- Added applied quota value metrics.
- Added account id to metrics.

## [0.2.0] - 2022-06-28

### Added

- Added feature to enable/disable operator. 

## [0.1.0] - 2022-06-28

[Unreleased]: https://github.com/giantswarm/aws-servicequotas-operator/compare/v0.5.0...HEAD
[0.5.0]: https://github.com/giantswarm/aws-servicequotas-operator/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/giantswarm/aws-servicequotas-operator/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/giantswarm/aws-servicequotas-operator/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/giantswarm/aws-servicequotas-operator/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/giantswarm/aws-servicequotas-operator/releases/tag/v0.1.0
