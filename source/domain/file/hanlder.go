package file

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"vcbiotech/microservice/telemetry"
	"vcbiotech/microservice/utils"

	"github.com/labstack/echo/v4"
)

type Repo interface {
	FindById(ctx context.Context, id uint64) (File, error)
}

type FileHandler struct {
	Repo Repo
}

func NewFileHandler(repo Repo) *FileHandler {
	return &FileHandler{
		Repo: repo,
	}
}

func (f *FileHandler) FindById(c echo.Context) error {
	logger := telemetry.SLogger(c.Request().Context())
	idParam := c.Param("id")
	const decimal = 10
	const bitSize = 64

	id, err := strconv.ParseUint(idParam, decimal, bitSize)
	if err != nil {
		errMsg := map[string]string{"Error": err.Error()}
		logger.Error("Could not parse id", errMsg)
		return c.JSON(http.StatusBadRequest, errMsg)
	}

	file, err := f.Repo.FindById(c.Request().Context(), id)
	if err != nil {
		errMsg := map[string]string{"Error": err.Error()}
		logger.Error("Could not find file", errMsg)
		return c.JSON(http.StatusInternalServerError, errMsg)
	}

	return c.JSON(http.StatusOK, file)
}

// Handler to render HTML template from uploaded file and data and then upload to the s3 bucket
func (fr *FileHandler) Insert(c echo.Context) error {
	err := c.Request().ParseMultipartForm(10 << 20)
	if err != nil {
		errMsg := map[string]string{"Error": "Unable to parse form"}
		return c.JSON(http.StatusBadRequest, errMsg)
	}

	templateFile, _, err := c.Request().FormFile("template")
	if err != nil {
		errMsg := map[string]string{"Error": "Error retrieving template file"}
		return c.JSON(http.StatusBadRequest, errMsg)
	}
	defer templateFile.Close()

	templateBytes, err := io.ReadAll(templateFile)
	if err != nil {
		errMsg := map[string]string{"Error": "Error reading template file"}
		return c.JSON(http.StatusBadRequest, errMsg)
	}

	// Get the JSON data
	jsonData := c.FormValue("jsonData")
	if jsonData == "" {
		errMsg := map[string]string{"Error": "jsonData field is required"}
		return c.JSON(http.StatusBadRequest, errMsg)
	}

	// Parse the JSON data into a map
	var templateData map[string]interface{}
	err = json.Unmarshal([]byte(jsonData), &templateData)
	if err != nil {
		errMsg := map[string]string{"Error": fmt.Sprintf("Failed to parse jsonData: %v", err)}
		return c.JSON(http.StatusBadRequest, errMsg)
	}

	templateData["BucketName"] = "name"

	//Parse the template
	formBuf, contentType, err := utils.ParseTemplate(templateBytes, templateData)
	if err != nil {
		errMsg := map[string]string{"Error": fmt.Sprintf("Failed to parse template: %v", err)}
		return c.JSON(http.StatusInternalServerError, errMsg)
	}

	// Send to PDF conversion service goteb
	pdfBuf, err := utils.ParseTemplateToPDF(formBuf, contentType)
	if err != nil {
		errMsg := map[string]string{"Error": fmt.Sprintf("Failed to convert to PDF: %v", err)}
		return c.JSON(http.StatusInternalServerError, errMsg)
	}

	// Generate a unique filename for the PDF
	pdfFilename := fmt.Sprintf("generated-pdf-%s.pdf", time.Now().Format("20060102-150405"))

	s3Client := utils.NewS3Client("test-file-manager-2025")

	// Upload the PDF to S3
	err = s3Client.UploadObject(pdfFilename, &pdfBuf)
	if err != nil {
		errMsg := map[string]string{"Error": fmt.Sprintf("Failed to upload PDF to S3: %v", err)}
		return c.JSON(http.StatusInternalServerError, errMsg)
	}

	// Create the user message as JSON
	userMessage := map[string]string{
		"message": "File has been successfully created and uploaded to S3.",
	}

	return c.JSON(http.StatusOK, userMessage)
}
