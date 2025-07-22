package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func ReqeuestBase[T any, U any](urlAppendix, method string, apiErrorCh chan error, apiResponseCh chan T, body U) {
	// TODO: steup envs
	// urlPrefix := os.Getenv("API_URL")
	urlPrefix := "localhost/api/v1/internal/identity"
	fullUrl := urlPrefix + urlAppendix

	var bodyReader io.Reader
	if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			apiErrorCh <- fmt.Errorf("Failed to marshal request body: %s", err)
			return
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}
	httpClient := &http.Client{}
	req, err := http.NewRequest(method, fullUrl, bodyReader)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "internal_token_admin_123")

	resp, err := httpClient.Do(req)
	if err != nil {
		apiErrorCh <- err
		return
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		apiErrorCh <- err
		return
	}

	if resp.StatusCode < 200 || resp.StatusCode > 300 {
		apiErrorCh <- fmt.Errorf("HTTP error: %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
		return
	}

	// if empty response
	if len(responseBody) == 0 {
		var nothing T
		apiResponseCh <- nothing
		return
	}

	var apiResponse T

	if err := json.Unmarshal(responseBody, &apiResponse); err != nil {
		apiErrorCh <- err
		return
	}

	apiResponseCh <- apiResponse
}
