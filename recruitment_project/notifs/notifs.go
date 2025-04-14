package notifs

import (
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"

	"github.com/FrancoisDuvet/friendly-chainsaw/recruitment_project/middleware"
	"github.com/FrancoisDuvet/friendly-chainsaw/recruitment_project/users"
	"github.com/gin-gonic/gin"
)


type Notification struct {
	Recipient string
	Subject   string
	Body      string
}

func SendEmail(notification Notification) error {
	from := os.Getenv("SMTP_EMAIL")
	password := os.Getenv("SMTP_PASSWORD")
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	auth := smtp.PlainAuth("", from, password, smtpHost)
	message := []byte(fmt.Sprintf("Subject: %s\n\n%s", notification.Subject, notification.Body))

	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{notification.Recipient}, message)
	if err != nil {
		log.Printf("Failed to send email to %s: %v", notification.Recipient, err)
		return err
	}

	log.Printf("Email sent to %s", notification.Recipient)
	return nil
}

//sends a job alert to an applicant
func NotifyJobAlert(applicantID, companyName, jobTitle string) {
	applicant, exists := users.Applicants[applicantID]
	if !exists {
		log.Printf("Applicant with ID %s not found", applicantID)
		return
	}

	subject := "New Job Alert"
	body := fmt.Sprintf("A new job titled '%s' has been posted by %s. Check it out on the platform!", jobTitle, companyName)

	notification := Notification{
		Recipient: applicant.Email,
		Subject:   subject,
		Body:      body,
	}

	go func() {
		if err := SendEmail(notification); err != nil {
			log.Printf("Failed to send job alert to %s: %v", applicant.Email, err)
		}
	}()
	for _, applicant := range users.Applicants {
		for _, followed := range applicant.Following {
			for _, job := range users.Jobs {
				if followed == job.CompanyID {
					NotifyJobAlert(applicant.ID, users.Companies[job.CompanyID].Title, job.Title)
				}
			}
		}
	}
	
}

//sends an application status update to an applicant
func NotifyApplicationStatus(applicantID, jobTitle, status string) {
	applicant, exists := users.Applicants[applicantID]
	if !exists {
		log.Printf("Applicant with ID %s not found", applicantID)
		return
	}

	subject := "Application Status Update"
	body := fmt.Sprintf("Your application for the job '%s' has been updated to: %s.", jobTitle, status)

	notification := Notification{
		Recipient: applicant.Email,
		Subject:   subject,
		Body:      body,
	}

	go func() {
		if err := SendEmail(notification); err != nil {
			log.Printf("Failed to send application status update to %s: %v", applicant.Email, err)
		}
	}()
}

//Super Admin Dashboard & Recruiter Approval Route
func SetupNotificationRoutes(r *gin.Engine) {
	admin := r.Group("/admin", middleware.RequireSession(), middleware.RequireRole("superadmin"))

	// Admin dashboard
	admin.GET("/dashboard", func(c *gin.Context) {
		var pending []users.Recruiter
		for _, rec := range users.Recruiters {
			if !rec.IsApproved {
				pending = append(pending, rec)
			}
		}

		c.HTML(http.StatusOK, "superadmin_dashboard.html", gin.H{
			"Recruiters": pending,
		})
	})

	// Approve recruiter by email
	admin.GET("/approve", func(c *gin.Context) {
		email := c.Query("email")
		for id, rec := range users.Recruiters {
			if rec.Email == email {
				rec.IsApproved = true
				users.Recruiters[id] = rec

				//Notify the recruiter via email
				go SendEmail(Notification{
					Recipient: rec.Email,
					Subject:   "Account Approved",
					Body:      "Your recruiter account has been approved by the Super Admin.",
				})

				c.JSON(http.StatusOK, gin.H{"message": "Recruiter approved successfully."})
				return
			}
		}
		c.JSON(http.StatusNotFound, gin.H{"message": "Recruiter not found."})
	})
}
