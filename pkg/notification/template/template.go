// Copyright 2018-2023 CERN
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// In applying this license, CERN does not waive the privileges and immunities
// granted to it by virtue of its status as an Intergovernmental Organization
// or submit itself to any jurisdiction.

package template

import (
	"bytes"
	"errors"
	"fmt"
	htmlTemplate "html/template"
	"io"
	"os"
	"regexp"
	textTemplate "text/template"

	"github.com/cs3org/reva/pkg/notification/handler"
	"github.com/mitchellh/mapstructure"
)

type TemplateType int

const (
	Subject TemplateType = iota
	Body
)

type RegistrationRequest struct {
	Name            string `mapstructure:"name" json:"name"`
	Handler         string `mapstructure:"handler" json:"handler"`
	BodyTmplPath    string `mapstructure:"body_template_path" json:"body_template_path"`
	SubjectTmplPath string `mapstructure:"subject_template_path" json:"subject_template_path"`
	Persistent      bool   `mapstructure:"persistent" json:"persistent"`
}

type Template struct {
	Name        string
	Handler     handler.Handler
	Persistent  bool
	tmplSubject *textTemplate.Template
	tmplBody    *htmlTemplate.Template
}

type TemplateFileNotFoundError struct {
	TemplateFileName string
	Err              error
}

func (t *TemplateFileNotFoundError) Error() string {
	return fmt.Sprintf("template file %s not found", t.TemplateFileName)
}

// New creates a new Template from a RegistrationRequest.
// The last return parameter is used for template deletion from the kv store.
func New(m map[string]interface{}, hs map[string]handler.Handler) (*Template, string, error, bool) {
	rr := &RegistrationRequest{}
	if err := mapstructure.Decode(m, rr); err != nil {
		return nil, rr.Name, err, false
	}

	h, ok := hs[rr.Handler]
	if !ok {
		return nil, rr.Name, fmt.Errorf("unknown handler %s", rr.Handler), false
	}

	tmplSubject, err := parseTmplFile(rr.SubjectTmplPath, Subject)
	if err != nil {
		_, isTemplateFileNotFoundError := err.(*TemplateFileNotFoundError)
		return nil, rr.Name, err, isTemplateFileNotFoundError
	}

	tmplBody, err := parseTmplFile(rr.BodyTmplPath, Body)
	if err != nil {
		_, isTemplateFileNotFoundError := err.(*TemplateFileNotFoundError)
		return nil, rr.Name, err, isTemplateFileNotFoundError
	}

	t := &Template{
		Name:        rr.Name,
		Handler:     h,
		tmplSubject: tmplSubject.(*textTemplate.Template),
		tmplBody:    tmplBody.(*htmlTemplate.Template),
	}

	if ok := t.checkTemplateName(); !ok {
		return nil, rr.Name, fmt.Errorf("template name %s must contain only alphanumeric characters and hyphens", rr.Name), false
	}

	return t, rr.Name, nil, false
}

func (t *Template) RenderSubject(arguments map[string]interface{}) (string, error) {
	var buf bytes.Buffer
	err := t.tmplSubject.Execute(&buf, arguments)
	return buf.String(), err
}

func (t *Template) RenderBody(arguments map[string]interface{}) (string, error) {
	var buf bytes.Buffer
	err := t.tmplBody.Execute(&buf, arguments)
	return buf.String(), err
}

func (t *Template) checkTemplateName() bool {
	re := regexp.MustCompile(`[a-zA-Z0-9-]`)
	invalidChars := re.ReplaceAllString(t.Name, "")

	return len(invalidChars) == 0
}

func parseTmplFile(path string, t TemplateType) (interface{}, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, &TemplateFileNotFoundError{
			TemplateFileName: path,
			Err:              err,
		}
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, &TemplateFileNotFoundError{
			TemplateFileName: path,
			Err:              err,
		}
	}

	switch t {
	case Subject:
		tmpl, err := textTemplate.New("subject").Parse(string(data))
		if err != nil {
			return nil, err
		}

		return tmpl, nil
	case Body:
		tmpl, err := htmlTemplate.New("body").Parse(string(data))
		if err != nil {
			return nil, err
		}

		return tmpl, nil
	}

	return nil, errors.New("unknown template type")
}
