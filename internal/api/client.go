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

// CreateObject creates a new object in IcePanel.
func (c *IcePanelClient) CreateObject(ctx context.Context, lc, ver string, obj *Object, dryRun bool) error {
	if dryRun {
		log.Printf("[Dry-Run] Would create object: %+v in landscape %s, version %s", obj, lc, ver)
		return nil
	}
	b, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("failed to marshal object: %w", err)
	}
	url := fmt.Sprintf("%s/landscapes/%s/versions/%s/model/objects", c.baseURL, lc, ver)
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
	return nil
}

// UpdateObject updates an existing object in IcePanel by handle ID.
func (c *IcePanelClient) UpdateObject(ctx context.Context, lc, ver, handle string, obj *Object, dryRun bool) error {
	if dryRun {
		log.Printf("[Dry-Run] Would update object %s: %+v in landscape %s, version %s", handle, obj, lc, ver)
		return nil
	}
	b, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("failed to marshal object: %w", err)
	}
	url := fmt.Sprintf("%s/landscapes/%s/versions/%s/model/objects/%s", c.baseURL, lc, ver, handle)
	resp, err := c.call(ctx, "PUT", url, b)
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
	return nil
}

// GetObject retrieves an object by handle ID.
func (c *IcePanelClient) GetObject(ctx context.Context, lc, ver, handle string) (*Object, error) {
	url := fmt.Sprintf("%s/landscapes/%s/versions/%s/model/objects/%s", c.baseURL, lc, ver, handle)
	resp, err := c.call(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			log.Printf("Error closing response body: %v", cerr)
		}
	}()
	if resp.StatusCode >= 300 {
		return nil, errors.New(resp.Status)
	}
	var obj Object
	if err := json.NewDecoder(resp.Body).Decode(&obj); err != nil {
		return nil, err
	}
	return &obj, nil
}

// ListObjects retrieves all objects for a given landscape and version.
func (c *IcePanelClient) ListObjects(ctx context.Context, lc, ver string) ([]*Object, error) {
	url := fmt.Sprintf("%s/landscapes/%s/versions/%s/model/objects?per=1000", c.baseURL, lc, ver)
	resp, err := c.call(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			log.Printf("Error closing response body: %v", cerr)
		}
	}()
	if resp.StatusCode >= 300 {
		return nil, errors.New(resp.Status)
	}
	var out struct {
		Data []*Object `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out.Data, nil
}

// Equal compares two IcePanel objects for logical equality (ignores Handle).
func (o *Object) Equal(other *Object) bool {
	if o == nil || other == nil {
		return o == other
	}
	if o.Name != other.Name || o.Desc != other.Desc || o.Type != other.Type {
		return false
	}
	if len(o.Props) != len(other.Props) {
		return false
	}
	for k, v := range o.Props {
		if ov, ok := other.Props[k]; !ok || !equalInterface(v, ov) {
			return false
		}
	}
	return true
}

// Diff returns a map of fields that differ between two objects (ignores Handle).
func (o *Object) Diff(other *Object) map[string][2]interface{} {
	diff := make(map[string][2]interface{})
	if o == nil || other == nil {
		diff["nil"] = [2]interface{}{o, other}
		return diff
	}
	if o.Name != other.Name {
		diff["Name"] = [2]interface{}{o.Name, other.Name}
	}
	if o.Desc != other.Desc {
		diff["Desc"] = [2]interface{}{o.Desc, other.Desc}
	}
	if o.Type != other.Type {
		diff["Type"] = [2]interface{}{o.Type, other.Type}
	}
	// Compare Props
	for k, v := range o.Props {
		if ov, ok := other.Props[k]; !ok || !equalInterface(v, ov) {
			diff["Props."+k] = [2]interface{}{v, ov}
		}
	}
	for k, v := range other.Props {
		if _, ok := o.Props[k]; !ok {
			diff["Props."+k] = [2]interface{}{nil, v}
		}
	}
	return diff
}

// equalInterface compares two interface{} values for equality.
func equalInterface(a, b interface{}) bool {
	aj, errA := json.Marshal(a)
	bj, errB := json.Marshal(b)
	if errA != nil || errB != nil {
		return false
	}
	return bytes.Equal(aj, bj)
}

// ValidateLandscapeVersion checks if the given landscape and version exist in IcePanel.
func (c *IcePanelClient) ValidateLandscapeVersion(ctx context.Context, lc, ver string) error {
	// Check landscape
	url := fmt.Sprintf("%s/landscapes/%s", c.baseURL, lc)
	resp, err := c.call(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to check landscape: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			log.Printf("Error closing response body: %v", cerr)
		}
	}()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("landscape %s not found (status %s)", lc, resp.Status)
	}
	// Check version
	url = fmt.Sprintf("%s/landscapes/%s/versions/%s", c.baseURL, lc, ver)
	resp, err = c.call(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to check version: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			log.Printf("Error closing response body: %v", cerr)
		}
	}()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("version %s not found in landscape %s (status %s)", ver, lc, resp.Status)
	}
	return nil
}

// WipeVersionIfRequested wipes the version if the wipe flag is true.
func (c *IcePanelClient) WipeVersionIfRequested(ctx context.Context, lc, ver string, wipe bool) error {
	if !wipe {
		log.Printf("Skipping wipe for landscape %s, version %s", lc, ver)
		return nil
	}
	log.Printf("Wiping IcePanel version: landscape=%s, version=%s", lc, ver)
	return c.WipeVersion(ctx, lc, ver)
}
