package main

import (
    "log"
    "html/template"
	"strings"
    "github.com/FrancoisDuvet/friendly-chainsaw/recruitment_project/middleware"
    "github.com/FrancoisDuvet/friendly-chainsaw/recruitment_project/auth"
    "github.com/FrancoisDuvet/friendly-chainsaw/recruitment_project/jpost"
    "github.com/FrancoisDuvet/friendly-chainsaw/recruitment_project/interview"
    "github.com/FrancoisDuvet/friendly-chainsaw/recruitment_project/resume"
    "github.com/FrancoisDuvet/friendly-chainsaw/recruitment_project/cvupload"
    "github.com/FrancoisDuvet/friendly-chainsaw/recruitment_project/notifs"
    "github.com/FrancoisDuvet/friendly-chainsaw/recruitment_project/users"
    "github.com/gin-gonic/gin"
    "github.com/joho/godotenv"
)

var funcMap = template.FuncMap{
    "join": strings.Join,
}

func init() {
    middleware.InitSessionStore([]byte("your-secret-key"))
    
    err := godotenv.Load(".env")
    if err != nil {
        log.Fatalf("Error loading .env file")
    }

}

func setupRouter() *gin.Engine {
    r := gin.Default()
    r.SetFuncMap(funcMap) // 🔥 Enables {{join ...}} in templates
    r.LoadHTMLGlob("templates/*.html")

    auth.SetupAuthRoutes()

    jpost.SetupJobRoutes(r)

    interview.SetupInterviewRoutes(r)

    resume.SetupResumeRoutes(r)

    cvupload.SetupCVUploadRoutes(r)

    notifs.SetupNotificationRoutes(r)

    users.SetupUserRoutes(r)

    return r
}

func main() {
    // Initialize the router
    r := setupRouter()

    // Startserver
    if err := r.Run(":8080"); err != nil {
        log.Fatal("Failed to run server:", err)
    }
}