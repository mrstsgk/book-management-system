# CLAUDE.md

Guidance for Claude Code (claude.ai/code) and agents working in this repository.

## Language

- **AI entrypoints** (this file, `AGENTS.md`, and similar agent-only prompts): write in **English**.
- **Human-facing policy** (`docs/**`, `frontend/architecture.md`, `backend/architecture.md`, `packages/*/README.md`, PR/branch docs people read): write in **Japanese**.
- Do not duplicate long Japanese policy into this file; `@`-import or point to the Japanese source of truth.

## Most Important Rule: What Goes Where

- **Source code** writes **How** — the mechanism, the logic.
- **Test code** writes **What** — the expected behavior/spec.
- **Commit log** writes **Why** — the motivation for the change.
- **Code comments** write **Why not** — rejected alternatives, non-obvious constraints.
- Exception: the types, fields and functions listed per layer for the backend and the frontend, and any **complex function** in any layer, get a one-line **What** summary as a **doc comment** (Go `//` starting with the identifier, TSDoc `/** */`) in **Japanese** — no constraints or conditions, which the code expresses. Generated files are out of scope.

Detail (Japanese source of truth):

@docs/rules/code-comments.md

## Architecture

Docs are MECE. Repo boundaries → `docs/architecture.md`. Backend detail → `backend/architecture.md`. Frontend detail → `frontend/architecture.md` (Japanese; humans read these).

@docs/architecture.md
@backend/architecture.md
@frontend/architecture.md

## Before Changing Backend Go Code

Before generating/changing backend code, answer: (1) what is this change's purpose, (2) is it a command or a query, (3) which UseCase does it belong to, (4) which model owns the responsibility (Domain / Read Model / Persistence Model — don't blur them), (5) does it mix a new change-reason into an existing class, (6) does it needlessly share a Domain Model with a Read Model, (7) does DB/Framework detail leak into Domain, (8) can the blast radius be made smaller.

If existing code violates these, don't blindly preserve it — check compatibility with callers, then propose the more purpose-driven, lower-blast-radius design. Detail (design principles, CQRS/Command DTO/Read Model rules): `backend/architecture.md` §2.

## Design guidelines by channel (when to read)

public (the reading list anyone can view; smartphone first) follows **Apple HIG (iOS)**; admin (the owner's management screens; PC only) follows the **Digital Agency Design System**. Detail: `frontend/architecture.md` §4 (Japanese).

Apple HIG (public):

- First time building a screen structure or control (navigation, tabs, sheets, inputs) → read the relevant HIG page; put its URL in the PR
- Layout/copy only with existing components → usually not needed (read if review feedback asks for it)
- Stuck on behavior / a11y / how it looks on iPhone, or review feedback → read
- HIG: https://developer.apple.com/jp/design/human-interface-guidelines (iOS: https://developer.apple.com/jp/design/human-interface-guidelines/designing-for-ios)

Digital Agency DS (admin; `packages/ui`):

- **New** component into `packages/ui` → read official usage; put the usage URL in the PR (one line)
- Layout/copy only with existing `packages/ui` components → usually not needed (Storybook / existing use is source of truth; read if review feedback asks for it)
- Stuck on behavior / a11y, or review feedback → read usage
- Reading every component’s usage every time → **do not** (excessive)

Official: https://design.digital.go.jp/dads/ (policy: https://www.digital.go.jp/policies/servicedesign/designsystem)

## Development Workflow Rules (Team)

Before creating a PR or changing its review stage, follow the human-confirmed review classification workflow in `AGENTS.md` and `docs/rules/git-pull-request.md`. Prepare a recommendation; wait for an explicit choice or delegation before creating the PR or applying a classification label. For explicit delegation, record the delegated decision and rationale as specified in the PR policy.

Team Git conventions (Japanese human-facing docs). Production branch is `main`.

@docs/rules/development-flow.md
@docs/rules/git-branch.md
@docs/rules/git-commit-message.md
@docs/rules/git-pull-request.md

## Testing, DB Docs

Japanese sources of truth. The DB-docs rule is enforced by the Stop hook `.claude/hooks/require-db-docs-on-migration.sh` (bypass: `SKIP_DB_DOCS_CHECK=1`); missing tests are checked by the CI `Test Presence` workflow.

- Behavior-changing source changes ship with colocated tests (`*_test.go` / `*.test.ts(x)`)
- `backend/migrations/*.sql` changes ship with `docs/db/backend-schema.{json,md}` updates

@docs/rules/testing.md
@docs/rules/db-documentation.md

## Development Process (AI-assisted Agile)

Claude Code owns Engineering (How). The owner of this repository is also its customer and decides requirements (What).

### Principles

- Keep Technical Discovery (investigate) separate from Design Decisions (decide).
- Do not resolve requirement ambiguity by technical judgment alone. Report points needing the owner's decision to the user.
- Do not reuse `mock/` code as production code.

### Legacy Evidence

- Treat legacy project code only as Evidence (constraints, migration, compatibility). Never add or change requirements because of legacy behavior.

@AGENTS.md
