package utils

import (
	"bytes"
	"file-manager/config"
	"fmt"
	"html/template"
	"io"
	"mime/multipart"
	"net/http"
	"reflect"
)

func ParseTemplate(templateBytes []byte, templateData map[string]interface{}) (bytes.Buffer, string, error) {
	// Define custom functions used in the template
	funcMap := template.FuncMap{
		"lenSafe": func(v any) int {
			if v == nil {
				return 0
			}
			val := reflect.ValueOf(v)
			switch val.Kind() {
			case reflect.Slice, reflect.Array, reflect.Map, reflect.String:
				return val.Len()
			default:
				return 0
			}
		},
	}

	//Parse the template with custom functions
	tmpl, err := template.New("uploaded").Funcs(funcMap).Parse(string(templateBytes))
	if err != nil {
		return bytes.Buffer{}, "", err
	}

	// Render HTML to buffer
	var htmlBuf bytes.Buffer
	err = tmpl.Execute(&htmlBuf, templateData)
	if err != nil {
		return bytes.Buffer{}, "", err
	}

	// Create multipart form for PDF service goteb
	var formBuf bytes.Buffer
	writer := multipart.NewWriter(&formBuf)
	part, err := writer.CreateFormFile("files", "index.html")
	if err != nil {
		return bytes.Buffer{}, "", err
	}
	_, err = io.Copy(part, &htmlBuf)
	if err != nil {
		return bytes.Buffer{}, "", err
	}

	writer.Close()

	contentType := writer.FormDataContentType()

	return formBuf, contentType, nil
}

func ParseTemplateToPDF(formBuf bytes.Buffer, contentType string) (bytes.Buffer, error) {

	// Load configuration to get the Gotenberg URL
	cfg := config.LoadConfig()

	// Send to PDF conversion service goteb
	pdfServiceURL := fmt.Sprintf("%s/forms/chromium/convert/html", cfg.GotenbergURL)
	req, err := http.NewRequest("POST", pdfServiceURL, &formBuf)
	if err != nil {
		return bytes.Buffer{}, err
	}
	req.Header.Set("Content-Type", contentType)

	// Create http client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return bytes.Buffer{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return bytes.Buffer{}, fmt.Errorf("PDF service returned error: %s", resp.Status)
	}

	// Read the PDF into a buffer
	var pdfBuf bytes.Buffer
	_, err = io.Copy(&pdfBuf, resp.Body)
	if err != nil {
		return bytes.Buffer{}, err
	}

	return pdfBuf, nil
}
