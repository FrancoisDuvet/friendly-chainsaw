package users

import (
	"fmt"
	"html/template"
	"strings"
	"net/http"
	"github.com/gorilla/sessions"
	"github.com/FrancoisDuvet/friendly-chainsaw/recruitment_project/middleware"
	"github.com/gin-gonic/gin"
	"github.com/FrancoisDuvet/friendly-chainsaw/recruitment_project/notifs"
)

var store = sessions.NewCookieStore([]byte("secret-key")) // Replace "secret-key" with a secure key
var Recruiters = map[string]Recruiter{}
var Applicants = map[string]Applicant{}
var Companies = map[string]Company{}
var Jobs = []Job{
	{
		ID:          "job1",
		Title:       "Backend Developer",
		Description: "Build backend systems with Go.",
		Skills:      []string{"Go", "PostgreSQL", "Docker"},
		CompanyID:   "company_1",
	},
}

//Data Models
type Recruiter struct {
	ID          string
	Name        string
	Email       string
	CompanyID   string
	IsApproved  bool
	JobPostings []Job
}

type Applicant struct {
	ID          string
	Name        string
	Email       string
	Skills      []string
	Resume      string
	AppliedJobs []string
	Following   []string
	StatusMap   map[string]string // jobID → status (e.g., “Under Review”)
}


type Company struct {
	ID          string
	Title       string
	Description string
	Logo        string
	IsApproved  bool
}

type Job struct {
	ID          string
	Title       string
	Description string
	Skills      []string
	CompanyID   string
}

func followCompanyHandler(c *gin.Context) {
	session, _ := store.Get(c.Request, "session")
	email := session.Values["user_email"].(string)
	companyID := c.Query("company_id")

	applicant, exists := Applicants[email]
	if !exists {
		c.String(http.StatusNotFound, "Applicant not found")
		return
	}

	// Avoid duplicate follows
	for _, cid := range applicant.Following {
		if cid == companyID {
			c.String(http.StatusOK, "Already following")
			return
		}
	}

	applicant.Following = append(applicant.Following, companyID)
	Applicants[email] = applicant

	c.String(http.StatusOK, "Company followed!")
}

//Recruiter Dashboard HTML rendering
func recruiterDashboard(c *gin.Context) {
	email := c.MustGet("user_email").(string)

	var recruiter Recruiter
	for _, r := range Recruiters {
		if r.Email == email {
			recruiter = r
			break
		}
	}

	if !recruiter.IsApproved {
		c.String(http.StatusForbidden, "Recruiter not approved")
		return
	}

	c.HTML(http.StatusOK, "recruiter_dashboard.html", gin.H{
		"Name":        recruiter.Name,
		"JobPostings": recruiter.JobPostings,
	})
}

// Applicant Dashboard HTML rendering
func applicantDashboard(c *gin.Context) {
	email := c.MustGet("user_email").(string)

	var applicant Applicant
	for _, a := range Applicants {
		if a.Email == email {
			applicant = a
			break
		}
	}

	c.HTML(http.StatusOK, "applicant_dashboard.html", gin.H{
		"Name": applicant.Name,
		"Jobs": Jobs,
	})
}

func createRecruiter(c *gin.Context) {
	c.Request.ParseForm()
	name := c.PostForm("name")
	email := c.PostForm("email")
	companyTitle := c.PostForm("company_title")
	companyDescription := c.PostForm("company_description")
	companyLogo := c.PostForm("company_logo")

	companyID := fmt.Sprintf("company_%d", len(Companies)+1)
	Companies[companyID] = Company{
		ID:          companyID,
		Title:       companyTitle,
		Description: companyDescription,
		Logo:        companyLogo,
		IsApproved:  false,
	}

	recruiterID := fmt.Sprintf("recruiter_%d", len(Recruiters)+1)
	Recruiters[recruiterID] = Recruiter{
		ID:         recruiterID,
		Name:       name,
		Email:      email,
		CompanyID:  companyID,
		IsApproved: false,
	}

	c.String(http.StatusCreated, "Recruiter account created. Awaiting Super Admin approval.")
}


func createApplicant(c *gin.Context) {
	c.Request.ParseForm()
	name := c.PostForm("name")
	email := c.PostForm("email")
	skills := c.PostFormArray("skills")

	applicantID := fmt.Sprintf("applicant_%d", len(Applicants)+1)
	Applicants[applicantID] = Applicant{
		ID:     applicantID,
		Name:   name,
		Email:  email,
		Skills: skills,
	}

	c.String(http.StatusCreated, "Applicant account created successfully.")
}

func updateApplicationStatusHandler(c *gin.Context) {
	jobID := c.PostForm("job_id")
	applicantID := c.PostForm("applicant_id")
	newStatus := c.PostForm("status")

	app, exists := Applicants[applicantID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Applicant not found"})
		return
	}

	if app.StatusMap == nil {
		app.StatusMap = make(map[string]string)
	}
	app.StatusMap[jobID] = newStatus
	Applicants[applicantID] = app

	go notifs.NotifyApplicationStatus(applicantID, jobID, newStatus)
	c.JSON(http.StatusOK, gin.H{"message": "Status updated!"})
}


//for template (function map)
var funcMap = template.FuncMap{
	"join": strings.Join,
}

func SetupUserRoutes(r *gin.Engine) {
    userRoutes := r.Group("/", middleware.RequireSession())
    userRoutes.GET("/recruiter/dashboard", recruiterDashboard)
    userRoutes.GET("/applicant/dashboard", applicantDashboard)
    r.POST("/recruiter/create", createRecruiter)
    r.POST("/applicant/create", createApplicant)
	r.POST("/recruiter/update-status", updateApplicationStatusHandler)
	r.POST("/applicant/follow", followCompanyHandler)
}