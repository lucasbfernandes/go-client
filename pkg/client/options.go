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

package client

import (
	"github.com/atomix/go-client/pkg/client/peer"
	"os"
	"time"
)

func applyOptions(opts ...Option) *options {
	options := &options{
		namespace:      os.Getenv("ATOMIX_NAMESPACE"),
		scope:          os.Getenv("ATOMIX_SCOPE"),
		peerPort:       8080,
		sessionTimeout: 1 * time.Minute,
	}
	for _, opt := range opts {
		opt.apply(options)
	}
	return options
}

type options struct {
	memberID       string
	peerHost       string
	peerPort       int
	services       []peer.Service
	joinTimeout    *time.Duration
	scope          string
	namespace      string
	sessionTimeout time.Duration
}

// Option provides a client option
type Option interface {
	apply(options *options)
}

// WithMemberID configures the client's member ID
func WithMemberID(memberID string) Option {
	return &memberIDOption{id: memberID}
}

type memberIDOption struct {
	id string
}

func (o *memberIDOption) apply(options *options) {
	options.memberID = o.id
	if options.peerHost == "" {
		options.peerHost = o.id
	}
}

// WithPeerHost configures the client's peer host
func WithPeerHost(host string) Option {
	return &peerHostOption{host: host}
}

type peerHostOption struct {
	host string
}

func (o *peerHostOption) apply(options *options) {
	options.peerHost = o.host
}

// WithPeerPort configures the client's peer port
func WithPeerPort(port int) Option {
	return &peerPortOption{port: port}
}

type peerPortOption struct {
	port int
}

func (o *peerPortOption) apply(options *options) {
	options.peerPort = o.port
}

// WithService configures a peer-to-peer service
func WithService(service peer.Service) Option {
	return &serviceOption{
		service: service,
	}
}

type serviceOption struct {
	service peer.Service
}

func (o *serviceOption) apply(options *options) {
	if options.services == nil {
		options.services = make([]peer.Service, 0)
	}
	options.services = append(options.services, o.service)
}

// WithJoinTimeout configures the client's join timeout
func WithJoinTimeout(timeout time.Duration) Option {
	return &joinTimeoutOption{timeout: timeout}
}

type joinTimeoutOption struct {
	timeout time.Duration
}

func (o *joinTimeoutOption) apply(options *options) {
	options.joinTimeout = &o.timeout
}

type scopeOption struct {
	scope string
}

func (o *scopeOption) apply(options *options) {
	options.scope = o.scope
}

// WithApplication configures the application name for the client
// Deprecated: Use WithScope instead
func WithApplication(application string) Option {
	return &scopeOption{scope: application}
}

// WithScope configures the application scope for the client
func WithScope(scope string) Option {
	return &scopeOption{scope: scope}
}

type namespaceOption struct {
	namespace string
}

func (o *namespaceOption) apply(options *options) {
	options.namespace = o.namespace
}

// WithNamespace configures the client's partition group namespace
func WithNamespace(namespace string) Option {
	return &namespaceOption{namespace: namespace}
}

type sessionTimeoutOption struct {
	timeout time.Duration
}

func (s *sessionTimeoutOption) apply(options *options) {
	options.sessionTimeout = s.timeout
}

// WithSessionTimeout sets the session timeout for the client
func WithSessionTimeout(timeout time.Duration) Option {
	return &sessionTimeoutOption{
		timeout: timeout,
	}
}
