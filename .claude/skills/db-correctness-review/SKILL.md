---
name: db-correctness-review
description: Adversarial MySQL/PostgreSQL/ProxySQL review for dbdeployer changes.
disable-model-invocation: true
---

# db correctness review

Invoke `/db-core-expertise` first if it is available, then review the change for database correctness.

## Checklist

- Database semantics
- Lifecycle behavior
- Packaging and environment assumptions
- Topology and routing behavior
- Operator edge cases

## Findings format

- Correctness Risks
- Edge Cases Checked
- Recommended Follow-up
