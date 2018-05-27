package sshmgr

import (
	"fmt"
	"time"

	"golang.org/x/crypto/ssh"
)

// SSHConfig type
type SSHConfig struct {
	name        string
	NetAddr     string
	Port        string
	User        string
	Password    string
	Key         []byte
	DialTimeout time.Duration
	Deadline    time.Duration
}

// NewConfig creates a SSHConfig with the specified parameters, default port and timeout
func NewConfig(netaddr, port, user, pass string, key []byte, timeout, deadline time.Duration) SSHConfig {
	return SSHConfig{
		name:        fmt.Sprintf("%s@%s:%s:%x:%x", user, netaddr, port, pass, key),
		NetAddr:     netaddr,
		Port:        port,
		User:        user,
		Password:    pass,
		Key:         key,
		DialTimeout: timeout,
		Deadline:    deadline,
	}
}

// newSSHClientConfig creates a ssh.ClientConfig from a SSHConfig
func newSSHClientConfig(config SSHConfig) (*ssh.ClientConfig, error) {
	if config.User == "" {
		return nil, fmt.Errorf("empty username")
	}

	if config.Password == "" && len(config.Key) == 0 {
		return nil, fmt.Errorf("empty password and Key")
	}

	var auths []ssh.AuthMethod

	if config.Password != "" {
		auths = append(auths, ssh.Password(config.Password))
	}

	if len(config.Key) > 0 {
		key, err := ssh.ParsePrivateKey(config.Key)
		if err != nil {
			return nil, err
		}
		auths = append(auths, ssh.PublicKeys(key))
	}

	return &ssh.ClientConfig{
		User:            config.User,
		Auth:            auths,
		Timeout:         config.DialTimeout,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}, nil
}
