package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/taqlyn/agents/internal/api"
	"github.com/taqlyn/agents/internal/auth"
	"github.com/taqlyn/agents/internal/config"
	"github.com/taqlyn/agents/internal/workspace"
)

type Deps struct {
	Cfg   config.Config
	Store *auth.Store
	API   *api.Client
}

func Register(server *mcp.Server, d Deps) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "auth_status",
		Description: "Current Taqlyn MCP login, environment scope (sandbox|production|both), and permission (read|write).",
	}, d.authStatus)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "auth_login",
		Description: "Log in to the Taqlyn account. Request only the scopes you need. Default sandbox+write. Production write requires confirmProductionWrite=true. If the API returns mfaRequired, call auth_verify_totp next.",
	}, d.authLogin)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "auth_verify_totp",
		Description: "Finish login when auth_login returned mfaRequired. Ask the user once for their TOTP code.",
	}, d.authVerifyTOTP)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "auth_set_scope",
		Description: "Change sandbox/production and read/write without re-entering the password. Production write requires confirmProductionWrite=true.",
	}, d.authSetScope)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "auth_logout",
		Description: "Revoke the Taqlyn session and delete the local MCP token file.",
	}, d.authLogout)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "health",
		Description: "Ping TAQLYN_API_URL /health. Does not require login.",
	}, d.health)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "inspect_workspace",
		Description: "Detect Android / iOS / React Native / Flutter from a project root. No Taqlyn login required.",
	}, d.inspectWorkspace)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "integration_plan",
		Description: "One-shot integrate helper: inspect the project, list Taqlyn apps if logged in, and return the next bind/credential/SDK steps. Prefer this before asking the user questions.",
	}, d.integrationPlan)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_apps",
		Description: "List Taqlyn apps in the logged-in organization.",
	}, d.listApps)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_app",
		Description: "Create a Taqlyn app (write). Sandbox and production hosts are booted automatically.",
	}, d.createApp)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_app",
		Description: "App plus platforms and public credentials (clientId, publicKeyId) for the allowed environment. Never returns private PEM.",
	}, d.getApp)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bind_android",
		Description: "Bind Android package name + SHA-256 fingerprints (write). Include upload key AND Play App Signing certs.",
	}, d.bindAndroid)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bind_ios",
		Description: "Bind iOS bundleId, teamId, and numeric Apple App ID (write).",
	}, d.bindIOS)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bind_web",
		Description: "Bind web baseUrl and pathTemplate for an environment (write).",
	}, d.bindWeb)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_links",
		Description: "List short links for an app×environment.",
	}, d.listLinks)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_link",
		Description: "Create a short / deferred link (write). mode: web_only | app_aware | deferred_app.",
	}, d.createLink)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_link",
		Description: "Get one short link by id.",
	}, d.getLink)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_link_stats",
		Description: "Click/match funnel for a link. Use to debug deferred matching.",
	}, d.getLinkStats)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_app_stats",
		Description: "Org app funnel stats.",
	}, d.getAppStats)
}

type envIn struct {
	Environment string `json:"environment,omitempty" jsonschema:"sandbox or production; defaults to sandbox"`
}

func (d Deps) snap() auth.Snapshot { return d.Store.Get() }

func (d Deps) requireRead(env string) (auth.Snapshot, auth.Environment, error) {
	s := d.snap()
	e := auth.ParseEnvironment(env)
	if e == "" || e == auth.EnvBoth {
		e = auth.EnvSandbox
		if env == "" && s.Environment == auth.EnvProduction {
			e = auth.EnvProduction
		}
	}
	if err := s.AllowsEnv(e); err != nil {
		return s, e, err
	}
	return s, e, nil
}

func (d Deps) requireWrite(env string) (auth.Snapshot, auth.Environment, error) {
	s, e, err := d.requireRead(env)
	if err != nil {
		return s, e, err
	}
	if err := s.AllowsWrite(); err != nil {
		return s, e, err
	}
	return s, e, nil
}

func text(v any) (*mcp.CallToolResult, any, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, nil, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(b)}},
	}, v, nil
}

func asString(m map[string]any, key string) string {
	v, _ := m[key].(string)
	return v
}

func filterCreds(creds []map[string]any, env auth.Environment) []map[string]any {
	if env == auth.EnvBoth {
		return creds
	}
	out := make([]map[string]any, 0, len(creds))
	for _, c := range creds {
		if strings.EqualFold(fmt.Sprint(c["env"]), string(env)) {
			out = append(out, c)
		}
	}
	return out
}

type emptyIn struct{}

type statusOut struct {
	Authenticated bool   `json:"authenticated"`
	APIURL        string `json:"apiUrl"`
	Email         string `json:"email,omitempty"`
	UserID        string `json:"userId,omitempty"`
	OrgID         string `json:"orgId,omitempty"`
	Role          string `json:"role,omitempty"`
	Environment   string `json:"environment,omitempty"`
	Permission    string `json:"permission,omitempty"`
	TokenFile     string `json:"tokenFile"`
}

func (d Deps) authStatus(ctx context.Context, _ *mcp.CallToolRequest, _ emptyIn) (*mcp.CallToolResult, statusOut, error) {
	s := d.snap()
	out := statusOut{
		Authenticated: s.Token != "",
		APIURL:        firstNonEmpty(s.APIURL, d.Cfg.APIURL),
		Email:         s.Email,
		UserID:        s.UserID,
		OrgID:         s.OrgID,
		Role:          s.Role,
		Environment:   string(s.Environment),
		Permission:    string(s.Permission),
		TokenFile:     d.Store.Path(),
	}
	res, _, err := text(out)
	return res, out, err
}

type loginIn struct {
	Email                  string `json:"email" jsonschema:"Taqlyn account email"`
	Password               string `json:"password" jsonschema:"Taqlyn account password"`
	Environment            string `json:"environment,omitempty" jsonschema:"sandbox, production, or both. Default sandbox"`
	Permission             string `json:"permission,omitempty" jsonschema:"read or write. Default write"`
	ConfirmProductionWrite bool   `json:"confirmProductionWrite,omitempty" jsonschema:"Required when environment includes production AND permission is write"`
}

func (d Deps) authLogin(ctx context.Context, _ *mcp.CallToolRequest, in loginIn) (*mcp.CallToolResult, any, error) {
	env := auth.ParseEnvironment(in.Environment)
	if env == "" {
		env = auth.EnvSandbox
	}
	perm := auth.ParsePermission(in.Permission)
	if perm == "" {
		perm = auth.PermWrite
	}
	if err := confirmProdWrite(env, perm, in.ConfirmProductionWrite); err != nil {
		return nil, nil, err
	}
	out, err := d.API.Login(ctx, in.Email, in.Password)
	if err != nil {
		return nil, nil, err
	}
	if mfa, _ := out["mfaRequired"].(bool); mfa {
		return text(map[string]any{
			"mfaRequired": true,
			"mfaToken":    out["mfaToken"],
			"mfaMethods":  out["mfaMethods"],
			"next":        "Ask the user for a TOTP code once, then call auth_verify_totp. Do not complete login until MFA succeeds.",
			"environment": env,
			"permission":  perm,
		})
	}
	if err := d.persistSession(out, env, perm); err != nil {
		return nil, nil, err
	}
	return d.authStatus(ctx, nil, emptyIn{})
}

type totpIn struct {
	MFAToken               string `json:"mfaToken" jsonschema:"mfaToken from auth_login"`
	Code                   string `json:"code" jsonschema:"6-digit TOTP from the user"`
	Environment            string `json:"environment,omitempty"`
	Permission             string `json:"permission,omitempty"`
	ConfirmProductionWrite bool   `json:"confirmProductionWrite,omitempty"`
}

func (d Deps) authVerifyTOTP(ctx context.Context, _ *mcp.CallToolRequest, in totpIn) (*mcp.CallToolResult, any, error) {
	env := auth.ParseEnvironment(in.Environment)
	if env == "" {
		env = auth.EnvSandbox
	}
	perm := auth.ParsePermission(in.Permission)
	if perm == "" {
		perm = auth.PermWrite
	}
	if err := confirmProdWrite(env, perm, in.ConfirmProductionWrite); err != nil {
		return nil, nil, err
	}
	out, err := d.API.VerifyTOTP(ctx, in.MFAToken, in.Code)
	if err != nil {
		return nil, nil, err
	}
	if err := d.persistSession(out, env, perm); err != nil {
		return nil, nil, err
	}
	return d.authStatus(ctx, nil, emptyIn{})
}

type scopeIn struct {
	Environment            string `json:"environment" jsonschema:"sandbox, production, or both"`
	Permission             string `json:"permission" jsonschema:"read or write"`
	ConfirmProductionWrite bool   `json:"confirmProductionWrite,omitempty"`
}

func (d Deps) authSetScope(ctx context.Context, _ *mcp.CallToolRequest, in scopeIn) (*mcp.CallToolResult, any, error) {
	s := d.snap()
	if s.Token == "" {
		return nil, nil, auth.ErrNotAuthenticated
	}
	env := auth.ParseEnvironment(in.Environment)
	perm := auth.ParsePermission(in.Permission)
	if env == "" || perm == "" {
		return nil, nil, fmt.Errorf("environment must be sandbox|production|both and permission read|write")
	}
	if err := confirmProdWrite(env, perm, in.ConfirmProductionWrite); err != nil {
		return nil, nil, err
	}
	s.Environment = env
	s.Permission = perm
	if err := d.Store.Put(s); err != nil {
		return nil, nil, err
	}
	return d.authStatus(ctx, nil, emptyIn{})
}

func (d Deps) authLogout(ctx context.Context, _ *mcp.CallToolRequest, _ emptyIn) (*mcp.CallToolResult, any, error) {
	if d.snap().Token != "" {
		_ = d.API.Logout(ctx)
	}
	if err := d.Store.Clear(); err != nil {
		return nil, nil, err
	}
	return text(map[string]any{"ok": true})
}

func (d Deps) persistSession(out map[string]any, env auth.Environment, perm auth.Permission) error {
	tok := asString(out, "token")
	if tok == "" {
		return fmt.Errorf("login did not return a session token")
	}
	return d.Store.Put(auth.Snapshot{
		APIURL:      d.Cfg.APIURL,
		Token:       tok,
		UserID:      asString(out, "userId"),
		OrgID:       asString(out, "orgId"),
		Email:       asString(out, "email"),
		Role:        asString(out, "role"),
		Environment: env,
		Permission:  perm,
	})
}

func confirmProdWrite(env auth.Environment, perm auth.Permission, confirm bool) error {
	prod := env == auth.EnvProduction || env == auth.EnvBoth
	if prod && perm == auth.PermWrite && !confirm {
		return fmt.Errorf("production write requires confirmProductionWrite=true")
	}
	return nil
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

func (d Deps) health(ctx context.Context, _ *mcp.CallToolRequest, _ emptyIn) (*mcp.CallToolResult, any, error) {
	out, err := d.API.Health(ctx)
	if err != nil {
		return nil, nil, err
	}
	out["apiUrl"] = d.Cfg.APIURL
	return text(out)
}

type inspectIn struct {
	Root string `json:"root" jsonschema:"Absolute path to the mobile/web app repository"`
}

func (d Deps) inspectWorkspace(_ context.Context, _ *mcp.CallToolRequest, in inspectIn) (*mcp.CallToolResult, any, error) {
	root, err := workspaceRoot(in.Root)
	if err != nil {
		return nil, nil, err
	}
	r, err := workspace.Inspect(root)
	if err != nil {
		return nil, nil, err
	}
	return text(r)
}

type planIn struct {
	Root    string `json:"root" jsonschema:"Project root to inspect"`
	AppName string `json:"appName,omitempty" jsonschema:"Preferred Taqlyn app name when creating"`
}

func (d Deps) integrationPlan(ctx context.Context, _ *mcp.CallToolRequest, in planIn) (*mcp.CallToolResult, any, error) {
	root, err := workspaceRoot(in.Root)
	if err != nil {
		return nil, nil, err
	}
	ws, err := workspace.Inspect(root)
	if err != nil {
		return nil, nil, err
	}
	plan := map[string]any{
		"workspace": ws,
		"next":      []string{},
	}
	s := d.snap()
	if s.Token == "" {
		plan["next"] = []string{
			"Not logged in. Ask the user once for Taqlyn email+password (sandbox+write is enough to integrate).",
			"Call auth_login, then call integration_plan again.",
		}
		return text(plan)
	}
	apps, err := d.API.ListApps(ctx)
	if err != nil {
		return nil, nil, err
	}
	plan["apps"] = apps
	next := []string{}
	name := strings.TrimSpace(in.AppName)
	if name == "" {
		name = filepathBase(root)
	}
	if len(apps) == 0 {
		next = append(next, "No apps yet. Call create_app with name="+name+" (write).")
	} else {
		next = append(next, "Reuse an existing app with get_app, or create_app if none match this project.")
	}
	if containsPlat(ws.Platforms, workspace.Android) {
		next = append(next, "bind_android with packageName="+emptyDash(ws.PackageName)+" and SHA-256 fingerprints from the upload keystore and Play App Signing.")
	}
	if containsPlat(ws.Platforms, workspace.IOS) {
		next = append(next, "bind_ios with bundleId="+emptyDash(ws.BundleID)+" plus Team ID and numeric Apple App ID from App Store Connect.")
	}
	if containsPlat(ws.Platforms, workspace.ReactNative) || containsPlat(ws.Platforms, workspace.Flutter) {
		next = append(next, "Complete native Android and iOS binds on the host projects (or Expo config plugin).")
	}
	next = append(next, "Call get_app and use sandbox clientId + publicKeyId only in the mobile SDK. Never put the private PEM on device.")
	next = append(next, "Wire SdkCore configure → resolveDeferred → setReadyForNavigation → observe → consume. Pair with the platform navigation adapter if the app already has a router.")
	plan["next"] = next
	plan["environment"] = s.Environment
	plan["permission"] = s.Permission
	return text(plan)
}

func workspaceRoot(root string) (string, error) {
	root = strings.TrimSpace(root)
	if root != "" {
		return root, nil
	}
	if w := strings.TrimSpace(os.Getenv("TAQLYN_WORKSPACE")); w != "" {
		return w, nil
	}
	if st, err := os.Stat("/workspace"); err == nil && st.IsDir() {
		return "/workspace", nil
	}
	return "", fmt.Errorf("root is required (or set TAQLYN_WORKSPACE / mount the project at /workspace)")
}

func filepathBase(p string) string {
	p = strings.TrimRight(p, "/")
	if i := strings.LastIndex(p, "/"); i >= 0 {
		return p[i+1:]
	}
	return p
}

func emptyDash(s string) string {
	if s == "" {
		return "(ask user once if inspect_workspace could not read it)"
	}
	return s
}

func containsPlat(ps []workspace.Platform, p workspace.Platform) bool {
	for _, x := range ps {
		if x == p {
			return true
		}
	}
	return false
}

func (d Deps) listApps(ctx context.Context, _ *mcp.CallToolRequest, _ emptyIn) (*mcp.CallToolResult, any, error) {
	if _, _, err := d.requireRead(""); err != nil {
		return nil, nil, err
	}
	out, err := d.API.ListApps(ctx)
	if err != nil {
		return nil, nil, err
	}
	return text(out)
}

type createAppIn struct {
	Name string `json:"name" jsonschema:"App display name"`
}

func (d Deps) createApp(ctx context.Context, _ *mcp.CallToolRequest, in createAppIn) (*mcp.CallToolResult, any, error) {
	if _, _, err := d.requireWrite(""); err != nil {
		return nil, nil, err
	}
	if strings.TrimSpace(in.Name) == "" {
		return nil, nil, fmt.Errorf("name is required")
	}
	out, err := d.API.CreateApp(ctx, in.Name, nil)
	if err != nil {
		return nil, nil, err
	}
	return text(out)
}

type appIn struct {
	AppID       string `json:"appId" jsonschema:"Taqlyn app id"`
	Environment string `json:"environment,omitempty"`
}

func (d Deps) getApp(ctx context.Context, _ *mcp.CallToolRequest, in appIn) (*mcp.CallToolResult, any, error) {
	_, env, err := d.requireRead(in.Environment)
	if err != nil {
		return nil, nil, err
	}
	if in.AppID == "" {
		return nil, nil, fmt.Errorf("appId is required")
	}
	plats, err := d.API.AppPlatforms(ctx, in.AppID, string(env))
	if err != nil {
		return nil, nil, err
	}
	creds, err := d.API.Credentials(ctx, in.AppID)
	if err != nil {
		return nil, nil, err
	}
	return text(map[string]any{
		"appId":       in.AppID,
		"environment": env,
		"platforms":   plats,
		"credentials": filterCreds(creds, env),
		"note":        "Use clientId + publicKeyId in the mobile SDK. Private PEM stays on the server.",
	})
}

type bindAndroidIn struct {
	AppID               string   `json:"appId"`
	PackageName         string   `json:"packageName"`
	SHA256Fingerprints  []string `json:"sha256Fingerprints" jsonschema:"Upload key and Play App Signing SHA-256"`
}

func (d Deps) bindAndroid(ctx context.Context, _ *mcp.CallToolRequest, in bindAndroidIn) (*mcp.CallToolResult, any, error) {
	if _, _, err := d.requireWrite(""); err != nil {
		return nil, nil, err
	}
	out, err := d.API.BindAndroid(ctx, in.AppID, map[string]any{
		"packageName":        in.PackageName,
		"sha256Fingerprints": in.SHA256Fingerprints,
	})
	if err != nil {
		return nil, nil, err
	}
	return text(out)
}

type bindIOSIn struct {
	AppID    string `json:"appId"`
	BundleID string `json:"bundleId"`
	TeamID   string `json:"teamId"`
	AppleID  string `json:"appleId" jsonschema:"Numeric Apple App ID"`
}

func (d Deps) bindIOS(ctx context.Context, _ *mcp.CallToolRequest, in bindIOSIn) (*mcp.CallToolResult, any, error) {
	if _, _, err := d.requireWrite(""); err != nil {
		return nil, nil, err
	}
	out, err := d.API.BindIOS(ctx, in.AppID, map[string]any{
		"bundleId": in.BundleID,
		"teamId":   in.TeamID,
		"appleId":  in.AppleID,
	})
	if err != nil {
		return nil, nil, err
	}
	return text(out)
}

type bindWebIn struct {
	AppID          string   `json:"appId"`
	Environment    string   `json:"environment,omitempty"`
	BaseURL        string   `json:"baseUrl"`
	PathTemplate   string   `json:"pathTemplate"`
	QueryForwarding string  `json:"queryForwarding,omitempty"`
	AuthorizedOrigins []string `json:"authorizedOrigins,omitempty"`
}

func (d Deps) bindWeb(ctx context.Context, _ *mcp.CallToolRequest, in bindWebIn) (*mcp.CallToolResult, any, error) {
	_, env, err := d.requireWrite(in.Environment)
	if err != nil {
		return nil, nil, err
	}
	body := map[string]any{
		"env":          env,
		"baseUrl":      in.BaseURL,
		"pathTemplate": in.PathTemplate,
	}
	if in.QueryForwarding != "" {
		body["queryForwarding"] = in.QueryForwarding
	}
	if len(in.AuthorizedOrigins) > 0 {
		body["authorizedOrigins"] = in.AuthorizedOrigins
	}
	out, err := d.API.BindWeb(ctx, in.AppID, body)
	if err != nil {
		return nil, nil, err
	}
	return text(out)
}

type listLinksIn struct {
	AppID       string `json:"appId"`
	Environment string `json:"environment,omitempty"`
}

func (d Deps) listLinks(ctx context.Context, _ *mcp.CallToolRequest, in listLinksIn) (*mcp.CallToolResult, any, error) {
	_, env, err := d.requireRead(in.Environment)
	if err != nil {
		return nil, nil, err
	}
	out, err := d.API.ListLinks(ctx, in.AppID, string(env))
	if err != nil {
		return nil, nil, err
	}
	return text(out)
}

type createLinkIn struct {
	AppID           string            `json:"appId"`
	Environment     string            `json:"environment,omitempty"`
	DestinationWeb  string            `json:"destinationWeb"`
	DestinationPath string            `json:"destinationPath,omitempty"`
	Mode            string            `json:"mode,omitempty" jsonschema:"web_only, app_aware, or deferred_app"`
	Params          map[string]string `json:"params,omitempty"`
}

func (d Deps) createLink(ctx context.Context, _ *mcp.CallToolRequest, in createLinkIn) (*mcp.CallToolResult, any, error) {
	_, env, err := d.requireWrite(in.Environment)
	if err != nil {
		return nil, nil, err
	}
	body := map[string]any{
		"appId":          in.AppID,
		"env":            env,
		"destinationWeb": in.DestinationWeb,
	}
	if in.DestinationPath != "" {
		body["destinationPath"] = in.DestinationPath
	}
	if in.Mode != "" {
		body["mode"] = in.Mode
	}
	if in.Params != nil {
		body["params"] = in.Params
	}
	out, err := d.API.CreateLink(ctx, body)
	if err != nil {
		return nil, nil, err
	}
	return text(out)
}

type idIn struct {
	ID string `json:"id"`
}

func (d Deps) getLink(ctx context.Context, _ *mcp.CallToolRequest, in idIn) (*mcp.CallToolResult, any, error) {
	if _, _, err := d.requireRead(""); err != nil {
		return nil, nil, err
	}
	out, err := d.API.GetLink(ctx, in.ID)
	if err != nil {
		return nil, nil, err
	}
	return text(out)
}

type linkIDIn struct {
	LinkID string `json:"linkId"`
}

func (d Deps) getLinkStats(ctx context.Context, _ *mcp.CallToolRequest, in linkIDIn) (*mcp.CallToolResult, any, error) {
	if _, _, err := d.requireRead(""); err != nil {
		return nil, nil, err
	}
	out, err := d.API.LinkStats(ctx, in.LinkID)
	if err != nil {
		return nil, nil, err
	}
	return text(out)
}

func (d Deps) getAppStats(ctx context.Context, _ *mcp.CallToolRequest, in appIn) (*mcp.CallToolResult, any, error) {
	if _, _, err := d.requireRead(in.Environment); err != nil {
		return nil, nil, err
	}
	out, err := d.API.AppStats(ctx, in.AppID)
	if err != nil {
		return nil, nil, err
	}
	return text(out)
}
