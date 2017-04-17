package sshmgr

import (
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// SSHSession type
type SSHSession struct {
	*ssh.Session
	clientName string
	manager    *SSHManager
}

// Close the session and notify the manager
func (s *SSHSession) Close() error {
	err := s.Session.Close()
	s.manager.notifySessionClose(s.clientName)
	return err
}

// SFTPSession type
type SFTPSession struct {
	*sftp.Client
	clientName string
	manager    *SSHManager
}

// Close the session and notify the manager
func (s *SFTPSession) Close() error {
	err := s.Client.Close()
	s.manager.notifySessionClose(s.clientName)
	return err
}
