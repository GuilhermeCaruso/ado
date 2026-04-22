package main

import (
	"embed"
	"flag"
	"fmt"
	"log"

	"github.com/GuilhermeCaruso/ado/tools/ui/internal/server"
)

//go:embed web
var webFS embed.FS

func main() {
	port := flag.String("port", "8080", "Port to listen on")
	flag.StringVar(port, "p", "8080", "Port (shorthand)")
	flag.Parse()

	addr := ":" + *port
	fmt.Printf("ADO UI → http://localhost%s\n", addr)

	if err := server.New(addr, webFS).Start(); err != nil {
		log.Fatal(err)
	}
}
