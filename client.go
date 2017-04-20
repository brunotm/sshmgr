package sshmgr

import (
	"fmt"
	"sync/atomic"

	"golang.org/x/crypto/ssh"
)

type sshClient struct {
	*ssh.Client
	refs int32
}

func (c *sshClient) incr() int32 {
	return atomic.AddInt32(&c.refs, 1)
}

func (c *sshClient) decr() int32 {
	return atomic.AddInt32(&c.refs, -1)
}

// NewSSHClient creates a new ssh.Client from the given config
func newSSHClient(config *SSHConfig) (*sshClient, error) {
	sshConfig, err := newSSHClientConfig(config)
	if err != nil {
		return nil, err
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", config.NetAddr, config.Port), sshConfig)
	if err != nil {
		return nil, err
	}
	return &sshClient{Client: client}, nil
}
