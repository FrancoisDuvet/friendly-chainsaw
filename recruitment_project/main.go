package main

import (
    "log"

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

func init() {
    // Load environment variables from .env file
    err := godotenv.Load(".env")
    if err != nil {
        log.Fatalf("Error loading .env file")
    }

    // Initialize database connections for all modules
    auth.InitDB()
    jpost.InitDB()
    interview.InitDB()
    resume.InitDB()
    users.InitDB()
}

func setupRouter() *gin.Engine {
    r := gin.Default()

    // Authentication routes
    auth.SetupAuthRoutes(r)

    // Job posting and application routes
    jpost.SetupJobRoutes(r)

    // Interview scheduling routes
    interview.SetupInterviewRoutes(r)

    // Resume upload and parsing routes
    resume.SetupResumeRoutes(r)

    // CV upload routes
    cvupload.SetupCVUploadRoutes(r)

    // Notification routes (if needed)
    notifs.SetupNotificationRoutes(r)

    // User management routes
    users.SetupUserRoutes(r)

    return r
}

func main() {
    // Initialize the router
    r := setupRouter()

    // Start the server
    if err := r.Run(":8080"); err != nil {
        log.Fatal("Failed to run server:", err)
    }
}