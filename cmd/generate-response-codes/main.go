package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"text/template"
)

type PaymentResponseCode struct {
	Code        string
	Description string
}

func toConstName(desc string, code string) string {
	// Pad code with zeros to 4 digits
	paddedCode := fmt.Sprintf("%04s", code)

	// Remove any special characters and convert to title case
	words := strings.Fields(desc)
	for i, word := range words {
		// Clean the word of any special characters
		word = regexp.MustCompile(`[^a-zA-Z0-9]+`).ReplaceAllString(word, "")
		words[i] = strings.Title(strings.ToLower(word))
	}
	return "Code" + paddedCode + strings.Join(words, "")
}

const outputTemplate = `// Code generated by generate-response-codes/main.go; DO NOT EDIT.

package api2c2p

import "fmt"

// PaymentResponseCode represents a 2C2P response code
type PaymentResponseCode string

// Description returns a human-readable description of the response code
func (c PaymentResponseCode) Description() string {
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
	{{toConstName .Description .Code}} PaymentResponseCode = "{{.Code}}" // {{.Description}}
	{{- end}}
)
`

func main() {
	// Read the response codes from the CSV file
	file, err := os.Open("docs/2c2p/response-code-payment.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	codes, err := reader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	// Format each code as a 4-digit string with leading zeros
	for i, code := range codes {
		for j, c := range code {
			codes[i][j] = fmt.Sprintf("%04s", c)
		}
	}

	// Parse rows
	var paymentCodes []PaymentResponseCode
	for i, row := range codes {
		if i == 0 || len(row) < 2 { // Skip header row and invalid rows
			continue
		}
		code := strings.TrimSpace(row[0])
		desc := strings.TrimSpace(row[1])
		if code == "" || desc == "" {
			continue
		}
		paymentCodes = append(paymentCodes, PaymentResponseCode{
			Code:        code,
			Description: desc,
		})
	}

	if len(paymentCodes) == 0 {
		log.Fatal("No response codes found in CSV")
	}

	// Generate Go code
	tmpl, err := template.New("codes").Funcs(template.FuncMap{"toConstName": func(desc string, code string) string { return toConstName(desc, code) }}).Parse(outputTemplate)
	if err != nil {
		log.Fatalf("Error parsing template: %v", err)
	}

	outputPath := "payment_response_codes.go"
	f, err := os.Create(outputPath)
	if err != nil {
		log.Fatalf("Error creating output file: %v", err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, paymentCodes); err != nil {
		log.Fatalf("Error executing template: %v", err)
	}

	fmt.Printf("Generated %s with %d response codes\n", outputPath, len(paymentCodes))
}
