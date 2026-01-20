package internal

import (
	"log"
	"os"
	"os/exec"
	"time"
)

func RunHooks(cfgName string, hooks []Hook) {
	for _, h := range hooks {
		go func(h Hook) {
			waitDuration := waitUntilWindow(h)
			if waitDuration > 0 {
				log.Printf("[%s] Hook '%s' will run in %s (outside time window)", cfgName, h.Cmd, waitDuration)
				time.Sleep(waitDuration)
			}

			log.Printf("[%s] Running hook: %s", cfgName, h.Cmd)
			cmd := exec.Command("sh", "-c", h.Cmd)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			if err := cmd.Run(); err != nil {
				log.Printf("[%s] Hook failed: %s, error: %v", cfgName, h.Cmd, err)
			}
		}(h)
	}
}

func waitUntilWindow(h Hook) time.Duration {
	now := time.Now()
	start := parseHookTime(h.RunAfter, now)
	end := parseHookTime(h.RunBefore, now)

	if !start.IsZero() && now.Before(start) {
		return start.Sub(now)
	}

	if !end.IsZero() && now.After(end) {
		return start.Add(24 * time.Hour).Sub(now)
	}

	return 0
}

func parseHookTime(s string, ref time.Time) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, err := time.Parse("15:04", s)
	if err != nil {
		log.Printf("Invalid hook time '%s': %v", s, err)
		return time.Time{}
	}
	return time.Date(ref.Year(), ref.Month(), ref.Day(), t.Hour(), t.Minute(), 0, 0, ref.Location())
}
