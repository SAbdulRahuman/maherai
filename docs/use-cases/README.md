# Maher AI - QuantOps — Use Cases

> **Status:** Draft  
> **Last Updated:** 2026-04-06

## Overview

This directory contains all use case definitions for Maher AI - QuantOps. Each use case
follows a standardized format and is also tracked as a GitHub Issue using the
**Use Case Definition** template.

## Use Case Index

| ID | Title | Actor(s) | Status | Priority |
|----|-------|----------|--------|----------|
| UC-001 | *Trader views real-time stock dashboard* | Retail Trader | Draft | P0 |
| UC-002 | *Trader receives Maher AI buy/sell recommendation* | Retail Trader, Maher AI Agent | Draft | P0 |
| UC-003 | *System ingests and scores news sentiment* | News API, Sentiment Analyzer | Draft | P1 |
| UC-004 | *Trader configures smart alerts* | Retail Trader | Draft | P1 |
| UC-005 | *Developer integrates via API* | Fintech Developer | Draft | P2 |
| UC-006 | *Admin monitors system health* | Platform Admin | Draft | P2 |

## How to Define a Use Case

1. Open a GitHub Issue using the **Use Case Definition** template
2. Assign a unique ID (e.g., `UC-007`)
3. Follow the standard format: Actors → Preconditions → Main Flow → Postconditions
4. Link to related features and architecture decisions
5. Once approved, document the detailed version in this directory

## Use Case Diagram

```
                     ┌─────────────────────────┐
                     │   Maher AI - QuantOps   │
                     │                         │
  ┌──────────┐       │  ┌───────────────────┐  │
  │  Retail  │──────►│  │ View Dashboard    │  │
  │  Trader  │──────►│  │ Get AI Insights   │  │
  │          │──────►│  │ Configure Alerts  │  │
  └──────────┘       │  └───────────────────┘  │
                     │                         │
  ┌──────────┐       │  ┌───────────────────┐  │
  │  Fintech │──────►│  │ API Integration   │  │
  │Developer │──────►│  │ Data Streaming    │  │
  └──────────┘       │  └───────────────────┘  │
                     │                         │
  ┌──────────┐       │  ┌───────────────────┐  │
  │ Platform │──────►│  │ Monitor Health    │  │
  │  Admin   │──────►│  │ Manage Users      │  │
  └──────────┘       │  └───────────────────┘  │
                     │                         │
  ┌──────────┐       │  ┌───────────────────┐  │
  │Enterprise│──────►│  │ Custom AI Models  │  │
  │  Client  │──────►│  │ Private Deploy    │  │
  └──────────┘       │  └───────────────────┘  │
                     └─────────────────────────┘
```

## Personas

| Persona | Description | Primary Use Cases |
|---------|-------------|-------------------|
| **Retail Trader** | Individual who actively trades stocks, needs real-time data and AI insights | UC-001, UC-002, UC-004 |
| **Individual Investor** | Long-term investor who wants portfolio monitoring and trend alerts | UC-001, UC-004 |
| **Fintech Developer** | Builds trading apps, needs API access to market data and AI signals | UC-005 |
| **Platform Admin** | Manages Maher AI platform infrastructure and user accounts | UC-006 |
| **Enterprise Client** | Hedge fund or institution needing custom AI models and private deployment | Custom |

---

*Use cases will be detailed as issues are created using the Use Case Definition template.*
