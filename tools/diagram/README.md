# diagram

Generates Mermaid diagrams from an ADO domain index YAML file.

## Requirements

Go 1.22+

## Usage

```bash
go run . --mode <mode> <path-to-domain.yaml>
```

| Mode | Description |
|------|-------------|
| `overview` | Flat high-level view — nodes and connections, no labels |
| `standard` | Subgraphs by ownership, collapsed edges, grouped stores (default) |
| `detailed` | One edge per route with full `METHOD /path`, collections per store |

Output is written as a `.mmd` file alongside the input YAML, inside a folder named after the domain.

## Build

```bash
go build -o diagram .
./diagram --mode <mode> <path-to-domain.yaml>
```

## Example

```bash
go run . --mode overview  ../../example/domain/ecommerce.yaml
go run . --mode standard  ../../example/domain/ecommerce.yaml
go run . --mode detailed  ../../example/domain/ecommerce.yaml
```

See [`example/`](example/) for the generated output from the ecommerce domain.
