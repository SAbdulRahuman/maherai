# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability in this project, please report it responsibly.

**DO NOT** open a public GitHub issue for security vulnerabilities.

### How to Report

1. Email: [security contact to be defined]
2. Include a detailed description of the vulnerability
3. Provide steps to reproduce if possible
4. Allow reasonable time for a fix before public disclosure

## Supported Versions

| Version | Supported |
|---------|-----------|
| Latest  | Yes       |

## Security Practices

This project follows security-first principles critical for financial
technology platforms:

- No API keys, secrets, or credentials in source code
- All dependencies regularly audited for vulnerabilities
- Input validation on all user-facing and API endpoints
- Data encryption at rest and in transit
- Market data API keys managed via Kubernetes Secrets / Vault
- Rate limiting and authentication on all public APIs
- Financial data handled with appropriate access controls
- Container images scanned for vulnerabilities before deployment
- Network policies enforced in Kubernetes clusters

## Response Timeline

- **Acknowledgment:** Within 48 hours
- **Assessment:** Within 1 week
- **Resolution:** Based on severity (Critical: 72h, High: 1 week, Medium: 2 weeks)
