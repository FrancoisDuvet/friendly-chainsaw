package notifs

import (
    "fmt"
    "log"
    "net/smtp"
    "os"
)

// Notification struct
type Notification struct {
    Recipient string
    Subject   string
    Body      string
}

// SendEmail sends an email notification
func SendEmail(notification Notification) error {
    from := os.Getenv("SMTP_EMAIL")
    password := os.Getenv("SMTP_PASSWORD")
    smtpHost := "smtp.gmail.com"
    smtpPort := "587"

    // Set up authentication
    auth := smtp.PlainAuth("", from, password, smtpHost)

    // Create the email message
    message := []byte(fmt.Sprintf("Subject: %s\n\n%s", notification.Subject, notification.Body))

    // Send the email
    err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{notification.Recipient}, message)
    if err != nil {
        log.Printf("Failed to send email to %s: %v", notification.Recipient, err)
        return err
    }

    log.Printf("Email sent to %s", notification.Recipient)
    return nil
}

// NotifyJobAlert sends a job alert to an applicant
func NotifyJobAlert(applicantEmail, companyName, jobTitle string) {
    subject := "New Job Alert"
    body := fmt.Sprintf("A new job titled '%s' has been posted by %s. Check it out on the platform!", jobTitle, companyName)

    notification := Notification{
        Recipient: applicantEmail,
        Subject:   subject,
        Body:      body,
    }

    // Send the email asynchronously
    go func() {
        if err := SendEmail(notification); err != nil {
            log.Printf("Failed to send job alert to %s: %v", applicantEmail, err)
        }
    }()
}

// NotifyApplicationStatus sends an application status update to an applicant
func NotifyApplicationStatus(applicantEmail, jobTitle, status string) {
    subject := "Application Status Update"
    body := fmt.Sprintf("Your application for the job '%s' has been updated to: %s.", jobTitle, status)

    notification := Notification{
        Recipient: applicantEmail,
        Subject:   subject,
        Body:      body,
    }

    // Send the email asynchronously
    go func() {
        if err := SendEmail(notification); err != nil {
            log.Printf("Failed to send application status update to %s: %v", applicantEmail, err)
        }
    }()
	}