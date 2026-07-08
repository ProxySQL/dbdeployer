---
paths:
  - cmd/**/*
  - providers/**/*
  - sandbox/**/*
  - ops/**/*
  - docs/**/*
  - .github/workflows/**/*
---

## Provider Surface Guidance
Review MySQL, PostgreSQL, and ProxySQL behavior as correctness-sensitive.

Check for version differences, package layout assumptions, startup ordering, auth defaults, port allocation, and replication semantics before changing behavior.

For ProxySQL work, verify the admin port and MySQL port pairing and make sure configuration changes preserve the intended routing behavior.

When behavior changes, update the relevant docs in `docs/`, `README.md`, and `CONTRIBUTING.md` alongside the code.
