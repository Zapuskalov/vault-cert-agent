package internal

import (
	"log"
	"os"
	"os/exec"
)

func RunHooks(cfgName string, cmds []string) {
	for _, c := range cmds {
		log.Printf("[%s] Running hook: %s", cfgName, c)

		cmd := exec.Command("sh", "-c", c)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			log.Printf("[%s] Hook failed: %s, error: %v", cfgName, c, err)
		}
	}
}
