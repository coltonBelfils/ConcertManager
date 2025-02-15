package htmlTemplate

import (
	"fmt"
	"github.com/cockroachdb/errors"
	"html/template"
	"path/filepath"
	"sync"
)

var (
	templates = map[string]*template.Template{}
	baseMutex sync.Mutex
)

func GetTemplate(name string) (*template.Template, error) {
	baseMutex.Lock()
	defer baseMutex.Unlock()

	files, globErr := filepath.Glob("./static/templates/*.gohtml")
	if globErr != nil {
		return nil, errors.Wrap(globErr, "globing all ./static/templates/*.gohtml failed")
	}

	files = append(files, fmt.Sprintf("./static/%s", name))

	fmt.Printf("templates: %+v\n", files)

	var tmpl *template.Template

	if t, ok := templates[name]; ok {
		tmpl = t
	} else {
		var parseErr error
		tmpl, parseErr = template.New(name).Funcs(
			template.FuncMap{
				"dStr":   dStr,
				"dInt":   dInt,
				"dFloat": dFloat,
				"dBool":  dBool,
			}).ParseFiles(files...)
		if parseErr != nil {
			return nil, parseErr
		}

		templates[name] = tmpl
	}

	return tmpl, nil
}

func dStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func dInt(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}

func dFloat(f *float64) float64 {
	if f == nil {
		return 0
	}
	return *f
}

func dBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}
