package jpost

import (
    "fmt"
    "net/http"

    "github.com/gin-gonic/gin"
)

// Mock database
var jobs = []Job{}
var applications = []Application{}

type Job struct {
    ID          string   `json:"id"`
    Title       string   `json:"title"`
    Description string   `json:"description"`
    Skills      []string `json:"skills"`
    CompanyID   string   `json:"company_id"`
}

type Application struct {
    ID        string `json:"id"`
    JobID     string `json:"job_id"`
    Applicant string `json:"applicant"`
    Status    string `json:"status"` // e.g., "Applied", "Interview Scheduled", "Offered"
}

// Recruiter posts a new job
func postJobHandler(c *gin.Context) {
    var job Job
    if err := c.ShouldBindJSON(&job); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job data"})
        return
    }

    // Generate a unique job ID
    job.ID = fmt.Sprintf("job_%d", len(jobs)+1)
    jobs = append(jobs, job)

    c.JSON(http.StatusCreated, gin.H{"message": "Job posted successfully", "job": job})
}

// Applicant views all job postings
func viewJobsHandler(c *gin.Context) {
    c.JSON(http.StatusOK, jobs)
}

// Applicant applies for a job
func applyJobHandler(c *gin.Context) {
    var application Application
    if err := c.ShouldBindJSON(&application); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid application data"})
        return
    }

    // Generate a unique application ID
    application.ID = fmt.Sprintf("application_%d", len(applications)+1)
    application.Status = "Applied"
    applications = append(applications, application)

    c.JSON(http.StatusCreated, gin.H{"message": "Application submitted successfully", "application": application})
}

// Applicant views their applications
func viewApplicationsHandler(c *gin.Context) {
    applicant := c.Query("applicant")
    var userApplications []Application

    for _, app := range applications {
        if app.Applicant == applicant {
            userApplications = append(userApplications, app)
        }
    }

    c.JSON(http.StatusOK, userApplications)
}

func jpostnapply() {
    r := gin.Default()

    // Job posting routes
    r.POST("/recruiter/jobs", postJobHandler) // Recruiter posts a job
    r.GET("/jobs", viewJobsHandler)          // Applicant views all jobs

    // Job application routes
    r.POST("/applicant/apply", applyJobHandler)       // Applicant applies for a job
    r.GET("/applicant/applications", viewApplicationsHandler) // Applicant views their applications

    fmt.Println("Server started at :8080")
    r.Run(":8080")
}