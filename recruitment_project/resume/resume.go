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
	"github.com/gin-gonic/gin"
	"github.com/pdfcpu/pdfcpu/pkg/api"
)

const uploadDir = "./uploads"

func init() {
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.Mkdir(uploadDir, os.ModePerm)
	}
}

type Resume struct {
	Name       string   `json:"name"`
	Email      string   `json:"email"`
	Phone      string   `json:"phone"`
	Skills     []string `json:"skills"`
	Education  string   `json:"education"`
	Experience string   `json:"experience"`
}

func SetupResumeRoutes(r *gin.Engine) {
	r.POST("/recruiter/parse-resume", uploadAndParseResumeHandler)
	// For applicants
	r.GET("/applicant/upload-resume", func(c *gin.Context) {
		c.HTML(http.StatusOK, "upload_resume.html", nil)
		r.POST("/recruiter/summarize-resume", summarizeResumeHandler)
	})

	// For recruiters
	r.GET("/recruiter/resume-parser", func(c *gin.Context) {
		c.HTML(http.StatusOK, "upload_resume.html", gin.H{
			"RecruiterMode": true,
		})
	})

}

// Recruiter uploads + parses resume using Gemini
func uploadAndParseResumeHandler(c *gin.Context) {
	file, err := c.FormFile("resume")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to upload file"})
		return
	}

	if !strings.HasSuffix(file.Filename, ".pdf") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only PDF files are allowed"})
		return
	}

	filePath := filepath.Join(uploadDir, file.Filename)
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	text, err := extractTextFromPDF(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to extract text"})
		return
	}

	resumeData, err := callGeminiAPI(text)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gemini parsing failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Resume parsed", "data": resumeData})
}

func extractTextFromPDF(filePath string) (string, error) {
	tempFile, err := ioutil.TempFile("", "extracted-*.txt")
	if err != nil {
		return "", err
	}
	defer os.Remove(tempFile.Name())

	if err := api.ExtractContentFile(filePath, tempFile.Name(), nil, nil); err != nil {
		return "", err
	}

	content, err := ioutil.ReadFile(tempFile.Name())
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// parsing
func callGeminiAPI(text string) (*Resume, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("missing Gemini API key")
	}

	// Gemini model: text-bison or gemini-pro
	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent?key=" + apiKey

	reqBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{
						"text": fmt.Sprintf(`Extract structured resume info as JSON:
						
%s

Output format:
{
  "name": "...",
  "email": "...",
  "phone": "...",
  "skills": [...],
  "education": "...",
  "experience": "..."
}`, text),
					},
				},
			},
		},
	}

	jsonData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Gemini API error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("Gemini error: %s", string(body))
	}

	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result.Candidates) == 0 {
		return nil, fmt.Errorf("No candidates returned")
	}

	rawJSON := result.Candidates[0].Content.Parts[0].Text

	// Parse raw JSON
	var resume Resume
	if err := json.Unmarshal([]byte(rawJSON), &resume); err != nil {
		return nil, fmt.Errorf("Invalid JSON from Gemini: %s", rawJSON)
	}

	return &resume, nil
}

func summarizeResumeHandler(c *gin.Context) {
	file, err := c.FormFile("resume")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Resume not uploaded"})
		return
	}

	path := filepath.Join("uploads", file.Filename)
	if err := c.SaveUploadedFile(file, path); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Upload failed"})
		return
	}

	text, err := extractTextFromPDF(path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Text extraction failed"})
		return
	}

	summary, err := getGeminiSummary(text)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}
func getGeminiSummary(resumeText string) (map[string]interface{}, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent?key=" + apiKey

	prompt := fmt.Sprintf(`
Please analyze the following resume and summarize the candidate's key information for a recruiter.
Return only the important information in JSON format using this structure:

{
  "name": "Full Name",
  "email": "Email Address",
  "phone": "Phone Number",
  "summary": "2–3 line overview of experience and skills",
  "skills": ["Skill 1", "Skill 2"],
  "education": "Highest relevant education",
  "experience": "Recent jobs, years, and companies"
}

Resume Content:
%s`, resumeText)

	body := map[string]interface{}{
		"contents": []map[string]interface{}{
			{"parts": []map[string]string{{"text": prompt}}},
		},
	}
	jsonData, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", url, bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				}
			}
		}
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result.Candidates) == 0 {
		return nil, fmt.Errorf("No response from Gemini")
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(result.Candidates[0].Content.Parts[0].Text), &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse JSON from Gemini output")
	}

	return parsed, nil
}

func ValidateResume(path string) error {
    text, err := extractTextFromPDF(path)
    if err != nil {
        return fmt.Errorf("PDF parsing failed: %v", err)
    }

    required := []string{"name", "skills", "education"}
    for _, field := range required {
        if !strings.Contains(strings.ToLower(text), field) {
            return fmt.Errorf("missing field: %s", field)
        }
    }

    return nil
}

func parseResumeSummary(summary string) *Resume {
	lines := strings.Split(summary, "\n")
	res := &Resume{}

	for _, line := range lines {
		if strings.Contains(line, "Name:") {
			res.Name = strings.TrimSpace(strings.SplitN(line, ":", 2)[1])
		} else if strings.Contains(line, "Email:") {
			res.Email = strings.TrimSpace(strings.SplitN(line, ":", 2)[1])
		} else if strings.Contains(line, "Phone:") {
			res.Phone = strings.TrimSpace(strings.SplitN(line, ":", 2)[1])
		} else if strings.Contains(line, "Skills:") {
			res.Skills = strings.Split(strings.TrimSpace(strings.SplitN(line, ":", 2)[1]), ",")
		} else if strings.Contains(line, "Education:") {
			res.Education = strings.TrimSpace(strings.SplitN(line, ":", 2)[1])
		} else if strings.Contains(line, "Experience:") {
			res.Experience = strings.TrimSpace(strings.SplitN(line, ":", 2)[1])
		}
	}
	return res
}
