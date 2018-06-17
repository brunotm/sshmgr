package sshmgr

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/brunotm/sshmgr/locker"
	"github.com/pkg/sftp"
)

var (
	errManagerClosed = errors.New("manager closed")
)

// Manager for shared ssh and sftp clients
type Manager struct {
	mtx        sync.RWMutex
	gcInterval time.Duration
	clientTTL  int64
	locker     *locker.Locker
	clients    map[string]*Client
	closeChan  chan struct{}
}

// New creates a new Manager.
// clientTTL specifies the maximum amount of time after which it was last accessed that client
// will be kept alive in the manager without open references.
// The client last access time is updated when the client is released
// gcInterval specifies the interval the manager will try to remove unused clients
func New(clientTTL, gcInterval time.Duration) (manager *Manager) {
	manager = &Manager{
		mtx:        sync.RWMutex{},
		gcInterval: gcInterval,
		clientTTL:  int64(clientTTL.Seconds()),
		locker:     locker.New(),
		clients:    map[string]*Client{},
		closeChan:  make(chan struct{}),
	}

	go manager.gc()
	return manager
}

// Close all running clients and shutdown the manager
func (m *Manager) Close() {
	close(m.closeChan)
}

func (m *Manager) getClient(id string) (client *Client) {
	m.mtx.RLock()
	client = m.clients[id]
	m.mtx.RUnlock()
	return client
}

func (m *Manager) delClient(id string) {
	m.mtx.RLock()
	delete(m.clients, id)
	m.mtx.RUnlock()
}

func (m *Manager) setClient(id string, client *Client) {
	m.mtx.RLock()
	m.clients[id] = client
	m.mtx.RUnlock()
}

// SSHClient returns an active managed client or create a new one on demand.
// Clients must be closed after usage so they can be removed when there are no references
func (m *Manager) SSHClient(config ClientConfig) (client *Client, err error) {

	select {
	case <-m.closeChan:
		return nil, errManagerClosed
	default:
	}

	id := config.id()
	m.locker.Lock(id)
	defer m.locker.Unlock(id)

	// Get a client for this config
	client = m.getClient(config.id())

	if client != nil {
		// Check if client is valid
		_, _, err = client.client.SendRequest("sshmgr", true, nil)
		if err == nil {
			client.incr()
			client.conn.SetDeadline(time.Now().Add(config.ConnDeadline))
			return client, nil
		}
		m.delClient(id)
	}

	if client, err = newClient(config); err != nil {
		return nil, err
	}

	// Add the client to the manager, increment the reference count
	// and set the current deadline
	client.incr()
	m.setClient(id, client)

	client.conn.SetDeadline(time.Now().Add(config.ConnDeadline))
	return client, nil
}

// SFTPClient creates a session from a active managed client or create a new one on demand.
// Clients must be closed after usage so they can be removed when they have no references
func (m *Manager) SFTPClient(config ClientConfig) (session *SFTPClient, err error) {

	// Get a client for this config
	client, err := m.SSHClient(config)
	if err != nil {
		return nil, err
	}

	// Create a SFTP session
	sftpClient, err := sftp.NewClient(client.client)
	if err != nil {
		return nil, err
	}

	msftp := &SFTPClient{}
	msftp.client = client
	msftp.Client = sftpClient

	return msftp, nil
}

func (m *Manager) gc() {
	ticker := time.NewTicker(m.gcInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.closeChan:
			m.collect(true)
			return

		case <-ticker.C:
			m.collect(false)
		}
	}
}

// collect unreferenced and expired clients
func (m *Manager) collect(shutdown bool) {
	now := time.Now().Unix()

	m.mtx.Lock()
	for id := range m.clients {
		m.locker.Lock(id)
		client := m.clients[id]

		if shutdown {
			delete(m.clients, id)
			client.Close()
			continue
		}

		if client.refcount() == 0 {
			if (now - atomic.LoadInt64(&client.atime)) >= m.clientTTL {
				delete(m.clients, id)
				client.Close()
			}
		}
		m.locker.Unlock(id)
	}
	m.mtx.Unlock()
}
