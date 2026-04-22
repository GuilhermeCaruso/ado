# ui

> **MVP / work in progress** — this is an early test version, not the final expected result.

Web UI for ADO — visualize Mermaid diagrams and generate skill prompts.

## Requirements

Go 1.22+

## Usage

```bash
go run . [--port 8080]
```

Open `http://localhost:8080` in your browser.

## Build

```bash
go build -o ui .
./ui --port 8080
```

## Features

| Tab | Description |
|-----|-------------|
| **Viewer** | Load a `.mmd` file from your machine, render and navigate the architecture diagram (pan + zoom) |
| **Generator** | Fill in a form and get a ready-to-use prompt for the `index-service` or `index-domain` skill |
