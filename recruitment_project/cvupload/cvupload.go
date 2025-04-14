package cvupload

import (
    "net/http"
    "os"
    "path/filepath"
    "strings"
    "github.com/gin-gonic/gin"
    "github.com/FrancoisDuvet/friendly-chainsaw/recruitment_project/resume"
)

const uploadDir = "./uploads"

func init() {
    if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
        os.Mkdir(uploadDir, os.ModePerm)
    }
}

func SetupCVUploadRoutes(r *gin.Engine) {
    r.POST("/applicant/upload-cv", uploadCVHandler)
}

func uploadCVHandler(c *gin.Context) {
    file, err := c.FormFile("cv")
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
        return
    }

    if !strings.HasSuffix(file.Filename, ".pdf") {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Only PDF files allowed"})
        return
    }

    filePath := filepath.Join(uploadDir, file.Filename)
    if err := c.SaveUploadedFile(file, filePath); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
        return
    }

    if err := resume.ValidateResume(filePath); err != nil {
        os.Remove(filePath)
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "CV uploaded and validated!"})
}
