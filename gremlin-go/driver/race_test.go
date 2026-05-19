/*
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.
*/

package gremlingo

// Reproducers for TINKERPOP-2845 race conditions in the Go GLV driver.
//
// The ticket documents three unsynchronized field accesses surfaced by
// `go test -race` against a server that closes a connection mid-flight:
//
//   Race 1A: channelResultSet.setError writes .err (resultSet.go) while a
//            caller blocked in channelResultSet.One reads .err.
//   Race 1B: connection.errorCallback writes .state (connection.go) while a
//            caller in Client.Close -> connection.close reads/writes .state.
//   Race 2 : connection.errorCallback writes .state while a subsequent caller
//            Submit -> loadBalancingPool.getLeastUsedConnection reads .state.
//
// These tests trigger the same code paths using a local mock WebSocket
// server that accepts the upgrade, reads one message, then drops the TCP
// connection. That produces a websocket close (no status) on the client's
// read loop -- the same failure mode the JIRA's CLOSE_CONNECTION_REQUEST_ID
// produces in gremlin-socket-server. No external dependencies required.

import (
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

// newCloseOnReceiveServer starts a local WS server that upgrades the
// connection, reads exactly one client message, then closes the underlying
// TCP socket without sending a WebSocket close frame. The client's read loop
// observes this as a close-with-no-status error and proceeds into
// readErrorHandler.
func newCloseOnReceiveServer(t *testing.T) (url string, shutdown func()) {
	t.Helper()
	upgrader := websocket.Upgrader{
		CheckOrigin: func(*http.Request) bool { return true },
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		_, _, _ = c.ReadMessage()
		if tcp, ok := c.UnderlyingConn().(*net.TCPConn); ok {
			_ = tcp.SetLinger(0)
		}
		_ = c.Close()
	})
	srv := httptest.NewServer(handler)
	url = "ws" + strings.TrimPrefix(srv.URL, "http") + "/gremlin"
	return url, srv.Close
}

// TestRace_TINKERPOP_2845_ResultSetClosedByServer reproduces Race 1
// (1A on channelResultSet.err and 1B on connection.state) from the ticket.
// Run with `go test -race -run TestRace_TINKERPOP_2845_ResultSetClosedByServer`.
func TestRace_TINKERPOP_2845_ResultSetClosedByServer(t *testing.T) {
	url, shutdown := newCloseOnReceiveServer(t)
	defer shutdown()

	client, err := NewClient(url)
	assert.Nil(t, err)
	assert.NotNil(t, client)

	rs, err := client.Submit("1")
	assert.Nil(t, err)
	assert.NotNil(t, rs)

	// Caller-side read of channelResultSet.err and the result channel; this
	// races with the read loop's readErrorHandler -> setError write when the
	// server-side close arrives.
	_, _, _ = rs.One()

	// Client.Close walks the pool and calls connection.close, which reads
	// connection.state; this races with errorCallback's write to the same
	// field on the same connection.
	client.Close()
}

// TestRace_TINKERPOP_2845_PoolStateConcurrentAccess reproduces Race 2 from
// the ticket. The first Submit triggers a server-side close, after which the
// read loop's errorCallback writes connection.state = closedDueToError. A
// subsequent Submit goes through loadBalancingPool.getLeastUsedConnection,
// which reads connection.state for every pool member without any
// synchronization against the errorCallback write.
//
// The first ResultSet's read is wrapped in channelResultSet.channelMutex to
// match the JIRA's Race 2 reproduction, suppressing Race 1A so the pool
// race surfaces in isolation under -race.
func TestRace_TINKERPOP_2845_PoolStateConcurrentAccess(t *testing.T) {
	url, shutdown := newCloseOnReceiveServer(t)
	defer shutdown()

	client, err := NewClient(url)
	assert.Nil(t, err)
	assert.NotNil(t, client)
	defer client.Close()

	rs1, err := client.Submit("1")
	assert.Nil(t, err)
	assert.NotNil(t, rs1)

	// Give the read loop time to hit the close error and run the error
	// callback. The race detector still fires even if the writes have
	// finished, because the happens-before edges aren't established.
	time.Sleep(200 * time.Millisecond)

	// JIRA's Race 1A workaround: lock channelMutex around the One() read so
	// the report we get is the pool-level race, not the resultSet race.
	chRS := rs1.(*channelResultSet)
	chRS.channelMutex.Lock()
	_, _, _ = rs1.One()
	chRS.channelMutex.Unlock()

	// Second Submit walks loadBalancingPool.getLeastUsedConnection, which
	// reads connection.state for the now-broken first connection -- racing
	// with the still-completing errorCallback write.
	rs2, _ := client.Submit("1")
	if rs2 != nil {
		_, _, _ = rs2.One()
	}
}
