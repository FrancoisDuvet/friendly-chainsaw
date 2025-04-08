package resume

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "os"
    "path/filepath"
    "strings"
    "gorm.io/gorm"
    "github.com/gin-gonic/gin"
    "github.com/pdfcpu/pdfcpu/pkg/api"
)

// Directory to store uploaded resumes
const uploadDir = "./uploads"

// Ensure the upload directory exists
func init() {
    if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
        os.Mkdir(uploadDir, os.ModePerm)
    }
}

// Resume struct for parsed data
type Resume struct {
    Name       string   `json:"name"`
    Email      string   `json:"email"`
    Phone      string   `json:"phone"`
    Skills     []string `json:"skills"`
    Education  string   `json:"education"`
    Experience string   `json:"experience"`
}

// Upload and parse resume handler
func uploadAndParseResumeHandler(c *gin.Context) {
    // Parse the uploaded file
    file, err := c.FormFile("resume")
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

    // Extract text from the PDF
    text, err := extractTextFromPDF(filePath)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to extract text from PDF"})
        return
    }

    // Call Google Gemini API for parsing
    parsedResume, err := callGeminiAPI(text)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse resume"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Resume parsed successfully", "data": parsedResume})
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

// Call Google Gemini API for resume parsing
func callGeminiAPI(resumeText string) (*Resume, error) {
    apiKey := os.Getenv("GEMINI_API_KEY")
    if apiKey == "" {
        return nil, fmt.Errorf("Gemini API key not found in environment variables")
    }

    // Prepare the request payload
    payload := map[string]string{
        "prompt": fmt.Sprintf("Summarize the following resume and extract key details such as name, email, phone, skills, education, and experience:\n\n%s", resumeText),
    }
    payloadBytes, err := json.Marshal(payload)
    if err != nil {
        return nil, err
    }

    // Make the API request
    req, err := http.NewRequest("POST", "https://gemini.googleapis.com/v1/summarize", bytes.NewBuffer(payloadBytes))
    if err != nil {
        return nil, err
    }
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    // Read and parse the response
    if resp.StatusCode != http.StatusOK {
        body, _ := ioutil.ReadAll(resp.Body)
        return nil, fmt.Errorf("Gemini API error: %s", string(body))
    }

    var result struct {
        Summary string `json:"summary"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    // Parse the summary into a Resume struct
    parsedResume := parseResumeSummary(result.Summary)
    return parsedResume, nil
}

// Parse the summary text into a Resume struct
func parseResumeSummary(summary string) *Resume {
    // This is a simple example. You can improve this logic to extract details more accurately.
    lines := strings.Split(summary, "\n")
    resume := &Resume{}

    for _, line := range lines {
        if strings.Contains(strings.ToLower(line), "name:") {
            resume.Name = strings.TrimSpace(strings.Split(line, ":")[1])
        } else if strings.Contains(strings.ToLower(line), "email:") {
            resume.Email = strings.TrimSpace(strings.Split(line, ":")[1])
        } else if strings.Contains(strings.ToLower(line), "phone:") {
            resume.Phone = strings.TrimSpace(strings.Split(line, ":")[1])
        } else if strings.Contains(strings.ToLower(line), "skills:") {
            resume.Skills = strings.Split(strings.TrimSpace(strings.Split(line, ":")[1]), ",")
        } else if strings.Contains(strings.ToLower(line), "education:") {
            resume.Education = strings.TrimSpace(strings.Split(line, ":")[1])
        } else if strings.Contains(strings.ToLower(line), "experience:") {
            resume.Experience = strings.TrimSpace(strings.Split(line, ":")[1])
        }
    }

    return resume
}

func SetupResumeRoutes(r *gin.Engine) {

    // Resume upload and parsing route
    r.POST("/recruiter/parse-resume", uploadAndParseResumeHandler)

    fmt.Println("Server started at :8080")
    r.Run(":8080")
}