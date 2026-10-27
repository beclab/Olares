# `bfl`

## Overview

BFL provides backend APIs for the system frontend. One instance runs in
`os-frontend`; each request identifies its user through the request context.

Deployment resources are in `.olares/config/cluster/deploy/bfl.yaml` and are
packaged into `os-framework`. BFL does not create personal volumes or initialize
user directories.

The network controllers manage the shared entrance addresses, default domain,
certificate and reverse proxy. Application custom domains and user settings
remain associated with their respective users.
