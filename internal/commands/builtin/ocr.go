package builtin

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"neuroshell/internal/commands"
	"neuroshell/internal/commands/printing"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"

	"github.com/gen2brain/go-fitz"
)

// OCRCommand implements the \ocr command for converting PDF to text using DeepInfra OCR API.
// It processes PDF files by converting pages to images and sending them to OCR services.
type OCRCommand struct{}

// Name returns the command name "ocr" for registration and lookup.
func (c *OCRCommand) Name() string {
	return "ocr"
}

// ParseMode returns ParseModeKeyValue for key-value argument parsing.
func (c *OCRCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the ocr command does.
func (c *OCRCommand) Description() string {
	return "Convert PDF to text/markdown using DeepInfra OCR API"
}

// Usage returns the syntax and usage examples for the ocr command.
func (c *OCRCommand) Usage() string {
	return "\\ocr[pdf=path/to/file.pdf, output=output.md, api_key=..., model=..., max_tokens=4500, pages=1-3]"
}

// HelpInfo returns structured help information for the ocr command.
func (c *OCRCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       c.Usage(),
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "pdf",
				Description: "Path to the PDF file to process",
				Required:    true,
				Type:        "string",
			},
			{
				Name:        "output",
				Description: "Output markdown file path (defaults to {pdf_name}_ocr.md)",
				Required:    false,
				Type:        "string",
			},
			{
				Name:        "api_key",
				Description: "DeepInfra API key (can use env var DEEPINFRA_API_KEY)",
				Required:    false,
				Type:        "string",
			},
			{
				Name:        "model",
				Description: "OCR model to use (defaults to allenai/olmOCR-7B-0825)",
				Required:    false,
				Type:        "string",
				Default:     "allenai/olmOCR-7B-0825",
			},
			{
				Name:        "max_tokens",
				Description: "Maximum tokens for response",
				Required:    false,
				Type:        "integer",
				Default:     "4500",
			},
			{
				Name:        "temperature",
				Description: "Temperature for response generation",
				Required:    false,
				Type:        "float",
				Default:     "0.0",
			},
			{
				Name:        "pages",
				Description: "Specific pages to process (e.g., '1-3,5' or 'all')",
				Required:    false,
				Type:        "string",
				Default:     "all",
			},
			{
				Name:        "to",
				Description: "Variable name to store the OCR result",
				Required:    false,
				Type:        "string",
				Default:     "_output",
			},
			{
				Name:        "silent",
				Description: "Suppress output display",
				Required:    false,
				Type:        "boolean",
				Default:     "false",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Description: "OCR process entire PDF with default settings",
				Command:     "\\ocr[pdf=document.pdf]",
			},
			{
				Description: "OCR specific pages and save to custom output file",
				Command:     "\\ocr[pdf=document.pdf, pages=1-3, output=first_pages.md]",
			},
			{
				Description: "OCR with custom API key and model",
				Command:     "\\ocr[pdf=document.pdf, api_key=your_key, model=custom-ocr-model]",
			},
		},
	}
}

// OCRRequest represents the request structure for DeepInfra OCR API
type OCRRequest struct {
	Model       string       `json:"model"`
	Messages    []OCRMessage `json:"messages"`
	MaxTokens   int          `json:"max_tokens"`
	Temperature float64      `json:"temperature"`
}

// OCRMessage represents a message in the OCR request
type OCRMessage struct {
	Role    string       `json:"role"`
	Content []OCRContent `json:"content"`
}

// OCRContent represents content in an OCR message (text or image)
type OCRContent struct {
	Type     string    `json:"type"`
	Text     string    `json:"text,omitempty"`
	ImageURL *ImageURL `json:"image_url,omitempty"`
}

// ImageURL represents an image URL with base64 data
type ImageURL struct {
	URL string `json:"url"`
}

// OCRResponse represents the response from DeepInfra OCR API
type OCRResponse struct {
	Choices []OCRChoice `json:"choices"`
	Error   *OCRError   `json:"error,omitempty"`
}

// OCRChoice represents a choice in the OCR response
type OCRChoice struct {
	Message OCRResponseMessage `json:"message"`
}

// OCRResponseMessage represents the message in an OCR response choice
type OCRResponseMessage struct {
	Content string `json:"content"`
}

// OCRError represents an error in the OCR response
type OCRError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

// Execute processes the OCR command with the given options and arguments.
func (c *OCRCommand) Execute(options map[string]string, _ string) error {
	printer := printing.NewDefaultPrinter()

	// Extract and validate options
	config, err := c.extractOptions(options)
	if err != nil {
		return fmt.Errorf("failed to extract options: %w", err)
	}

	// Validate PDF file exists
	if _, err := os.Stat(config.PDFPath); os.IsNotExist(err) {
		return fmt.Errorf("PDF file does not exist: %s", config.PDFPath)
	}

	printer.Printf("Starting OCR processing: %s\n", config.PDFPath)

	// TODO: Implement PDF to PNG conversion
	pageImages, err := c.convertPDFToImages(config.PDFPath, config.Pages)
	if err != nil {
		return fmt.Errorf("failed to convert PDF to images: %w", err)
	}

	printer.Printf("Converted pages to images: %d pages\n", len(pageImages))

	// Process each page through OCR API
	var allResults []string
	httpServiceInterface, err := services.GetGlobalRegistry().GetService("http_request")
	if err != nil {
		return fmt.Errorf("failed to get HTTP service: %w", err)
	}
	httpService, ok := httpServiceInterface.(*services.HTTPRequestService)
	if !ok {
		return fmt.Errorf("HTTP service type assertion failed")
	}

	for i, imageData := range pageImages {
		result, err := c.processPageOCR(httpService, imageData, config)
		if err != nil {
			return fmt.Errorf("failed to process page %d: %w", i+1, err)
		}
		allResults = append(allResults, result)
		printer.Printf("Processed page %d ✓\n", i+1)
	}

	// Combine results and save to file
	combinedResult := strings.Join(allResults, "\n\n---\n\n")

	err = c.saveToFile(config.OutputPath, combinedResult)
	if err != nil {
		return fmt.Errorf("failed to save OCR result: %w", err)
	}

	// Store in variable
	variableService, err := services.GetGlobalVariableService()
	if err != nil {
		return fmt.Errorf("failed to get variable service: %w", err)
	}

	// Use SetSystemVariable for underscore-prefixed variables
	if strings.HasPrefix(config.ToVariable, "_") {
		err = variableService.SetSystemVariable(config.ToVariable, combinedResult)
	} else {
		err = variableService.Set(config.ToVariable, combinedResult)
	}
	if err != nil {
		return fmt.Errorf("failed to set variable %s: %w", config.ToVariable, err)
	}

	// Store output file path
	err = variableService.SetSystemVariable("_ocr_file", config.OutputPath)
	if err != nil {
		return fmt.Errorf("failed to set _ocr_file variable: %w", err)
	}

	if !config.Silent {
		printer.Printf("OCR completed successfully: %s\n", config.OutputPath)
		printer.Printf("Result stored in variable: %s\n", config.ToVariable)
	}

	return nil
}

// OCRConfig holds the configuration for OCR processing
type OCRConfig struct {
	PDFPath     string
	OutputPath  string
	APIKey      string
	Model       string
	MaxTokens   int
	Temperature float64
	Pages       string
	ToVariable  string
	Silent      bool
}

// extractOptions extracts and validates options from the command
func (c *OCRCommand) extractOptions(options map[string]string) (*OCRConfig, error) {
	config := &OCRConfig{
		Model:       "allenai/olmOCR-7B-0825",
		MaxTokens:   4500,
		Temperature: 0.0,
		Pages:       "all",
		ToVariable:  "_output",
		Silent:      false,
	}

	// Required: PDF path
	pdfPath, exists := options["pdf"]
	if !exists || pdfPath == "" {
		return nil, fmt.Errorf("pdf option is required")
	}
	config.PDFPath = pdfPath

	// Optional: Output path (default to PDF name + _ocr.md)
	if outputPath, exists := options["output"]; exists && outputPath != "" {
		config.OutputPath = outputPath
	} else {
		baseName := strings.TrimSuffix(filepath.Base(pdfPath), filepath.Ext(pdfPath))
		config.OutputPath = baseName + "_ocr.md"
	}

	// Optional: API key (check env var if not provided)
	if apiKey, exists := options["api_key"]; exists && apiKey != "" {
		config.APIKey = apiKey
	} else {
		config.APIKey = os.Getenv("DEEPINFRA_API_KEY")
		if config.APIKey == "" {
			return nil, fmt.Errorf("API key required: set api_key option or DEEPINFRA_API_KEY environment variable")
		}
	}

	// Optional: Model
	if model, exists := options["model"]; exists && model != "" {
		config.Model = model
	}

	// Optional: Max tokens
	if maxTokensStr, exists := options["max_tokens"]; exists && maxTokensStr != "" {
		maxTokens, err := strconv.Atoi(maxTokensStr)
		if err != nil {
			return nil, fmt.Errorf("invalid max_tokens value: %s", maxTokensStr)
		}
		config.MaxTokens = maxTokens
	}

	// Optional: Temperature
	if tempStr, exists := options["temperature"]; exists && tempStr != "" {
		temp, err := strconv.ParseFloat(tempStr, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid temperature value: %s", tempStr)
		}
		config.Temperature = temp
	}

	// Optional: Pages
	if pages, exists := options["pages"]; exists && pages != "" {
		config.Pages = pages
	}

	// Optional: To variable
	if toVar, exists := options["to"]; exists && toVar != "" {
		config.ToVariable = toVar
	}

	// Optional: Silent
	if silentStr, exists := options["silent"]; exists && silentStr != "" {
		silent, err := strconv.ParseBool(silentStr)
		if err != nil {
			return nil, fmt.Errorf("invalid silent value: %s", silentStr)
		}
		config.Silent = silent
	}

	return config, nil
}

// convertPDFToImages converts PDF pages to base64-encoded PNG images
func (c *OCRCommand) convertPDFToImages(pdfPath string, pagesSpec string) ([]string, error) {
	// Open PDF document
	doc, err := fitz.New(pdfPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open PDF document: %w", err)
	}
	defer func() {
		_ = doc.Close() // Ignore error on close
	}()

	// Get total number of pages
	totalPages := doc.NumPage()
	if totalPages == 0 {
		return nil, fmt.Errorf("PDF document has no pages")
	}

	// Parse page specification
	pageNumbers, err := c.parsePageSpec(pagesSpec, totalPages)
	if err != nil {
		return nil, fmt.Errorf("invalid page specification: %w", err)
	}

	// Convert each page to base64-encoded PNG
	var base64Images []string
	for _, pageNum := range pageNumbers {
		// Convert page to image (pageNum is 0-based in go-fitz)
		img, err := doc.Image(pageNum - 1)
		if err != nil {
			return nil, fmt.Errorf("failed to convert page %d to image: %w", pageNum, err)
		}

		// Encode image as PNG and then base64
		var buf bytes.Buffer
		err = png.Encode(&buf, img)
		if err != nil {
			return nil, fmt.Errorf("failed to encode page %d as PNG: %w", pageNum, err)
		}

		// Convert to base64
		base64Data := base64.StdEncoding.EncodeToString(buf.Bytes())
		base64Images = append(base64Images, base64Data)
	}

	return base64Images, nil
}

// parsePageSpec parses page specification string and returns slice of page numbers
func (c *OCRCommand) parsePageSpec(pagesSpec string, totalPages int) ([]int, error) {
	if pagesSpec == "" || pagesSpec == "all" {
		// Return all pages
		pages := make([]int, totalPages)
		for i := 0; i < totalPages; i++ {
			pages[i] = i + 1
		}
		return pages, nil
	}

	var pageNumbers []int
	ranges := strings.Split(pagesSpec, ",")

	for _, rangeStr := range ranges {
		rangeStr = strings.TrimSpace(rangeStr)
		if rangeStr == "" {
			continue
		}

		// Check if it's a range (e.g., "1-3") or single page
		if strings.Contains(rangeStr, "-") {
			parts := strings.Split(rangeStr, "-")
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid range format: %s", rangeStr)
			}

			start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
			if err != nil {
				return nil, fmt.Errorf("invalid start page: %s", parts[0])
			}

			end, err := strconv.Atoi(strings.TrimSpace(parts[1]))
			if err != nil {
				return nil, fmt.Errorf("invalid end page: %s", parts[1])
			}

			if start < 1 || start > totalPages {
				return nil, fmt.Errorf("start page %d out of range (1-%d)", start, totalPages)
			}

			if end < 1 || end > totalPages {
				return nil, fmt.Errorf("end page %d out of range (1-%d)", end, totalPages)
			}

			if start > end {
				return nil, fmt.Errorf("start page %d greater than end page %d", start, end)
			}

			// Add all pages in range
			for i := start; i <= end; i++ {
				pageNumbers = append(pageNumbers, i)
			}
		} else {
			// Single page
			pageNum, err := strconv.Atoi(rangeStr)
			if err != nil {
				return nil, fmt.Errorf("invalid page number: %s", rangeStr)
			}

			if pageNum < 1 || pageNum > totalPages {
				return nil, fmt.Errorf("page %d out of range (1-%d)", pageNum, totalPages)
			}

			pageNumbers = append(pageNumbers, pageNum)
		}
	}

	// Remove duplicates and sort
	uniquePages := make(map[int]bool)
	var result []int
	for _, page := range pageNumbers {
		if !uniquePages[page] {
			uniquePages[page] = true
			result = append(result, page)
		}
	}

	return result, nil
}

// processPageOCR sends a page image to DeepInfra OCR API and returns the result
func (c *OCRCommand) processPageOCR(httpService *services.HTTPRequestService, imageData string, config *OCRConfig) (string, error) {
	// Create OCR request
	request := OCRRequest{
		Model: config.Model,
		Messages: []OCRMessage{
			{
				Role: "user",
				Content: []OCRContent{
					{
						Type: "image_url",
						ImageURL: &ImageURL{
							URL: fmt.Sprintf("data:image/png;base64,%s", imageData),
						},
					},
					{
						Type: "text",
						Text: c.getOCRPrompt(),
					},
				},
			},
		},
		MaxTokens:   config.MaxTokens,
		Temperature: config.Temperature,
	}

	// Marshal request to JSON
	requestBody, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Prepare HTTP headers
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", config.APIKey),
	}

	// Send request to DeepInfra API
	response, err := httpService.Post(
		"https://api.deepinfra.com/v1/openai/chat/completions",
		string(requestBody),
		headers,
	)
	if err != nil {
		return "", fmt.Errorf("failed to send request to DeepInfra: %w", err)
	}

	// Check HTTP status code first
	if response.StatusCode != 200 {
		return "", fmt.Errorf("API returned status %d: %s", response.StatusCode, response.Body)
	}

	// Parse response
	var ocrResponse OCRResponse
	err = json.Unmarshal([]byte(response.Body), &ocrResponse)
	if err != nil {
		return "", fmt.Errorf("failed to parse response (body: %s): %w", response.Body, err)
	}

	// Check for API errors
	if ocrResponse.Error != nil {
		return "", fmt.Errorf("API error: %s (%s)", ocrResponse.Error.Message, ocrResponse.Error.Type)
	}

	// Extract content
	if len(ocrResponse.Choices) == 0 {
		return "", fmt.Errorf("no choices in API response (full response: %s)", response.Body)
	}

	return ocrResponse.Choices[0].Message.Content, nil
}

// getOCRPrompt returns the prompt text for OCR processing
func (c *OCRCommand) getOCRPrompt() string {
	return "Attached is one page of a document that you must process. " +
		"Just return the plain text representation of this document as if you were reading it naturally. " +
		"Convert equations to LaTeX and tables to markdown.\n" +
		"Return your output as markdown, with a front matter section on top specifying values for the " +
		"primary_language, is_rotation_valid, rotation_correction, is_table, and is_diagram parameters."
}

// saveToFile saves the OCR result to a file
func (c *OCRCommand) saveToFile(filePath string, content string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write content to file
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", filePath, err)
	}

	return nil
}

// IsReadOnly returns false as the ocr command creates files and makes HTTP requests.
func (c *OCRCommand) IsReadOnly() bool {
	return false
}

// Register the OCR command
func init() {
	if err := commands.GetGlobalRegistry().Register(&OCRCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register ocr command: %v", err))
	}
}
