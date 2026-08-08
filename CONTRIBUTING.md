# Contributing to Fake Network Operator (FNO)

First off, thank you for considering contributing to the Fake Network Operator! It is community members like you that make FNO a great tool for simulating high-performance AI/ML network topologies in Kubernetes without physical GPU/RDMA hardware.

Please take a moment to review this document before submitting your first pull request or opening an issue.

---

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [How Can I Contribute?](#how-can-i-contribute)
  - [Reporting Bugs](#reporting-bugs)
  - [Suggesting Features](#suggesting-features)
  - [Pull Request Process](#pull-request-process)
- [Development Setup](#development-setup)
  - [Prerequisites](#prerequisites)
  - [Building the Project](#building-the-project)
  - [Running Locally on KinD](#running-locally-on-kind)
  - [Running Tests](#running-tests)
- [Code Style & Formatting](#code-style--formatting)
- [Commit Message Format](#commit-message-format)
- [Contact & Support](#contact--support)

---

## Code of Conduct

This project and everyone participating in it is governed by the [Fake Network Operator Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code. Please report unacceptable behavior to [muhmmadayanashiq@gmail.com](mailto:muhmmadayanashiq@gmail.com).

---

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please search existing issues to check if the problem has already been reported. When creating a bug report, please include as many details as possible:

1. Submit reports via [GitHub Issues](https://github.com/fake-network-operator/fake-network-operator/issues).
2. **Use a clear and descriptive title.**
3. **Describe the exact steps to reproduce the issue.**
4. **Include relevant logs, Custom Resources (CRDs), and environment details** (Kubernetes version, KinD version, Go version).
5. **Describe the observed behavior** versus **what you expected to happen**.

### Suggesting Features

Feature requests and enhancement suggestions are tracked on [GitHub Issues](https://github.com/fake-network-operator/fake-network-operator/issues).

When proposing a feature:
- Provide a clear and detailed explanation of the proposed feature.
- Explain why this feature would be useful for FNO users (e.g., simulating a new network topology, RoCE/InfiniBand device mock, or CNI plugin interop).
- Outline prospective implementation details or architecture if available.

### Pull Request Process

1. **Fork** the repository on GitHub: [https://github.com/fake-network-operator/fake-network-operator](https://github.com/fake-network-operator/fake-network-operator).
2. **Clone** your fork locally and create your feature branch:
   ```bash
   git checkout -b feature/my-amazing-feature
   ```
3. Make your code changes, ensuring they comply with our style guidelines and pass all test suites.
4. **Commit** your changes following our [Commit Message Format](#commit-message-format).
5. **Push** to your branch on GitHub and **open a Pull Request** against the `main` branch.
6. Participate in code review discussions and address any requested updates.

---

## Development Setup

### Prerequisites

To build and develop Fake Network Operator locally, ensure you have the following installed:

- **Go**: 1.23+
- **Docker**: 20.10+ (or compatible container runtime)
- **KinD (Kubernetes in Docker)**: v0.20.0+
- **Helm**: 3.x
- **kubectl**: v1.28+

### Building the Project

Build container images locally using Make:

```bash
make docker-build
```

### Running Locally on KinD

Launch a local KinD cluster, load the operator docker image, and deploy via Helm using the following single sequence:

```bash
make kind-create && make kind-load && make helm-install
```

To clean up your local development environment:

```bash
make kind-delete
```

### Running Tests

Execute unit and controller tests prior to submitting your Pull Request:

```bash
make test
```

---

## Code Style & Formatting

Fake Network Operator adheres to standard Go code guidelines and Kubernetes ecosystem conventions:

- **Formatting**: All Go files must be formatted with `gofmt` (or `goimports`).
- **Linting**: Code must pass `golangci-lint` clean of warnings.
- **Godoc**: Exported functions, types, and constants must have concise Godoc comments.

Run formatting and linters locally:

```bash
gofmt -s -w .
golangci-lint run
```

---

## Commit Message Format

We enforce [Conventional Commits](https://www.conventionalcommits.org/) format for all commit messages to ensure clear git history and automated changelog generation.

Format: `<type>(<scope>): <description>`

Allowed types:
- `feat:` — A new feature or capability
- `fix:` — A bug fix
- `docs:` — Documentation updates
- `style:` — Code formatting, whitespace, or missing semicolon changes
- `refactor:` — Code changes that neither fix bugs nor add features
- `test:` — Adding or updating test cases
- `chore:` — Maintenance tasks, build script or dependency updates

Example commit messages:
```bash
git commit -m "feat(api): add NetworkTopology CRD definition"
git commit -m "fix(controller): resolve interface indexing deadlock"
git commit -m "docs: update development setup prerequisites"
```

---

## Contact & Support

If you have questions, ideas, or need guidance on contributing, please contact the maintainer:

- **Maintainer**: Mohammad Ayan
- **Email**: [muhmmadayanashiq@gmail.com](mailto:muhmmadayanashiq@gmail.com)
- **Repository**: [https://github.com/fake-network-operator/fake-network-operator](https://github.com/fake-network-operator/fake-network-operator)
