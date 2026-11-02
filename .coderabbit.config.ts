import { defineConfig, type CodeRabbitConfig } from "@coderabbitai/config";

// Keep existing review rules when switching profiles.
const baseConfig = {
  "language": "ja-JP",
  "reviews": {
    "profile": "chill",
    "auto_review": {
      "enabled": true,
      "drafts": false,
      "base_branches": [
        "develop",
        "main"
      ]
    },
    "path_filters": [
      "!**/node_modules/**",
      "!**/dist/**",
      "!**/storybook-static/**",
      "!**/pnpm-lock.yaml",
      "!**/package-lock.json",
      "!**/go.sum",
      "!backend/api/gen/**",
      "!frontend/*/src/api/generated/**"
    ],
    "path_instructions": [
      {
        "path": "**",
        "instructions": `# General review

Branch, commit, and PR conventions follow \`docs/rules/git-branch.md\` / \`docs/rules/git-commit-message.md\` / \`docs/rules/git-pull-request.md\` (Japanese, source of truth). Write review comments in Japanese.

## Commit messages
- \`<type>: <subject>\` (type is feat/fix/chore/docs/refactor/test/hotfix; subject is at most 50 characters, in Japanese, present tense, no trailing period)
- Mapping between type and branch prefix: feat→\`feat/\`, fix→\`fix/\`, chore→\`chore/\`, docs→\`docs/\`, refactor→\`refactor/\`, hotfix→\`hotfix/\`.
  \`test\` is the exception with no branch prefix (do not use a \`test/\` prefix; bundle added tests into the target feature's \`feat/\`/\`fix/\` branch and use \`test:\` only for the commit)

## Branches (Git Flow)
- Working branches (\`feat/\`, \`fix/\`, \`refactor/\`, \`chore/\`, \`docs/\`) must branch from \`develop\` and open PRs to \`develop\`
- PRs to \`main\` are only releases from \`develop\` or \`hotfix/*\`. \`hotfix/*\` branches from \`main\` and is applied to both \`main\` and \`develop\`
- Branch names must not contain an issue number (names like \`feat/issue-83-xxx\` are to be flagged)

## Security
- Are secrets (API keys, passwords, tokens, DB connection strings, signing keys, etc.) included in logs, error responses, or commits?
- Are SQL values bound via placeholders/ORM? Is any query built by string concatenation (SQL injection)?
- Is external input embedded into HTML/shell commands without escaping (XSS, command injection)?
- Does a new endpoint unintentionally bypass authentication/authorization (admin channel: session cookie + same-origin check)?
- Is a user-supplied ID used as-is to access another user's resource (IDOR, missing authorization)? Check whether an ownership check exists
- Does the CORS config allow a wildcard (\`*\`) together with credentialed requests (cookies, etc.)?
- Are internal error messages or stack traces returned to the client as-is (exposure of internals)?

## Performance
- Is a DB query called individually inside a loop (N+1)? Check whether it can be replaced with a bulk fetch (JOIN / IN clause / batch)
- Does a list API allow a query with no upper bound (no LIMIT / page-size cap)?
- Are only the needed fields fetched and returned (fetching everything with a needlessly large payload)?
- Does the request hot path run heavy synchronous work (large file I/O, synchronous external API calls) serially? Check for room to make it async, cache it, or batch it

## Design principles (SOLID / KISS / YAGNI / DRY)
- Single responsibility: does one function/class carry so much that it changes for several different reasons?
- Open-closed: can new behavior be added by extension (adding a new implementation) rather than by bolting more if-branches onto existing code?
- Liskov substitution: does an interface implementation/subtype break the contract (how errors are returned, handling of null, strengthening preconditions / weakening postconditions)?
- Interface segregation: is an interface bloated so that implementers are forced to implement methods the caller does not use?
- Dependency inversion: do upper layers depend on abstractions (interfaces) defined in the domain layer rather than directly on concrete lower-layer implementations? (details in the \`backend/**\` section)
- KISS: is an abstraction/generalization that is not needed now added (unused extension points, excessive design patterns)?
- YAGNI: is something added for the future that is not requested now (unused flags, generic functions called from nowhere, empty domain packages)?
- Reinvention: is something hand-written that the standard library, an already-installed dependency, or a native platform feature (HTML input types, CSS, DB constraints) already provides?
- Dependencies: is a new dependency added for what a few lines of code could do?
- Dead flexibility: are there config values, options, or parameters that never change or that no caller uses?
- Exception: interfaces required by \`backend/architecture.md\` (Domain Repository/Query/ExternalGateway IF, UseCase IF + \`Impl\`) are intentional even with a single implementation; do not flag them under KISS/YAGNI.
- DRY: is the same logic copy-pasted in several places? Also check that over-sharing has not hurt readability or changeability

## Shell script pitfalls
- Under \`set -e\`, an arithmetic increment such as \`((var++))\` returns a non-zero exit status when the result happens to be 0 and stops the script unintentionally. Use \`var=$((var + 1))\`
- Code that deletes/lists using a path or prefix taken from an environment variable must validate at startup that it is non-empty and not the root path, so that an empty string or root path does not make it target everything
`
      },
      {
        "path": "backend/**",
        "instructions": `Go / Echo monolith. The source of truth for policy is \`backend/architecture.md\` (Japanese). Do not duplicate findings already covered by mechanical checks such as \`go vet\`.

## Architecture (onion + tactical DDD + CQRS)
Layout is by business area: \`internal/<area>/{domain,usecase/{command,query},infrastructure/{postgres,gateway},presentation/http/<channel>}/\` (areas: book, tag, auth; channels: public, admin; technical foundation in \`internal/platform/\`, shared errors in \`internal/domainerr/\`). Flag the following boundary violations:
- Does code import another area's \`infrastructure\` or Persistence Model directly? Is business logic placed in \`platform\` / \`domainerr\`? Are business rules duplicated per channel (e.g. \`internal/admin/domain\`)?
- **Keep Handlers thin**: only conversion generated type → UseCase → generated type (input validation is already done by the OpenAPI validation middleware). Is business logic or **authorization** written in the Handler (presentation)? (authorization is the UseCase's responsibility)
- **Error → HTTP status mapping is centralized in the shared \`httpserver.HTTPErrorHandler\`**. Are \`if errors.Is(...)\` status branches scattered across Handlers? When a sentinel error is added to \`domainerr\`, is the \`HTTPErrorHandler\` mapping updated too?
- Are input constraints hand-written in a Handler? (constraints go in the \`openapi/\` YAML and are left to the validation middleware)
- \`domain\` depends on no other layer. Does domain import infrastructure / presentation / GORM?
- Does a UseCase depend on the domain's **Repository IF (write) / Query IF (read)** and not import \`infrastructure/postgres\` directly? Are writes in \`<area>/usecase/command\` and reads in \`<area>/usecase/query\`?
- Are business constraints such as character limits guaranteed by a domain VO as well, not only by OpenAPI constraints?
- **Does a Repository / Query implementation contain business rules?** It should only hide persistence details and move domain objects in and out
- **Is the domain model anemic?** Do invariants that Entities / VOs should own leak into UseCases?
- **The transaction boundary belongs to the Command (UseCase)**. Are updates needing strict consistency (a book and its tags, optimistic lock by version) in a single transaction? Does an individual Repository implementation commit on its own?
- Is business logic duplicated per channel (\`/api\` public / \`/api/admin\` admin)? \`cmd/public-api\` and \`cmd/admin-api\` should differ only in wiring and public routes
- DI is hand-written in \`cmd/*/main.go\` (config → platform → infrastructure → usecase → presentation). When a new UseCase / Handler is added, is the wiring added too?
- Is an error response shape other than the common \`ErrorResponse\` invented?
- In a PR that changes \`openapi/\` YAML, are \`openapi/dist/**\`, \`backend/api/gen/**\` (\`make -C backend openapi\`), and the FE generated output \`frontend/*/src/api/generated/**\` (\`pnpm -C frontend gen:api\`) regenerated (\`docs/architecture.md\` §4)? \`openapi/dist/**\`, \`backend/api/gen/**\`, and \`frontend/*/src/api/generated/**\` are generated, so flag manual edits. Do UseCase / Domain / Infrastructure import \`api/gen\`?
- Does binding from the body in an HTTP handler overwrite values such as IDs that come from path parameters?

## Changeability principles (\`backend/architecture.md\` §2 design principles / CQRS)
- **Purpose-driven naming, one class one purpose**: does a name express purpose rather than existence? Is a single struct / UseCase mixing multiple reasons to change? Flag vague shared names such as \`Common\` / \`Util\` / \`Manager\` / \`Helper\` / \`BaseService\`, and forced merging of processing that changes for different reasons
- **Do not conflate the three models**: Domain Model (Entity / VO in \`<area>/domain\`) / Read Model (read-only; \`XxxReadModel\` / \`XxxView\` / \`XxxListItem\` etc.) / Persistence Model (the \`gorm\`-tagged model in \`<area>/infrastructure/postgres\`). Is a DB table structure or GORM model used as-is as a Domain Model? Does a Read Model return write-side Entity / VO types, or is the same struct reused for both update and read?
- **Domain does not depend on external technology**: does domain reference Echo / Request·Response / GORM / SQL / DB-specific types / external API SDKs? Is external communication done via Repository / Query / ExternalGateway IFs (defined in domain)? Are business rules placed in Controllers, DTOs, or ORM models?
- **Writes use a Command DTO**: is the boundary passed as a Command DTO from \`<area>/usecase/command\`, converted from DTO to domain model inside the UseCase before being handed to the Repository? Are Request structs, ORM Entities, or \`echo.Context\` passed to UseCase / Domain? Does a Command DTO contain business logic (validation, state transitions, DB / external API access)?
- **Reads center on Read Models**: do they return a shape optimized for display/search? Are write-side DTOs or domain models forcibly reused?
- **Confine business rules to VOs and Entities**: is the same range check (e.g. \`if rating >= 1 && rating <= 5\`) scattered across several call sites? Does the VO constructor make it impossible to create an invalid state?
- **External API integration** (openBD): is it called from a UseCase via the domain's ExternalGateway IF, with the implementation in \`<area>/infrastructure/gateway\`? Do UseCase / Domain depend directly on a concrete HTTP client?
- **Is a DB Entity (Persistence Model) used as an API response?** Is an ORM Entity used as a UseCase input?
- **Is an existing design that violates these principles followed unconditionally?** After checking compatibility with callers, propose a design with a clear purpose and a small blast radius
- Blast radius: does a single change ripple into unrelated features (e.g. adding a reason to change to a shared struct)? Propose a more purpose-aligned design with a smaller blast radius

## Persistence / migrations
- Is GORM AutoMigrate used? (the source of truth for schema is \`backend/migrations/*.sql\`, golang-migrate). Is migration run at application startup?

## Tests
- Require \`*_test.go\` in the same directory for behavior-changing changes. Check that they verify boundary values (both exactly-at-limit and over-limit), that the target's state does not change on failure, and that lower-layer errors propagate (see \`docs/rules/testing.md\`)
- Does each layer verify the right thing? Domain = exceptions / invariants / state transitions / boundary values (no dependency on DB or HTTP) / UseCase = use-case scenarios (with Fakes injected) / Adapter = contract tests for ExternalGateway (external-service mock) and only for Repository / Query with complex SQL (JOIN, subquery, aggregation, locking or multi-statement transactions, DB-specific features); do not ask for contract tests on simple reads/writes / Presentation (\`*_integration_test.go\`, the only HTTP-level test; do not ask for a separate Fake-based Handler test) = on a real DB, at least one case per HTTP status the API can return (500 via a cancelled request context), every response validated against the spec, inserted rows checked field by field, and for each constrained input field the boundary values from the acceptance criteria / cmd (\`cmd/<service>/main_test.go\`) = routes respond through \`buildEcho\` and a spec violation returns 400 / E2E = only the minimum important business flows (do not duplicate what Domain/UseCase can verify). When a Query contract test exists, check that it verifies the mapping to the Read Model and that it does not depend on write-side Domain Models
`
      },
      {
        "path": "backend/migrations/**",
        "instructions": `golang-migrate SQL (PostgreSQL).
- Is there a \`.down.sql\` paired with each \`.up.sql\`, and does down correctly roll back up?
- Is an already-merged migration file rewritten? (changes go in a new numbered file)
- For changes that affect existing data, such as adding NOT NULL or changing a type, is the handling of existing rows (default value / backfill) considered?
- If the table structure changed, is \`docs/db/backend-schema.json\` / \`.md\` updated too? (see \`docs/rules/db-documentation.md\`)
`
      },
      {
        "path": "openapi/**",
        "instructions": `The source of truth for the API contract (OpenAPI Design First). The rules are defined in \`docs/architecture.md\` §4 (Japanese). Do not duplicate findings already covered by Redocly lint.
- Generated output (\`openapi/dist/**\`, \`backend/api/gen/**\`, \`frontend/*/src/api/generated/**\`) is out of review scope. However, check that the same PR contains tests for the code that implements/uses the added or changed operations (the BE Handler, the FE hook/screen), and review those tests (CI \`Test Presence\` checks for per-operation tests. \`docs/rules/testing.md\`)
- When the YAML changes, are \`openapi/dist/**\` (bundle), \`backend/api/gen/**\` (\`make -C backend openapi\`), and \`frontend/*/src/api/generated/**\` (\`pnpm -C frontend gen:api\`) regenerated in the same PR? Flag manual edits of \`openapi/dist/**\`
- Is \`operationId\` the lowerCamel of the BE UseCase type name it calls, without \`Usecase\` (\`SearchBooksUsecase\` → \`searchBooks\`; verbs follow the UseCase naming rule: \`read\` for one item, \`search\` for any list, \`create\` / \`update\` / \`delete\` or a business verb such as \`register\` / \`rename\`, not \`list\` / \`get\`), and unique within the channel? Does \`summary\` start with the same verb as \`operationId\` (\`searchBooks\` → \`Search books\`, \`readBook\` → \`Read a book\`)?
- Are only statuses that are actually returned written? Do required fields have \`required\`?
- Is \`enum\` attached only to items whose value set is guaranteed by a DB constraint or the Domain? (an unguaranteed enum becomes a contract violation on an out-of-range value)
- If \`common/schemas/ErrorResponse.yaml\` / \`FieldError.yaml\` change, is the hand-written \`ErrorResponse\` in \`backend/internal/platform/httpserver\` fixed in the same PR?
- For breaking changes such as removing an existing field, making it required, or changing its type, does the PR explain the OpenAPI CI breaking-change report (warning only until the first release)?
`
      },
      {
        "path": "frontend/**",
        "instructions": `React + Vite + TypeScript (pnpm workspace). The source of truth for policy is \`frontend/architecture.md\` (Japanese). Do not duplicate findings already covered by ESLint / Prettier / \`tsc\` / Vitest.

## Architecture
- Has code assuming Next.js / SSR / RSC been brought in? (this is a Vite SPA). Is app code placed in \`.js\` / \`.jsx\`?
- The design reference differs per channel: public follows **Apple HIG (iOS)**, admin follows the **Digital Agency Design System** (\`packages/ui\` + \`@digital-go-jp/tailwind-theme-plugin\`) (\`frontend/architecture.md\` §4). Has another UI library (shadcn, etc.) been introduced without agreement? Does public consider tap targets (44pt or larger) and safe areas?
- Only DS themes / imported components, OpenAPI generated types, and pure functions (amounts, dates) may live in \`packages/\`. Are domain-bearing composite components shared across apps?
- **No imports between features** (dependencies are one-way: shared → features → app; composition is done in \`app/\`)
- Is logic (fetching, transformation, screen state) in \`hooks/\`, with components receiving props and only rendering? Is data fetching or complex branching piling up inside components?
- Is TanStack Query the source of truth for server state, with no duplicate storage in Zustand? Is business data or frequently changing values put in Context?
- Are API types/clients hand-written? \`src/api/generated/**\` (Orval) is generated by \`pnpm gen:api\`, so flag manual edits (change the contract and regenerate)
- Is public mobile-first (mobile first, single column), and is admin laid out for PC?
- Storybook is a state catalog of Presentational components that are self-contained via props. Are there stories for whole pages or stories wired directly to production Providers / APIs?

## Tests
- Require \`*.test.ts(x)\` in the same directory for behavior-changing changes. Check that they verify the main states (loading, empty, success, error), boundary values, and error propagation. Are self-written child components over-mocked? Is a huge Snapshot relied on? (see \`frontend/architecture.md\` and \`docs/rules/testing.md\`)
`
      },
      {
        "path": "**/*{_test.go,.test.ts,.test.tsx}",
        "instructions": `Test files. Review for missing failure-path coverage (\`docs/rules/testing.md\`).
- For each changed behavior, inspect all related test files together. Flag the behavior only when those tests collectively cover only the success case, and ask whether they cover: invalid input, boundary values (exactly at the limit and one over), an empty / zero-item case, and a lower-layer error propagating
- Handler: does every 4xx/5xx status the OpenAPI spec lists for the operation have a test? Is the response checked against the spec? For every input constraint written in the operation's YAML (\`required\`, lengths, ranges, \`enum\`, \`pattern\`, \`format\`, item counts, etc.), is there one passing value and one rejected value (400 with \`errors[{field, rule}]\`) taken from the acceptance criteria? If the YAML uses a constraint kind for the first time, is it added to the \`internal/platform/httpserver\` test spec?
- Failure must leave state unchanged; a validation error must not call lower layers. Are both asserted?
- Frontend: are the API-error and input-error states tested, not only the success screen?
- Are tests derived from the acceptance criteria / OpenAPI rather than from what the code currently does? Flag assertions that merely mirror the implementation
- High-value viewpoints beyond boundary values. Raise one only when the changed code actually involves it, and name the concrete missing case:
  - Validation failures (expected user mistakes, not system errors): for each required / type / format / length / range / correlation rule, is the exact expected error (status and \`errors[{field, rule}]\`) tested? Is the invariant also tested at the Domain VO constructor, not only via OpenAPI?
  - Business-rule violations (valid input that the business forbids: duplicate ISBN, ISBN not in the external catalog, duplicate tag name, not found, version mismatch): does each map to the expected \`ErrInvalid\` / \`ErrNotFound\` / \`ErrConflict\` (and 4xx), with state unchanged?
  - Equivalence partitioning / decision table: when behavior depends on a combination of conditions (e.g. keyword × tag filter × paging), is each meaningful combination covered, not only the all-true one?
  - State transitions: for an entity with a status, are invalid transitions rejected and tested (e.g. using an expired or logged-out session)?
  - Concurrency: for read-modify-write with a version column, is there a test that a stale version is rejected?
  - Idempotency: is a retried or double-submitted request safe (same result, no duplicate side effect)?
  - Authorization: is an unauthenticated request, and a cross-origin write, rejected?
  - Date and time: exactly at a deadline / validity start and end, across midnight, timezone (JST vs UTC), leap day
  - External failure: for Gateway / external API calls, are timeout, error response, and partial failure tested, including that state is not left half-updated?
  - Empty / zero / null / duplicated input, and the last page of a list
- Do not ask for tests on trivial code; name the specific missing case instead of a generic "add more tests"
`
      }
    ]
  }
} satisfies CodeRabbitConfig;

const reviewLabels = ["review:design", "review:standard", "review:critical", "review:skip"];

const designInstructions = `Review the current design stage: responsibilities, dependency direction, consistency with agreed requirements, API contracts, and exceptional cases.
Do not request production implementation, generated code, or implementation tests merely because they are absent during design review. This stage-specific rule takes precedence over path instructions requiring those artifacts in the same PR. Those requirements apply before implementation is merged.
Do not decide unresolved requirements; identify them as decisions for humans.`;

const implementationInstructions = `Review conformance to the agreed design, defects, boundary values, and missing behavior tests.
Propose changes to an agreed design only when a concrete defect or requirement/repository-rule violation justifies them. Name the evidence and impact; do not reopen design decisions based only on preference.
Distinguish necessary fixes from optional improvements. Human approval and existing test requirements still apply.`;

export default defineConfig((ctx) => {
  const selected = (ctx.pr?.labels ?? []).filter((label) => reviewLabels.includes(label));
  const classified = selected.length === 1;
  const category = classified ? selected[0] : undefined;
  const isDesign = category === "review:design";
  const isSkipped = category === "review:skip";
  const isLocal = ctx.invocation.source === "cli" && ctx.pr === null;
  const stageInstructions = !classified
    ? "PR review classification is missing or conflicting. Ask the human to choose exactly one review:design, review:standard, review:critical, or review:skip label; never choose or apply it yourself. For local CLI review without PR context, review the supplied changes without inferring a PR classification."
    : isDesign
      ? designInstructions
      : implementationInstructions + (category === "review:critical"
        ? "\nFor affected critical behavior, also inspect failure paths, concurrency, idempotency, authorization, and data consistency. Name concrete risks; do not demand unrelated checks."
        : "");

  return {
    ...baseConfig,
    reviews: {
      ...baseConfig.reviews,
      profile: isDesign || category === "review:critical" ? "assertive" : "chill",
      auto_apply_labels: false,
      auto_review: {
        ...baseConfig.reviews.auto_review,
        enabled: (classified && !isSkipped) || isLocal,
        drafts: classified && isDesign,
      },
      path_instructions: baseConfig.reviews.path_instructions.map((rule) => ({
        ...rule,
        instructions: stageInstructions + "\n\n" + rule.instructions,
      })),
    },
  };
});
