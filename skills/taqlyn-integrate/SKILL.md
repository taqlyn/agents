---
name: taqlyn-integrate
description: Integrates Taqlyn deferred deep links into Android, iOS, React Native, or Flutter apps with the Taqlyn MCP. Use when adding Taqlyn, App Links, Universal Links, deferred matching, SdkCore, or creating a Taqlyn app/bind/credentials from an agent.
---

# Integrate Taqlyn

Do the work with **Taqlyn MCP tools**. Ask the user only for secrets inspect_workspace cannot see (Play SHA-256, iOS Team ID / Apple App ID, password / TOTP).

## Workflow

Copy and track:

```
Integrate:
- [ ] 1. MCP reachable (`health`)
- [ ] 2. Logged in (`auth_status` / `auth_login`)
- [ ] 3. `integration_plan` with the project root
- [ ] 4. `create_app` or reuse `list_apps`
- [ ] 5. Bind platforms
- [ ] 6. `get_app` → sandbox `clientId` + `publicKeyId` only
- [ ] 7. Install SDK + wire configure → resolve → ready → observe → consume
- [ ] 8. Optional navigation adapter
- [ ] 9. Optional `create_link` (deferred_app) for a test URL
```

### 1. MCP

If tools are missing, load [taqlyn-mcp](../taqlyn-mcp/SKILL.md). Then `health`.

### 2. Auth

Call `auth_status`. If not authenticated, ask **once** for email + password.

```
auth_login
  environment: sandbox
  permission: write
```

Production write needs `confirmProductionWrite: true`. If `mfaRequired`, ask once for TOTP → `auth_verify_totp`.

Never put the private PEM on a device, in git, or in MCP logs.

### 3. Plan

`inspect_workspace` + `integration_plan` on the app repo root. Follow `next`. Do not re-ask for package name / bundle id when the plan already has them.

### 4–5. App + binds

- Android: upload-key SHA-256 **and** Play App Signing SHA-256.
- iOS: Associated Domains host from `get_app` boot hosts after bind.
- RN / Flutter: bind native Android and iOS (or Expo config plugin).

### 6–7. SDK

Public ids only (`app_test_…`, `pk_test_…`). After splash / sign-in:

1. `resolveDeferred()`
2. `setReadyForNavigation(true)`
3. Map `path` / `params`, then `consume(linkId)`

If the router already handles https, set `linkProcessingMode` to deferred-only.

Docs: project `apps/docs` — [Android](../../../../apps/docs/content/platforms/android.md), [iOS](../../../../apps/docs/content/platforms/ios.md), [React Native](../../../../apps/docs/content/platforms/react-native.md), [Flutter](../../../../apps/docs/content/platforms/flutter.md). Fall back to https://docs.taqlyn.com if those paths are not in this workspace.

### 8. Navigation adapters (optional)

- Android: `com.taqlyn.nav:navigation2`
- iOS: `TaqlynNavSwiftUI`
- Expo Router: `@taqlyn/nav-expo-router`
- React Navigation: `@taqlyn/nav-react-navigation`
- Flutter: `taqlyn_nav_go_router`

Navigate **once** per `linkId`, then consume.
