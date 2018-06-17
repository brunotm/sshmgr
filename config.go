package sshmgr

import (
	"fmt"
	"strconv"
	"time"

	"github.com/cespare/xxhash"
	"golang.org/x/crypto/ssh"
)

// ClientConfig parameters for getting ssh or sftp clients from the manager
type ClientConfig struct {
	// NetAddr specifies the host ip or name
	NetAddr string

	// Port specifies the host port to connect to.
	// Defaults to 22 if empty
	Port string

	// User to authenticate as
	User string

	// Password to authenticate with
	Password string

	// Key to authenticate with
	Key []byte

	// IgnoreHostKey specifies whether to use InsecureIgnoreHostKey as the
	// HostKeyCallback to disable host key verification
	IgnoreHostKey bool

	// Deadline to be used in the underlying net.Conn.
	// Specified as a time.Duration so its set as the sum of the current time
	// and the ConnDeadline when the connection is established or to upgrade the
	// deadline when reusing a client
	ConnDeadline time.Duration

	// DialTimeout
	DialTimeout time.Duration
}

// id returns this Config ID
func (c ClientConfig) id() (id string) {
	return strconv.FormatUint(xxhash.Sum64String(
		fmt.Sprint(c.User, c.NetAddr, c.Port, c.Password, c.Key)), 10)
}

// newSSHClientConfig creates a ssh.ClientConfig from a ClientConfig
func newSSHClientConfig(config ClientConfig) (c *ssh.ClientConfig, err error) {
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

	c = &ssh.ClientConfig{}
	c.SetDefaults()
	c.User = config.User
	c.Auth = auths
	c.Timeout = config.DialTimeout

	if config.IgnoreHostKey {
		c.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	}

	// Reverse order of available ciphers to prevent early failure in negotiation
	// with older ssh server versions
	for i := len(c.Ciphers)/2 - 1; i >= 0; i-- {
		opp := len(c.Ciphers) - 1 - i
		c.Ciphers[i], c.Ciphers[opp] = c.Ciphers[opp], c.Ciphers[i]
	}

	return c, nil
}
