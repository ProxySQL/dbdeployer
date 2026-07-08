# dbdeployer Specialized Claude Code Agent Design

Date: 2026-03-31
Status: Approved for implementation planning
Primary host: Claude Code
Scope: `dbdeployer` reference implementation plus a reusable database-expertise layer

## Summary

This design defines a specialized Claude Code agent for `dbdeployer` that is execution-oriented, highly autonomous, and optimized first for:

1. test matrix design and execution
2. database correctness review and edge-case discovery

The system should help with feature development, end-to-end review, testing, documentation, and reference-manual work related to `dbdeployer`, while remaining reusable across other database-oriented projects later.

The recommended design is a two-layer system:

- a reusable database-expertise layer outside the `dbdeployer` repo
- a `dbdeployer` operating layer inside `~/dbdeployer/.claude/`

The agent is presented to the user as one primary maintainer agent, but internally it must follow enforced role-based phases rather than behaving like a free-form generic coding assistant.

## Goals

- Create a Claude Code setup that behaves like a disciplined `dbdeployer` maintainer.
- Allow high-autonomy execution inside `~/dbdeployer`.
- Prioritize verification and DB-correctness review over rapid but weak completion.
- Support both local developer-machine execution and stronger Linux-runner verification.
- Keep domain knowledge portable beyond `dbdeployer`.
- Ensure docs and reference material stay aligned with behavior changes.

## Non-Goals

- Building a large multi-agent swarm.
- Building a plugin or MCP-heavy platform in v1.
- Encoding every database fact into a single giant prompt or handbook.
- Treating live web access as the primary knowledge source.

## Requirements Chosen During Brainstorming

- Primary host: Claude Code
- Expertise source: repo + curated knowledge + live web
- Autonomy: high
- Operating model: small agent system implemented as one agent with enforced role-based phases
- Deliverable strategy: both repo-local and reusable, with repo-local value first
- Initial optimization priorities:
  1. test execution and matrix design
  2. DB correctness review and edge-case hunting
- Verification environments: both mixed local machines and a dedicated Linux runner path
- Completion policy: strict
- Knowledge placement: split between reusable external knowledge and `dbdeployer`-specific repo knowledge

## Architecture

### Layer 1: Reusable Database Expertise

This layer lives outside `~/dbdeployer`, ideally in a separate repository or managed knowledge directory, and is exposed to Claude Code through user-level assets under `~/.claude/`.

It should contain concise, maintainable knowledge files and workflows for:

- MySQL operational behavior
- PostgreSQL packaging and runtime behavior
- ProxySQL routing, admin, and runtime behavior
- cross-provider comparison notes
- version-specific pitfalls
- replication and topology edge cases
- testing heuristics and verification playbooks
- documentation and reference-writing standards

This layer is reusable across projects and should avoid `dbdeployer`-specific implementation details.

### Layer 2: dbdeployer Operating Layer

This layer lives in `~/dbdeployer/.claude/` and is versioned with the project.

It should contain:

- `CLAUDE.md` with project memory, architecture summary, command surfaces, test entrypoints, and completion rules
- focused skills for maintainer workflows
- slash commands for frequent review and verification tasks
- hooks that enforce verification and documentation discipline

This layer captures `dbdeployer` architecture and operating conventions, including provider boundaries, relevant scripts, doc locations, and repo-specific risk points.

## Execution Model

The user interacts with one primary `dbdeployer maintainer` agent. Internally, the agent must pass through fixed phases before it can declare a task complete.

The phases are:

1. task framing
2. implementation
3. DB correctness review
4. verification review
5. docs/manual sync
6. completion gate

This structure is intentional. The same agent may implement and review, but it must switch roles explicitly so that implementation assumptions are challenged before completion.

## Phase Definitions

### 1. Task Framing

The agent classifies the task before touching code:

- feature
- bug
- provider behavior change
- test-only change
- docs/manual change
- mixed change

It must also identify affected surfaces, such as:

- MySQL
- PostgreSQL
- ProxySQL
- provider registry
- CLI and flags
- sandbox templates
- docs and reference manual
- test matrix

### 2. Implementation

The agent may design and edit freely, but it must make assumptions explicit:

- version assumptions
- OS and package assumptions
- provider behavior assumptions
- expected existing test coverage

### 3. DB Correctness Review

The agent must switch from builder to adversarial reviewer and ask whether the change matches actual database behavior.

The review must explicitly check for:

- MySQL, PostgreSQL, or ProxySQL behavior mismatches
- version-specific differences
- startup and lifecycle ordering issues
- replication, authentication, routing, and packaging differences
- operator-facing edge cases such as missing binaries, port collisions, config-path differences, and partial setup failures

### 4. Verification Review

The agent selects and runs the strongest required verification path:

- fast local checks for quick iteration
- full Linux-runner validation for strict confirmation

Under the chosen strict policy, the agent may not claim completion without running the relevant checks for the change it made. If the environment prevents full verification, it must stop short of claiming completion and report the exact gap.

### 5. Docs/Manual Sync

If behavior, flags, support statements, installation flows, examples, or failure modes changed, documentation must be updated in the same task.

This includes, when relevant:

- quickstarts
- provider guides
- reference/manual pages
- examples
- caveats and operator notes

### 6. Completion Gate

Before completion, the agent must report:

- what changed
- what was verified
- what edge cases were checked
- what documentation was updated
- what residual risk remains, if any

## v1 Deliverables

Version 1 should stay narrow and operationally useful.

### Repo-Local Deliverables in `~/dbdeployer/.claude/`

- `CLAUDE.md`
- 3-4 focused skills, likely including:
  - `dbdeployer-maintainer`
  - `db-correctness-review`
  - `verification-matrix`
  - `docs-reference-sync`
- a small set of slash commands for recurring workflows
- hooks for:
  - verification-completion discipline
  - docs-update reminders on behavior-sensitive changes
  - warnings around destructive cleanup or reset actions

### Reusable Knowledge Deliverables

- MySQL notes
- PostgreSQL notes
- ProxySQL notes
- cross-provider notes
- edge-case checklists
- verification playbooks
- documentation/reference-writing guidance

The knowledge should be concise and structured. The goal is retrieval and disciplined execution, not bulk accumulation.

## Live Web Policy

Live web access is allowed and useful, but only as a supplemental source.

It should be used when facts may have changed or require verification, such as:

- upstream release behavior
- package names and installation flows
- official MySQL, PostgreSQL, or ProxySQL documentation
- issue trackers or release notes directly relevant to the task

The agent should prefer repo knowledge and curated knowledge first, then consult the web when temporal instability or missing context requires it.

## Recommended Path

### Stage 1: Repo-Local Operating System

Build the `~/dbdeployer/.claude/` layer first so Claude Code becomes a disciplined `dbdeployer` maintainer immediately.

Deliverables:

- `CLAUDE.md`
- focused skills
- a few slash commands
- basic hooks for verification and docs/test guardrails

### Stage 2: Reusable Database Expertise Layer

Extract or author the reusable cross-project database knowledge in a separate repo or managed knowledge directory and connect it to Claude Code at the user level.

Deliverables:

- concise DB notes
- edge-case checklists
- verification heuristics
- docs/reference standards

### Stage 3: Selective Automation

Only after the workflow proves useful in practice, add targeted automation such as:

- helper scripts for choosing verification paths
- stronger hooks on risky file classes
- a local retrieval helper or MCP service if a real need emerges
- automation that suggests documentation updates from changed surfaces

## Trade-Offs Considered

### Lean Repo-Local Specialist

Fastest to build and easiest to evolve, but weaker portability and weaker separation between reusable expertise and `dbdeployer`-specific rules.

### Full Multi-Agent System

Potentially stronger coverage, but too much coordination cost for v1 and too easy to over-engineer.

### Recommended Hybrid

The chosen design captures most of the practical benefit of specialization while keeping the system maintainable and reusable.

## Success Criteria

The design is successful if the resulting Claude Code setup:

- consistently runs stronger verification than a generic coding agent would
- catches DB-behavior and topology edge cases before completion
- updates docs when behavior changes
- remains usable on both local machines and a Linux verification runner
- can be extended into other DB-oriented projects without being rewritten from scratch

## Open Implementation Questions

These are implementation questions, not design blockers:

- the exact file layout under `~/dbdeployer/.claude/`
- the exact hook triggers and severity levels
- whether slash commands, skills, or both should own each workflow
- how the reusable knowledge repo is physically synchronized into the Claude user environment

These will be resolved during implementation planning.
