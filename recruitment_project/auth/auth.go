package auth

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"

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

var store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))

const (
    RoleSuperAdmin = "super_admin"
    RoleRecruiter  = "recruiter"
    RoleApplicant  = "applicant"
)

type User struct {
    ID    string `gorm:"primaryKey"`
    Name  string
    Email string `gorm:"unique"`
    Role  string
}

// Database connection
var db *gorm.DB

func ConnectDB() *gorm.DB {
    dsn := "host=localhost user=postgres password=mysecretpassword dbname=postgres port=5431 sslmode=disable"
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        panic("failed to connect to db")
    }
    return database
}

//redirection to Google OAuth)
func loginHandler(c *gin.Context) {
    url := googleOAuthConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
    c.Redirect(http.StatusTemporaryRedirect, url)
}

// for OAuth
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
    var user User
    result := db.First(&user, "email = ?", userInfo.Email)
    if result.Error != nil {
        if result.Error == gorm.ErrRecordNotFound {
            // Assign default role (applicant) for new users
            user = User{
                ID:    userInfo.ID,
                Name:  userInfo.Name,
                Email: userInfo.Email,
                Role:  RoleApplicant,
            }
            db.Create(&user)
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
            return
        }
    }

    // Save user info in session
    session, _ := store.Get(c.Request, "session")
    session.Values["user_email"] = user.Email
    session.Values["user_role"] = user.Role
    session.Save(c.Request, c.Writer)

    switch user.Role {
    case RoleSuperAdmin:
        c.Redirect(http.StatusSeeOther, "/admin/dashboard")
    case RoleRecruiter:
        c.Redirect(http.StatusSeeOther, "/recruiter/dashboard?id=" + user.ID)
    case RoleApplicant:
        c.Redirect(http.StatusSeeOther, "/applicant/dashboard?id=" + user.ID)
    default:
        c.Redirect(http.StatusSeeOther, "/login") // fallback
    }
    
}

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

func approveRecruiterHandler(c *gin.Context) {
    email := c.Query("email")
    var user User
    result := db.First(&user, "email = ?", email)
    if result.Error != nil {
        if result.Error == gorm.ErrRecordNotFound {
            c.JSON(http.StatusNotFound, gin.H{"error": "Recruiter not found"})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        }
        return
    }

    if user.Role != RoleRecruiter {
        c.JSON(http.StatusBadRequest, gin.H{"error": "User is not a recruiter"})
        return
    }

    // Approve recruiter
    user.Role = RoleRecruiter
    db.Save(&user)
    c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Recruiter %s approved successfully.", email)})
}

func init() {
    db = ConnectDB()
    db.AutoMigrate(&User{}) // Migrate the User struct to the database
}

func SetupAuthRoutes() {
    r := gin.Default()
    r.LoadHTMLGlob("templates/*.html") // Load HTML templates

r.GET("/login", func(c *gin.Context) {
    c.HTML(http.StatusOK, "login.html", nil)
})

    // routes
    r.GET("/auth/login", loginHandler)
    r.GET("/auth/callback", callbackHandler)
    r.GET("/dashboard", dashboardHandler)
    admin := r.Group("/admin", roleMiddleware(RoleSuperAdmin))
    admin.GET("/approve", approveRecruiterHandler)

    fmt.Println("Server started at :8080")
    log.Fatal(r.Run(":8080"))
}