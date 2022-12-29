// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build !windows
// +build !windows

package dcgmreceiver

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestNewDcgmClientOnInitializationError(t *testing.T) {
	realDcgmInit := dcgmInit
	defer func() { dcgmInit = realDcgmInit }()
	dcgmInit = func(args ...string) (func(), error) {
		return nil, fmt.Errorf("No DCGM client library *OR* No DCGM connection")
	}

	client, err := newClient(createDefaultConfig().(*Config), zaptest.NewLogger(t))
	require.Regexp(t, ".*Unable to connect.*", err)
	require.Nil(t, client)
}