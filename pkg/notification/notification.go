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

package notification

import (
	"fmt"
	"regexp"

	"github.com/cs3org/reva/pkg/notification/template"
)

type Notification struct {
	TemplateName string
	template     template.Template
	Ref          string
	Recipients   []string
}

type Manager interface {
	// UpsertNotification insert or updates a notification.
	UpsertNotification(n Notification) error
	// GetNotification reads a notification.
	GetNotification(ref string) (*Notification, error)
	// DeleteNotification deletes a notifcation.
	DeleteNotification(ref string) error
}

type NotificationNotFoundError struct {
	Ref string
}

type InvalidNotificationError struct {
	Ref string
	Msg string
}

func (nnfe *NotificationNotFoundError) Error() string {
	return fmt.Sprintf("notification %s not found", nnfe.Ref)
}

func (ine *InvalidNotificationError) Error() string {
	return ine.Msg
}

func (n *Notification) InitTemplate(t template.Template) {
	n.template = t
}

func (n *Notification) Send(sender string, templateData map[string]interface{}) error {
	subject, err := n.template.RenderSubject(templateData)
	if err != nil {
		return err
	}

	body, err := n.template.RenderBody(templateData)
	if err != nil {
		return err
	}

	for _, recipient := range n.Recipients {
		err := n.template.Handler.Send(sender, recipient, subject, body)
		if err != nil {
			return err
		}
	}

	return nil
}

func (n *Notification) CheckNotification() error {
	re := regexp.MustCompile(`[a-zA-Z0-9-]`)
	invalidTemplateName := re.ReplaceAllString(n.TemplateName, "")

	if len(n.Ref) == 0 {
		return &InvalidNotificationError{
			Ref: n.Ref,
			Msg: "empty ref",
		}
	}

	if len(invalidTemplateName) > 0 {
		return &InvalidNotificationError{
			Ref: n.Ref,
			Msg: fmt.Sprintf("invalid template name %s", n.TemplateName),
		}
	}

	if len(n.Recipients) == 0 {
		return &InvalidNotificationError{
			Ref: n.Ref,
			Msg: "empty recipient list",
		}
	}

	return nil
}
