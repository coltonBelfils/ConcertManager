package htmlTemplate

import (
	"fmt"
	"html/template"
	"sync"
)

var (
	templates = map[string]*template.Template{}
	baseMutex sync.Mutex
)

func GetTemplate(name string) (*template.Template, error) {
	baseMutex.Lock()
	defer baseMutex.Unlock()

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
			}).ParseFiles(fmt.Sprintf("./static/%s", name))
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
