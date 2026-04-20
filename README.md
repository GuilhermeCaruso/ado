# ADO — Agent-Driven Organization

[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

ADO is a structured indexing system for software architectures built for AI agents.

When an agent receives a task in an unfamiliar domain, the first thing it does is explore repositories, trace routes, and map dependencies. This costs tokens, takes time, and the agent can never be sure it found everything.

ADO works like a search engine index: scrape once, query many times.

| | Without index | With ADO |
|---|---|---|
| Context tokens | ~60,000 | ~6,000 |
| Coverage | Partial | Complete |
| Time to start | High | Low |

---

## How it works

ADO has two schemas and two skills:

```
spec/v1/
├── domain-schema.yaml    # full domain index (services, team, stores, topics)
└── service-schema.yaml   # service self-declaration (ado.yaml)

skills/
├── index-domain/         # scrapes multiple repos and builds a domain index
└── index-service/        # analyzes the current repo and generates its ado.yaml
```

**Domain index** — a single YAML file mapping an entire domain: services, routes, events, data stores, dependencies, and team context. Consumed by AI agents before any task.

**Service self-declaration (`ado.yaml`)** — each service declares itself by placing an `ado.yaml` at its root. When `index-domain` runs, it reads these files directly instead of scraping.

Priority when building a domain index:
1. `ado.yaml` — service self-declaration (highest trust)
2. `service-snapshot.yaml` — technical snapshot
3. Full repository scraping — fallback

---

## Getting started

### Index a domain

Copy the `index-domain` skill to your agent setup:

```
skills/index-domain/SKILL.md
skills/index-domain/template.yml
```

Run the skill and follow the interactive questionnaire. It scrapes your repositories and generates a domain index YAML.

### Self-declare a service

Copy the `index-service` skill to the service repository:

```
skills/index-service/SKILL.md
skills/index-service/template.yml
```

Run the skill inside the repo. It generates an `ado.yaml` at the root.

### Generate diagrams

```bash
pip install -r tools/diagram/requirements.txt

python tools/diagram/main.py --mode overview  my-domain.yaml
python tools/diagram/main.py --mode standard  my-domain.yaml
python tools/diagram/main.py --mode detailed  my-domain.yaml
```

Diagrams are written as `.mmd` (Mermaid) files alongside the input YAML.

---

## Examples

```
example/
├── self-index/
│   └── ado.yaml                      # single service self-declaration
└── domain/
    ├── ecommerce.yaml                 # full domain index
    ├── order-service/ado.yaml
    └── catalog-service/ado.yaml
```

---

## Skills compatibility

ADO skills are plain Markdown files with no platform-specific syntax. They work with any LLM that can follow instructions: Claude, GPT-4, Gemini, Cursor, and others.

To use with Claude Code, copy the skill folder to `.claude/skills/` in your project.

---

## Project structure

```
ado/
├── spec/v1/
│   ├── domain-schema.yaml
│   └── service-schema.yaml
├── skills/
│   ├── index-domain/
│   └── index-service/
├── tools/
│   └── diagram/
└── example/
```

---

## Roadmap

- [ ] **CLI** — single binary to run skills, generate diagrams, and validate index files from the terminal
- [ ] **UI** — web interface to fill domain/service forms, visualize diagrams, and export YAML without writing by hand
- [ ] **MCP server** — expose the domain index as tools (`get_service`, `find_route`, `list_dependencies`) so any MCP-compatible agent can query it at runtime
- [ ] **GitHub Action** — automatically regenerate `ado.yaml` on push and open a PR when the index is outdated
- [ ] **VS Code extension** — inline validation of `ado.yaml` and domain index files against the spec schemas
- [ ] **Index registry** — public hub to share and discover domain indexes across organizations

---

## Contributing

1. Fork the repository
2. Create a branch: `git checkout -b feat/your-contribution`
3. Open a pull request

---

## License

Apache 2.0 — see [LICENSE](LICENSE).
