package httpserver

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/ilpaka/landmark_app/backend/auth-service/internal/config"
)

const swaggerUIPage = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>Auth API — Swagger UI</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui.css" crossorigin="anonymous" />
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui-bundle.js" crossorigin="anonymous"></script>
  <script>
    window.onload = function () {
      window.ui = SwaggerUIBundle({
        url: "/openapi/auth.yaml",
        dom_id: "#swagger-ui",
        presets: [SwaggerUIBundle.presets.apis],
        layout: "BaseLayout",
      });
    };
  </script>
</body>
</html>`

// findAuthOpenAPISpec returns an absolute path to auth.yaml, or "" if not found.
func findAuthOpenAPISpec(cfg *config.Config) string {
	if p := strings.TrimSpace(cfg.OpenAPIAuthYAML); p != "" {
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			if abs, err := filepath.Abs(p); err == nil {
				return abs
			}
			return p
		}
	}
	candidates := []string{
		"api/openapi/auth.yaml",
		"../api/openapi/auth.yaml",
		filepath.Join("..", "..", "api", "openapi", "auth.yaml"),
		filepath.Join("backend", "auth-service", "api", "openapi", "auth.yaml"),
	}
	for _, c := range candidates {
		if st, err := os.Stat(c); err == nil && !st.IsDir() {
			abs, err := filepath.Abs(c)
			if err != nil {
				return c
			}
			return abs
		}
	}
	return ""
}

func shouldExposeSwagger(cfg *config.Config) bool {
	if cfg.Env == "production" {
		return cfg.DevSwagger
	}
	return true
}

// mountSwaggerUI registers GET /docs (Swagger UI) and GET /openapi/auth.yaml when specPath is non-empty.
func mountSwaggerUI(engine *gin.Engine, specPath string) {
	if specPath == "" {
		engine.GET("/docs", func(c *gin.Context) {
			c.Data(http.StatusNotFound, "text/html; charset=utf-8", []byte(`<!DOCTYPE html><html><head><meta charset="utf-8"><title>Auth API docs</title></head><body>
<p>OpenAPI file not found. Set <code>OPENAPI_AUTH_YAML</code> to the absolute path of <code>api/openapi/auth.yaml</code> inside the auth-service module, or run from the repository root / <code>backend/auth-service/</code> so a default relative path resolves.</p>
</body></html>`))
		})
		return
	}
	engine.GET("/docs", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(swaggerUIPage))
	})
	engine.GET("/openapi/auth.yaml", func(c *gin.Context) {
		c.File(specPath)
	})
	engine.GET("/dev/openapi/auth.yaml", func(c *gin.Context) {
		c.File(specPath)
	})
	engine.GET("/dev/docs", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/docs")
	})
}
