package test

import (
	"fmt"
	"os"
	"os/exec"
)

func startInstance() (*exec.Cmd, error) {
	fmt.Println("start instance")
	install := exec.Command("go", "install")
	install.Dir = "."

	if out, err := install.CombinedOutput(); err != nil {
		return nil, err
	} else if len(out) != 0 {
		// Successful go install is silent
		return nil, fmt.Errorf(string(out))
	}

	app := exec.Command("cumulus", "run")
	app.Stdout = os.Stdout
	app.Stderr = os.Stderr

	fmt.Println("about to start")
	err := app.Start()
	fmt.Println("started")
	return app, err
}
