package registry

import (
	"encoding/json"
	"fmt"

	"github.com/cs3org/reva/pkg/notification/handler"
	"github.com/cs3org/reva/pkg/notification/template"
	"github.com/pkg/errors"
)

type Registry struct {
	store map[string]template.Template
}

// New returns a new Template Registry.
func New() *Registry {
	r := &Registry{
		store: make(map[string]template.Template),
	}

	return r
}

func (r *Registry) Put(tb []byte, hs map[string]handler.Handler) (*string, error, bool) {
	var data map[string]interface{}

	err := json.Unmarshal(tb, &data)
	if err != nil {
		return nil, errors.Wrapf(err, "template registration unmarshall failed"), false
	}

	t, name, err, delete := template.New(data, hs)
	if err != nil {
		return &name, errors.Wrapf(err, "template %s registration failed", name), delete
	}

	r.store[t.Name] = *t
	return &t.Name, nil, false
}

func (r *Registry) Get(n string) (*template.Template, error) {
	if t, ok := r.store[n]; ok {
		return &t, nil
	} else {
		return nil, fmt.Errorf("template %s not found", n)
	}
}
