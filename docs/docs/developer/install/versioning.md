---
outline: [2, 3]
description: Version management system in Olares covering release types, branch strategies and upgrade specifications. Details about semantic versioning implementation and compatibility.
---
# Olares versioning

The Olares versioning and release process is designed to provide clear version definitions and stable upgrade paths. This document outlines Olares' versioning rules, release types, branch management practices, and upgrade guidelines.

## Versioning rules

Olares primarily follows the [Semantic versioning specification](https://semver.org/), which defines versions in format `Major.Minor.Patch[-PreReleaseVersion]`. For example, `1.11.0-rc.0` represents a release candidate for version `1.11.0`.

Versions are ordered as follows:  
  `1.0.0-alpha` < `1.0.0-alpha.1` < `1.0.0-alpha.beta` < `1.0.0-beta` < `1.0.0-beta.2` < `1.0.0-beta.11` < `1.0.0-rc.1` < `1.0.0`


## Release types

Olares offers three types of releases: **Stable**, **Release Candidate (RC)**, and **Daily build**.

### Stable releases

Stable releases are thoroughly tested versions suitable for production environments. The official one-line installation command always defaults to the latest stable version.
- **Release cadence**: Monthly
- **Examples**: `v1.10.5`, `v1.11.0`, `v1.11.1`, `v1.12.0`

### Release Candidate (RC) releases

RC releases are pre-release versions for testing before a stable release. RC releases may go through several iterations before being promoted to stable.
- **Release cadence**: Based on testing status
- **Examples**: `v1.11.0.rc.0`, `v1.11.0.rc.1`

### Daily build releases

Daily build releases, or daily builds are automatically generated from the `main` branch every day at 2:00 AM (UTC+8), with the build date embedded in the version name. These versions reflect the latest code changes and are intended for development and testing purposes. Daily builds are often unstable and not suitable for production use.
- **Release cadence**: Daily
- **Examples**: `v1.12.0-20241201`

## Release branch management

During the `1.x` phase, Olares follows a structured monthly release cadence:

1. At the end of each month, a release branch (e.g., `release-1.11`) is created from the `main` branch.

2. An initial RC version (e.g., `v1.11.0.rc.0`) is released from the new release branch. Additional RC versions may follow as testing progresses.

3. The `main` branch transitions to the next version (e.g., from `v1.11` to `v1.12`).

4. Issues identified in the stable version are addressed through patch releases (e.g., `v1.11.1`), based on the corresponding release branch.

Developers can submit pull requests (PRs) to both the `main` branch and the relevant release branch as needed.

## Upgrade policies and compatibility

Olares is in a rapid development phase with frequent feature additions and changes. To ensure stability and avoid unexpected issues, Olares follows a controlled upgrade policy:

- **Upgrades within the same minor version**:  
   
   Upgrading within the same minor version (e.g., `1.4.0` to `1.4.2`) is fully supported. These updates typically include bug fixes or small improvements that do not affect compatibility.

- **Upgrades across minor versions**:  
   
   Upgrading from one minor version to another (e.g., `1.4.x` to `1.5.x`) involves more significant changes, such as new features or architectural updates. Because these changes may not be backward-compatible, automatic upgrades are **not supported**. Instead, users must manually uninstall the existing version before installing the newer version.

The type of release you are using also determines what upgrades are possible:

- **Stable**: You can only upgrade to newer stable releases, ensuring maximum reliability and stability.
- **RC**: You can upgrade to newer RC versions or stable releases as they become available.
- **Daily build**: You can upgrade to newer daily build, RC, or stable releases.

::: tip Future upgrade policy
In the next major release, Olares plans to fully support seamless automatic upgrades for all versions within the same major version.
:::
