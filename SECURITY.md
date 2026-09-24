# Security Policy

## Reporting a vulnerability

Security issues in Totipo should be reported privately.

Please **do not open a public GitHub issue, pull request, or discussion** containing details of a suspected vulnerability.

Use GitHub's **private vulnerability reporting** for this repository:

1. Open the repository's **Security** tab.
2. Select **Report a vulnerability**.
3. Submit the report through the private security advisory form.

A useful report includes, where applicable:

* the affected specification section, commit, tag, test vector, or tool;
* a description of the security impact;
* the assumptions or conditions required to reproduce the issue;
* a proof of concept or example demonstrating the issue;
* affected implementations, if known; and
* a suggested mitigation or specification change, if you have one.

Please avoid including secrets, credentials, or unrelated personal information in reports.

## What should be reported privately

Examples of issues that should be treated as security-sensitive include:

* weaknesses in the cryptographic design or key derivation;
* violations of confidentiality, integrity, or authentication properties;
* nonce, salt, or key-reuse conditions that could compromise security;
* downgrade or algorithm-confusion attacks;
* ambiguous encoding, parsing, or canonicalization rules that could produce security-relevant differences between implementations;
* malformed inputs that can bypass required validation;
* specification inconsistencies that could lead conforming implementations to behave insecurely;
* incorrect test vectors or conformance tests that could cause insecure behavior to be accepted as conforming; and
* vulnerabilities in repository tooling where exploitation could affect the specification, conformance results, generated artifacts, or project integrity.

Ordinary specification questions, editorial mistakes, interoperability problems without a security impact, and feature proposals may be reported through normal public GitHub issues.

If you are unsure whether an issue has security implications, prefer private reporting.

## Disclosure

Please allow the maintainers an opportunity to investigate and address a reported vulnerability before publishing technical details.

The project will coordinate disclosure with reporters when practical. Depending on the issue, resolution may involve changes to the specification, test vectors, conformance tests, tooling, implementations, or published security guidance.

No fixed response or remediation timeline is guaranteed.

## Project status

Totipo is under active development.

The specification, reference and conformance work, test vectors, and repository tooling should not be interpreted as having received an independent security audit unless explicitly stated otherwise.

Implementers are responsible for evaluating the security of their own implementations and the cryptographic libraries and platforms on which they depend.
