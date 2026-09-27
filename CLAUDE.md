# CLAUDE.md

Guidance for Claude Code (claude.ai/code) and agents working in this repository.
Cursor equivalents live in `.cursor/rules/*.mdc` (`alwaysApply: true`).

## Language

- **AI entrypoints** (this file, `.cursor/rules/*.mdc`, and similar agent-only prompts): write in **English**.
- **Human-facing policy** (`docs/**`, `frontend/architecture.md`, `backend/architecture.md`, `packages/*/README.md`, PR/branch docs people read): write in **Japanese**.
- Do not duplicate long Japanese policy into this file; `@`-import or point to the Japanese source of truth.

## Most Important Rule: What Goes Where

- **Source code** writes **How** — the mechanism, the logic.
- **Test code** writes **What** — the expected behavior/spec.
- **Commit log** writes **Why** — the motivation for the change.
- **Code comments** write **Why not** — rejected alternatives, non-obvious constraints.
- Two exceptions write **What** instead, and — unlike other code comments — are written in **Japanese**:
  1. A **Repository** (the write-side persistence port and its concrete implementation) always gets a one-line What comment on the type/class itself and on each of its methods. This is the only class-level comment requirement — classes elsewhere follow the Why-not default above.
  2. Any **complex method**, in any layer (non-obvious control flow, several branches, a non-trivial algorithm), gets a one-line What comment, regardless of whether it's a Repository.

## Architecture

Docs are MECE. Repo boundaries → `docs/architecture.md`. Backend detail → `backend/architecture.md`. Frontend detail → `frontend/architecture.md` (Japanese; humans read these).

@docs/architecture.md
@backend/architecture.md
@frontend/architecture.md

## Before Changing Backend Go Code

Before generating/changing backend code, answer: (1) what is this change's purpose, (2) is it a command or a query, (3) which UseCase does it belong to, (4) which model owns the responsibility (Domain / Read Model / Persistence Model — don't blur them), (5) does it mix a new change-reason into an existing class, (6) does it needlessly share a Domain Model with a Read Model, (7) does DB/Framework detail leak into Domain, (8) can the blast radius be made smaller.

If existing code violates these, don't blindly preserve it — check compatibility with callers, then propose the more purpose-driven, lower-blast-radius design. Detail (design principles, CQRS/Command DTO/Read Model rules): `backend/architecture.md` §2.

## Digital Agency DS (when to read usage)

Follow this when implementing (detail: `frontend/architecture.md` §2.1, Japanese).

- **New** component into `packages/ui` → read official usage; put the usage URL in the PR (one line)
- Layout/copy only with existing `packages/ui` components → do **not** read usage (Storybook / existing use is source of truth)
- Stuck on behavior / a11y, or review feedback → read usage
- Reading every component’s usage every time → **do not** (excessive)

Official: https://design.digital.go.jp/dads/

## Development Workflow Rules (Team)

Team Git conventions (Japanese human-facing docs). Production branch is `main`.

@docs/rules/git-branch.md
@docs/rules/git-commit-message.md
@docs/rules/git-pull-request.md

## Decision Records, Testing, DB Docs

Japanese sources of truth. Stop hooks in `.claude/hooks/` enforce the mechanically checkable parts (bypass env vars are documented in each rule).

- Non-trivial judgment calls (library/infra choice, dismissed review suggestion, root-cause decision) → `docs/adr/YYYY-MM-DD-<slug>.md`
- Behavior-changing source changes ship with colocated tests (`*_test.go` / `*.test.ts(x)`)
- `backend/migrations/*.sql` changes ship with `docs/db/backend-schema.{json,md}` updates

@docs/rules/adr.md
@docs/rules/testing.md
@docs/rules/db-documentation.md
