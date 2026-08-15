package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client is the thin Taqlyn HTTP wrapper (session Bearer). Feature tools import this, not net/http.
type Client struct {
	BaseURL    string
	HTTP       *http.Client
	UserAgent  string
	token      func() string
}

func New(baseURL string, token func() string) *Client {
	return &Client{
		BaseURL:   strings.TrimRight(baseURL, "/"),
		HTTP:      &http.Client{Timeout: 30 * time.Second},
		UserAgent: "taqlyn-mcp/0.1",
		token:     token,
	}
}

type Error struct {
	Status  int    `json:"status"`
	Code    string `json:"error"`
	Message string `json:"message"`
}

func (e Error) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("taqlyn api %d: %s (%s)", e.Status, e.Message, e.Code)
	}
	if e.Message != "" {
		return fmt.Sprintf("taqlyn api %d: %s", e.Status, e.Message)
	}
	return fmt.Sprintf("taqlyn api %d", e.Status)
}

func (c *Client) Health(ctx context.Context) (map[string]any, error) {
	var out map[string]any
	if err := c.do(ctx, http.MethodGet, "/health", nil, nil, false, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) Login(ctx context.Context, email, password string) (map[string]any, error) {
	var out map[string]any
	err := c.do(ctx, http.MethodPost, "/v1/auth/login", map[string]string{
		"email":    email,
		"password": password,
	}, nil, false, &out)
	return out, err
}

func (c *Client) VerifyTOTP(ctx context.Context, mfaToken, code string) (map[string]any, error) {
	var out map[string]any
	err := c.do(ctx, http.MethodPost, "/v1/auth/mfa/totp/verify", map[string]string{
		"mfaToken": mfaToken,
		"code":     code,
	}, nil, false, &out)
	return out, err
}

func (c *Client) Logout(ctx context.Context) error {
	return c.do(ctx, http.MethodPost, "/v1/auth/logout", nil, nil, true, nil)
}

func (c *Client) Me(ctx context.Context) (map[string]any, error) {
	var out map[string]any
	if err := c.do(ctx, http.MethodGet, "/v1/auth/me", nil, nil, true, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) ListApps(ctx context.Context) ([]map[string]any, error) {
	var out []map[string]any
	if err := c.do(ctx, http.MethodGet, "/v1/session/apps", nil, nil, true, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) CreateApp(ctx context.Context, name string, platforms []string) (map[string]any, error) {
	body := map[string]any{"name": name}
	if len(platforms) > 0 {
		body["platforms"] = platforms
	}
	var out map[string]any
	if err := c.do(ctx, http.MethodPost, "/v1/session/apps", body, nil, true, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) AppPlatforms(ctx context.Context, appID, env string) (map[string]any, error) {
	q := url.Values{}
	if env != "" {
		q.Set("env", env)
	}
	var out map[string]any
	if err := c.do(ctx, http.MethodGet, "/v1/session/apps/"+url.PathEscape(appID)+"/platforms", nil, q, true, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) BindAndroid(ctx context.Context, appID string, body map[string]any) (map[string]any, error) {
	var out map[string]any
	path := "/v1/session/apps/" + url.PathEscape(appID) + "/platforms/android"
	if err := c.do(ctx, http.MethodPut, path, body, nil, true, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) BindIOS(ctx context.Context, appID string, body map[string]any) (map[string]any, error) {
	var out map[string]any
	path := "/v1/session/apps/" + url.PathEscape(appID) + "/platforms/ios"
	if err := c.do(ctx, http.MethodPut, path, body, nil, true, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) BindWeb(ctx context.Context, appID string, body map[string]any) (map[string]any, error) {
	var out map[string]any
	path := "/v1/session/apps/" + url.PathEscape(appID) + "/platforms/web"
	if err := c.do(ctx, http.MethodPut, path, body, nil, true, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) Credentials(ctx context.Context, appID string) ([]map[string]any, error) {
	var out []map[string]any
	path := "/v1/session/apps/" + url.PathEscape(appID) + "/credentials"
	if err := c.do(ctx, http.MethodGet, path, nil, nil, true, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) ListLinks(ctx context.Context, appID, env string) ([]map[string]any, error) {
	q := url.Values{}
	q.Set("appId", appID)
	if env != "" {
		q.Set("env", env)
	}
	var out []map[string]any
	if err := c.do(ctx, http.MethodGet, "/v1/session/short-links", nil, q, true, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) CreateLink(ctx context.Context, body map[string]any) (map[string]any, error) {
	var out map[string]any
	if err := c.do(ctx, http.MethodPost, "/v1/session/short-links", body, nil, true, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetLink(ctx context.Context, id string) (map[string]any, error) {
	var out map[string]any
	if err := c.do(ctx, http.MethodGet, "/v1/session/short-links/"+url.PathEscape(id), nil, nil, true, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) LinkStats(ctx context.Context, linkID string) (map[string]any, error) {
	var out map[string]any
	if err := c.do(ctx, http.MethodGet, "/v1/stats/links/"+url.PathEscape(linkID), nil, nil, true, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) AppStats(ctx context.Context, appID string) (map[string]any, error) {
	var out map[string]any
	if err := c.do(ctx, http.MethodGet, "/v1/stats/apps/"+url.PathEscape(appID), nil, nil, true, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) do(ctx context.Context, method, path string, body any, query url.Values, auth bool, out any) error {
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		rdr = bytes.NewReader(b)
	}
	u := c.BaseURL + path
	if query != nil {
		if enc := query.Encode(); enc != "" {
			u += "?" + enc
		}
	}
	req, err := http.NewRequestWithContext(ctx, method, u, rdr)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if auth {
		tok := ""
		if c.token != nil {
			tok = c.token()
		}
		if tok == "" {
			return fmt.Errorf("not authenticated: call auth_login")
		}
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 300 {
		var apiErr Error
		_ = json.Unmarshal(raw, &apiErr)
		apiErr.Status = resp.StatusCode
		if apiErr.Message == "" {
			apiErr.Message = strings.TrimSpace(string(raw))
		}
		return apiErr
	}
	if out == nil || resp.StatusCode == http.StatusNoContent || len(bytes.TrimSpace(raw)) == 0 {
		return nil
	}
	return json.Unmarshal(raw, out)
}
