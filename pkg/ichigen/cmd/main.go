package main

import (
	"flag"
	"fmt"
	"ichi-go/pkg/ichigen"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "generate", "g":
		handleGenerate(os.Args[2:])
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func handleGenerate(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: ichigen generate <type> <name> [options]")
		fmt.Println("Example: ichigen g controller product --domain=catalog")
		os.Exit(1)
	}

	genType := args[0]
	name := args[1]

	fs := flag.NewFlagSet("generate", flag.ExitOnError)
	domain := fs.String("domain", "", "Domain/module name (required)")
	crud := fs.Bool("crud", false, "Generate CRUD operations")
	fs.Parse(args[2:])

	if *domain == "" {
		fmt.Println("Error: --domain flag is required")
		os.Exit(1)
	}

	config := ichigen.Config{
		Type:   genType,
		Name:   name,
		Domain: *domain,
		CRUD:   *crud,
	}

	if err := ichigen.Generate(config); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Generated %s: %s in domain %s\n", genType, name, *domain)
}

func printUsage() {
	usage := `
Go Schematic Generator for ichi-go

Usage:
  ichigen g <type> <name> --domain=<domain> [--crud]

Types:
  controller   Generate HTTP controller
  service      Generate service layer
  repository   Generate repository
  validator    Generate validator
  dto          Generate DTOs
  full         Generate complete stack

Examples:
  ichigen g controller product --domain=catalog
  ichigen g full product --domain=catalog --crud
`
	fmt.Println(strings.TrimSpace(usage))
}
