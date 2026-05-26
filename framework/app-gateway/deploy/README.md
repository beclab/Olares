# Gateway routing configuration

## Overview

This directory holds Kubernetes manifests generated during platform install or upgrade. They are applied automatically with **Olares platform install or upgrade**; end users and tenant admins do not edit these files directly.

## What customers need to know

- **Access URLs** (domain/Host) are assigned when apps or entrances are provisioned and match the generated routes automatically.
- **Path rules** stay aligned with each app's design; the platform does not add a global path prefix at the gateway layer.
- After install, upgrade, or removal, routes update through the platform workflow. If access fails, use platform support with app name, URL, and timestamp rather than editing low-level gateway config by hand.

## Support

For gateway and routing issues, see Olares platform docs and release notes, or contact support with app name, access URL, and time of failure.
