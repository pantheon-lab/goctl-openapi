package main

import (
	"bytes"
	"html/template"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
)

type Opts func(*swaggerConfig)

// SwaggerOpts configures the Doc middlewares.
type swaggerConfig struct {
	// SpecURL the url to find the spec for
	SpecURL string
	// SwaggerHost for the js that generates the swagger ui site
	SwaggerHost string
}

func Doc(basePath, env string, opts ...Opts) http.HandlerFunc {
	config := &swaggerConfig{
		SpecURL:     basePath + "-json",
		SwaggerHost: "https://petstore.swagger.io",
	}
	for _, opt := range opts {
		opt(config)
	}

	tmpl := template.Must(template.New("swaggerdoc").Parse(swaggerTemplateV3))
	buf := bytes.NewBuffer(nil)
	err := tmpl.Execute(buf, config)
	uiHTML := buf.Bytes()

	needPermission := false
	if env == "prod" {
		needPermission = true
	}

	return func(rw http.ResponseWriter, r *http.Request) {
		if err != nil {
			httpx.Error(rw, err)
			return
		}
		if r.URL.Path == basePath {
			if needPermission {
				rw.WriteHeader(http.StatusOK)
				rw.Header().Set("Content-Type", "text/plain")
				_, err = rw.Write([]byte("Swagger not open on prod"))
				if err != nil {
					httpx.Error(rw, err)
				}
				return
			}

			rw.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, err = rw.Write(uiHTML)
			if err != nil {
				httpx.Error(rw, err)
				return
			}

			rw.WriteHeader(http.StatusOK)
			return
		}
	}
}

const swaggerTemplateV3 = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>API documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
    <style>
      html { box-sizing: border-box; overflow: -moz-scrollbars-vertical; overflow-y: scroll; }
      *, *:before, *:after { box-sizing: inherit; }
      body { margin:0; background: #fafafa; }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-standalone-preset.js"></script>
    <script>
    window.onload = function() {
      const ui = SwaggerUIBundle({
        url: "{{ .SpecURL }}",
        dom_id: "#swagger-ui",
        deepLinking: true,
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIStandalonePreset
        ],
        plugins: [
          SwaggerUIBundle.plugins.DownloadUrl
        ],
        layout: "StandaloneLayout",
        validatorUrl: null,
      })
      window.ui = ui
    }
    </script>
</body>
</html>`
