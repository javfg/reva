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

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/cs3org/reva/pkg/logger"
	"github.com/cs3org/reva/pkg/notification/notificationhelper"
	"github.com/cs3org/reva/pkg/notification/trigger"
)

var notificationTriggerCommand = func() *command {
	defaultNatsAddress := "cbox-nats-01.cern.ch"
	defaultSender := "noreply@cernbox.cern.ch"
	defaultTemplateData := "{}"

	cmd := newCommand("notification-trigger")
	cmd.Description = func() string { return "trigger a registered notification" }
	cmd.Usage = func() string { return "Usage:  notification-trigger -nats-address <address> [-flags] <ref>" }

	natsAddress := cmd.String("nats-address", defaultNatsAddress, "nats server address")
	sender := cmd.String("sender", defaultSender, "the sender address")
	templateData := cmd.String("template-data", defaultTemplateData, "a json object with data to fill in the template")

	cmd.ResetFlags = func() {
		*natsAddress = defaultNatsAddress
		*sender = defaultSender
		*templateData = defaultTemplateData
	}

	cmd.Action = func(w ...io.Writer) error {
		if cmd.NArg() < 1 {
			return errors.New("Invalid arguments: " + cmd.Usage())
		}

		opt := logger.WithWriter(os.Stdout, logger.ConsoleMode)
		l := logger.New(opt)

		ref := cmd.Args()[0]

		td := make(map[string]any)
		json.Unmarshal([]byte(*templateData), &td)

		fmt.Printf("Sending trigger with ref %s\n", ref)

		nhc := notificationhelper.NotificationHelperConfig{
			NatsAddress: *natsAddress,
		}

		notificationHelper := &notificationhelper.NotificationHelper{
			Name: "reva",
			Conf: &nhc,
			Log:  l,
		}

		if err := notificationHelper.Start(nil, *l); err != nil {
			fmt.Println("error initializing notification helper")
		}

		notificationHelper.TriggerNotification(&trigger.Trigger{
			Ref:          ref,
			Sender:       *sender,
			TemplateData: td,
		})

		return nil
	}

	return cmd
}
