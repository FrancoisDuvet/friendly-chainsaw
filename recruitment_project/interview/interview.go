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
)

// Mock database
var interviews = []Interview{}

type Interview struct {
    ID           string    `json:"id"`
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
    interviews = append(interviews, interview)

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
    for i, interview := range interviews {
        if interview.ID == response.InterviewID {
            if response.Action == "accept" {
                interviews[i].Status = "Accepted"
                go sendEmail(interview.Recruiter, "Interview Accepted", fmt.Sprintf("The applicant has accepted the interview scheduled on %s.", interview.ScheduledAt))
                c.JSON(http.StatusOK, gin.H{"message": "Interview accepted"})
                return
            } else if response.Action == "propose" {
                interviews[i].Status = "Rescheduled"
                interviews[i].ProposedTime = response.ProposedTime
                go sendEmail(interview.Recruiter, "Interview Reschedule Proposed", fmt.Sprintf("The applicant has proposed a new time: %s.", response.ProposedTime))
                c.JSON(http.StatusOK, gin.H{"message": "Proposed a new time for the interview"})
                return
            }
        }
    }

    c.JSON(http.StatusNotFound, gin.H{"error": "Interview not found"})
}

// View all interviews (Recruiter or Applicant)
func viewInterviewsHandler(c *gin.Context) {
    user := c.Query("user")
    var userInterviews []Interview

    for _, interview := range interviews {
        if interview.Recruiter == user || interview.Applicant == user {
            userInterviews = append(userInterviews, interview)
        }
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

func interview() {
    r := gin.Default()

    // Interview scheduling routes
    r.POST("/recruiter/schedule", scheduleInterviewHandler) // Recruiter schedules an interview
    r.POST("/applicant/respond", respondToInterviewHandler) // Applicant responds to an interview
    r.GET("/interviews", viewInterviewsHandler)             // View all interviews for a user

    fmt.Println("Server started at :8080")
    r.Run(":8080")
}