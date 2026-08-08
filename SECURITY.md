# Security Policy

The Fake Network Operator (FNO) project takes security issues seriously. We appreciate the efforts of security researchers and community members who work with us to identify and report vulnerabilities responsibly.

---

## Supported Versions

Only the latest release line of Fake Network Operator receives security updates and vulnerability patches.

| Version | Supported          | Notes                                  |
| ------- | ------------------ | -------------------------------------- |
| 0.1.x   | :white_check_mark: | Active release line                    |
| < 0.1.0 | :x:                | Unsupported pre-release / experimental |

---

## Reporting a Vulnerability

If you discover a security vulnerability in Fake Network Operator, **please do not open a public GitHub issue**. Instead, report it directly to the maintainer via private email:

- **Email**: [muhmmadayanashiq@gmail.com](mailto:muhmmadayanashiq@gmail.com)
- **Subject**: `[SECURITY VULNERABILITY] Fake Network Operator`

Please include the following information in your report:
- Type of issue (e.g., privilege escalation, container breakout, injection vulnerability)
- Complete steps to reproduce the issue
- Affected components, CRDs, or configurations
- Any potential mitigation or proposed fix, if available

---

## Response Timeline

We commit to handling security reports promptly according to the following timeline:

- **Acknowledgment**: Within **48 hours** of receiving your report, we will acknowledge receipt.
- **Initial Assessment**: Within **7 days**, we will complete an initial triage and risk assessment, providing feedback on validity and remediation plans.
- **Patch & Disclosure**: Fixes will be prepared privately, validated against our test suite, and published along with a security advisory.

---

## Coordinated Disclosure Policy

We follow a **Coordinated Vulnerability Disclosure** process:

1. **Private Remediation**: Once a vulnerability is validated, we develop and test a patch privately.
2. **Testing & Verification**: The fix is verified to ensure it resolves the security issue without impacting cluster operations.
3. **Public Release**: We issue a security advisory and release a patched version of FNO.
4. **Credit**: We will publicly acknowledge and credit researchers who responsibly report vulnerabilities, unless they request anonymity.

---

## Security Scope & Architectural Limitations

> [!WARNING]
> **Fake Network Operator (FNO) is a simulation tool** designed to mock physical network hardware (e.g., InfiniBand/RoCE adapters, SR-IOV interfaces, and high-speed interconnect topologies) inside Kubernetes clusters for testing, CI/CD, and control-plane benchmarking.
>
> **FNO must NOT be used as a security boundary or isolation mechanism.**
> Simulated devices and mock network interfaces do not enforce hardware-level tenant isolation, cryptographic payload protection, or physical network isolation. Deploy FNO exclusively in non-production, development, or sandboxed test environments.
