package sshmgr

import (
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// SSHSession type
type SSHSession struct {
	*ssh.Session
	client  *sshClient
	manager *SSHManager
}

// Close the session and notify the manager
func (s *SSHSession) Close() (err error) {
	err = s.Session.Close()
	s.manager.notifySessionClose(s.client)
	return err
}

// SFTPSession type
type SFTPSession struct {
	*sftp.Client
	client  *sshClient
	manager *SSHManager
}

// Close the session and notify the manager
func (s *SFTPSession) Close() (err error) {
	err = s.Client.Close()
	s.manager.notifySessionClose(s.client)
	return err
}
