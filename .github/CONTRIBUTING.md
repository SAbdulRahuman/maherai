# Contributing to Maher AI - QuantOps

Thank you for your interest in contributing! Maher AI - QuantOps is a real-time financial
intelligence platform combining market data, news signals, and AI agents to generate
actionable insights for traders, investors, and financial institutions.

"Maher" (ماهر) means expert — we're building the AI-powered financial expert.

## How to Contribute

### 1. Propose a Feature / Module Idea
- Open an issue using the **Feature / Module Idea** template
- Identify which platform layer and domain it belongs to
- The team will review and label it during triage

### 2. Define a Use Case
- Use the **Use Case Definition** template
- Follow the actor → precondition → flow → postcondition format
- Link it to a parent idea or feature

### 3. Suggest Architecture Changes
- Use the **Architecture Decision Record (ADR)** template
- Evaluate alternatives before proposing a decision
- ADRs are discussed and reviewed before acceptance

### 4. Request a Feature
- Use the **Feature Request** template
- Write a clear user story and acceptance criteria
- Attach wireframes, Grafana screenshots, or mockups when possible

### 5. Report a Bug
- Use the **Bug Report** template
- Include reproduction steps and environment details

### 6. Submit Code
1. Fork the repository
2. Create a feature branch: `feature/your-feature-name`
3. Follow the coding standards (see below)
4. Write tests for your changes
5. Submit a Pull Request using the PR template
6. Request review from at least one maintainer

## Branch Strategy

| Branch | Purpose |
|--------|---------|
| `main` | Production-ready, protected |
| `develop` | Integration branch for next release |
| `feature/*` | New features and modules |
| `fix/*` | Bug fixes |
| `data/*` | Data pipeline changes |
| `ai/*` | AI/ML model changes |
| `infra/*` | Infrastructure & Kubernetes changes |
| `docs/*` | Documentation updates |
| `adr/*` | Architecture decision records |

## Coding Standards

- Write clean, readable, self-documenting code
- Follow language-specific style guides (Python: PEP 8, JS/TS: ESLint)
- All public APIs must be documented with OpenAPI specs
- Prometheus metrics must follow naming conventions (`maher_*`)
- Kubernetes manifests must pass linting
- Docker images must use multi-stage builds and non-root users
- Security-first: never commit API keys, secrets, or credentials

## Review Process

1. All PRs require at least **1 approval**
2. Architecture-impacting PRs require **2 approvals** + ADR reference
3. AI model changes require benchmark results attached
4. CI checks must pass before merge
5. Squash-merge to keep history clean

## Code of Conduct

We are committed to providing a welcoming and inclusive environment.
All participants must adhere to our [Code of Conduct](CODE_OF_CONDUCT.md).

## Questions?

Open a **Discussion** for general questions, brainstorming, or technical RFCs.
