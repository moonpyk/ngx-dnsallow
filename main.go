package main

import (
	"os"

	"moonpyk.net/ngx-dnsallow/cmd"
)

func main() {
	os.Exit(cmd.Execute())
}
