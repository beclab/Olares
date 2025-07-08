# Olares CLI

This directory contains the code for **olares-cli**, the official command-line interface for administering an **Olares** cluster. It provides a modular, pipeline-based architecture for orchestrating complex system operations. See the full [Olares CLI Documentation](https://docs.olares.com/developer/install/cli-1.12/olares-cli.html) for command reference and tutorials.

Key responsibilities include: 
- **Cluster management**: Installing, upgrading, restarting, and maintaining an Olares cluster.
- **Node management**: Adding to or removing nodes from an Olares cluster.


## Execution Model

For most of the commands, `olares-cli` is executed through a four-tier hierarchy:

```
Pipeline ➜ Module ➜ Task ➜ Action
````

### Example: `install-olares` Pipeline

```text
Pipeline: Install Olares
├── ...other modules
└── Module: Bootstrap OS
    ├── ...other tasks
    ├── Task: Check Prerequisites
    │   └── Action: run-precheck.sh
    └── Task: Configure System
        └── Action: apply-sysctl
````


## Repository layout

```text
cli/
├── cmd/                  # Cobra command definitions
│   ├── main.go           # CLI entry point
│   └── ctl/
│       ├── root.go
│       ├── os/           # OS-level maintenance commands
│       ├── node/         # Cluster node operations
│       └── gpu/          # GPU management
└── pkg/
    ├── core/
    │   ├── action/       # Re-usable action primitives
    │   ├── module/       # Module abstractions
    │   ├── pipeline/     # Pipeline abstractions
    │   └── task/         # Task abstractions
    └── pipelines/        # Pre-built pipelines
    │   ├── ...           # actual modules and tasks for various commands and components
```


## Build from source

### Prerequisites

* **Go 1.24+**
* **GoReleaser** (optional, for cross-compiling and packaging)

### Sample commands

```bash
# Clone the repo and enter the CLI folder
cd cli

# 1) Build for the host OS/ARCH
go build -o olares-cli ./cmd/main.go

# 2) Cross-compile for Linux amd64 (from macOS, for example)
GOOS=linux GOARCH=amd64 go build -o olares-cli ./cmd/main.go

# 3) Produce multi-platform artifacts (tar.gz, checksums, etc.)
goreleaser release --snapshot --clean
```

---

## Development workflow

### Add a new command

1. Create the command file in `cmd/ctl/<category>/`.
2. Define a pipeline in `pkg/pipelines/`.
3. Implement modules & tasks inside the relevant `pkg/` sub-packages.


### Test your build

1. Upload the self-built `olares-cli` binary to a machine that's running Olares.
2. Replace the existing `olares-cli` binary on the machine using `sudo cp -f olares-cli /usr/local/bin`.
3. Execute arbitrary commands using `olares-cli`
