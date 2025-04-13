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
	"github.com/gorilla/sessions"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

// Connects to interview DB
func ConnectDB() *gorm.DB {
	dsn := "host=localhost user=postgres password=mysecretpassword dbname=postgres port=5431 sslmode=disable"
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect to db")
	}
	return database
}

type Interview struct {
	ID           string    `gorm:"primaryKey" json:"id"`
	JobID        string    `json:"job_id"`
	Applicant    string    `json:"applicant"`
	Recruiter    string    `json:"recruiter"`
	ScheduledAt  time.Time `json:"scheduled_at"`
	Status       string    `json:"status"` // Pending, Accepted, Rescheduled
	ProposedTime time.Time `json:"proposed_time,omitempty"`
}

// Recruiter schedules an interview
func scheduleInterviewHandler(c *gin.Context) {
	var interview Interview
	if err := c.ShouldBindJSON(&interview); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid interview data"})
		return
	}

	interview.ID = uuid.New().String()
	interview.Status = "Pending"

	if err := db.Create(&interview).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to schedule interview"})
		return
	}

	go sendEmail(interview.Applicant, "Interview Scheduled", fmt.Sprintf("An interview has been scheduled for you on %s.", interview.ScheduledAt))

	c.JSON(http.StatusCreated, gin.H{"message": "Interview scheduled successfully", "interview": interview})
}

// Applicant responds to the interview
func respondToInterviewHandler(c *gin.Context) {
	var response struct {
		InterviewID  string    `json:"interview_id"`
		Action       string    `json:"action"`       // "accept" or "propose"
		ProposedTime time.Time `json:"proposed_time"` // Only if "propose"
	}
	if err := c.ShouldBindJSON(&response); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid response data"})
		return
	}

	var interview Interview
	if err := db.First(&interview, "id = ?", response.InterviewID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Interview not found"})
		return
	}

	if response.Action == "accept" {
		interview.Status = "Accepted"
	} else if response.Action == "propose" {
		interview.Status = "Rescheduled"
		interview.ProposedTime = response.ProposedTime
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid action"})
		return
	}

	if err := db.Save(&interview).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update interview"})
		return
	}

	notifyMessage := "The applicant has accepted the interview."
	if response.Action == "propose" {
		notifyMessage = fmt.Sprintf("The applicant proposed a new time: %s.", response.ProposedTime)
	}

	go sendEmail(interview.Recruiter, fmt.Sprintf("Interview %s", interview.Status), notifyMessage)

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Interview %s", interview.Status)})
}

//(filtered by user email)
func viewInterviewsHandler(c *gin.Context) {
	user := c.Query("user")
	var userInterviews []Interview

	if err := db.Where("recruiter = ? OR applicant = ?", user, user).Find(&userInterviews).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch interviews"})
		return
	}

	c.JSON(http.StatusOK, userInterviews)
}

func sendEmail(to, subject, body string) {
	from := os.Getenv("SMTP_EMAIL")
	password := os.Getenv("SMTP_PASSWORD")
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	auth := smtp.PlainAuth("", from, password, smtpHost)
	message := []byte(fmt.Sprintf("Subject: %s\n\n%s", subject, body))

	if err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, message); err != nil {
		log.Printf("Failed to send email to %s: %v", to, err)
	} else {
		log.Printf("Email sent to %s", to)
	}
}
var store = sessions.NewCookieStore([]byte("secret-key"))

func SetupInterviewRoutes(r *gin.Engine) {
	r.GET("/recruiter/interview/new", func(c *gin.Context) {
        session, _ := store.Get(c.Request, "session")
        email := session.Values["user_email"].(string)
        c.HTML(http.StatusOK, "schedule_interview.html", gin.H{
            "RecruiterEmail": email,
        })
    })
    
    r.GET("/applicant/interview/respond", func(c *gin.Context) {
        c.HTML(http.StatusOK, "respond_interview.html", nil)
    })
    
    r.POST("/recruiter/schedule", scheduleInterviewHandler)
	r.POST("/applicant/respond", respondToInterviewHandler)
	r.GET("/interviews", viewInterviewsHandler)
}

//package init
func init() {
	db = ConnectDB()
	db.AutoMigrate(&Interview{})
}
