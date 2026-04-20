# diagram

Generates Mermaid diagrams from an ADO domain index YAML file.

## Requirements

```bash
pip install -r requirements.txt
```

## Usage

```bash
python main.py <path-to-domain.yaml> --mode <mode>
```

| Mode | Description |
|------|-------------|
| `overview` | Flat high-level view — nodes and connections, no labels |
| `standard` | Subgraphs by ownership, collapsed edges, grouped stores (default) |
| `detailed` | One edge per route with full `METHOD /path`, collections per store |

Output is written as a `.mmd` file alongside the input YAML, inside a folder named after the domain.

## Example

```bash
python main.py ../../example/domain/ecommerce.yaml --mode overview
python main.py ../../example/domain/ecommerce.yaml --mode standard
python main.py ../../example/domain/ecommerce.yaml --mode detailed
```

See [`example/`](example/) for the generated output from the ecommerce domain.
