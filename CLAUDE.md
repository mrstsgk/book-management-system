# CLAUDE.md

Guidance for Claude Code (claude.ai/code) and agents working in this repository.
Cursor equivalents live in `.cursor/rules/*.mdc` (`alwaysApply: true`). Rules shared with every agent are in `AGENTS.md`.

## Language

- **AI entrypoints** (this file, `AGENTS.md`, `.cursor/rules/*.mdc`, and similar agent-only prompts): write in **English**.
- **Human-facing policy** (`docs/**`, `frontend/architecture.md`, `backend/architecture.md`, `packages/*/README.md`, PR/branch docs people read): write in **Japanese**.
- Do not duplicate long Japanese policy into this file; `@`-import or point to the Japanese source of truth.
- All code comments (Go and TypeScript) are written in **Japanese**.

## Most Important Rule: What Goes Where

- **Source code** writes **How** — the mechanism, the logic.
- **Test code** writes **What** — the expected behavior/spec.
- **Commit log** writes **Why** — the motivation for the change.
- **Code comments** write **Why not** — rejected alternatives, non-obvious constraints.
- Exception: the types, fields and functions listed per layer for the backend and the frontend, and any **complex function** in any layer, get a one-line **What** summary as a **doc comment** (Go `//` starting with the identifier, TSDoc `/** */`) in **Japanese** — no constraints or conditions, which the code expresses. Generated files are out of scope.

Detail (Japanese source of truth):

@docs/rules/code-comments.md

## Architecture

Docs are MECE. Repo boundaries → `docs/architecture.md`. Backend detail → `backend/architecture.md`. Frontend detail → `frontend/architecture.md` (Japanese; humans read these). The reading order for humans is `README.md`.

@docs/architecture.md
@backend/architecture.md
@frontend/architecture.md

## Before Changing Backend Go Code

Before generating/changing backend code, answer: (1) what is this change's purpose, (2) is it a command or a query, (3) which UseCase does it belong to, (4) which model owns the responsibility (Domain / Read Model / Persistence Model — don't blur them), (5) does it mix a new change-reason into an existing class, (6) does it needlessly share a Domain Model with a Read Model, (7) does DB/Framework detail leak into Domain, (8) can the blast radius be made smaller.

If existing code violates these, don't blindly preserve it — check compatibility with callers, then propose the more purpose-driven, lower-blast-radius design. Detail (design principles, naming, CQRS/Command DTO/Read Model rules): `backend/architecture.md` §2. UseCases are named `<Verb><Subject>Usecase` (`ReadBookUsecase` / `SearchBooksUsecase` / `RegisterBookUsecase`).

## Digital Agency DS (when to read usage)

Follow this when implementing (detail: `frontend/architecture.md` §2.1, Japanese).

- **New** component into `packages/ui` → read official usage; put the usage URL in the PR (one line)
- Layout/copy only with existing `packages/ui` components → do **not** read usage (Storybook / existing use is source of truth)
- Stuck on behavior / a11y, or review feedback → read usage
- Reading every component’s usage every time → **do not** (excessive)

Official: https://design.digital.go.jp/dads/

## Development Workflow Rules (Team)

Before creating a PR or changing its review stage, follow the human-confirmed review classification workflow in `AGENTS.md` and `docs/rules/git-pull-request.md`. Prepare a recommendation; wait for an explicit choice or delegation before creating the PR or applying a classification label. For explicit delegation, record the delegated decision and rationale as specified in the PR policy.

Team Git conventions (Japanese human-facing docs). Production branch is `main`; working branches start from and merge into `develop`.

@docs/rules/development-flow.md
@docs/rules/git-branch.md
@docs/rules/git-commit-message.md
@docs/rules/git-pull-request.md

## Decision Records, Testing, DB Docs

Japanese sources of truth. Stop hooks in `.claude/hooks/` enforce the mechanically checkable parts (bypass env vars are documented in each rule); CI enforces the rest for every PR, including human-written ones.

- Non-trivial judgment calls (library/infra choice, dismissed review suggestion, root-cause decision) → `docs/adr/YYYY-MM-DD-<slug>.md` (Stop hook `require-adr-on-decision.sh`)
- Behavior-changing source changes ship with colocated tests (`*_test.go` / `*.test.ts(x)`). CI `Test Presence` fails when an API operation has no Handler test or no generated-MSW-handler test on the FE, and warns when a changed source has no test change in the same directory (`.github/scripts/check-api-tests.sh`, `warn-missing-tests.sh`)
- `backend/migrations/*.sql` changes ship with `docs/db/backend-schema.{json,md}` updates (Stop hook `require-db-docs-on-migration.sh`)
- Tests that need PostgreSQL fail (not skip) when it is unreachable: run `make db-up migrate-up` before `make test`

@docs/rules/adr.md
@docs/rules/testing.md
@docs/rules/db-documentation.md

## Development Process (AI-assisted)

Claude Code owns Engineering (How). Requirements (`docs/requirements.md`, `docs/specifications.md`) are the owner's.

- Keep Technical Discovery (investigate) separate from Design Decisions (decide).
- Do not resolve requirement ambiguity by technical judgment alone. Report points needing the owner's decision to the user.
- Do not reuse HTML mockup code (`docs/design/`) as production code.
- Treat legacy code only as evidence (constraints, compatibility). Never add or change requirements because of legacy behavior.

@AGENTS.md
