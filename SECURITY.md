# Security Policy

## Supported Versions

`monime-package-go` is pre-1.0 and under active development. Security fixes are
applied to the latest released minor version and to `main`.

| Version | Supported |
| ------- | --------- |
| latest `v0.x` | ✅ |
| older `v0.x` | ❌ |

## Reporting a Vulnerability

**Please do not report security vulnerabilities through public GitHub issues,
pull requests, or discussions.**

Instead, report them privately so we can address the issue before it is publicly
disclosed:

- Use GitHub's **[Private vulnerability reporting](https://github.com/Walon-Foundation/monime-package-go/security/advisories/new)**
  (Security → Advisories → "Report a vulnerability"), or
- Email the maintainers at **pay.peeap@gmail.com** with the subject
  `SECURITY: monime-package-go`.

Please include:

- A description of the vulnerability and its impact.
- Steps to reproduce (a minimal code sample is ideal).
- Affected version(s) / commit, and your environment (Go version, OS).
- Any suggested remediation, if you have one.

## What to Expect

- We will acknowledge your report within **3 business days**.
- We will investigate and keep you informed of progress.
- Once a fix is ready, we will coordinate a release and, where appropriate,
  publish a GitHub Security Advisory crediting you (unless you prefer to remain
  anonymous).

## Scope and Handling Credentials

This SDK talks to the Monime API using your **access token** and **space id**.
A few practices that keep your integration safe:

- Never hardcode credentials. Prefer the environment variables
  `MONIME_ACCESS_TOKEN` and `MONIME_SPACE_ID`.
- Never commit credentials, `.env` files, or real request/response payloads that
  contain tokens to version control.
- Rotate any token that may have been exposed via logs, screenshots, or commits.

Vulnerabilities in the **Monime API itself** (as opposed to this SDK) should be
reported to Monime directly via <https://monime.io>.
