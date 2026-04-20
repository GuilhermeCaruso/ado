# index-domain

Generates a structured ADO domain index YAML by scraping GitHub repositories. The output is a complete service map — readable by humans and AI agents alike — following the [ADO domain schema](../../spec/v1/domain-schema.yaml). The `template.yml` alongside this skill is a reference template for the output format.

---

## Input

The user provides:
1. **Output name** — the base filename (e.g. `payments`, `lending`)
2. **Repositories** — one or more, each with a `url` and `ownership_type`
   - `ownership_type`: `owner` (your team) | `internal` (another internal team) | `partner` (external API/partner)

---

## Process

### Step 1 — Collect domain context

Before scraping anything, ask the user all questions in a single message:

---

**Welcome to index-domain!** I'll guide you through indexing your services. First, let me gather some context.

**Domain**
1. What is the domain name? *(e.g. `payments`, `hr`, `catalog`)*
2. Describe this domain in 2–4 sentences — what problem does it solve, who are the users, what business value does it deliver?
3. What is the business area? *(e.g. `financial-products`, `core-banking`)*
4. What is the team's Slack channel or main communication channel?

**Team**
5. What is the team name?
6. What is the team mission? *(one sentence)*
7. List the team members in this format (one per line):
   ```
   Full Name | role | github_handle | @slack_handle
   ```
   Roles: `tech_lead`, `senior_engineer`, `engineer`, `product_manager`, `data_engineer`, `sre`

**Repositories**
8. List the repositories to index, one per line:
   ```
   https://github.com/org/repo | ownership_type
   ```

**Output**
9. What should the output file be named? *(e.g. `payments` → generates `payments.yaml`)*

---

Wait for all answers. If any required field is missing or ambiguous, ask a follow-up before proceeding.

---

### Step 2 — Check for existing index files

Before scraping each repository, check for pre-built index files at the repo root. Use the following priority order:

**Priority 1 — `ado.yaml`** (ADO self-declaration)

Try: `<repo-url>/blob/main/ado.yaml` (fallback: `/blob/master/ado.yaml`).

If found, this is the highest-trust source — maintained by the service team in full ADO format:
- Use all fields directly: `routes` (with UUIDs), `events`, `stores`, `description`, `tech_stack`, `docs`
- Preserve route UUIDs exactly as declared — never reassign them
- Skip scraping entirely for this service
- Only resolve cross-service relationships (`callers`, dependency UUID matching) at domain level in Step 4

**Priority 2 — `service-snapshot.yaml`** (technical snapshot)

Try: `<repo-url>/blob/main/service-snapshot.yaml` (fallback: `/blob/master/service-snapshot.yaml`).

If found and no `ado.yaml` exists, use it as the technical base:
- `routes` — use routes (id, method, path, headers, params, body, stores). Still scrape for `description` per route.
- `events.published` / `events.subscribed` — use directly.
- `stores` — use including `fields` when present.
- `tech_stack`, `port`, `healthcheck` — use directly.

**Priority 3 — Full scraping**

If neither file exists, proceed with full repository scraping as described in Step 3.

**Always resolve regardless of source:**
- `dependencies` and `callers` (require cross-service analysis across all repos)
- `criticality` (provided by user or inferred — never blindly trusted from files)
- `ownership_type` (provided by the user)

Note found sources in the final report: `"ado.yaml found — used directly for <service-name>"` or `"service-snapshot.yaml found — used as base for <service-name>"`.

---

### Step 3 — Scrape each repository

For each repository, analyze:

**A. Repository metadata**
- Fetch the GitHub repo page and README
- Extract: description, primary language, framework, status (`active` / `deprecated`), criticality clue

**B. HTTP routes / API surface**
- Look for router files, controller files, OpenAPI/Swagger specs, route registration, handler files
- Common patterns:
  - Go/Fiber: `app.Get`, `app.Post`, `v1.Get` in `main.go`, `routes.go`, `router.go`
  - NestJS: `@Get`, `@Post`, `@Controller` decorators
  - Express: `router.get`, `router.post`
  - OpenAPI: `paths:` in `swagger.yaml` / `openapi.yaml`
- Per route extract: `path`, `method`, `description`, `headers`, `params`, `body` fields

**C. Events (Kafka / async)**
- Look for: topic names, producer calls, consumer decorators, event subscriptions
- Common patterns:
  - Go: `kafka.Publish`, `topic :=`, `ProduceTopic`
  - NestJS: `@EventPattern`, `@EventSubscription`, `client.emit`
  - Config files: `topics:` blocks
- Per event extract: name (topic), producer/consumer role, other services involved

**D. Data stores**
- Look for: MongoDB collection names, Redis key prefixes, database config files
- Common patterns: `Collection("name")`, `db.collection`, `REDIS_PREFIX`, `keyPrefix`
- Extract: store type, collection names and their purpose

**E. Dependencies on other services**
- Look for: HTTP client calls, service URLs in config/env, `axios.get`, `http.Get`, gRPC calls
- Map which routes call which other services

**F. Docs references**
- Look for: `README.md`, `docs/AI_CONTEXT.md`, `docs/ARCHITECTURE.md`

---

### Step 4 — Resolve cross-service relationships

After scraping all repos:

1. **Match dependencies** — if service A calls service B, route IDs must cross-reference correctly
2. **Match events** — if service A publishes an event that service B subscribes to, both `published.consumers` and `subscribed` blocks must be consistent
3. **Assign stable UUIDs** to every route (v4 format: `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`)
4. **Build the `data` block** — consolidate all stores and topics across services

---

### Step 5 — Generate the YAML file

Produce a single file named `<output-name>.yaml` following **exactly** the structure of [template.yml](template.yml).

Key rules:
- Every route must have a unique **v4 UUID** `id`. Never use short IDs or sequential numbers.
- `dependencies[].routes` references route UUIDs of the **target** service
- `callers[].routes` references route UUIDs of the **caller** service (inverse of dependencies)
- `events.published[].consumers` lists service names that subscribe to this event
- `events.subscribed[]` groups by source service with event names and affected stores
- `stores` inside routes lists which store + collection each route reads or writes
- `criticality`: `critical` | `high` | `medium` | `low`
- `ownership_type`: `owner` | `internal` | `partner`
- `status`: `active` | `deprecated` | `maintenance`

For `internal` or `partner` services, mark unknown fields as `unknown`. Focus on what is visibly consumed by the owned services.

**YAML safety rules:**
1. Quote any string containing `: ` (colon + space) — e.g. `"Manages orders (ex: cart, checkout)"`
2. Quote strings containing `#`, `[`, `]`, `{`, `}`, `*`, `&`, `!`, `|`, `>`
3. Prefer block scalars (`description: >`) for long text

---

### Step 6 — Validate and report

After writing the file:
1. Verify all UUID cross-references are consistent
2. Verify published events match subscribed events
3. Report:
   - Services indexed: N owner, N internal, N partner
   - Routes found: total count
   - Events found: N topics
   - Stores found: N stores
   - Sources used: for each service, whether `ado.yaml`, `service-snapshot.yaml`, or full scraping was used
   - Fields left as `TODO` that the user should fill manually

---

## Output

A single `<output-name>.yaml` file in the current directory.

If the ADO diagram tool is available (`tools/diagram/main.py`), generate diagrams for all three modes:
```
python tools/diagram/main.py --mode overview <output-name>.yaml
python tools/diagram/main.py --mode standard <output-name>.yaml
python tools/diagram/main.py --mode detailed <output-name>.yaml
```

If not available, skip and inform the user.

---

## Quality bar

A good index file:
- Has **descriptions on every service, route, and event** — not just field names
- Has **body/params documented** for all POST/PUT/PATCH routes
- Has **store references** on routes that read or write data
- Is **readable standalone** — an AI agent with no other context should understand what each service does and how they relate

If scraping gives insufficient data, use `TODO:` as the value rather than guessing.
