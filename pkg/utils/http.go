package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// RequestConfig holds the configuration for making an HTTP request
type RequestConfig struct {
	Method  string            // HTTP method (GET, POST, etc.)
	URL     string            // Request URL
	Headers map[string]string // Request headers
	Body    interface{}       // Request body (will be JSON marshaled)
}

// SendRequest sends an HTTP request based on the provided configuration
// and returns the response. The caller is responsible for closing the response body.
//
// Example:
//
//	config := RequestConfig{
//	    Method: "GET",
//	    URL:    "https://api.example.com/data",
//	    Headers: map[string]string{
//	        "Authorization": "Bearer token",
//	    },
//	}
//	resp, err := SendRequest(config)
//	if err != nil {
//	    log.Printf("Error making request: %v", err)
//	    return
//	}
//	defer resp.Body.Close()
//	// Process response...
func SendRequest(config RequestConfig) (*http.Response, error) {
	// Create request body if provided
	var body io.Reader
	if config.Body != nil {
		jsonData, err := json.Marshal(config.Body)
		if err != nil {
			return nil, fmt.Errorf("error marshaling request body: %w", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	// Create the request
	req, err := http.NewRequest(config.Method, config.URL, body)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Set headers
	for key, value := range config.Headers {
		req.Header.Set(key, value)
	}

	// If body is JSON and Content-Type is not set, set it to application/json
	if config.Body != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}

	return resp, nil
}

// SendRequestAndParseResponse sends an HTTP request and parses the response into the provided result.
// It automatically handles response body closing and error checking.
//
// Examples:
//
//  1. Simple GET request with response parsing:
//     type UserResponse struct {
//     ID    string `json:"id"`
//     Login string `json:"login"`
//     }
//     var response UserResponse
//     config := RequestConfig{
//     Method: "GET",
//     URL:    "https://api.example.com/user",
//     Headers: map[string]string{
//     "Authorization": "Bearer token",
//     },
//     }
//     err := SendRequestAndParseResponse(config, &response)
//
//  2. POST request with JSON body:
//     config := RequestConfig{
//     Method: "POST",
//     URL:    "https://api.example.com/create",
//     Headers: map[string]string{
//     "Authorization": "Bearer token",
//     },
//     Body: map[string]string{
//     "name": "example",
//     },
//     }
//     err := SendRequestAndParseResponse(config, nil)
//
//  3. Request with complex response structure:
//     type Response struct {
//     Data []struct {
//     ID   string `json:"id"`
//     Name string `json:"name"`
//     } `json:"data"`
//     }
//     var result Response
//     config := RequestConfig{
//     Method: "GET",
//     URL:    "https://api.example.com/items",
//     }
//     err := SendRequestAndParseResponse(config, &result)
func SendRequestAndParseResponse(config RequestConfig, result any) error {
	resp, err := SendRequest(config)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	// Check if response is successful
	if resp.StatusCode >= 400 {
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response if result is provided
	if result != nil {
		if err := json.Unmarshal(body, result); err != nil {
			return fmt.Errorf("error parsing response: %w", err)
		}
	}

	return nil
}

