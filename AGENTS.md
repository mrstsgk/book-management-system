# AGENTS.md

Shared rules for AI agents (Claude Code, Codex, etc.) working in this repository.

## Pull request review classification

- Before creating any PR (including Draft), inspect the changes and propose one category with a concrete reason: design/specification (`review:design`, assertive), standard implementation (`review:standard`, chill), critical implementation (`review:critical`, assertive), or no review (`review:skip`, CodeRabbit auto review disabled; human approval and CI still required).
- Ask the human to choose. If the human explicitly delegates the choice (for example, "お任せ" or "assertiveで進めて"), choose within that instruction and record the delegation and rationale; do not ask again. An assertive implementation uses `review:critical`, not the design-stage exemption. Do not infer a decision from silence or a generic request to create a PR. An explicit choice already made for this PR remains valid; do not ask again unless its scope/stage changes.
- Prepare the code, tests and PR body before waiting. After the human decides, create a Draft PR, apply exactly one classification label, verify it, then mark ready if appropriate. Record the human's choice and reason in the PR body. Do not select or change classification labels without an explicit human choice or delegation.
- When a design Draft gains implementation, ask the human to reclassify it before requesting the implementation review. Do not retain the design-stage exemption for implementation/tests.
- Follow `docs/rules/git-pull-request.md` for classification, label setup, review initiation and verification. The AI that creates the PR conducts this interview; CodeRabbit does not conduct it.
- Before creating a PR for implementation, bug fixes, or refactoring after detailed design, run the local checks (`make fmt-check lint vuln test` in `backend/`, `pnpm typecheck && pnpm lint && pnpm test` in `frontend/`) and check for drift against the governing docs: Claude Code runs the `spec-drift-reviewer` agent; other agents follow the steps in `.claude/agents/spec-drift-reviewer.md`. Report drift to the human before creating the PR.

## Go

- Write code using the language and standard-library features available at the Go version in `backend/go.mod`; `make lint` (golangci-lint `modernize`) enforces this.

## Requirements

- The source of truth for requirements is `docs/requirements/status-map.md` and the GitHub Issues labeled `論点` (open questions)
  - `status-map.md`: current-position map (feature groups, releases, progress of requirements definition)
  - `論点` issues: one issue per question; open = undecided, closed = decided, label `保留` = on hold. The title starts with the question ID (`Q-xxx`)
- Tasks and progress are managed in GitHub Issues. Check the related `論点` issues
- Do not edit `status-map.md` unless the owner asks. Do not create, edit, or close `論点` issues unless the owner asks (investigation comments are allowed)
- Post investigation results for an issue as a comment on the related `論点` issue (key points and a recommendation). Only a long investigation (with evidence and sources) also goes to `docs/research/{folder}/{file}.md`, linked from the comment. Pick the folder by the rules in `docs/research/README.md`; do not invent names. A comment or note is material for a decision, not a decision. Do not change the status of the `論点` issue (labels, open/closed)
- After an investigation, report to the user where you wrote what (the issue comment, and the file if any)
- If you find an ambiguity that needs the owner's decision, do not decide it yourself; report it to the user
- End-to-end flow from requirements to release (outline design, detailed design in OpenAPI, TDD, PR): `docs/rules/development-flow.md` (Japanese)
- Outline design (概要設計) is one file per feature at `docs/design/{feature}.md`, following `docs/design/_template.md`. `{feature}` matches `mock/{feature}/`

## Mocks

Rules for building HTML mocks under `mock/` for confirming requirements with the owner.

### Positioning

- A mock is a tool for confirming requirements. It is not production code. Never reuse mock code in production (`frontend/`)
- The version the owner agreed to is the source of truth for screens (appearance, information layout, operations, wording, per-state display). Versions not yet agreed are proposals
- The source of truth for business rules is `docs/requirements/`. If you notice a contradiction or gap in business rules while building a mock, do not resolve it in the mock; report it to the user

### What to read before building

1. The target story in `docs/requirements/status-map.md` and its GitHub Issue (purpose, acceptance criteria)
2. Related `論点` issues (reflect decided ones; do not fill undecided ones by guessing)
3. Instructions from the user

### Do not start from the screen

Organize in this order before building the mock.

User Purpose → Task → Concept → State → Mock

- User Purpose: not "what should this screen make the user do" but "what does the user want to achieve". Example: not "view the book list" but "find a book worth reading next"
- Task: the actions needed to reach the purpose. Think about it before the number of screens or the screen structure
- Concept: the concepts the user understands and operates. Do not use DB table names or API DTOs as-is as UI concepts

### Location and versions

- `mock/{feature}/v{n}/index.html` (e.g. `mock/admin-screens/v1/index.html`)
- Do not overwrite old versions. Create a new version only when the agreed screens change. Minor fixes before agreement are made within the same version
- Show `{feature} Mock v{n} / YYYY-MM-DD` at the top of the screen

### States to represent

Do not build only the normal display. Represent the following states relevant to the target story.

- Data present / zero items / many items / long strings
- Operable / not operable (show the reason on screen too)
- Input error / system error

For each state, make clear what is visible, what can be pressed, and what can be done next.

### Information design and operation

- For each screen, check what to look at first, what the most important information is, and whether the next action is clear
- Make it operable beyond appearance: button press, input, selection, back, cancel, confirm, complete, error, not-operable, modal, and screen transitions

### Usage environment

- public: smartphone first. Operable with a thumb; limit the information on one screen
- admin: PC only. No smartphone/tablet layout needed

### Shared design rules

- admin screens follow `docs/rules/ui-design.md` (button roles, typography, button sizes, name length for master data) in both mocks and production. Do not reuse mock code. public screens are out of its scope (they follow Apple HIG).

### Accessibility

- public mocks follow Apple's Human Interface Guidelines (iOS); admin mocks follow the Digital Agency Design System usage guidance
- Do not distinguish buttons and links by appearance alone
- Do not indicate errors by color alone
- Give input fields labels
- Keep tap targets large enough

### How to build

- One file (HTML / CSS / JS), dummy data, no API connection is fine
- Make the main operations actually operable from start to finish
- Reduce information that is unnecessary for the purpose rather than adding more
- Use the words users use on screen. Do not expose DB or API names on screen
- For undecided parts, do not guess a single answer; let the viewer switch between 2–3 options (show the option name on screen)
- Give each input/output element `data-item="{No.}"` matching the No. in the outline design's item table. Show number badges only when the URL has `?annotate=1`. Screenshots for `docs/design/{feature}/images/` are taken with `?annotate=1`
- Input validation messages follow the four rules in `frontend/architecture.md` §4 (入力チェックのメッセージ): do not blame the user, state the format or condition, avoid system terms, and show them right below or beside the field

### Do not

- Fill undecided issues by guessing
- Change requirements (business rules, stories) on your own
- Bring the production frontend structure into the mock
- Decide the DB, API, or architecture
- Copy existing screens as-is without judgment

### Report after building

After building a mock, report the following to the user.

- Ambiguities and contradictions found while building (and whether they concern a story, a business rule, or an issue)
- Places where multiple options were provided to avoid guessing
- States that could not be represented

### After the owner's review

If the review produces a spec change, do not just fix the mock and stop. Report to the user and reconcile with the related `論点` issues (what you may edit follows the "Requirements" section).

## React

- Follow `frontend/architecture.md` (Vite SPA; no Next.js/SSR/RSC patterns). Hook and render purity rules are enforced by `pnpm lint` (`eslint-plugin-react-hooks` recommended).
