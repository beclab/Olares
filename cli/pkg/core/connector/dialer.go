/*
 Copyright 2021 The KubeSphere Authors.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package connector

import (
	"sync"
	"time"

	"bytetrade.io/web3os/installer/pkg/core/logger"
)

type Dialer struct {
	lock        sync.Mutex
	connections map[string]Connection
}

func NewDialer() *Dialer {
	return &Dialer{
		connections: make(map[string]Connection),
	}
}

func (d *Dialer) Connect(host Host) (Connection, error) {
	var err error

	d.lock.Lock()
	defer d.lock.Unlock()
	conn, ok := d.connections[host.GetName()]
	if !ok {
		opts := Cfg{
			Username:   host.GetUser(),
			Port:       host.GetPort(),
			Address:    host.GetAddress(),
			Password:   host.GetPassword(),
			PrivateKey: host.GetPrivateKey(),
			KeyFile:    host.GetPrivateKeyPath(),
			Timeout:    time.Duration(host.GetTimeout()) * time.Second,
		}
		conn, err = NewConnection(opts)
		if err != nil {
			return nil, err
		}
		d.connections[host.GetName()] = conn
	}

	return conn, nil
}

func (d *Dialer) Close(host Host) {
	conn, ok := d.connections[host.GetName()]
	if !ok {
		return
	}

	conn.Close()
	logger.Debugf("close connection %s", host.GetName())

	c := conn.(*connection)
	d.forgetConnection(c)
}

func (d *Dialer) forgetConnection(conn *connection) {
	d.lock.Lock()
	defer d.lock.Unlock()

	for k := range d.connections {
		if d.connections[k] == conn {
			delete(d.connections, k)
		}
	}
}
