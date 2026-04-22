package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	mode := flag.String("mode", "standard", "Diagram mode: overview | standard | detailed")
	flag.StringVar(mode, "m", "standard", "Diagram mode (shorthand)")
	flag.Parse()

	if *mode != "overview" && *mode != "standard" && *mode != "detailed" {
		fmt.Fprintf(os.Stderr, "error: invalid mode %q — must be overview, standard, or detailed\n", *mode)
		os.Exit(1)
	}

	path := flag.Arg(0)
	if path == "" {
		path = "template.yaml"
	}

	cfg, err := loadConfig(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	var diagram string
	switch *mode {
	case "overview":
		diagram = generateOverview(cfg)
	case "detailed":
		diagram = generateDetailed(cfg)
	default:
		diagram = generateStandard(cfg)
	}

	base := strings.TrimSuffix(path, filepath.Ext(path))
	if err := os.MkdirAll(base, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	outPath := filepath.Join(base, *mode+".mmd")
	if err := os.WriteFile(outPath, []byte(diagram), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "wrote %s\n", outPath)
}
