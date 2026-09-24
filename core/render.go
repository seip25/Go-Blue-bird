package core

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/seip25/Go-Blue-bird/templates"
)

var (
	templatesMap  = make(map[string]*template.Template)
	templatesOnce sync.Once
)

func InitTemplates() {
	templatesOnce.Do(func() {
		err := fs.WalkDir(templates.FS, ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if !d.IsDir() && strings.HasSuffix(path, ".html") && strings.HasPrefix(path, "pages/") {
				tmpl, parseErr := template.ParseFS(templates.FS, "layouts/base.html", path)
				if parseErr != nil {
					return fmt.Errorf("failed to parse template '%s': %w", path, parseErr)
				}

				templatesMap[path] = tmpl

				relToPages := strings.TrimPrefix(path, "pages/")
				templatesMap[relToPages] = tmpl

				withoutExt := strings.TrimSuffix(relToPages, filepath.Ext(relToPages))
				templatesMap[withoutExt] = tmpl
			}
			return nil
		})

		if err != nil {
			panic(fmt.Sprintf("Failed to initialize templates: %v", err))
		}
	})
}

func Render(c *gin.Context, status int, templateName string, data interface{}) {
	InitTemplates()

	tmpl, exists := templatesMap[templateName]
	if !exists && !strings.HasSuffix(templateName, ".html") {
		tmpl, exists = templatesMap[templateName+".html"]
	}

	if !exists {
		c.String(http.StatusInternalServerError, "Template not found: '%s'", templateName)
		return
	}

	var buf bytes.Buffer
	targetBlock := "base"
	if tmpl.Lookup("base") == nil {
		targetBlock = templateName
	}

	if err := tmpl.ExecuteTemplate(&buf, targetBlock, data); err != nil {
		c.String(http.StatusInternalServerError, "Template render error ('%s'): %v", templateName, err)
		return
	}

	c.Data(status, "text/html; charset=utf-8", buf.Bytes())
}
