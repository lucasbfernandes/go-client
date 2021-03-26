// Copyright 2019-present Open Networking Foundation.
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

package counter

import (
	"context"
	"github.com/lucasbfernandes/go-client/pkg/client/errors"
	"github.com/lucasbfernandes/go-client/pkg/client/primitive"
	"github.com/lucasbfernandes/go-client/pkg/client/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCounterOperations(t *testing.T) {
	partitions, closers := test.StartTestPartitions(3)
	defer test.StopTestPartitions(closers)

	sessions, err := test.OpenSessions(partitions)
	assert.NoError(t, err)
	defer test.CloseSessions(sessions)

	name := primitive.NewName("default", "test", "default", "test")
	counter, err := New(context.TODO(), name, sessions)
	assert.NoError(t, err)
	assert.NotNil(t, counter)

	value, err := counter.Get(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, int64(0), value)

	err = counter.Set(context.TODO(), 1)
	assert.NoError(t, err)

	value, err = counter.Get(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, int64(1), value)

	err = counter.Set(context.TODO(), -1)
	assert.NoError(t, err)

	value, err = counter.Get(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, int64(-1), value)

	value, err = counter.Increment(context.TODO(), 1)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), value)

	value, err = counter.Decrement(context.TODO(), 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(-10), value)

	value, err = counter.Get(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, int64(-10), value)

	value, err = counter.Increment(context.TODO(), 20)
	assert.NoError(t, err)
	assert.Equal(t, int64(10), value)

	err = counter.Close(context.Background())
	assert.NoError(t, err)

	counter1, err := New(context.TODO(), name, sessions)
	assert.NoError(t, err)

	counter2, err := New(context.TODO(), name, sessions)
	assert.NoError(t, err)

	value, err = counter1.Get(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, int64(10), value)

	err = counter1.Close(context.Background())
	assert.NoError(t, err)

	err = counter1.Delete(context.Background())
	assert.NoError(t, err)

	err = counter2.Delete(context.Background())
	assert.Error(t, err)
	assert.True(t, errors.IsNotFound(err))

	counter, err = New(context.TODO(), name, sessions)
	assert.NoError(t, err)

	value, err = counter.Get(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, int64(0), value)
}
