package kulascope

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type sendJob struct {
	cfg     Config
	payload CreateLogRequest
}

var sendQueue chan sendJob

func startSenderWorkers(num int) {
	sendQueue = make(chan sendJob, 50_000)
	for i := 0; i < num; i++ {
		go senderWorker()
	}
}

func senderWorker() {
	for job := range sendQueue {
		sendWithRetry(job)
	}
}

func sendWithRetry(job sendJob) {
	const maxRetries = 5

	for attempt := 0; attempt <= maxRetries; attempt++ {
		err := trySend(job.cfg, job.payload)
		if err == nil {
			return
		}

		if attempt == maxRetries {
			baseLogger.Error().Err(err).Msg("failed to send log after retries")
			return
		}

		wait := time.Duration(1<<attempt) * time.Second
		time.Sleep(wait)
	}
}

func trySend(cfg Config, payload CreateLogRequest) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	var url string
	if cfg.Environment == Staging {
		url = "https://api.staging.kulawise.com/kulascope/logs"
	} else {
		url = "https://api.kulawise.com/kulascope/logs"
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", cfg.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return fmt.Errorf("server error: %d", resp.StatusCode)
	}

	return nil
}
