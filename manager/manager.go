package manager

/*
Package manager wraps sshmgr.Manager with a client ttl of 10 seconds
and a gc interval of 5 seconds
*/

import (
	"time"

	"github.com/brunotm/sshmgr"
)

const (
	clientTTL  time.Duration = time.Second * 10
	gcInterval time.Duration = time.Second * 5
)

func init() {
	manager = sshmgr.New(clientTTL, gcInterval)
}

// manager is the package default ssh manager
var manager *sshmgr.Manager

// SSHClient creates or return a existing client
func SSHClient(config sshmgr.ClientConfig) (client *sshmgr.Client, err error) {
	return manager.SSHClient(config)
}

// SFTPClient creates or return a existing client
func SFTPClient(config sshmgr.ClientConfig) (client *sshmgr.SFTPClient, err error) {
	return manager.SFTPClient(config)
}
