// Copyright 2018-2021 CERN
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

package machine

import (
	"context"

	authpb "github.com/cs3org/go-cs3apis/cs3/auth/provider/v1beta1"
	user "github.com/cs3org/go-cs3apis/cs3/identity/user/v1beta1"
	rpc "github.com/cs3org/go-cs3apis/cs3/rpc/v1beta1"
	"github.com/cs3org/reva/pkg/auth"
	"github.com/cs3org/reva/pkg/auth/manager/registry"
	"github.com/cs3org/reva/pkg/auth/scope"
	"github.com/cs3org/reva/pkg/errtypes"
	"github.com/cs3org/reva/pkg/rgrpc/todo/pool"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

// 'machine' is an authentication method used to impersonate users.
// To impersonate the given user it's only needed an api-key, saved
// in a config file.

type manager struct {
	APIKey      string `mapstructure:"api_key"`
	GatewayAddr string `mapstructure:"gateway_addr"`
}

func init() {
	registry.Register("machine", New)
}

// Configure parses the map conf
func (m *manager) Configure(conf map[string]interface{}) error {
	err := mapstructure.Decode(conf, m)
	if err != nil {
		return errors.Wrap(err, "error decoding conf")
	}
	return nil
}

// New creates a new manager for the 'machine' authentication
func New(conf map[string]interface{}) (auth.Manager, error) {
	m := &manager{}
	err := m.Configure(conf)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// Authenticate impersonate an user if the provided secret is equal to the api-key
func (m *manager) Authenticate(ctx context.Context, username, secret string) (*user.User, map[string]*authpb.Scope, error) {
	if m.APIKey != secret {
		return nil, nil, errtypes.InvalidCredentials("")
	}

	gtw, err := pool.GetGatewayServiceClient(m.GatewayAddr)
	if err != nil {
		return nil, nil, err
	}

	userResponse, err := gtw.GetUserByClaim(ctx, &user.GetUserByClaimRequest{
		Claim: "username",
		Value: username,
	})

	switch {
	case err != nil:
		return nil, nil, err
	case userResponse.Status.Code == rpc.Code_CODE_NOT_FOUND:
		return nil, nil, errtypes.NotFound(userResponse.Status.Message)
	case userResponse.Status.Code != rpc.Code_CODE_OK:
		return nil, nil, errtypes.InternalError(userResponse.Status.Message)
	}

	scope, err := scope.AddOwnerScope(nil)
	if err != nil {
		return nil, nil, err
	}

	return userResponse.GetUser(), scope, nil

}