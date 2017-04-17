/*
Package sshmgr is a goroutine safe manager for SSH clients sharing between ssh/sftp sessions

It makes possible to share and reutilize existing client connections
for the same host `made with the same user and port` between multiple goroutines.

This is useful when yout application relies on SSH/SFTP for interacting with several
hosts and not spawn multiple connections to the same hosts saving resources on both sides.

Clients are reference counted per session, and automatically closed/removed from the manager when all dependent sessions are closed.

	Usage:

		package main

		import (
			"github.com/brunotm/sshmgr"
		)

		func main() {
			config := sshmgr.NewConfig("hostA.domain.com", "user", "password", "or_key_file_path")
			sshSession, err := sshmgr.Manager.GetSSHSession(config)
			if err != nil {
				panic(err)
			}
			defer sshSession.Close()

			data, err := sshSession.CombinedOutput("uptime")
			if err != nil {
				panic(err)
			}

			fmt.Printf("%s: %s", config.NetAddr, string(data))
		}
*/
package sshmgr

import (
	"fmt"
	"sync"

	"github.com/brunotm/sshmgr/locker"
	"github.com/pkg/sftp"
)

// Manager is the package default ssh manager
var Manager = NewManager()

// SSHManager manage ssh clients and sessions.
// Clients are reference counted per session and removed from manager when the refcount reaches 0
type SSHManager struct {
	mtx     *sync.Mutex
	locker  *locker.Locker
	clients map[string]*managedClient
}

// GetSSHSession creates a session from a active managed client or create a new one on demand
func (m *SSHManager) GetSSHSession(config *SSHConfig) (*SSHSession, error) {
	clientName := fmt.Sprintf("%s@%s:%s", config.User, config.NetAddr, config.Port)
	m.locker.Lock(clientName)
	defer m.locker.Unlock(clientName)

	// Found a active client, try to create a new session from it
	if client := m.clients[clientName]; client != nil {
		session, err := client.client.NewSession()
		if err != nil {
			return nil, err
		}
		client.incr()
		return &SSHSession{clientName: clientName, manager: m, Session: session}, nil
	}

	// Create a new client and session
	client, err := newManagedClient(config)
	if err != nil {
		return nil, err
	}

	session, err := client.client.NewSession()
	if err != nil {
		return nil, err
	}
	client.incr()

	m.mtx.Lock()
	m.clients[clientName] = client
	m.mtx.Unlock()

	return &SSHSession{clientName: clientName, manager: m, Session: session}, nil
}

// GetSFTPSession creates a session from a active managed client or create a new one on demand
func (m *SSHManager) GetSFTPSession(config *SSHConfig) (*SFTPSession, error) {
	clientName := fmt.Sprintf("%s@%s:%s", config.User, config.NetAddr, config.Port)
	m.locker.Lock(clientName)
	defer m.locker.Unlock(clientName)

	// Found a active client, try to create a new session from it
	if client := m.clients[clientName]; client != nil {
		session, err := sftp.NewClient(client.client)
		if err != nil {
			return nil, err
		}
		client.incr()
		return &SFTPSession{clientName: clientName, manager: m, Client: session}, nil
	}

	// Create a new client and session
	client, err := newManagedClient(config)
	if err != nil {
		return nil, err
	}

	session, err := sftp.NewClient(client.client)
	if err != nil {
		return nil, err
	}
	client.incr()

	m.mtx.Lock()
	m.clients[clientName] = client
	m.mtx.Unlock()

	return &SFTPSession{clientName: clientName, manager: m, Client: session}, nil
}

func (m *SSHManager) notifySessionClose(clientName string) {
	m.locker.Lock(clientName)
	defer m.locker.Unlock(clientName)

	client, ok := m.clients[clientName]
	if !ok {
		// We should never get here
		panic(fmt.Sprintf("Client not found: %s", clientName))
	}

	if client.decr() == 0 {
		defer client.client.Close()
		m.mtx.Lock()
		delete(m.clients, clientName)
		m.mtx.Unlock()

	}
}

// NewManager creates a new SSHManager
func NewManager() *SSHManager {
	return &SSHManager{&sync.Mutex{}, locker.New(), map[string]*managedClient{}}
}
