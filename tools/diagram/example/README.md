# Diagram examples

These files are the Mermaid output generated from [`example/domain/ecommerce.yaml`](../../../example/domain/ecommerce.yaml).

Each file corresponds to one diagram mode:

| File | Mode | Description |
|------|------|-------------|
| `overview.mmd` | `overview` | Flat high-level view — nodes and connections, no labels |
| `standard.mmd` | `standard` | Subgraphs by ownership, collapsed edges, grouped stores |
| `detailed.mmd` | `detailed` | One edge per route with full `METHOD /path`, collections per store |

To regenerate:

```bash
python tools/diagram/main.py --mode overview  example/domain/ecommerce.yaml
python tools/diagram/main.py --mode standard  example/domain/ecommerce.yaml
python tools/diagram/main.py --mode detailed  example/domain/ecommerce.yaml
```
