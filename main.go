package main

import "cgpd/cmd"

var Version = "dev"

func main() {
	cmd.SetVersion(Version)
	cmd.Execute()
}
