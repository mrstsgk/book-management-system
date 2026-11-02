# AGENTS.md

Shared rules for AI agents (Claude Code, Cursor, Codex, etc.) working in this repository. Claude Code also reads `CLAUDE.md`; Cursor also reads `.cursor/rules/*.mdc`. Human-facing policy is Japanese under `docs/`, `backend/architecture.md` and `frontend/architecture.md`; this file only points at it.

## Pull request review classification

- Before creating any PR (including Draft), inspect the changes and propose one category with a concrete reason: design/specification (`review:design`, assertive), standard implementation (`review:standard`, chill), critical implementation (`review:critical`, assertive), or no review (`review:skip`, CodeRabbit auto review disabled; human approval and CI still required).
- Ask the human to choose. If the human explicitly delegates the choice (for example, "お任せ" or "assertiveで進めて"), choose within that instruction and record the delegation and rationale; do not ask again. An assertive implementation uses `review:critical`, not the design-stage exemption. Do not infer a decision from silence or a generic request to create a PR. An explicit choice already made for this PR remains valid; do not ask again unless its scope/stage changes.
- Prepare the code, tests and PR body before waiting. After the human decides, create a Draft PR, apply exactly one classification label, verify it, then mark ready if appropriate. Record the human's choice and reason in the PR body. Do not select or change classification labels without an explicit human choice or delegation.
- When a design Draft gains implementation, ask the human to reclassify it before requesting the implementation review. Do not retain the design-stage exemption for implementation/tests.
- Follow `docs/rules/git-pull-request.md` for classification, label setup, review initiation and verification. The AI that creates the PR conducts this interview; CodeRabbit does not conduct it.
- Before creating a PR for implementation, bug fixes, or refactoring, run the local checks (`make fmt-check lint vuln test` in `backend/` with the DB up, `pnpm typecheck && pnpm lint && pnpm format:check && pnpm test` in `frontend/`, `.github/scripts/check-api-tests.sh` at the root) and check for drift against the governing docs: Claude Code runs the `spec-drift-reviewer` agent; other agents follow the steps in `.claude/agents/spec-drift-reviewer.md`. Report drift to the human before creating the PR.
- Review findings (CodeRabbit or human) are examined adversarially before acting; one finding = one commit, and reply in the thread with the commit ID (`docs/rules/git-pull-request.md`).

## Requirements and design

- Sources of truth: `docs/requirements.md` (why, for whom), `docs/specifications.md` (what the system and the process must satisfy), per-feature design under `docs/superpowers/specs/` with HTML mockups under `docs/design/`, implementation plans under `docs/superpowers/plans/`, and decisions under `docs/adr/` (`docs/rules/adr.md`). The end-to-end flow is `docs/rules/development-flow.md` (Japanese).
- Do not resolve requirement ambiguity by technical judgment alone. If a point needs the owner's decision (wording, what to show, what to allow), report it instead of deciding.
- Do not reuse HTML mockup code (`docs/design/**/*.html`) as production code.
- Do not add or change requirements because of a legacy behavior; treat legacy code only as evidence (constraints, compatibility).

## Go

- Write code using the language and standard-library features available at the Go version in `backend/go.mod`; `make lint` (golangci-lint `modernize`) enforces this.
- Before changing backend code, run the checklist in `CLAUDE.md` §"Before Changing Backend Go Code". Naming and layer rules: `backend/architecture.md` §2.

## React

- Follow `frontend/architecture.md` (Vite SPA; no Next.js/SSR/RSC patterns). Hook and render purity rules are enforced by `pnpm lint` (`eslint-plugin-react-hooks` recommended + `@eslint-react`). Import direction, `@/` imports and named exports are enforced by ESLint too; when adding a feature directory, add it to `frontend/eslint.config.mjs`.
- Naming: `frontend/architecture.md` §9. Validation message wording: §2.2.

## Comments and tests

- Comments: `docs/rules/code-comments.md` (Why-not by default; a one-line Japanese What doc comment for the listed types/functions). All comments are Japanese.
- Tests: `docs/rules/testing.md`. CI `Test Presence` fails when an API operation has no Handler test / no generated-MSW-handler test on the FE side, and warns when a changed source has no test change in the same directory.
