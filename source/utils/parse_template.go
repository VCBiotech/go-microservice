package utils

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"mime/multipart"
	"net/http"
)

func ParseTemplate(templateBytes []byte, templateData map[string]interface{}) (bytes.Buffer, string, error) {
	//Parse the template
	tmpl, err := template.New("uploaded").Parse(string(templateBytes))
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

	// Send to PDF conversion service goteb
	pdfServiceURL := "http://localhost:3001/forms/chromium/convert/html"
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
