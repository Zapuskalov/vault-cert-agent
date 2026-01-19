package internal

import (
	"log"
	"sync"
	"time"
)

func RunOnce(cfg Config) error {
	client, err := VaultClient(cfg)
	if err != nil {
		return err
	}

	ok, err := IssueCert(client, cfg)
	if ok && len(cfg.Hooks.PostIssue) > 0 {
		RunHooks(cfg.Name, cfg.Hooks.PostIssue)
	}

	return err
}

func RunDaemon() error {
	var wg sync.WaitGroup

	for _, cfg := range Cfgs {
		wg.Add(1)

		go func(cfg Config) {
			defer wg.Done()
			runWorker(cfg)
		}(cfg)
	}

	wg.Wait()
	return nil
}

func runWorker(cfg Config) {
	interval, err := time.ParseDuration(cfg.Daemon.CheckInterval)
	if err != nil || interval <= 0 {
		log.Printf("[%s] Invalid check_interval '%s', defaulting to 1m", cfg.Name, cfg.Daemon.CheckInterval)
		interval = time.Minute
	}

	for {
		if NeedsRenew(cfg) {
			log.Printf("[%s] Certificate needs renewal, issuing...", cfg.Name)
			if err := RunOnce(cfg); err != nil {
				log.Printf("[%s] Error issuing certificate: %v", cfg.Name, err)
			}
		}
		time.Sleep(interval)
	}
}
