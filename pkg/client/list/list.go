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

package list

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/atomix/atomix-api/proto/atomix/headers"
	api "github.com/atomix/atomix-api/proto/atomix/list"
	"github.com/atomix/atomix-go-client/pkg/client/primitive"
	"github.com/atomix/atomix-go-client/pkg/client/session"
	"github.com/atomix/atomix-go-client/pkg/client/util"
	"github.com/atomix/atomix-go-client/pkg/client/util/net"
	"google.golang.org/grpc"
	"time"
)

// Type is the list type
const Type primitive.Type = "list"

// Client provides an API for creating Lists
type Client interface {
	// GetList gets the List instance of the given name
	GetList(ctx context.Context, name string, opts ...session.Option) (List, error)
}

// List provides a distributed list data structure
// The list values are defines as strings. To store more complex types in the list, encode values to strings e.g.
// using base 64 encoding.
type List interface {
	primitive.Primitive

	// Append pushes a value on to the end of the list
	Append(ctx context.Context, value []byte) error

	// Insert inserts a value at the given index
	Insert(ctx context.Context, index int, value []byte) error

	// Set sets the value at the given index
	Set(ctx context.Context, index int, value []byte) error

	// Get gets the value at the given index
	Get(ctx context.Context, index int) ([]byte, error)

	// Remove removes and returns the value at the given index
	Remove(ctx context.Context, index int) ([]byte, error)

	// Len gets the length of the list
	Len(ctx context.Context) (int, error)

	// Slice returns a slice of the list from the given start index to the given end index
	Slice(ctx context.Context, from int, to int) (List, error)

	// SliceFrom returns a slice of the list from the given index
	SliceFrom(ctx context.Context, from int) (List, error)

	// SliceTo returns a slice of the list to the given index
	SliceTo(ctx context.Context, to int) (List, error)

	// Items iterates through the values in the list
	// This is a non-blocking method. If the method returns without error, values will be pushed on to the
	// given channel and the channel will be closed once all values have been read from the list.
	Items(ctx context.Context, ch chan<- []byte) error

	// Watch watches the list for changes
	// This is a non-blocking method. If the method returns without error, list events will be pushed onto
	// the given channel.
	Watch(ctx context.Context, ch chan<- *Event, opts ...WatchOption) error

	// Clear removes all values from the list
	Clear(ctx context.Context) error
}

// EventType is the type for a list Event
type EventType string

const (
	// EventNone indicates the event is not a change event
	EventNone EventType = ""

	// EventInserted indicates a value was added to the list
	EventInserted EventType = "added"

	// EventRemoved indicates a value was removed from the list
	EventRemoved EventType = "removed"
)

// Event is a list change event
type Event struct {
	// Type indicates the event type
	Type EventType

	// Index is the index at which the event occurred
	Index int

	// Value is the value that was changed
	Value []byte
}

// New creates a new list primitive
func New(ctx context.Context, name primitive.Name, partitions []net.Address, opts ...session.Option) (List, error) {
	i, err := util.GetPartitionIndex(name.Name, len(partitions))
	if err != nil {
		return nil, err
	}
	return newList(ctx, name, partitions[i], opts...)
}

// newList creates a new list for the given partition
func newList(ctx context.Context, name primitive.Name, address net.Address, opts ...session.Option) (*list, error) {
	sess, err := session.New(ctx, name, address, &sessionHandler{}, opts...)
	if err != nil {
		return nil, err
	}
	return &list{
		name:    name,
		session: sess,
	}, nil
}

// list is the single partition implementation of List
type list struct {
	name    primitive.Name
	session *session.Session
}

func (l *list) Name() primitive.Name {
	return l.name
}

func (l *list) Append(ctx context.Context, value []byte) error {
	_, err := l.session.DoCommand(ctx, func(ctx context.Context, conn *grpc.ClientConn, header *headers.RequestHeader) (*headers.ResponseHeader, interface{}, error) {
		client := api.NewListServiceClient(conn)
		request := &api.AppendRequest{
			Header: header,
			Value:  base64.StdEncoding.EncodeToString(value),
		}
		response, err := client.Append(ctx, request)
		if err != nil {
			return nil, nil, err
		}
		return response.Header, response, nil
	})
	return err
}

func (l *list) Insert(ctx context.Context, index int, value []byte) error {
	response, err := l.session.DoCommand(ctx, func(ctx context.Context, conn *grpc.ClientConn, header *headers.RequestHeader) (*headers.ResponseHeader, interface{}, error) {
		client := api.NewListServiceClient(conn)
		request := &api.InsertRequest{
			Header: header,
			Index:  uint32(index),
			Value:  base64.StdEncoding.EncodeToString(value),
		}
		response, err := client.Insert(ctx, request)
		if err != nil {
			return nil, nil, err
		}
		return response.Header, response, nil
	})
	if err != nil {
		return err
	}

	switch response.(*api.InsertResponse).Status {
	case api.ResponseStatus_OUT_OF_BOUNDS:
		return errors.New("index out of bounds")
	default:
		return nil
	}
}

func (l *list) Set(ctx context.Context, index int, value []byte) error {
	response, err := l.session.DoCommand(ctx, func(ctx context.Context, conn *grpc.ClientConn, header *headers.RequestHeader) (*headers.ResponseHeader, interface{}, error) {
		client := api.NewListServiceClient(conn)
		request := &api.SetRequest{
			Header: header,
			Index:  uint32(index),
			Value:  base64.StdEncoding.EncodeToString(value),
		}
		response, err := client.Set(ctx, request)
		if err != nil {
			return nil, nil, err
		}
		return response.Header, response, nil
	})
	if err != nil {
		return err
	}

	switch response.(*api.SetResponse).Status {
	case api.ResponseStatus_OUT_OF_BOUNDS:
		return errors.New("index out of bounds")
	default:
		return nil
	}
}

func (l *list) Get(ctx context.Context, index int) ([]byte, error) {
	r, err := l.session.DoQuery(ctx, func(ctx context.Context, conn *grpc.ClientConn, header *headers.RequestHeader) (*headers.ResponseHeader, interface{}, error) {
		client := api.NewListServiceClient(conn)
		request := &api.GetRequest{
			Header: header,
			Index:  uint32(index),
		}
		response, err := client.Get(ctx, request)
		if err != nil {
			return nil, nil, err
		}
		return response.Header, response, nil
	})
	if err != nil {
		return nil, err
	}

	response := r.(*api.GetResponse)
	switch response.Status {
	case api.ResponseStatus_OUT_OF_BOUNDS:
		return nil, errors.New("index out of bounds")
	default:
		return base64.StdEncoding.DecodeString(response.Value)
	}
}

func (l *list) Remove(ctx context.Context, index int) ([]byte, error) {
	r, err := l.session.DoCommand(ctx, func(ctx context.Context, conn *grpc.ClientConn, header *headers.RequestHeader) (*headers.ResponseHeader, interface{}, error) {
		client := api.NewListServiceClient(conn)
		request := &api.RemoveRequest{
			Header: header,
			Index:  uint32(index),
		}
		response, err := client.Remove(ctx, request)
		if err != nil {
			return nil, nil, err
		}
		return response.Header, response, nil
	})
	if err != nil {
		return nil, err
	}

	response := r.(*api.RemoveResponse)
	switch response.Status {
	case api.ResponseStatus_OUT_OF_BOUNDS:
		return nil, errors.New("index out of bounds")
	default:
		return base64.StdEncoding.DecodeString(response.Value)
	}
}

func (l *list) Len(ctx context.Context) (int, error) {
	response, err := l.session.DoQuery(ctx, func(ctx context.Context, conn *grpc.ClientConn, header *headers.RequestHeader) (*headers.ResponseHeader, interface{}, error) {
		client := api.NewListServiceClient(conn)
		request := &api.SizeRequest{
			Header: header,
		}
		response, err := client.Size(ctx, request)
		if err != nil {
			return nil, nil, err
		}
		return response.Header, response, nil
	})
	if err != nil {
		return 0, err
	}
	return int(response.(*api.SizeResponse).Size_), nil
}

func (l *list) Items(ctx context.Context, ch chan<- []byte) error {
	stream, err := l.session.DoQueryStream(ctx, func(ctx context.Context, conn *grpc.ClientConn, header *headers.RequestHeader) (interface{}, error) {
		client := api.NewListServiceClient(conn)
		request := &api.IterateRequest{
			Header: header,
		}
		return client.Iterate(ctx, request)
	}, func(responses interface{}) (*headers.ResponseHeader, interface{}, error) {
		response, err := responses.(api.ListService_IterateClient).Recv()
		if err != nil {
			return nil, nil, err
		}
		return response.Header, response, nil
	})
	if err != nil {
		return err
	}

	go func() {
		defer close(ch)
		for event := range stream {
			response := event.(*api.IterateResponse)
			if bytes, err := base64.StdEncoding.DecodeString(response.Value); err == nil {
				ch <- bytes
			}
		}
	}()
	return nil
}

func (l *list) Watch(ctx context.Context, ch chan<- *Event, opts ...WatchOption) error {
	stream, err := l.session.DoCommandStream(ctx, func(ctx context.Context, conn *grpc.ClientConn, header *headers.RequestHeader) (interface{}, error) {
		client := api.NewListServiceClient(conn)
		request := &api.EventRequest{
			Header: header,
		}
		for _, opt := range opts {
			opt.beforeWatch(request)
		}
		return client.Events(ctx, request)
	}, func(responses interface{}) (*headers.ResponseHeader, interface{}, error) {
		response, err := responses.(api.ListService_EventsClient).Recv()
		if err != nil {
			return nil, nil, err
		}
		for _, opt := range opts {
			opt.afterWatch(response)
		}
		return response.Header, response, nil
	})
	if err != nil {
		return err
	}

	select {
	case event, ok := <-stream:
		if !ok {
			return errors.New("watch handshake failed")
		}
		response := event.(*api.EventResponse)
		if response.Type != api.EventResponse_OPEN {
			return fmt.Errorf("expected handshake response, received %v", response)
		}
	case <-time.After(15 * time.Second):
		return errors.New("handshake timed out")
	}

	go func() {
		defer close(ch)
		for event := range stream {
			response := event.(*api.EventResponse)
			// If this is a normal event (not a handshake response), write the event to the watch channel
			if response.Type != api.EventResponse_OPEN {
				var t EventType
				switch response.Type {
				case api.EventResponse_NONE:
					t = EventNone
				case api.EventResponse_ADDED:
					t = EventInserted
				case api.EventResponse_REMOVED:
					t = EventRemoved
				}

				if bytes, err := base64.StdEncoding.DecodeString(response.Value); err == nil {
					ch <- &Event{
						Type:  t,
						Index: int(response.Index),
						Value: bytes,
					}
				}
			}
		}
	}()
	return nil
}

func (l *list) Slice(ctx context.Context, from int, to int) (List, error) {
	return &slicedList{
		from: &from,
		to:   &to,
		list: l,
	}, nil
}

func (l *list) SliceFrom(ctx context.Context, from int) (List, error) {
	return &slicedList{
		from: &from,
		list: l,
	}, nil
}

func (l *list) SliceTo(ctx context.Context, to int) (List, error) {
	return &slicedList{
		to:   &to,
		list: l,
	}, nil
}

func (l *list) Clear(ctx context.Context) error {
	_, err := l.session.DoCommand(ctx, func(ctx context.Context, conn *grpc.ClientConn, header *headers.RequestHeader) (*headers.ResponseHeader, interface{}, error) {
		client := api.NewListServiceClient(conn)
		request := &api.ClearRequest{
			Header: header,
		}
		response, err := client.Clear(ctx, request)
		if err != nil {
			return nil, nil, err
		}
		return response.Header, response, nil
	})
	return err
}

func (l *list) Close() error {
	return l.session.Close()
}

func (l *list) Delete() error {
	return l.session.Delete()
}
