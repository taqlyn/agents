---
name: taqlyn-debug
description: Debug Taqlyn deferred matching, App Links, Universal Links, first-open resolve, and link stats using the Taqlyn MCP. Use when a Taqlyn link opens the wrong place, deferred is empty, AASA/assetlinks fail, or clicks do not match.
---

# Debug Taqlyn

Use MCP read scope (`sandbox` is enough unless the failing link is live).

## Workflow

```
Debug:
- [ ] 1. `auth_status` / `auth_login` (read is enough)
- [ ] 2. `get_link` or `list_links` for the failing code
- [ ] 3. `get_link_stats` (clicks vs matches)
- [ ] 4. `get_app` platforms + public credentials for the env
- [ ] 5. `inspect_workspace` vs bound package / bundle
```

## How to read stats

| clicks | matches | Meaning |
|--------|---------|---------|
| 0 | 0 | Link never opened, wrong host, or cached edge miss |
| >0 | 0 | Click recorded, first-open resolve failed or already consumed |
| >0 | >0 | Match happened; bug is likely navigation / consume / wrong env SDK ids |

## Common fixes

- **Android autoVerify / Play:** bind **both** upload-key and Play App Signing SHA-256. Host in the intent filter must match the Taqlyn boot host for that env.
- **iOS Universal Links:** Associated Domains `applinks:HOST` for the env host. Team ID + numeric Apple App ID must match App Store Connect.
- **Deferred empty on second launch:** expected after `consume`. First cold start after store install only.
- **Warm https double-handled:** set `linkProcessingMode` deferred-only when the native router already consumes the URL.
- **Wrong env:** sandbox ids (`app_test_`, `pk_test_`) against a production host (or the reverse). `get_app` for the same env as the link.
- **Emulator deferred:** Android first-open after Play needs a **device**. App Links can be tested on emulator.

Do not rotate credentials unless the user asked. MCP `get_app` never returns private PEM.
