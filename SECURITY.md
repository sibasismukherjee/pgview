# Security Policy

## Supported versions

Only the latest release receives security fixes. Older versions are not patched.

| Version | Supported |
|---------|-----------|
| Latest (`v0.4.x`) | Yes |
| Older releases | No |

## Reporting a vulnerability

**Please do not open a public GitHub Issue for security vulnerabilities.**
Public disclosure before a fix is available puts all users at risk.

### Preferred channel — GitHub private vulnerability reporting

Use GitHub's built-in private reporting:

1. Go to <https://github.com/sibasismukherjee/pgview/security/advisories/new>
2. Fill in a title, description, and (if known) severity
3. Submit — only you and the maintainer can see it until it is published

This is the fastest path to a coordinated fix and public advisory.

### Alternative — maintainer profile contact

If you are unable to use GitHub's advisory flow, use the contact address listed
on the maintainer's GitHub profile: <https://github.com/sibasismukherjee>.

## What to include

A useful report covers:

- **Affected version(s)** — output of `pgview -version`
- **Description** — what the vulnerability is and how it can be exploited
- **Reproduction steps** — the minimal steps or command line needed to trigger it
- **Impact** — what an attacker could achieve (data exposure, code execution, etc.)
- **Fix suggestion** (optional) — if you have one

## Response timeline

| Milestone | Target |
|-----------|--------|
| Acknowledgement | Within 3 business days |
| Triage and severity assessment | Within 7 days |
| Patch or mitigation published | Within 30 days for critical/high; 90 days for lower severity |
| Public advisory | At the same time as or shortly after the fix |

If a deadline cannot be met, the maintainer will communicate the delay and a
revised timeline before the original deadline passes.

## Scope

In scope for this policy:

- The `pgview` CLI binary and all Go source code in this repository
- Connection handling and credential management (`internal/db/`)
- SQL injection via filter DSL, row editor, or any user-supplied input that
  reaches a PostgreSQL query

Out of scope:

- Vulnerabilities in the PostgreSQL server itself
- Third-party dependencies (report those upstream; you may note them here so
  pgview can update the dependency)
- Issues that require physical access to the user's machine

## Disclosure policy

pgview follows **coordinated disclosure**: the maintainer and reporter work
together on a fix before any public announcement. Once a fix is released, a
GitHub Security Advisory will be published with full details and credit to the
reporter (unless they prefer to remain anonymous).

## Credits

Reporters who responsibly disclose valid vulnerabilities will be credited in the
Security Advisory and in the release changelog entry, with their preferred name
or handle.
