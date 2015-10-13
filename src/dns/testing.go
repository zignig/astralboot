// Local backend grabbed from etcd backend

// Copyright (c) 2014 The SkyDNS Authors. All rights reserved.
// Use of this source code is governed by The MIT License (MIT) that can be
// found in the LICENSE file.

// Package etcd provides the default SkyDNS server Backend implementation,
// which looks up records stored under the `/skydns` key in etcd when queried.
package main

import (
	"errors"
	"fmt"

	"github.com/skynetservices/skydns/msg"
	"strings"
)

// Config represents configuration for the Etcd backend - these values
// should be taken directly from server.Config
type Config struct {
	Ttl      uint32
	Priority uint16
}

type Backend struct {
	config *Config
}

// NewBackend returns a new Backend for SkyDNS, backed by etcd.
func NewBackend() *Backend {
	return &Backend{}
}

func (g *Backend) Records(name string, exact bool) ([]msg.Service, error) {
	path, star := msg.PathWithWildcard(name)
	bits := strings.Split(path, "/")
	fmt.Println(bits, path, star)
	srv := msg.Service{}
	srv.Host = "192.168.5.1"
	l := make([]msg.Service, 0)
	l = append(l, srv)
	return l, nil // errors.New("FAIL")
}

func (g *Backend) ReverseRecord(name string) (*msg.Service, error) {
	path, star := msg.PathWithWildcard(name)
	fmt.Println(path, star)
	return nil, errors.New("FAIL REVERSE")
}
