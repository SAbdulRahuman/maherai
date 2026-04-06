# Contributing to Maher AI — QuantOps

Thank you for your interest in contributing! Maher AI — QuantOps is a real-time financial
intelligence platform combining market data, news signals, and AI agents to generate
actionable insights for traders, investors, and financial institutions.

"Maher" (ماهر) means **expert** — we're building the AI-powered financial expert.

## Quick Links

- [Architecture Overview](../docs/architecture/README.md)
- [Development Guide](../docs/development/README.md)
- [API Design](../docs/api/README.md)
- [Roadmap](../docs/roadmap/README.md)
- [Code of Conduct](CODE_OF_CONDUCT.md)

## How to Contribute

### 1. Set Up Your Development Environment

Follow the [Development Guide](../docs/development/README.md) to get started locally.

### 2. Find Something to Work On

- Browse [open issues](../../issues) labelled `good-first-issue` or `help-wanted`
- Check the [Roadmap](../docs/roadmap/README.md) for upcoming work
- Join [Discussions](../../discussions) to propose ideas

### 3. Propose a Feature / Module Idea
- Open an issue using the **Feature / Module Idea** template
- Identify which platform layer and domain it belongs to
- The team will review and label it during triage

### 4. Define a Use Case
- Use the **Use Case Definition** template
- Follow the actor → precondition → flow → postcondition format
- Link it to a parent idea or feature

### 5. Suggest Architecture Changes
- Use the **Architecture Decision Record (ADR)** template
- Evaluate alternatives before proposing a decision
- ADRs are discussed and reviewed before acceptance

### 6. Request a Feature
- Use the **Feature Request** template
- Write a clear user story and acceptance criteria
- Attach wireframes, Grafana screenshots, or mockups when possible

### 7. Report a Bug
- Use the **Bug Report** template
- Include reproduction steps and environment details

### 8. Submit Code

```bash
# 1. Fork the repository
# 2. Clone your fork
git clone https://github.com/<your-username>/maherai.git
cd maherai

# 3. Add upstream remote
git remote add upstream https://github.com/seenimoa/maherai.git

# 4. Create a feature branch
git checkout -b feature/your-feature-name

# 5. Make your changes, write tests
# 6. Run linting and tests
make lint
make test

# 7. Commit with conventional commit message
git commit -m "feat(ai): add confidence scoring to Maher recommendations"

# 8. Push and create a Pull Request
git push origin feature/your-feature-name
```

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
| `release/*` | Release candidates |

## Commit Convention

We use [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

Types: feat, fix, docs, test, chore, refactor, perf, ci, build
Scopes: ai, market, api, dashboard, infra, sentiment, alerts, gateway
```

Examples:
```
feat(ai): add confidence scoring to Maher recommendations
fix(market): handle NSE API rate limit errors gracefully
docs(api): add OpenAPI spec for insights endpoint
test(sentiment): add unit tests for sentiment analyzer
ci(workflows): add container image security scanning
```

## Coding Standards

- Write clean, readable, self-documenting code
- **Python:** PEP 8, type hints, Google docstrings (`ruff` + `black` + `mypy`)
- **TypeScript:** ESLint + Prettier, strict mode, functional React components
- All public APIs must be documented with OpenAPI specs
- Prometheus metrics must follow naming conventions (`maher_*`)
- Kubernetes manifests must pass linting
- Docker images must use multi-stage builds and non-root users
- **Never commit API keys, secrets, or credentials**

See the full [Development Guide](../docs/development/README.md) for detailed standards.

## Review Process

1. All PRs require at least **1 approval**
2. Architecture-impacting PRs require **2 approvals** + ADR reference
3. AI model changes require benchmark results attached
4. CI checks must pass before merge
5. Squash-merge to keep history clean
6. PRs should be small and focused (< 500 lines ideally)

## Testing Requirements

- All new code must include tests
- Unit test coverage minimum: 80% (AI), 85% (API), 70% (frontend)
- Integration tests for cross-service interactions
- See [Testing section](../docs/development/README.md#testing) in Development Guide

## Code of Conduct

All participants must adhere to our [Code of Conduct](CODE_OF_CONDUCT.md).

## Questions?

- Open a **Discussion** for general questions, brainstorming, or technical RFCs
- Check existing [Discussions](../../discussions) before creating a new one
- Tag issues with appropriate labels for visibility
