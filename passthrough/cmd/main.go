package main

import (
	"github.com/docker/docker-credential-helpers/credentials"
	"github.com/docker/docker-credential-helpers/passthrough"
)

func main() {
	credentials.Serve(passthrough.Passthrough{})
}
