package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os/exec"
	"strings"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/namsral/flag"
)

var (
	e                *echo.Echo
	t                *Template
	COMMIT_LONG_SHA  string
	COMMIT_SHORT_SHA string
	VERSION          string
	PORT             int
	PUBLIC_URL       string
	ENV              string
)

func execCmd(name string, args ...string) string {
	cmd := exec.Command(name, args...)
	stdout, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	return strings.TrimSpace(string(stdout))
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	flag.StringVar(&PUBLIC_URL, "PUBLIC_URL", "delphi.market", "Public URL of website")
	flag.IntVar(&PORT, "PORT", 4321, "Server port")
	flag.StringVar(&ENV, "ENV", "development", "Specify for which environment files should be built")
	flag.Parse()
	e = echo.New()
	t = &Template{
		templates: template.Must(template.New("").Funcs(FuncMap).ParseGlob("pages/**.html")),
	}
	COMMIT_LONG_SHA = execCmd("git", "rev-parse", "HEAD")
	COMMIT_SHORT_SHA = execCmd("git", "rev-parse", "--short", "HEAD")
	VERSION = fmt.Sprintf("v0.0.0+%s", COMMIT_SHORT_SHA)
	log.Printf("Running commit %s", COMMIT_SHORT_SHA)
	log.Printf("Public URL: %s", PUBLIC_URL)
	log.Printf("Environment: %s", ENV)
}

func main() {
	e.Static("/", "public")
	e.Renderer = t
	e.GET("/", index)
	e.GET("/login", login)
	e.GET("/api/login", verifyLogin)
	e.GET("/api/session", checkSession)
	e.POST("/logout", logout)
	e.GET("/market/:id", sessionGuard(market))
	e.POST("/market/:id/order", sessionGuard(order))
	e.GET("/invoice/:id", sessionGuard(invoice))
	e.GET("/api/invoice/:id", sessionGuard(invoiceStatus))
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format:           "${time_custom} ${method} ${uri} ${status}\n",
		CustomTimeFormat: "2006-01-02 15:04:05.00000-0700",
	}))
	e.Use(sessionHandler)
	e.HTTPErrorHandler = httpErrorHandler
	err := e.Start(fmt.Sprintf("%s:%d", "127.0.0.1", PORT))
	if err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
