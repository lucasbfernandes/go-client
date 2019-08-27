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

package set

import (
	"context"
	"github.com/atomix/atomix-api/proto/atomix/headers"
	api "github.com/atomix/atomix-api/proto/atomix/set"
	"github.com/atomix/atomix-go-client/pkg/client/primitive"
	"github.com/atomix/atomix-go-client/pkg/client/session"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"testing"
	"time"
)

func TestSession(t *testing.T) {
	client := &testSetServiceClient{
		create:    make(chan bool, 1),
		keepAlive: make(chan bool),
		close:     make(chan bool, 1),
	}
	name := primitive.NewName("default", "test", "default", "test")
	handler := &sessionHandler{client}
	sess, err := session.New(context.TODO(), name, handler, session.WithTimeout(2*time.Second))
	assert.NoError(t, err)
	assert.True(t, client.created())
	assert.Equal(t, "default", sess.Name.Namespace)
	assert.Equal(t, "test", sess.Name.Name)
	assert.Equal(t, uint64(1), sess.SessionID)
	assert.Equal(t, uint64(10), sess.GetState().Index)
	assert.True(t, client.keptAlive())
	err = sess.Close()
	assert.NoError(t, err)
	assert.True(t, client.closed())
}

type testSetServiceClient struct {
	create    chan bool
	keepAlive chan bool
	close     chan bool
}

func (c *testSetServiceClient) created() bool {
	return <-c.create
}

func (c *testSetServiceClient) keptAlive() bool {
	return <-c.keepAlive
}

func (c *testSetServiceClient) closed() bool {
	return <-c.close
}

func (c *testSetServiceClient) Create(ctx context.Context, request *api.CreateRequest, opts ...grpc.CallOption) (*api.CreateResponse, error) {
	c.create <- true
	return &api.CreateResponse{
		Header: &headers.ResponseHeader{
			SessionID: uint64(1),
			Index:     uint64(10),
		},
	}, nil
}

func (c *testSetServiceClient) KeepAlive(ctx context.Context, request *api.KeepAliveRequest, opts ...grpc.CallOption) (*api.KeepAliveResponse, error) {
	c.keepAlive <- true
	return &api.KeepAliveResponse{
		Header: &headers.ResponseHeader{
			SessionID: request.Header.SessionID,
		},
	}, nil
}

func (c *testSetServiceClient) Close(ctx context.Context, request *api.CloseRequest, opts ...grpc.CallOption) (*api.CloseResponse, error) {
	c.close <- true
	return &api.CloseResponse{
		Header: &headers.ResponseHeader{
			SessionID: request.Header.SessionID,
		},
	}, nil
}

func (c *testSetServiceClient) Size(ctx context.Context, in *api.SizeRequest, opts ...grpc.CallOption) (*api.SizeResponse, error) {
	panic("implement me")
}

func (c *testSetServiceClient) Contains(ctx context.Context, in *api.ContainsRequest, opts ...grpc.CallOption) (*api.ContainsResponse, error) {
	panic("implement me")
}

func (c *testSetServiceClient) Add(ctx context.Context, in *api.AddRequest, opts ...grpc.CallOption) (*api.AddResponse, error) {
	panic("implement me")
}

func (c *testSetServiceClient) Remove(ctx context.Context, in *api.RemoveRequest, opts ...grpc.CallOption) (*api.RemoveResponse, error) {
	panic("implement me")
}

func (c *testSetServiceClient) Clear(ctx context.Context, in *api.ClearRequest, opts ...grpc.CallOption) (*api.ClearResponse, error) {
	panic("implement me")
}

func (c *testSetServiceClient) Events(ctx context.Context, in *api.EventRequest, opts ...grpc.CallOption) (api.SetService_EventsClient, error) {
	panic("implement me")
}

func (c *testSetServiceClient) Iterate(ctx context.Context, in *api.IterateRequest, opts ...grpc.CallOption) (api.SetService_IterateClient, error) {
	panic("implement me")
}