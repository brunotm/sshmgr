package sshmgr

/*
Package sshmgr is a goroutine safe manager for SSH clients sharing between ssh/sftp sessions

It makes possible to share and reutilize existing client connections
for the same host `made with the same user and port` between multiple sessions and goroutines.

Clients are reference counted per session, and automatically closed/removed from the manager when all dependent sessions are closed.

	Usage:

		package main

		import (
			"io/ioutil"
			"github.com/brunotm/sshmgr"
		)

		func main() {
			key, err := ioutil.ReadFile("path to key file")
			if err != nil {
				panic(err)
			}

			config := sshmgr.NewConfig("hostA.domain.com", "port", "user", "password", key)
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

import (
	"fmt"
	"sync"
	"time"

	"github.com/brunotm/sshmgr/locker"
	"github.com/pkg/sftp"
)

func init() {
	Manager = NewManager()
}

// Manager is the package default ssh manager
var Manager *SSHManager

// SSHManager manage ssh clients and sessions.
// Clients are reference counted per session and removed from manager when the refcount reaches 0
type SSHManager struct {
	mtx     sync.RWMutex
	locker  *locker.Locker
	clients map[string]*sshClient
}

// getClient an existing client from manager client map
func (m *SSHManager) getClient(name string) (client *sshClient) {
	m.mtx.RLock()
	client = m.clients[name]
	m.mtx.RUnlock()
	return client
}

// addClient client to the manager
func (m *SSHManager) addClient(client *sshClient) {
	m.mtx.Lock()
	m.clients[client.config.name] = client
	m.mtx.Unlock()
}

// delClient an existing client from manager
func (m *SSHManager) delClient(name string) {
	m.mtx.Lock()
	delete(m.clients, name)
	m.mtx.Unlock()
}

// GetSSHSession creates a session from an active managed client or create a new one on demand
func (m *SSHManager) GetSSHSession(config SSHConfig) (session *SSHSession, err error) {
	m.locker.Lock(config.name)
	defer m.locker.Unlock(config.name)

	// Get a client for this config
	var client *sshClient
	client = m.getClient(config.name)
	if client == nil {
		if client, err = newSSHClient(config); err != nil {
			return nil, err
		}
	}

	// Create a SSH session
	sess, err := client.NewSession()
	if err != nil {
		return nil, err
	}

	// Add the client to the manager
	// increment the reference count
	// and set the current deadline
	m.addClient(client)
	client.incr()
	client.conn.SetDeadline(time.Now().Add(config.Deadline))
	return &SSHSession{manager: m, Session: sess, client: client}, nil
}

// GetSFTPSession creates a session from a active managed client or create a new one on demand
func (m *SSHManager) GetSFTPSession(config SSHConfig) (session *SFTPSession, err error) {
	m.locker.Lock(config.name)
	defer m.locker.Unlock(config.name)

	// Get a client for this config
	var client *sshClient
	client = m.getClient(config.name)
	if client == nil {
		if client, err = newSSHClient(config); err != nil {
			return nil, err
		}
	}

	// Create a SFTP session
	sftpClient, err := sftp.NewClient(client.Client)
	if err != nil {
		return nil, err
	}

	// Add the client to the manager
	// increment the reference count
	// and set the current deadline
	m.addClient(client)
	client.incr()
	client.conn.SetDeadline(time.Now().Add(config.Deadline))
	return &SFTPSession{manager: m, Client: sftpClient, client: client}, nil
}

// notifySessionClose notifies the manager about the closing of a session
func (m *SSHManager) notifySessionClose(client *sshClient) {
	m.locker.Lock(client.config.name)
	defer m.locker.Unlock(client.config.name)

	m.mtx.RLock()
	managedClient := m.clients[client.config.name]
	m.mtx.RUnlock()

	if managedClient == nil || client != managedClient {
		// We should never get here
		panic(fmt.Sprintf("client not found in manager after close: %s", client.config.name))
	}

	if client.decr() == 0 {
		defer client.Close()
		m.mtx.Lock()
		delete(m.clients, client.config.name)
		m.mtx.Unlock()

	}
}

// NewManager creates a new SSHManager
func NewManager() *SSHManager {
	return &SSHManager{sync.RWMutex{}, locker.New(), map[string]*sshClient{}}
}
