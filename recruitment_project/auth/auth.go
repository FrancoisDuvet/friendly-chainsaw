package auth

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"

    "github.com/gin-gonic/gin"
    "github.com/gorilla/sessions"
    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
)

// OAuth2 configuration
var googleOAuthConfig = &oauth2.Config{
    ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
    ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
    RedirectURL:  "http://localhost:8080/auth/callback",
    Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
    Endpoint:     google.Endpoint,
}

// Session store
var store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))

// User roles
const (
    RoleSuperAdmin = "super_admin"
    RoleRecruiter  = "recruiter"
    RoleApplicant  = "applicant"
)

// Mock database
var users = map[string]User{}

type User struct {
    ID    string
    Name  string
    Email string
    Role  string
}

// Login handler (redirects to Google OAuth)
func loginHandler(c *gin.Context) {
    url := googleOAuthConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
    c.Redirect(http.StatusTemporaryRedirect, url)
}

// OAuth callback handler
func callbackHandler(c *gin.Context) {
    code := c.Query("code")
    if code == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Code not found"})
        return
    }

    token, err := googleOAuthConfig.Exchange(context.Background(), code)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange token"})
        return
    }

    client := googleOAuthConfig.Client(context.Background(), token)
    resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info"})
        return
    }
    defer resp.Body.Close()

    var userInfo struct {
        ID    string `json:"id"`
        Name  string `json:"name"`
        Email string `json:"email"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode user info"})
        return
    }

    // Check if user exists in the database
    user, exists := users[userInfo.Email]
    if !exists {
        // Assign default role (applicant) for new users
        user = User{
            ID:    userInfo.ID,
            Name:  userInfo.Name,
            Email: userInfo.Email,
            Role:  RoleApplicant,
        }
        users[userInfo.Email] = user
    }

    // Save user info in session
    session, _ := store.Get(c.Request, "session")
    session.Values["user_email"] = user.Email
    session.Values["user_role"] = user.Role
    session.Save(c.Request, c.Writer)

    c.Redirect(http.StatusSeeOther, "/dashboard")
}

// Middleware for Role-Based Access Control (RBAC)
func roleMiddleware(requiredRole string) gin.HandlerFunc {
    return func(c *gin.Context) {
        session, _ := store.Get(c.Request, "session")
        role, ok := session.Values["user_role"].(string)
        if !ok || role != requiredRole {
            c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
            c.Abort()
            return
        }
        c.Next()
    }
}

// Dashboard handler
func dashboardHandler(c *gin.Context) {
    session, _ := store.Get(c.Request, "session")
    email, _ := session.Values["user_email"].(string)
    role, _ := session.Values["user_role"].(string)

    c.JSON(http.StatusOK, gin.H{
        "message": "Welcome to your dashboard!",
        "email":   email,
        "role":    role,
    })
}

// Super Admin approval handler
func approveRecruiterHandler(c *gin.Context) {
    email := c.Query("email")
    user, exists := users[email]
    if !exists || user.Role != RoleRecruiter {
        c.JSON(http.StatusNotFound, gin.H{"error": "Recruiter not found"})
        return
    }

    // Approve recruiter
    user.Role = RoleRecruiter
    users[email] = user
    c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Recruiter %s approved successfully.", email)})
}

func initOAuth() {
    r := gin.Default()

    // Authentication routes
    r.GET("/auth/login", loginHandler)
    r.GET("/auth/callback", callbackHandler)

    // Dashboard route
    r.GET("/dashboard", dashboardHandler)

    // Super Admin routes
    admin := r.Group("/admin", roleMiddleware(RoleSuperAdmin))
    admin.GET("/approve", approveRecruiterHandler)

    fmt.Println("Server started at :8080")
    log.Fatal(r.Run(":8080"))
}