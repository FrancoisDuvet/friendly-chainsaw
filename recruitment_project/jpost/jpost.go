package jpost

import (
    "fmt"
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

// Database connection
var db *gorm.DB

func ConnectDB() *gorm.DB {
    dsn := "host=localhost user=postgres password=mysecretpassword dbname=postgres port=5431 sslmode=disable"
    database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        panic("failed to connect to db")
    }
    return database
}

type Job struct {
    ID          string   `gorm:"primaryKey" json:"id"`
    Title       string   `json:"title"`
    Description string   `json:"description"`
    Skills      []string `gorm:"type:text[]" json:"skills"` // PostgreSQL array type
    CompanyID   string   `json:"company_id"`
}

type Application struct {
    ID        string `gorm:"primaryKey" json:"id"`
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
    job.ID = uuid.New().String()

    // Save to database
    if err := db.Create(&job).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to post job"})
        return
    }

    c.JSON(http.StatusCreated, gin.H{"message": "Job posted successfully", "job": job})
}

// Applicant views all job postings
func viewJobsHandler(c *gin.Context) {
    var jobs []Job
    if err := db.Find(&jobs).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch jobs"})
        return
    }

    c.JSON(http.StatusOK, jobs)
}

// Applicant applies for a job
func applyJobHandler(c *gin.Context) {
    var application Application
    if err := c.ShouldBindJSON(&application); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid application data"})
        return
    }

    // Generate application ID
    application.ID = uuid.New().String()
    application.Status = "Applied"

    // Save to database
    if err := db.Create(&application).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit application"})
        return
    }

    c.JSON(http.StatusCreated, gin.H{"message": "Application submitted successfully", "application": application})
}

// Applicant views their applications
func viewApplicationsHandler(c *gin.Context) {
    applicant := c.Query("applicant")
    var userApplications []Application

    // Fetch applications for the applicant
    if err := db.Where("applicant = ?", applicant).Find(&userApplications).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch applications"})
        return
    }

    c.JSON(http.StatusOK, userApplications)
}

func init() {
    db = ConnectDB()
    db.AutoMigrate(&Job{}, &Application{}) // Migrate the Job and Application structs to the database
}

func SetupJobRoutes(r *gin.Engine) {

    // Job posting routes
    r.POST("/recruiter/jobs", postJobHandler) // Recruiter posts a job
    r.GET("/jobs", viewJobsHandler)          // Applicant views all jobs

    // Job application routes
    r.POST("/applicant/apply", applyJobHandler)       // Applicant applies for a job
    r.GET("/applicant/applications", viewApplicationsHandler) // Applicant views their applications

    fmt.Println("Server started at :8080")
    r.Run(":8080")
}