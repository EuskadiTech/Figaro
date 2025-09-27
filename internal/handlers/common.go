// Package handlers provides HTTP request handlers for Figaro application.
package handlers

import (
	"embed"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/EuskadiTech/Figaro/internal/auth"
	"github.com/EuskadiTech/Figaro/pkg/config"
	"github.com/gin-gonic/gin"
)

//go:embed static/*
var staticFS embed.FS

//go:embed templates/*
var templateFS embed.FS

// Handlers holds the application handlers and dependencies
type Handlers struct {
	Config *config.Config
}

// SessionData holds session information
type SessionData struct {
	Centro string
	Aula   string
}

// New creates a new Handlers instance
func New(cfg *config.Config) *Handlers {
	return &Handlers{
		Config: cfg,
	}
}

// loadTemplate loads a specific template by name from the embedded filesystem
func (h *Handlers) loadTemplate(templateName string) (*template.Template, error) {
	// Create a new template with base as the root
	tmpl := template.New("base.html")

	// Add template functions
	tmpl = tmpl.Funcs(template.FuncMap{
		"call": func(fn interface{}, args ...interface{}) interface{} {
			if f, ok := fn.(func(string) bool); ok && len(args) > 0 {
				if arg, ok := args[0].(string); ok {
					return f(arg)
				}
			}
			return false
		},
		"contains": func(slice interface{}, item string) bool {
			switch s := slice.(type) {
			case []string:
				for _, v := range s {
					if v == item {
						return true
					}
				}
			case string:
				return strings.Contains(s, item)
			}
			return false
		},
		"now": func() time.Time {
			return time.Now()
		},
		// Math functions for pagination
		"add": func(a, b int) int {
			return a + b
		},
		"sub": func(a, b int) int {
			return a - b
		},
		"mul": func(a, b int) int {
			return a * b
		},
		"min": func(a, b int) int {
			if a < b {
				return a
			}
			return b
		},
		"max": func(a, b int) int {
			if a > b {
				return a
			}
			return b
		},
		"seq": func(start, end int) []int {
			result := make([]int, 0, end-start+1)
			for i := start; i <= end; i++ {
				result = append(result, i)
			}
			return result
		},
		"slice": func(s string, start, length int) string {
			if start >= len(s) {
				return ""
			}
			end := start + length
			if end > len(s) {
				end = len(s)
			}
			return s[start:end]
		},
	})

	// First, always load base.html
	baseContent, err := templateFS.ReadFile("templates/base.html")
	if err != nil {
		return nil, err
	}

	// Parse base template
	tmpl, err = tmpl.Parse(string(baseContent))
	if err != nil {
		return nil, err
	}

	// Always load pagination.html
	paginationContent, err := templateFS.ReadFile("templates/pagination.html")
	if err != nil {
		return nil, err
	}
	tmpl, err = tmpl.Parse(string(paginationContent))
	if err != nil {
		return nil, err
	}

	// Then load the specific template if it's not base.html
	if templateName != "base.html" {
		content, err := templateFS.ReadFile("templates/" + templateName)
		if err != nil {
			return nil, err
		}

		// Parse the specific template
		tmpl, err = tmpl.Parse(string(content))
		if err != nil {
			return nil, err
		}
	}

	return tmpl, nil
}

// renderTemplate renders a template with the given data
func (h *Handlers) renderTemplate(c *gin.Context, templateName string, data interface{}) {
	tmpl, err := h.loadTemplate(templateName)
	if err != nil {
		c.String(http.StatusInternalServerError, "Template error: %v", err)
		return
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	// Always execute base.html as the root template
	err = tmpl.ExecuteTemplate(c.Writer, "base.html", data)
	if err != nil {
		c.String(http.StatusInternalServerError, "Template execution error: %v", err)
		return
	}
}

// getCommonData creates common template data
func (h *Handlers) getCommonData(c *gin.Context) gin.H {
	data := gin.H{
		"Flash":     c.Query("flash"),
		"PageTitle": "Figaró",
	}

	// Add user if logged in
	if user := auth.GetCurrentUser(c); user != nil {
		data["User"] = user

		// Add session info
		session := gin.H{}
		if centro, err := c.Cookie("centro"); err == nil {
			session["Centro"] = centro
		}
		if aula, err := c.Cookie("aula"); err == nil {
			session["Aula"] = aula
		}
		data["Session"] = session

		// Add permissions check function
		data["HasAccess"] = func(permission string) bool {
			return auth.UserHasAccess(c, permission)
		}
	}

	return data
}