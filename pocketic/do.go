package pocketic

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func (pic PocketIC) do(method, url string, input, output any) error {
	start := time.Now()
	for {
		if pic.timeout < time.Since(start) {
			return fmt.Errorf("timeout exceeded")
		}

		pic.logger.Printf("[POCKETIC] %s %s %+v", method, url, input)
		req, err := newRequest(method, url, input)
		if err != nil {
			return err
		}
		resp, err := pic.client.Do(req)
		if err != nil {
			return err
		}
		switch resp.StatusCode {
		case http.StatusOK, http.StatusCreated:
			if resp.Body == nil || output == nil {
				// No need to decode the response body.
				return nil
			}
			if err := json.NewDecoder(resp.Body).Decode(output); err != nil {
				return fmt.Errorf("failed to decode response body: %w", err)
			}
			return nil
		case http.StatusAccepted:
			var response startedOrBusyResponse
			if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
				return fmt.Errorf("failed to decode accepted response body: %w", err)
			}
			pic.logger.Printf("[POCKETIC] Accepted: %s %s", response.StateLabel, response.OpID)
			if method == http.MethodGet {
				continue
			}
			for {
				pic.logger.Printf("[POCKETIC] Waiting for %s %s", response.StateLabel, response.OpID)
				if pic.timeout < time.Since(start) {
					return fmt.Errorf("timeout exceeded")
				}

				req, err := newRequest(
					http.MethodGet,
					fmt.Sprintf(
						"%s/read_graph/%s/%s",
						pic.server.URL(),
						response.StateLabel,
						response.OpID,
					),
					nil,
				)
				if err != nil {
					return err
				}
				resp, err := pic.client.Do(req)
				if err != nil {
					return err
				}
				switch resp.StatusCode {
				case http.StatusOK, http.StatusCreated:
					if resp.Body == nil || output == nil {
						// No need to decode the response body.
						return nil
					}
					if err := json.NewDecoder(resp.Body).Decode(output); err != nil {
						return fmt.Errorf("failed to decode response body: %w", err)
					}
					return nil
				case http.StatusAccepted, http.StatusConflict:
				default:
					var errResp ErrorMessage
					if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
						return fmt.Errorf("failed to decode accepted/conflict response body: %w", err)
					}
					return errResp
				}
			}
		case http.StatusConflict:
			var response startedOrBusyResponse
			if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
				return fmt.Errorf("failed to decode conflict response body: %w", err)
			}
			pic.logger.Printf("[POCKETIC] Conflict: %s %s", response.StateLabel, response.OpID)
			time.Sleep(pic.delay) // Retry after a short delay.
			continue
		default:
			var errResp ErrorMessage
			if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
				return fmt.Errorf("failed to decode error response body: %w", err)
			}
			return errResp
		}
	}
}

type startedOrBusyResponse struct {
	StateLabel string `json:"state_label"`
	OpID       string `json:"op_id"`
}
