package template

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"path"
	"text/template"
)

func Must(t *Template, err error) *Template {
	if err != nil {
		panic(err)
	}
	return t
}

type Template struct {
	patterns []string
	root     string
	// fns is the mapping of names to functions to be used in templates
	fns template.FuncMap
	// tm holds a map of templates
	tm map[string]*template.Template
}

func New(fsys fs.FS, opts ...TemplateOption) (*Template, error) {
	t := Template{
		root:     ".",
		patterns: make([]string, 0),
		fns:      make(template.FuncMap),
		tm:       make(map[string]*template.Template),
	}
	for _, o := range opts {
		o(&t)
	}
	de, err := fs.ReadDir(fsys, t.root)
	if err != nil {
		return nil, fmt.Errorf("error reading directory: %w", err)
	}
	for _, e := range de {
		// for now we ignore sub directories
		if e.IsDir() {
			continue
		}
		/*
				// 2. ignore files with _*.html pattern
			if strings.HasPrefix(f.Name(), "_") || filepath.Ext(f.Name()) != tt.ext {
				continue
			}

			patterns := []string{path.Join(tt.root, f.Name())}
			if tt.partials {
				patterns = append(patterns, path.Join(tt.root, "_*"+tt.ext))
			}
		*/
		// if file doesn't start with a letter continue
		patterns := []string{path.Join(t.root, e.Name())}
		v, err := template.New(e.Name()).Funcs(t.fns).ParseFS(fsys, patterns...)
		if err != nil {
			return nil, fmt.Errorf("error parsing template: %w", err)
		}
		// 3. Add to the map
		t.tm[e.Name()] = v
	}
	return &t, nil
}

// Execute applies a parsed template to the specified data object, writing the output to wr
func (t *Template) Execute(wr io.Writer, name string, data any) error {
	v, ok := t.tm[name+".html"]
	if !ok {
		return fmt.Errorf("template with name %s not found", name)
	}
	err := v.Execute(wr, data)
	if err != nil {
		return fmt.Errorf("error writing to output: %w", err)
	}
	return nil
}

// ExecuteHTTP responds to an HTTP request. If there is an error, a 503 is returned
func (tt *Template) ExecuteHTTP(w http.ResponseWriter, r *http.Request, name string, data any) {
	if err := tt.Execute(w, name, data); err != nil {
		http.Error(w, "error writing partial "+err.Error(), http.StatusServiceUnavailable)
	}
}

// DefaultFuncs adds default functions to be used in a template.
//
//   - Map allows for a map to be passed into the pipeline inline of a template.
var DefaultFuncs = template.FuncMap{
	"map": func(pairs ...any) (map[string]any, error) {
		if len(pairs)%2 != 0 {
			return nil, errors.New("misaligned map")
		}

		m := make(map[string]any, len(pairs)/2)

		for i := 0; i < len(pairs); i += 2 {
			key, ok := pairs[i].(string)
			if !ok {
				return nil, fmt.Errorf("cannot use type %T as map key", pairs[i])
			}
			m[key] = pairs[i+1]
		}
		return m, nil
	},
}

type TemplateOption func(*Template)

func Patterns(patterns ...string) TemplateOption {
	return func(t *Template) {
		t.patterns = append(t.patterns, patterns...)
	}
}

func Root(root string) TemplateOption {
	return func(t *Template) { t.root = root }
}

func Funcs(funcs ...template.FuncMap) TemplateOption {
	return func(t *Template) {
		for _, fs := range funcs {
			for k, v := range fs {
				t.fns[k] = v
			}
		}
	}
}
