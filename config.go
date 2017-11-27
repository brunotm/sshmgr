package sshmgr

import (
	"fmt"
	"time"

	"golang.org/x/crypto/ssh"
)

const (
	defaultPort           = "22"
	defaultTimeoutSeconds = 5
)

// SSHConfig type
type SSHConfig struct {
	NetAddr        string `json:"netaddr,omitempty"`
	Port           string `json:"port,omitempty"`
	User           string `json:"ssh_user,omitempty"`
	Password       string `json:"ssh_password,omitempty"`
	Key            []byte `json:"ssh_key,omitempty"`
	TimeoutSeconds int    `json:"timeout_seconds,omitempty"`
}

// NewConfig creates a SSHConfig with the specified parameters, default port and timeout
func NewConfig(netaddr, user, pass string, key []byte) *SSHConfig {
	return &SSHConfig{netaddr, defaultPort, user, pass, key, defaultTimeoutSeconds}
}

// newSSHClientConfig creates a ssh.ClientConfig from a *SSHConfig
func newSSHClientConfig(config *SSHConfig) (*ssh.ClientConfig, error) {
	if config.User == "" {
		return nil, fmt.Errorf("Empty username")
	}

	if config.Password == "" && len(config.Key) == 0 {
		return nil, fmt.Errorf("Empty password and Key")
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
		Timeout:         time.Duration(config.TimeoutSeconds) * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}, nil
}
