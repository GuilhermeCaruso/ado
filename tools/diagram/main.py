#!/usr/bin/env python3
"""ADO diagram generator — produces Mermaid diagrams from ADO domain index YAML files."""

from __future__ import annotations

import argparse
import sys
from dataclasses import dataclass, field
from enum import Enum
from pathlib import Path
import yaml


# ---------------------------------------------------------------------------
# Data model
# ---------------------------------------------------------------------------

@dataclass
class Domain:
    name: str = ""
    description: str = ""
    business_area: str = ""
    slack_channel: str = ""


@dataclass
class TeamMember:
    name: str = ""
    role: str = ""
    github: str = ""
    slack: str = ""


@dataclass
class Team:
    name: str = ""
    mission: str = ""
    members: list[TeamMember] = field(default_factory=list)


@dataclass
class ServiceIndex:
    service: str = ""
    description: str = ""


@dataclass
class ContextBlock:
    domain: Domain = field(default_factory=Domain)
    team: Team = field(default_factory=Team)
    service_index: list[ServiceIndex] = field(default_factory=list)


@dataclass
class StoreUser:
    service: str = ""
    access: str = ""
    key_prefix: str = ""


@dataclass
class Store:
    type: str = ""
    description: str = ""
    used_by: list[StoreUser] = field(default_factory=list)
    key_prefix: str = ""


@dataclass
class Topic:
    description: str = ""
    producer: str = ""
    consumers: list[str] = field(default_factory=list)


@dataclass
class DataBlock:
    stores: dict[str, Store] = field(default_factory=dict)
    topics: dict[str, Topic] = field(default_factory=dict)


@dataclass
class StoreRef:
    store: str = ""
    collection: str = ""
    access: str = ""


@dataclass
class Route:
    id: str = ""
    path: str = ""
    method: str = ""
    stores: list[StoreRef] = field(default_factory=list)


@dataclass
class Dependency:
    service: str = ""
    routes: list[str] = field(default_factory=list)


@dataclass
class Caller:
    service: str = ""
    routes: list[str] = field(default_factory=list)


@dataclass
class PublishedEvent:
    name: str = ""
    consumers: list[str] = field(default_factory=list)


@dataclass
class SubscribedEvents:
    service: str = ""
    events: list[str] = field(default_factory=list)
    stores: list[StoreRef] = field(default_factory=list)


@dataclass
class Events:
    published: list[PublishedEvent] = field(default_factory=list)
    subscribed: list[SubscribedEvents] = field(default_factory=list)


@dataclass
class TechStack:
    language: str = ""
    framework: str = ""


@dataclass
class Service:
    ownership_type: str = ""
    status: str = ""
    criticality: str = ""
    description: str = ""
    tech_stack: TechStack = field(default_factory=TechStack)
    routes: list[Route] = field(default_factory=list)
    dependencies: list[Dependency] = field(default_factory=list)
    callers: list[Caller] = field(default_factory=list)
    events: Events = field(default_factory=Events)


@dataclass
class Config:
    context: ContextBlock = field(default_factory=ContextBlock)
    data: DataBlock = field(default_factory=DataBlock)
    services: dict[str, Service] = field(default_factory=dict)


# ---------------------------------------------------------------------------
# YAML loading
# ---------------------------------------------------------------------------

def _load_store_ref(raw: dict) -> StoreRef:
    return StoreRef(
        store=raw.get("store", ""),
        collection=raw.get("collection", ""),
        access=raw.get("access", ""),
    )


def _load_config(raw: dict) -> Config:
    ctx_raw = raw.get("context", {}) or {}
    domain_raw = ctx_raw.get("domain", {}) or {}
    team_raw = ctx_raw.get("team", {}) or {}

    domain = Domain(
        name=domain_raw.get("name", ""),
        description=domain_raw.get("description", ""),
        business_area=domain_raw.get("business_area", ""),
        slack_channel=domain_raw.get("slack_channel", ""),
    )

    members = [
        TeamMember(
            name=m.get("name", ""),
            role=m.get("role", ""),
            github=m.get("github", ""),
            slack=m.get("slack", ""),
        )
        for m in (team_raw.get("members") or [])
    ]

    team = Team(
        name=team_raw.get("name", ""),
        mission=team_raw.get("mission", ""),
        members=members,
    )

    service_index = [
        ServiceIndex(service=si.get("service", ""), description=si.get("description", ""))
        for si in (ctx_raw.get("service_index") or [])
    ]

    data_raw = raw.get("data", {}) or {}
    stores: dict[str, Store] = {}
    for name, s in (data_raw.get("stores") or {}).items():
        s = s or {}
        stores[name] = Store(
            type=s.get("type", ""),
            description=s.get("description", ""),
            key_prefix=s.get("key_prefix", ""),
            used_by=[
                StoreUser(
                    service=u.get("service", ""),
                    access=u.get("access", ""),
                    key_prefix=u.get("key_prefix", ""),
                )
                for u in (s.get("used_by") or [])
            ],
        )

    topics: dict[str, Topic] = {}
    for name, t in (data_raw.get("topics") or {}).items():
        t = t or {}
        topics[name] = Topic(
            description=t.get("description", ""),
            producer=t.get("producer", ""),
            consumers=list(t.get("consumers") or []),
        )

    services: dict[str, Service] = {}
    for name, svc in (raw.get("services") or {}).items():
        svc = svc or {}
        tech_raw = svc.get("tech_stack") or {}
        routes = [
            Route(
                id=r.get("id", ""),
                path=r.get("path", ""),
                method=r.get("method", ""),
                stores=[_load_store_ref(sr) for sr in (r.get("stores") or [])],
            )
            for r in (svc.get("routes") or [])
        ]
        dependencies = [
            Dependency(
                service=d.get("service", ""),
                routes=list(d.get("routes") or []),
            )
            for d in (svc.get("dependencies") or [])
        ]
        callers = [
            Caller(
                service=c.get("service", ""),
                routes=list(c.get("routes") or []),
            )
            for c in (svc.get("callers") or [])
        ]
        events_raw = svc.get("events") or {}
        published = [
            PublishedEvent(
                name=e.get("name", ""),
                consumers=list(e.get("consumers") or []),
            )
            for e in (events_raw.get("published") or [])
        ]
        subscribed = [
            SubscribedEvents(
                service=e.get("service", ""),
                events=list(e.get("events") or []),
                stores=[_load_store_ref(sr) for sr in (e.get("stores") or [])],
            )
            for e in (events_raw.get("subscribed") or [])
        ]
        services[name] = Service(
            ownership_type=svc.get("ownership_type", ""),
            status=svc.get("status", ""),
            criticality=svc.get("criticality", ""),
            description=svc.get("description", ""),
            tech_stack=TechStack(
                language=tech_raw.get("language", ""),
                framework=tech_raw.get("framework", ""),
            ),
            routes=routes,
            dependencies=dependencies,
            callers=callers,
            events=Events(published=published, subscribed=subscribed),
        )

    return Config(
        context=ContextBlock(domain=domain, team=team, service_index=service_index),
        data=DataBlock(stores=stores, topics=topics),
        services=services,
    )


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

class Mode(str, Enum):
    overview = "overview"
    standard = "standard"
    detailed = "detailed"


def _node_id(name: str) -> str:
    return name.replace("-", "_")


def _truncate(s: str, max_len: int) -> str:
    return s if len(s) <= max_len else s[:max_len] + "..."


def _service_node(name: str, svc: Service, indent: str) -> str:
    label = f"{name}\\n<i>{_truncate(svc.description, 40)}</i>"
    nid = _node_id(name)
    if svc.ownership_type == "partner":
        return f'{indent}{nid}(["\U0001f91d {label}"]):::partner\n'
    if svc.ownership_type == "internal":
        return f'{indent}{nid}[/"\U0001f3e2 {label}"/]:::internal\n'
    return f'{indent}{nid}["\u2699\ufe0f {label}"]:::owner\n'


def _services_by_type(services: dict[str, Service], ownership: str) -> list[str]:
    return sorted(n for n, s in services.items() if s.ownership_type == ownership)


def _build_route_index(cfg: Config) -> dict[str, dict[str, str]]:
    idx: dict[str, dict[str, str]] = {}
    for name, svc in cfg.services.items():
        idx[name] = {r.id: f"{r.method} {r.path}" for r in svc.routes}
    return idx


def _build_store_access(cfg: Config) -> dict[str, dict[str, bool]]:
    acc: dict[str, dict[str, bool]] = {}
    for name, svc in cfg.services.items():
        for r in svc.routes:
            for sr in r.stores:
                acc.setdefault(sr.store, {})[name] = True
        for sub in svc.events.subscribed:
            for sr in sub.stores:
                acc.setdefault(sr.store, {})[name] = True
    return acc


def _collapsed_http_label(methods: list[str]) -> str:
    if len(methods) == 1:
        return methods[0]
    unique = sorted(set(methods))
    return f"{len(methods)} calls ({', '.join(unique)})"


def _collapsed_event_label(events: list[str]) -> str:
    if len(events) == 1:
        parts = events[0].split(".")
        return f"event: {parts[-1]}"
    return f"{len(events)} events"


def _append_unique(lst: list[str], value: str) -> list[str]:
    if value not in lst:
        lst.append(value)
    return lst


# ---------------------------------------------------------------------------
# Diagram sections
# ---------------------------------------------------------------------------

def _write_frontmatter(parts: list[str], cfg: Config) -> None:
    domain = cfg.context.domain
    team = cfg.context.team
    if not domain.name and not team.name:
        return
    title = domain.name
    if team.name:
        title += f" · {team.name}"
    parts += ["---\n", f"title: {title}\n", "---\n\n"]


def _write_context_comments(parts: list[str], cfg: Config) -> None:
    domain = cfg.context.domain
    team = cfg.context.team
    if domain.name:
        line = f"    %% domain: {domain.name}"
        if domain.business_area:
            line += f" | area: {domain.business_area}"
        if domain.slack_channel:
            line += f" | slack: {domain.slack_channel}"
        parts.append(line + "\n")
    if team.name:
        leads = [
            f"{m.name} ({m.role})"
            for m in team.members
            if m.role in ("tech_lead", "product_manager")
        ]
        line = f"    %% team: {team.name}"
        if leads:
            line += f" | {', '.join(leads)}"
        parts.append(line + "\n")
    parts.append("\n")


def _write_styles(parts: list[str]) -> None:
    parts += [
        "\n    %% styles\n",
        "    classDef owner fill:#dbeafe,stroke:#3b82f6,color:#1e3a5f\n",
        "    classDef internal fill:#f0fdf4,stroke:#16a34a,color:#14532d\n",
        "    classDef partner fill:#fef9c3,stroke:#ca8a04,color:#713f12\n",
        "    classDef store fill:#f3f4f6,stroke:#6b7280,color:#1f2937\n",
    ]


def _write_ownership_subgraphs(parts: list[str], cfg: Config) -> None:
    owners = _services_by_type(cfg.services, "owner")
    internals = _services_by_type(cfg.services, "internal")
    partners = _services_by_type(cfg.services, "partner")

    parts += ['    subgraph owned[" Owned Services "]\n', "        direction TB\n"]
    for name in owners:
        parts.append(_service_node(name, cfg.services[name], "        "))
    parts.append("    end\n\n")

    if internals:
        parts += ['    subgraph internal_svcs[" Internal Services "]\n', "        direction TB\n"]
        for name in internals:
            parts.append(_service_node(name, cfg.services[name], "        "))
        parts.append("    end\n\n")

    if partners:
        parts += ['    subgraph partners[" Partners "]\n', "        direction TB\n"]
        for name in partners:
            parts.append(_service_node(name, cfg.services[name], "        "))
        parts.append("    end\n\n")


def _write_store_subgraph(parts: list[str], cfg: Config) -> None:
    if not cfg.data.stores:
        return
    parts += ['    subgraph stores[" Data Stores "]\n', "        direction LR\n"]
    for name in sorted(cfg.data.stores):
        store = cfg.data.stores[name]
        icon = "\u26a1" if store.type == "redis" else "\U0001f5c4\ufe0f"
        parts.append(f'        {_node_id(name)}_store[("{icon} {name}")]:::store\n')
    parts.append("    end\n\n")


def _write_store_edges(parts: list[str], cfg: Config, store_acc: dict[str, dict[str, bool]]) -> None:
    if not cfg.data.stores:
        return
    parts.append("\n    %% store access\n")
    for store_name in sorted(cfg.data.stores):
        store_node = f"{_node_id(store_name)}_store"
        for svc_name in sorted(store_acc.get(store_name, {})):
            parts.append(f"    {_node_id(svc_name)} --- {store_node}\n")


# ---------------------------------------------------------------------------
# Generators
# ---------------------------------------------------------------------------

def generate_overview(cfg: Config) -> str:
    parts: list[str] = []
    _write_frontmatter(parts, cfg)
    parts.append("flowchart LR\n\n")
    _write_context_comments(parts, cfg)

    for ownership in ("owner", "internal", "partner"):
        for name in _services_by_type(cfg.services, ownership):
            parts.append(_service_node(name, cfg.services[name], "    "))

    http_pairs: set[tuple[str, str]] = set()
    evt_pairs: set[tuple[str, str]] = set()

    for name, svc in cfg.services.items():
        for dep in svc.dependencies:
            if dep.service != name:
                http_pairs.add((_node_id(name), _node_id(dep.service)))
        for ev in svc.events.published:
            for consumer in ev.consumers:
                if consumer != name:
                    evt_pairs.add((_node_id(name), _node_id(consumer)))

    parts.append("\n    %% http calls\n")
    for frm, to in sorted(http_pairs):
        parts.append(f"    {frm} --> {to}\n")

    parts.append("\n    %% events\n")
    for frm, to in sorted(evt_pairs):
        parts.append(f"    {frm} -.-> {to}\n")

    _write_styles(parts)
    return "".join(parts)


def generate_standard(cfg: Config) -> str:
    parts: list[str] = []
    _write_frontmatter(parts, cfg)
    parts.append("flowchart TD\n\n")
    _write_context_comments(parts, cfg)

    route_idx = _build_route_index(cfg)
    store_acc = _build_store_access(cfg)

    _write_ownership_subgraphs(parts, cfg)
    _write_store_subgraph(parts, cfg)

    http_edges: dict[tuple[str, str], list[str]] = {}
    for name, svc in cfg.services.items():
        for dep in svc.dependencies:
            key = (_node_id(name), _node_id(dep.service))
            if not dep.routes:
                http_edges.setdefault(key, []).append("HTTP")
                continue
            for rid in dep.routes:
                label = route_idx.get(dep.service, {}).get(rid, rid[:8] + "...")
                method = label.split(" ")[0]
                http_edges.setdefault(key, []).append(method)

    evt_edges: dict[tuple[str, str], list[str]] = {}
    for name, svc in cfg.services.items():
        for ev in svc.events.published:
            for consumer in ev.consumers:
                if consumer != name:
                    key = (_node_id(name), _node_id(consumer))
                    evt_edges.setdefault(key, []).append(ev.name)

    parts.append("    %% http calls\n")
    for (frm, to), vals in sorted(http_edges.items()):
        parts.append(f'    {frm} -->|"{_collapsed_http_label(vals)}"| {to}\n')

    parts.append("\n    %% events\n")
    for (frm, to), vals in sorted(evt_edges.items()):
        parts.append(f'    {frm} -.->|"{_collapsed_event_label(vals)}"| {to}\n')

    _write_store_edges(parts, cfg, store_acc)
    _write_styles(parts)
    return "".join(parts)


def generate_detailed(cfg: Config) -> str:
    parts: list[str] = []
    _write_frontmatter(parts, cfg)
    parts.append("flowchart TD\n\n")
    _write_context_comments(parts, cfg)

    route_idx = _build_route_index(cfg)
    store_acc = _build_store_access(cfg)

    _write_ownership_subgraphs(parts, cfg)
    _write_store_subgraph(parts, cfg)

    parts.append("    %% http calls\n")
    for name in sorted(cfg.services):
        svc = cfg.services[name]
        for dep in svc.dependencies:
            frm, to = _node_id(name), _node_id(dep.service)
            if not dep.routes:
                parts.append(f'    {frm} -->|"HTTP"| {to}\n')
                continue
            for rid in dep.routes:
                label = route_idx.get(dep.service, {}).get(rid, rid[:8] + "...")
                parts.append(f'    {frm} -->|"{label}"| {to}\n')

    parts.append("\n    %% events\n")
    for name in sorted(cfg.services):
        svc = cfg.services[name]
        for ev in svc.events.published:
            frm = _node_id(name)
            for consumer in ev.consumers:
                if consumer != name:
                    parts.append(f'    {frm} -.->|"{ev.name}"| {_node_id(consumer)}\n')

    if cfg.data.stores:
        parts.append("\n    %% store access\n")
        store_detail: dict[tuple[str, str], list[str]] = {}
        for name, svc in cfg.services.items():
            for r in svc.routes:
                for sr in r.stores:
                    k = (name, sr.store)
                    entry = sr.collection + f" ({sr.access})" if sr.collection else sr.access
                    store_detail[k] = _append_unique(store_detail.get(k, []), entry)
            for sub in svc.events.subscribed:
                for sr in sub.stores:
                    k = (name, sr.store)
                    entry = sr.collection + f" ({sr.access})" if sr.collection else sr.access
                    store_detail[k] = _append_unique(store_detail.get(k, []), entry)

        for store_name in sorted(cfg.data.stores):
            store_node = f"{_node_id(store_name)}_store"
            for svc_name in sorted(store_acc.get(store_name, {})):
                detail = store_detail.get((svc_name, store_name), [])
                if not detail:
                    label = "access"
                elif len(detail) <= 2:
                    label = ", ".join(detail)
                else:
                    label = f"{len(detail)} collections"
                parts.append(f'    {_node_id(svc_name)} ---|"  {label}  "| {store_node}\n')

    _write_styles(parts)
    return "".join(parts)


# ---------------------------------------------------------------------------
# CLI
# ---------------------------------------------------------------------------

def main() -> None:
    parser = argparse.ArgumentParser(
        description="Generate Mermaid diagrams from an ADO domain index YAML file."
    )
    parser.add_argument("path", nargs="?", default="template.yaml", help="Path to the ADO YAML file")
    parser.add_argument(
        "--mode", "-m",
        choices=[m.value for m in Mode],
        default=Mode.standard.value,
        help="Diagram mode: overview | standard | detailed (default: standard)",
    )
    args = parser.parse_args()

    input_path = Path(args.path)
    if not input_path.exists():
        print(f"error: file not found: {input_path}", file=sys.stderr)
        sys.exit(1)

    raw = yaml.safe_load(input_path.read_text(encoding="utf-8")) or {}
    cfg = _load_config(raw)

    mode = Mode(args.mode)
    if mode == Mode.overview:
        diagram = generate_overview(cfg)
    elif mode == Mode.detailed:
        diagram = generate_detailed(cfg)
    else:
        diagram = generate_standard(cfg)

    out_dir = input_path.parent / input_path.stem
    out_dir.mkdir(parents=True, exist_ok=True)
    out_path = out_dir / f"{mode.value}.mmd"
    out_path.write_text(diagram, encoding="utf-8")
    print(f"wrote {out_path}", file=sys.stderr)


if __name__ == "__main__":
    main()
