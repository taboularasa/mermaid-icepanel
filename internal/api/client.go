// Package api provides client functionality for the IcePanel API.
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"mermaid-icepanel/internal/config"
)

// HTTPClient provides an interface for making HTTP requests.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// DefaultHTTPClient uses the standard http client.
type DefaultHTTPClient struct {
	Client *http.Client
}

// Do executes HTTP requests using the standard http client.
func (c *DefaultHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return c.Client.Do(req)
}

// Object represents an IcePanel object.
type Object struct {
	Handle string                 `json:"handleId"`
	Name   string                 `json:"name"`
	Desc   string                 `json:"description,omitempty"`
	Type   string                 `json:"type"` // actor, system, store, group
	Props  map[string]interface{} `json:"properties,omitempty"`
}

// Connection represents an IcePanel connection.
type Connection struct {
	Handle string `json:"handleId"`
	From   string `json:"fromId"`
	To     string `json:"toId"`
	Label  string `json:"name"`
}

// Diagram represents an IcePanel diagram.
type Diagram struct {
	Name        string        `json:"name"`
	Type        string        `json:"type"` // app-diagram
	Objects     []*Object     `json:"objects"`
	Connections []*Connection `json:"connections"`
}

// IcePanelClient handles communication with the IcePanel API.
type IcePanelClient struct {
	httpClient HTTPClient
	Token      string // Export Token field
	baseURL    string
}

// NewIcePanelClient creates a new client for IcePanel API.
func NewIcePanelClient(config *config.Config, client HTTPClient, token string) *IcePanelClient {
	// Use token from parameter or fall back to config default
	apiToken := token
	if apiToken == "" {
		apiToken = config.DefaultToken
	}

	return &IcePanelClient{
		httpClient: client,
		Token:      apiToken,
		baseURL:    config.APIBaseURL,
	}
}

func (c *IcePanelClient) call(ctx context.Context, method, url string, body []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return c.httpClient.Do(req)
}

func (c *IcePanelClient) listIDs(ctx context.Context, lc, ver, path string) ([]string, error) {
	url := fmt.Sprintf("%s/landscapes/%s/versions/%s/%s?per=1000", c.baseURL, lc, ver, path)
	resp, err := c.call(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			log.Printf("Error closing response body: %v", cerr)
		}
	}()
	var out struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	ids := make([]string, len(out.Data))
	for i, d := range out.Data {
		ids[i] = d.ID
	}
	return ids, nil
}

func (c *IcePanelClient) delAll(ctx context.Context, lc, ver, path string, ids []string) error {
	for _, id := range ids {
		url := fmt.Sprintf("%s/landscapes/%s/versions/%s/%s/%s", c.baseURL, lc, ver, path, id)
		resp, err := c.call(ctx, "DELETE", url, nil)
		if err != nil {
			return err
		}
		if cerr := resp.Body.Close(); cerr != nil {
			log.Printf("Error closing response body: %v", cerr)
		}
		if resp.StatusCode >= 300 {
			return fmt.Errorf("delete %s: status %s", id, resp.Status)
		}
	}
	return nil
}

// WipeVersion deletes all content in an IcePanel version.
func (c *IcePanelClient) WipeVersion(ctx context.Context, lc, ver string) error {
	// groups
	groups, err := c.listIDs(ctx, lc, ver, "diagram-groups")
	if err != nil {
		return err
	}
	if err := c.delAll(ctx, lc, ver, "diagram-groups", groups); err != nil {
		return err
	}
	// diagrams
	diags, err := c.listIDs(ctx, lc, ver, "diagrams")
	if err != nil {
		return err
	}
	if err := c.delAll(ctx, lc, ver, "diagrams", diags); err != nil {
		return err
	}
	// objects
	objs, err := c.listIDs(ctx, lc, ver, "model/objects")
	if err != nil {
		return err
	}
	if err := c.delAll(ctx, lc, ver, "model/objects", objs); err != nil {
		return err
	}
	// connections
	conns, err := c.listIDs(ctx, lc, ver, "model/connections")
	if err != nil {
		return err
	}
	return c.delAll(ctx, lc, ver, "model/connections", conns)
}

// PostDiagram uploads a diagram to IcePanel.
func (c *IcePanelClient) PostDiagram(ctx context.Context, lc, ver string, d *Diagram, verbose bool) error {
	b, err := json.Marshal(d)
	if err != nil {
		return fmt.Errorf("failed to marshal diagram: %w", err)
	}
	if verbose {
		log.Printf("POST payload: %s", string(b))
	}
	url := fmt.Sprintf("%s/landscapes/%s/versions/%s/diagrams", c.baseURL, lc, ver)
	resp, err := c.call(ctx, "POST", url, b)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			log.Printf("Error closing response body: %v", cerr)
		}
	}()
	if resp.StatusCode >= 300 {
		return errors.New(resp.Status)
	}
	var out struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}
	log.Printf("New diagram id %s", out.ID)
	return nil
}
