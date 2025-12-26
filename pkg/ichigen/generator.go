package ichigen

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed templates/*
var templatesFS embed.FS

type Config struct {
	Type   string
	Name   string
	Domain string
	CRUD   bool
}

type TemplateData struct {
	PackageName string
	StructName  string
	VarName     string
	Domain      string
	LowerName   string
	TableName   string
	HasCRUD     bool
}

func Generate(config Config) error {
	if config.Domain == "" {
		return fmt.Errorf("domain is required")
	}

	data := prepareTemplateData(config)

	switch config.Type {
	case "controller", "c":
		return generateController(data)
	case "service", "s":
		return generateService(data)
	case "repository", "r", "repo":
		return generateRepository(data)
	case "validator", "v":
		return generateValidator(data)
	case "dto", "d":
		return generateDTO(data)
	case "full", "f":
		return generateFull(data)
	default:
		return fmt.Errorf("unknown type: %s", config.Type)
	}
}

func prepareTemplateData(config Config) TemplateData {
	structName := toPascalCase(config.Name)
	varName := toCamelCase(config.Name)
	lowerName := strings.ToLower(config.Name)

	return TemplateData{
		PackageName: config.Domain,
		StructName:  structName,
		VarName:     varName,
		Domain:      config.Domain,
		LowerName:   lowerName,
		TableName:   toSnakeCase(config.Name) + "s",
		HasCRUD:     config.CRUD,
	}
}

func generateController(data TemplateData) error {
	path := filepath.Join("internal", "applications", data.Domain, "controller", data.LowerName+"_controller.go")
	return renderTemplate("templates/controller.tmpl", path, data)
}

func generateService(data TemplateData) error {
	path := filepath.Join("internal", "applications", data.Domain, "service", data.LowerName+"_service.go")
	return renderTemplate("templates/service.tmpl", path, data)
}

func generateRepository(data TemplateData) error {
	path := filepath.Join("internal", "applications", data.Domain, "repository", data.LowerName+"_repository.go")
	return renderTemplate("templates/repository.tmpl", path, data)
}

func generateValidator(data TemplateData) error {
	path := filepath.Join("internal", "applications", data.Domain, "validators", data.LowerName+"_validator.go")
	return renderTemplate("templates/validator.tmpl", path, data)
}

func generateDTO(data TemplateData) error {
	path := filepath.Join("internal", "applications", data.Domain, "dto", data.LowerName+"_dto.go")
	return renderTemplate("templates/dto.tmpl", path, data)
}

func generateProviders(data TemplateData) error {
	path := filepath.Join("internal", "applications", data.Domain, "providers.go")
	return renderTemplate("templates/providers.tmpl", path, data)
}

func generateRegistry(data TemplateData) error {
	path := filepath.Join("internal", "applications", data.Domain, "registry.go")
	return renderTemplate("templates/registry.tmpl", path, data)
}

func generateFull(data TemplateData) error {
	generators := []func(TemplateData) error{
		generateDTO,
		generateValidator,
		generateRepository,
		generateService,
		generateController,
		generateProviders,
		generateRegistry,
	}

	for _, gen := range generators {
		if err := gen(data); err != nil {
			return err
		}
	}

	fmt.Println("✓ Complete stack generated:")
	fmt.Println("  - DTO")
	fmt.Println("  - Validator")
	fmt.Println("  - Repository")
	fmt.Println("  - Service")
	fmt.Println("  - Controller")
	fmt.Println("  - Providers")
	fmt.Println("  - Registry")

	return nil
}

func renderTemplate(tmplPath, outputPath string, data TemplateData) error {
	tmplContent, err := templatesFS.ReadFile(tmplPath)
	if err != nil {
		return fmt.Errorf("failed to read template: %w", err)
	}

	tmpl, err := template.New("gen").Parse(string(tmplContent))
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	fmt.Printf("  Created: %s\n", outputPath)
	return nil
}

func toPascalCase(s string) string {
	words := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-' || r == ' '
	})
	for i, word := range words {
		words[i] = strings.Title(strings.ToLower(word))
	}
	return strings.Join(words, "")
}

func toCamelCase(s string) string {
	pascal := toPascalCase(s)
	if len(pascal) == 0 {
		return ""
	}
	return strings.ToLower(pascal[:1]) + pascal[1:]
}

func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}
