package main

import "github.com/universal-governance/ugc/cli"

var version = "dev"

func main() {
	cli.SetVersion(version)
	cli.Execute()
}
