package main

import (
	"fmt"
	"html"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

type ResponseCode struct {
	Code        string
	Description string
}

const outputTemplate = `// Code generated by generate-response-codes/main.go; DO NOT EDIT.

package api2c2p

import "fmt"

// ResponseCode represents a 2C2P response code
type ResponseCode string

// Description returns a human-readable description of the response code
func (c ResponseCode) Description() string {
	switch c {
	{{- range .}}
	case "{{.Code}}":
		return "{{.Description}}"
	{{- end}}
	default:
		return fmt.Sprintf("Unknown response code: %s", string(c))
	}
}

// Known response codes
const (
	{{- range .}}
	RespCode{{.Code}} ResponseCode = "{{.Code}}" // {{.Description}}
	{{- end}}
)
`

func main() {
	// Read HTML file
	htmlPath := filepath.Join("docs", "2c2p", "Payment Response Codes.html")
	content, err := os.ReadFile(htmlPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	// Find the table in the HTML content
	tableRegex := regexp.MustCompile(`<table>.*?<tbody>(.*?)</tbody>.*?</table>`)
	rowRegex := regexp.MustCompile(`<tr>.*?<td[^>]*>(.*?)</td>.*?<td[^>]*>(.*?)</td>.*?</tr>`)

	// Extract table content
	tableMatch := tableRegex.FindStringSubmatch(string(content))
	if len(tableMatch) < 2 {
		fmt.Fprintf(os.Stderr, "No table found in HTML\n")
		os.Exit(1)
	}

	// Parse rows
	var codes []ResponseCode
	rows := rowRegex.FindAllStringSubmatch(tableMatch[1], -1)
	for _, row := range rows {
		if len(row) < 3 {
			continue
		}
		code := strings.TrimSpace(html.UnescapeString(row[1]))
		desc := strings.TrimSpace(html.UnescapeString(row[2]))
		if code == "Code" || code == "" {
			continue
		}
		codes = append(codes, ResponseCode{
			Code:        code,
			Description: desc,
		})
	}

	if len(codes) == 0 {
		fmt.Fprintf(os.Stderr, "No response codes found in table\n")
		os.Exit(1)
	}

	// Generate Go code
	tmpl, err := template.New("codes").Parse(outputTemplate)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing template: %v\n", err)
		os.Exit(1)
	}

	outputPath := "payment_response_codes.go"
	f, err := os.Create(outputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	if err := tmpl.Execute(f, codes); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing template: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated %s with %d response codes\n", outputPath, len(codes))
}
