Package sshmgr is a goroutine safe manager for SSH clients sharing between ssh/sftp sessions
====

It makes possible to share and reutilize existing client connections for the same host `made with the same user and port` between multiple goroutines.</br>
This is useful when yout application relies on SSH/SFTP for interacting with several hosts and not spawn multiple connections to the same hosts saving resources on both sides.

Clients are reference counted per session, and automatically closed/removed from the manager when all dependent sessions are closed.

-----------------------------------------------------------
## Usage (with the package default Manager):

```go
package main

import (
	"github.com/brunotm/sshmgr"
)

func main() {
	config := sshmgr.NewConfig("hostA.domain.com", "user", "password", "or_key_file_path")
	sshSession, err := sshmgr.Manager.GetSSHSession(config)
	if err != nil {
		panic(err)
	}
	defer sshSession.Close()

	data, err := sshSession.CombinedOutput("uptime")
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s: %s", config.NetAddr, string(data))
}
```

Written by Bruno Moura <brunotm@gmail.com>