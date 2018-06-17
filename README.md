# Go sshmgr

[![Build Status](https://travis-ci.org/brunotm/sshmgr.svg?branch=master)](https://travis-ci.org/brunotm/sshmgr) [![Go Report Card](https://goreportcard.com/badge/github.com/brunotm/sshmgr)](https://goreportcard.com/report/github.com/brunotm/sshmgr)
====

### A goroutine safe manager for SSH and SFTP client sharing.

It makes possible to share and reutilize existing clients for the same host `made with the same user,port and credentials` between multiple goroutines.</br>

Clients are reference counted, and automatically closed/removed from the manager when they have no references and the client TTL is exceeded.

-----------------------------------------------------------
## Usage:

```go
package main

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/brunotm/sshmgr"
)

func main() {

	// Creates a manager with a client ttl of 10 seconds and
	// a GC interval of 5 seconds
	manager := sshmgr.New(time.Second*10, time.Second*5)
	defer manager.Close()

	key, err := ioutil.ReadFile("/path/to/key")
	if err != nil {
		panic(err)
	}

	config := sshmgr.ClientConfig{}
	config.NetAddr = "hosta"
	config.Port = "22"
	config.User = "root"
	config.Password = ""
	config.Key = key
	config.IgnoreHostKey = true
	config.ConnDeadline = time.Minute
	config.DialTimeout = time.Second * 5

	client, err := manager.SSHClient(config)
	if err != nil {
		panic(err)
	}
	// Must close the client when done.
	defer client.Close()

	data, err := client.CombinedOutput("uptime", nil)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s: %s", config.NetAddr, string(data))
}

```

Written by Bruno Moura <brunotm@gmail.com>