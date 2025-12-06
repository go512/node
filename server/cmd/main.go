package main

import (
	"fmt"
	"log"
	"node/cmd/cli"
	"os"
)

var (
	Version   = ""
	BuildTime = ""
	Commit    = ""
)

func version() string {
	return fmt.Sprintf(`
version: %s 
buildTime: %s 
commit: %s
`,
		Version,
		BuildTime,
		Commit)
}

func main() {
	app := cli.NewApp()
	app.Version = version()
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
