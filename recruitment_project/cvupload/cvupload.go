package main

import (
    "bytes"
    "fmt"
    "net/http"
    "os"
    "path/filepath"
    "strings"

    "github.com/gin-gonic/gin"
    "github.com/pdfcpu/pdfcpu/pkg/api"
)

// Directory to store uploaded CVs
const uploadDir = "./uploads"

// Ensure the upload directory exists
func init() {
    if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
        os.Mkdir(uploadDir, os.ModePerm)
    }
}

// Upload CV handler
func uploadCVHandler(c *gin.Context) {
    // Parse the uploaded file
    file, err := c.FormFile("cv")
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to upload file"})
        return
    }

    // Ensure the file is a PDF
    if !strings.HasSuffix(file.Filename, ".pdf") {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Only PDF files are allowed"})
        return
    }

    // Save the file to the uploads directory
    filePath := filepath.Join(uploadDir, file.Filename)
    if err := c.SaveUploadedFile(file, filePath); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
        return
    }

    // Validate the PDF contents
    if err := validateResume(filePath); err != nil {
        // Delete the invalid file
        os.Remove(filePath)
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "CV uploaded successfully"})
}

// Validate the resume for required fields
func validateResume(filePath string) error {
    // Extract text from the PDF
    text, err := extractTextFromPDF(filePath)
    if err != nil {
        return fmt.Errorf("Failed to parse PDF: %v", err)
    }

    // Check for required fields
    requiredFields := []string{"Name", "Skills", "Education"}
    for _, field := range requiredFields {
        if !strings.Contains(strings.ToLower(text), strings.ToLower(field)) {
            return fmt.Errorf("Resume is incomplete. Missing field: %s", field)
        }
    }

    return nil
}

// Extract text from a PDF file
func extractTextFromPDF(filePath string) (string, error) {
    var buf bytes.Buffer

    // Use pdfcpu to extract text from the PDF
    err := api.ExtractTextFile(filePath, &buf, nil)
    if err != nil {
        return "", err
    }

    return buf.String(), nil
}

func main() {
    r := gin.Default()

    // CV upload route
    r.POST("/applicant/upload-cv", uploadCVHandler)

    fmt.Println("Server started at :8080")
    r.Run(":8080")
}