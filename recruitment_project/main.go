package main

import (
	"html/template"
	"log"
	"strings"

	"github.com/FrancoisDuvet/friendly-chainsaw/recruitment_project/auth"
	"github.com/FrancoisDuvet/friendly-chainsaw/recruitment_project/cvupload"
	"github.com/FrancoisDuvet/friendly-chainsaw/recruitment_project/interview"
	"github.com/FrancoisDuvet/friendly-chainsaw/recruitment_project/jpost"
	"github.com/FrancoisDuvet/friendly-chainsaw/recruitment_project/middleware"
	"github.com/FrancoisDuvet/friendly-chainsaw/recruitment_project/notifs"
	"github.com/FrancoisDuvet/friendly-chainsaw/recruitment_project/resume"
	"github.com/FrancoisDuvet/friendly-chainsaw/recruitment_project/users"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var funcMap = template.FuncMap{
	"join": strings.Join,
}

func init() {
	// Initialize secure cookie sessions
	middleware.InitSessionStore([]byte("your-secret-key"))

	// Load .env variables
	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("Error loading .env file")
	}
}

func setupRouter() *gin.Engine {
	r := gin.Default()

	// Enable template functions (e.g., join)
	r.SetFuncMap(funcMap)
	r.LoadHTMLGlob("templates/*.html")

	// Route setups by module
	auth.SetupAuthRoutes(r)
	jpost.SetupJobRoutes(r)
	interview.SetupInterviewRoutes(r)
	resume.SetupResumeRoutes(r)
	cvupload.SetupCVUploadRoutes(r)
	notifs.SetupNotificationRoutes(r)
	users.SetupUserRoutes(r)

	return r
}

func main() {
	r := setupRouter()

	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to run server:", err)
	}
}
