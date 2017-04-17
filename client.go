package sshmgr

import (
	"fmt"
	"sync/atomic"

	"golang.org/x/crypto/ssh"
)

type managedClient struct {
	client *ssh.Client
	refs   int32
}

func (c *managedClient) incr() int32 {
	return atomic.AddInt32(&c.refs, 1)
}

func (c *managedClient) decr() int32 {
	return atomic.AddInt32(&c.refs, -1)
}

func newManagedClient(config *SSHConfig) (*managedClient, error) {
	client, err := newSSHClient(config)
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}
	return &managedClient{client: client}, nil
}

// NewSSHClient creates a new ssh.Client from the given config
func newSSHClient(config *SSHConfig) (*ssh.Client, error) {
	sshConfig, err := newSSHClientConfig(config)
	if err != nil {
		return nil, err
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", config.NetAddr, config.Port), sshConfig)
	if err != nil {
		return nil, fmt.Errorf("Failed to dial: %s", err)
	}
	return client, nil
}
