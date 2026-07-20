# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 0.1.x   | :white_check_mark: |

## Reporting a Vulnerability

If you discover a security vulnerability, please report it privately to avoid exposing it to the public.

**Do not open a public issue.**

Instead, use [GitHub private vulnerability reporting](https://github.com/sploitzberg/go-llm-project-structure/security/advisories/new). If private reporting is unavailable, contact the repository owner through a private channel listed on their GitHub profile; do not include vulnerability details in a public issue.

Please include:
- A description of the vulnerability
- Steps to reproduce the vulnerability
- Affected versions
- Any potential impact or exploit

We will:
- Acknowledge receipt of the report within 48 hours
- Provide a detailed response within 7 days
- Work with you to understand and resolve the issue
- Coordinate disclosure of the vulnerability

## Security Best Practices

This project follows security best practices including:
- Automated dependency scanning via Dependabot
- Security scanning via gosec
- Secret scanning in CI/CD pipeline
- Regular dependency updates

For more information on our security practices, see the [Guardrails & CI/CD](README.md#guardrails--cicd) section in the README.
