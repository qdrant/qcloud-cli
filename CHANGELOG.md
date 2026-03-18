# Changelog

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
