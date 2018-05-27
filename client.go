package sshmgr

import (
	"net"
	"sync/atomic"

	"golang.org/x/crypto/ssh"
)

type sshClient struct {
	*ssh.Client
	config SSHConfig
	conn   net.Conn
	refs   int32
}

func (c *sshClient) incr() int32 {
	return atomic.AddInt32(&c.refs, 1)
}

func (c *sshClient) decr() int32 {
	return atomic.AddInt32(&c.refs, -1)
}

// newSSHClient creates a new ssh.Client from the given config
func newSSHClient(config SSHConfig) (*sshClient, error) {
	addr := config.NetAddr + ":" + config.Port
	sshConfig, err := newSSHClientConfig(config)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialTimeout("tcp", addr, config.DialTimeout)
	if err != nil {
		return nil, err
	}
	c, chans, reqs, err := ssh.NewClientConn(conn, addr, sshConfig)
	if err != nil {
		return nil, err
	}

	sshClient := &sshClient{}
	sshClient.config = config
	sshClient.conn = conn
	sshClient.Client = ssh.NewClient(c, chans, reqs)
	return sshClient, nil
}
