package users

import (
    "fmt"
    "net/http"
)

// Mock database
var recruiters = map[string]Recruiter{}
var applicants = map[string]Applicant{}
var companies = map[string]Company{}

type Recruiter struct {
    ID          string
    Name        string
    Email       string
    CompanyID   string
    IsApproved  bool
    JobPostings []Job
}

type Applicant struct {
    ID       string
    Name     string
    Email    string
    Skills   []string
    Resume   string
    AppliedJobs []string
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

// Recruiter Dashboard Handler
func recruiterDashboard(w http.ResponseWriter, r *http.Request) {
    recruiterID := r.URL.Query().Get("id")
    recruiter, exists := recruiters[recruiterID]
    if !exists || !recruiter.IsApproved {
        http.Error(w, "Recruiter not found or not approved", http.StatusForbidden)
        return
    }

    fmt.Fprintf(w, "Welcome to the Recruiter Dashboard, %s\n", recruiter.Name)
    fmt.Fprintf(w, "Your Job Postings:\n")
    for _, job := range recruiter.JobPostings {
        fmt.Fprintf(w, "- %s: %s\n", job.Title, job.Description)
    }
}

// Applicant Dashboard Handler
func applicantDashboard(w http.ResponseWriter, r *http.Request) {
    applicantID := r.URL.Query().Get("id")
    applicant, exists := applicants[applicantID]
    if !exists {
        http.Error(w, "Applicant not found", http.StatusNotFound)
        return
    }

    fmt.Fprintf(w, "Welcome to the Applicant Dashboard, %s\n", applicant.Name)
    fmt.Fprintf(w, "Your Skills: %v\n", applicant.Skills)
    fmt.Fprintf(w, "Your Applied Jobs: %v\n", applicant.AppliedJobs)
}

// Create Recruiter Account Handler
func createRecruiter(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    // Parse form data
    r.ParseForm()
    name := r.FormValue("name")
    email := r.FormValue("email")
    companyTitle := r.FormValue("company_title")
    companyDescription := r.FormValue("company_description")
    companyLogo := r.FormValue("company_logo")

    // Create company
    companyID := fmt.Sprintf("company_%d", len(companies)+1)
    company := Company{
        ID:          companyID,
        Title:       companyTitle,
        Description: companyDescription,
        Logo:        companyLogo,
        IsApproved:  false, // Super Admin approval required
    }
    companies[companyID] = company

    // Create recruiter
    recruiterID := fmt.Sprintf("recruiter_%d", len(recruiters)+1)
    recruiter := Recruiter{
        ID:         recruiterID,
        Name:       name,
        Email:      email,
        CompanyID:  companyID,
        IsApproved: false, // Super Admin approval required
    }
    recruiters[recruiterID] = recruiter

    fmt.Fprintf(w, "Recruiter account created. Awaiting Super Admin approval.\n")
}

// Create Applicant Account Handler
func createApplicant(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    // Parse form data
    r.ParseForm()
    name := r.FormValue("name")
    email := r.FormValue("email")
    skills := r.Form["skills"]

    // Create applicant
    applicantID := fmt.Sprintf("applicant_%d", len(applicants)+1)
    applicant := Applicant{
        ID:     applicantID,
        Name:   name,
        Email:  email,
        Skills: skills,
    }
    applicants[applicantID] = applicant

    fmt.Fprintf(w, "Applicant account created successfully.\n")
}

func main() {
    http.HandleFunc("/recruiter/dashboard", recruiterDashboard)
    http.HandleFunc("/applicant/dashboard", applicantDashboard)
    http.HandleFunc("/recruiter/create", createRecruiter)
    http.HandleFunc("/applicant/create", createApplicant)

    fmt.Println("Server started at :8080")
    http.ListenAndServe(":8080", nil)
}