# Changelog

## [v0.12.0](https://github.com/qdrant/qcloud-cli/releases/tag/v0.12.0)

### Features

- add the option to use a helper command to retrieve api key securely instead of storing it in plain text (#59)

### Bug Fixes

- readme quickstart commands (#61)

## [v0.11.2](https://github.com/qdrant/qcloud-cli/releases/tag/v0.11.2)

### Bug Fixes

- poll immediately when waiting for cluster health (#56)

## [v0.11.1](https://github.com/qdrant/qcloud-cli/releases/tag/v0.11.1)

### Bug Fixes

- fix scale confirm prompt line order (#49)
- change ParseMillicores to use math.Round instead of truncation using a cast (#51)
- avoid double -dev in the version variable when building with make (#53)
- use getter methods to retrieve phase when scaling a cluster for nil safety (#54)

## [v0.11.0](https://github.com/qdrant/qcloud-cli/releases/tag/v0.11.0)

### Features

- add --multi-az to the cluster scale command (#45)

### Bug Fixes

- stricter context setting (#47)

## [v0.10.2](https://github.com/qdrant/qcloud-cli/releases/tag/v0.10.2)

### Bug Fixes

- add installation instructions in the readme, change the releases with a format that allows to download latest without the version in the filename (#43)

## [v0.10.1](https://github.com/qdrant/qcloud-cli/releases/tag/v0.10.1)

### Bug Fixes

- rework --label on cluster update to add/update and remove instead (#40)
- change the allow-ips to allow-ip, make it behave like the --label (#42)

## [v0.10.0](https://github.com/qdrant/qcloud-cli/releases/tag/v0.10.0)

### Features

- add remaining database configuration cluster update (#38)

## [v0.9.0](https://github.com/qdrant/qcloud-cli/releases/tag/v0.9.0)

### Features

- add --disk-performance flag to the 'cluster create' and 'cluster scale' commands (#35)

## [v0.8.0](https://github.com/qdrant/qcloud-cli/releases/tag/v0.8.0)

### Features

- cluster scale command (#33)

## [v0.7.1](https://github.com/qdrant/qcloud-cli/releases/tag/v0.7.1)

### Bug Fixes

- add command completion for cloud-provider and cloud-region flags for packages (#30)
- parse resources into a full struct so that maintaining resource quantities is easier (#32)

## [v0.7.0](https://github.com/qdrant/qcloud-cli/releases/tag/v0.7.0)

### Features

- move cloud-provider and cloud-region subcommands as root commands (#27)

## [v0.6.1](https://github.com/qdrant/qcloud-cli/releases/tag/v0.6.1)

### Bug Fixes

- package upgrade simplification (#24)

## [v0.6.0](https://github.com/qdrant/qcloud-cli/releases/tag/v0.6.0)

### Features

- add --gpu flag on cluster create that filters packages properly (#22)

## [v0.5.0](https://github.com/qdrant/qcloud-cli/releases/tag/v0.5.0)

### Features

- add self-uprade command to autoinstall releases (this won't work until the repo is public) (#20)

## [v0.4.1](https://github.com/qdrant/qcloud-cli/releases/tag/v0.4.1)

### Bug Fixes

- normalize resources to allow users to set cpu and ram in an easier (#16)
- package matching now uses cpu and ram (#18)

## [v0.4.0](https://github.com/qdrant/qcloud-cli/releases/tag/v0.4.0)

### Features

- add support for gpu and multi-az (#13)

### Bug Fixes

- clarify quick start commands for creating a cluster and hint the user to create a key (#15)

## [v0.3.1](https://github.com/qdrant/qcloud-cli/releases/tag/v0.3.1)

### Bug Fixes

- checkout so that it fetches tags in the build workflow (#10)
- correctly match dist files in the build workflow and fetch 0 to retrieve tag information so that goreleaser can see tags (#12)

## [v0.3.0](https://github.com/qdrant/qcloud-cli/releases/tag/v0.3.0)

### Features

- backup schedule commands (#5)

## [v0.2.0](https://github.com/qdrant/qcloud-cli/releases/tag/v0.2.0)

### Features

- add backup management commands (#1)
