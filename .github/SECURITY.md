# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability in this project, please report it responsibly.

**DO NOT** open a public GitHub issue for security vulnerabilities.

### How to Report

1. Email: **security@maherai.dev** (or use GitHub's private vulnerability reporting)
2. Include a detailed description of the vulnerability
3. Provide steps to reproduce if possible
4. Include the affected component and version
5. Allow reasonable time for a fix before public disclosure

### What to Include

- Description of the vulnerability
- Steps to reproduce
- Affected version(s) and component(s)
- Potential impact assessment
- Suggested fix (if any)

## Supported Versions

| Version | Supported |
|---------|-----------|
| Latest (main) | ✅ Yes |
| Develop branch | ✅ Yes (pre-release) |
| Older releases | ❌ Upgrade to latest |

## Security Practices

This project follows security-first principles critical for financial
technology platforms:

### Code & Dependencies

- No API keys, secrets, or credentials in source code
- All dependencies regularly audited for vulnerabilities (`dependabot`)
- Container images scanned for CVEs before deployment (Trivy/Snyk)
- Third-party libraries pinned to specific versions
- SAST (Static Application Security Testing) in CI pipeline

### API & Data Protection

- Input validation on all user-facing and API endpoints
- Data encryption at rest (AES-256) and in transit (TLS 1.3)
- Rate limiting and authentication on all public APIs
- CORS, CSRF, and XSS protection headers
- SQL injection prevention via parameterized queries
- Financial data handled with appropriate access controls

### Infrastructure

- Market data API keys managed via Kubernetes Secrets / HashiCorp Vault
- Container images run as non-root users with read-only filesystems
- Network policies enforced in Kubernetes clusters
- Pod security standards (restricted) applied
- Ingress TLS termination with auto-renewed certificates (cert-manager)
- RBAC (Role-Based Access Control) for all K8s resources

### AI/ML Security

- AI model inputs validated and sanitized
- LLM prompt injection prevention
- AI decisions logged with full context for audit trail
- No PII stored in AI training data
- Model artifacts stored in encrypted object storage

### Monitoring & Audit

- All authentication events logged
- API access logs retained for 90 days
- AI recommendation audit trail (inputs → decision → output)
- Real-time alerting on suspicious patterns (brute force, unusual API usage)
- Regular security review of Grafana dashboard access

## Response Timeline

| Severity | Acknowledgment | Assessment | Resolution |
|----------|---------------|------------|------------|
| **Critical** | 24 hours | 48 hours | 72 hours |
| **High** | 48 hours | 3 days | 1 week |
| **Medium** | 1 week | 1 week | 2 weeks |
| **Low** | 1 week | 2 weeks | Next release |

## Security Compliance Goals

| Standard | Target | Phase |
|----------|--------|-------|
| OWASP Top 10 | Full compliance | Phase 3 |
| CMA (Saudi Capital Market Authority) | Financial data guidelines | Phase 3 |
| SOC 2 Type I | Audit readiness | Phase 4 |
| ISO 27001 | Alignment | Phase 4 |

## Bug Bounty

A formal bug bounty program will be considered after Phase 3 (production launch).

---

Thank you for helping keep Maher AI — QuantOps and its users safe.
