package interview

import (
    "fmt"
    "log"
    "net/http"
    "net/smtp"
    "os"
    "time"

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

// Interview struct for database
type Interview struct {
    ID           string    `gorm:"primaryKey" json:"id"`
    JobID        string    `json:"job_id"`
    Applicant    string    `json:"applicant"`
    Recruiter    string    `json:"recruiter"`
    ScheduledAt  time.Time `json:"scheduled_at"`
    Status       string    `json:"status"` // e.g., "Pending", "Accepted", "Rescheduled"
    ProposedTime time.Time `json:"proposed_time,omitempty"`
}

// Schedule an interview (Recruiter)
func scheduleInterviewHandler(c *gin.Context) {
    var interview Interview
    if err := c.ShouldBindJSON(&interview); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid interview data"})
        return
    }

    // Generate a unique interview ID
    interview.ID = uuid.New().String()
    interview.Status = "Pending"

    // Save to database
    if err := db.Create(&interview).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to schedule interview"})
        return
    }

    // Send email notification to the applicant
    go sendEmail(interview.Applicant, "Interview Scheduled", fmt.Sprintf("An interview has been scheduled for you on %s.", interview.ScheduledAt))

    c.JSON(http.StatusCreated, gin.H{"message": "Interview scheduled successfully", "interview": interview})
}

// Applicant accepts or proposes a new time for the interview
func respondToInterviewHandler(c *gin.Context) {
    var response struct {
        InterviewID  string    `json:"interview_id"`
        Action       string    `json:"action"`       // "accept" or "propose"
        ProposedTime time.Time `json:"proposed_time"` // Only required if action is "propose"
    }
    if err := c.ShouldBindJSON(&response); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid response data"})
        return
    }

    // Find the interview
    var interview Interview
    if err := db.First(&interview, "id = ?", response.InterviewID).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            c.JSON(http.StatusNotFound, gin.H{"error": "Interview not found"})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        }
        return
    }

    // Update interview status
    if response.Action == "accept" {
        interview.Status = "Accepted"
        if err := db.Save(&interview).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update interview status"})
            return
        }
        go sendEmail(interview.Recruiter, "Interview Accepted", fmt.Sprintf("The applicant has accepted the interview scheduled on %s.", interview.ScheduledAt))
        c.JSON(http.StatusOK, gin.H{"message": "Interview accepted"})
    } else if response.Action == "propose" {
        interview.Status = "Rescheduled"
        interview.ProposedTime = response.ProposedTime
        if err := db.Save(&interview).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to propose new time"})
            return
        }
        go sendEmail(interview.Recruiter, "Interview Reschedule Proposed", fmt.Sprintf("The applicant has proposed a new time: %s.", response.ProposedTime))
        c.JSON(http.StatusOK, gin.H{"message": "Proposed a new time for the interview"})
    }
}

// View all interviews (Recruiter or Applicant)
func viewInterviewsHandler(c *gin.Context) {
    user := c.Query("user")
    var userInterviews []Interview

    // Fetch interviews for the user
    if err := db.Where("recruiter = ? OR applicant = ?", user, user).Find(&userInterviews).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch interviews"})
        return
    }

    c.JSON(http.StatusOK, userInterviews)
}

// Send email notification
func sendEmail(to, subject, body string) {
    from := os.Getenv("SMTP_EMAIL")
    password := os.Getenv("SMTP_PASSWORD")
    smtpHost := "smtp.gmail.com"
    smtpPort := "587"

    auth := smtp.PlainAuth("", from, password, smtpHost)
    message := []byte(fmt.Sprintf("Subject: %s\n\n%s", subject, body))

    err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, message)
    if err != nil {
        log.Printf("Failed to send email to %s: %v", to, err)
    } else {
        log.Printf("Email sent to %s", to)
    }
}

func init() {
    db = ConnectDB()
    db.AutoMigrate(&Interview{}) // Migrate the Interview struct to the database
}

func main() {
    r := gin.Default()

    // Interview scheduling routes
    r.POST("/recruiter/schedule", scheduleInterviewHandler) // Recruiter schedules an interview
    r.POST("/applicant/respond", respondToInterviewHandler) // Applicant responds to an interview
    r.GET("/interviews", viewInterviewsHandler)             // View all interviews for a user

    fmt.Println("Server started at :8080")
    r.Run(":8080")
}