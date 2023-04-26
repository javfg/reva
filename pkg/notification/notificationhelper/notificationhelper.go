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

package notificationhelper

import (
	"encoding/json"
	"fmt"

	"github.com/cs3org/reva/pkg/notification"
	"github.com/cs3org/reva/pkg/notification/template"
	"github.com/cs3org/reva/pkg/notification/trigger"
	"github.com/cs3org/reva/pkg/utils"
	"github.com/mitchellh/mapstructure"
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type NotificationHelper struct {
	Name string
	Conf *NotificationHelperConfig
	Log  *zerolog.Logger
	nc   nats.Conn
	js   nats.JetStreamContext
	kv   nats.KeyValue
}

type NotificationHelperConfig struct {
	NatsAddress string `mapstructure:"nats_address" docs:";The NATS server address."`
	NatsToken   string `mapstructure:"nats_token" docs:";The token to authenticate against the NATS server"`
	NatsStream  string `mapstructure:"nats_stream" docs:"reva-notifications;The notifications NATS stream."`
}

// Start initializes the notification helper
func (nh *NotificationHelper) Start(ts map[string]map[string]interface{}, log zerolog.Logger) error {
	nh.Conf.NatsStream = utils.WithDefault(nh.Conf.NatsStream, "reva-notifications")

	nc, err := nats.Connect(nh.Conf.NatsAddress, nats.Token(nh.Conf.NatsToken))
	if err != nil {
		return errors.Wrapf(err, "connection to nats server at '%s' failed", nh.Conf.NatsAddress)
	}
	nh.nc = *nc

	js, err := nc.JetStream(nats.PublishAsyncMaxPending(256))
	if err != nil {
		return errors.Wrap(err, "jetstream initialization failed")
	}

	stream, _ := js.StreamInfo(nh.Conf.NatsStream)
	if stream != nil {
		if _, err := js.AddStream(&nats.StreamConfig{
			Name: nh.Conf.NatsStream,
			Subjects: []string{
				fmt.Sprintf("%s.notification", nh.Conf.NatsStream),
				fmt.Sprintf("%s.trigger", nh.Conf.NatsStream),
			},
		}); err != nil {
			return errors.Wrap(err, "nats stream creation failed")
		}
	}

	nh.js = js

	bucketName := fmt.Sprintf("%s-template", nh.Conf.NatsStream)
	kv, err := nh.js.CreateKeyValue(&nats.KeyValueConfig{
		Bucket: bucketName,
	})
	if err != nil {
		return errors.Wrap(err, "template store creation failed")
	}

	nh.kv = kv

	if tCount, err := nh.RegisterTemplates(ts); err != nil {
		return errors.Wrapf(err, "template registration failed")
	} else {
		if tCount == 0 {
			nh.Log.Info().Str("service", nh.Name).Msg("no templates to register")
		} else {
			nh.Log.Info().Str("service", nh.Name).Msgf("%d notification templates successfully registered", tCount)
		}
	}

	return nil
}

// Stop stops the notification helper
func (nh *NotificationHelper) Stop() {
	if err := nh.nc.Drain(); err != nil {
		nh.Log.Fatal().Err(err)
	}
}

// RegisterTemplates registers multiple templates on the notification service
func (nh *NotificationHelper) RegisterTemplates(ts map[string]map[string]interface{}) (int, error) {
	if len(ts) == 0 {
		return 0, nil
	}

	tCount := 0
	var tc *map[string]template.RegistrationRequest
	if err := mapstructure.Decode(ts, &tc); err != nil {
		return 0, err
	}

	for _, c := range *tc {
		nh.RegisterTemplate(&c)
		tCount++
	}

	return tCount, nil
}

// RegisterTemplate send a Template Registration Request to the notification service
func (nh *NotificationHelper) RegisterTemplate(rr *template.RegistrationRequest) {
	if tb, err := json.Marshal(rr); err != nil {
		nh.Log.Error().Err(err).Msgf("template registration json marshalling failed")
	} else {
		_, err := nh.kv.Put(rr.Name, tb)
		if err != nil {
			nh.Log.Error().Err(err).Msgf("template registration publish failed")
		}
		nh.Log.Debug().Str("service", nh.Name).Msgf("%s template registration published", rr.Name)
	}
}

// RegisterNotification registers a notification in the notification service
func (nh *NotificationHelper) RegisterNotification(n *notification.Notification) {
	if nb, err := json.Marshal(n); err != nil {
		nh.Log.Error().Err(err).Msgf("notification registration json marshalling failed")
	} else {
		notificationSubject := fmt.Sprintf("%s.notification-register", nh.Conf.NatsStream)
		_, err := nh.js.Publish(notificationSubject, nb)
		if err != nil {
			nh.Log.Error().Err(err).Msgf("notification registration publish failed")
		}
		nh.Log.Debug().Str("service", nh.Name).Msgf("%s notification registration published", n.Ref)
	}
}

// UnregisterNotification unregisters a notification in the notification service
func (nh *NotificationHelper) UnregisterNotification(ref string) {
	notificationSubject := fmt.Sprintf("%s.notification-unregister", nh.Conf.NatsStream)
	_, err := nh.js.Publish(notificationSubject, []byte(ref))
	if err != nil {
		nh.Log.Error().Err(err).Msgf("notification registration publish failed")
	}
	nh.Log.Debug().Str("service", nh.Name).Msgf("%s notification unregistration published", ref)
}

// TriggerNotification sends a notification trigger to the notifications service
func (nh *NotificationHelper) TriggerNotification(tr *trigger.Trigger) {
	if trb, err := json.Marshal(tr); err != nil {
		nh.Log.Error().Err(err).Msgf("notification trigger json marshalling failed")
	} else {
		triggerSubject := fmt.Sprintf("%s.trigger", nh.Conf.NatsStream)
		_, err := nh.js.Publish(triggerSubject, trb)
		if err != nil {
			nh.Log.Error().Err(err).Msgf("notification trigger publish failed")
		}
		nh.Log.Debug().Str("service", nh.Name).Msgf("%s notification trigger published", tr.Ref)
	}
}
