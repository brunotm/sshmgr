package sshmgr

import (
	"errors"
	"io"
	"net"
	"sync/atomic"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

var (
	errClientClosed = errors.New("client already closed")
)

// Client is a shared managed ssh client
type Client struct {
	client *ssh.Client
	conn   net.Conn
	atime  int64
	refs   int32
}

// Close notifies the manager that this client can be removed
// if there is no more references to it
func (c *Client) Close() (err error) {
	if c.refcount() == 0 {
		return errClientClosed
	}

	c.updateAtime()
	c.decr()
	return nil
}

// CombinedOutput runs cmd on the remote host and returns its combined
// standard output and standard error.
func (c *Client) CombinedOutput(cmd string, envs map[string]string) (data []byte, err error) {
	s, err := c.client.NewSession()
	if err != nil {
		return nil, err
	}
	defer s.Close()

	for name := range envs {
		if err = s.Setenv(name, envs[name]); err != nil {
			return nil, err
		}
	}

	return s.CombinedOutput(cmd)
}

type readCloser struct {
	io.Reader
	s *ssh.Session
}

func (r readCloser) Close() (err error) {
	return r.s.Close()
}

// CombinedReader is like CombinedOutput but returns a io.Reader combining both stderr and stdout.
func (c *Client) CombinedReader(cmd string, envs map[string]string) (reader io.ReadCloser, err error) {
	s, err := c.client.NewSession()
	if err != nil {
		return nil, err
	}

	for name := range envs {
		if err = s.Setenv(name, envs[name]); err != nil {
			return nil, err
		}
	}

	stdout, err := s.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stderr, err := s.StderrPipe()
	if err != nil {
		return nil, err
	}

	if err = s.Run(cmd); err != nil {
		return nil, err
	}

	return readCloser{Reader: io.MultiReader(stdout, stderr), s: s}, nil
}

func (c *Client) incr() (r int32) {
	return atomic.AddInt32(&c.refs, 1)
}

func (c *Client) decr() (r int32) {
	return atomic.AddInt32(&c.refs, -1)
}

func (c *Client) updateAtime() {
	atomic.StoreInt64(&c.atime, time.Now().Unix())
}

func (c *Client) refcount() (r int32) {
	return atomic.LoadInt32(&c.refs)
}

// SFTPClient type
type SFTPClient struct {
	*sftp.Client
	client *Client
}

// Close the session and notify the manager
func (s *SFTPClient) Close() (err error) {
	return s.client.Close()
}

// Lock overrides the original client channel lock (does nothing)
func (s *SFTPClient) Lock() {
}

// Unlock overrides the original client channel lock (does nothing)
func (s *SFTPClient) Unlock() {
}

// newClient creates a new ssh.Client from the given config
func newClient(config ClientConfig) (client *Client, err error) {
	if config.Port == "" {
		config.Port = "22"
	}
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

	client = &Client{}
	client.conn = conn
	client.client = ssh.NewClient(c, chans, reqs)
	return client, nil
}
