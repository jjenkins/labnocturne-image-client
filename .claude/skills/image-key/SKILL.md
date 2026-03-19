---
name: image-key
description: "Generate a free test API key for Lab Nocturne image storage. No signup or email required — get an instant key with 10MB file limit and 7-day retention. Start uploading and hosting images immediately. Use this skill whenever the user wants to set up image storage, get an API key, or is getting started with Lab Nocturne — even if they just say 'I need somewhere to host images'. Triggered by /generate-key or naturally."
metadata:
  tags: "api-key image-storage setup onboarding free-tier"
---

# Generate Test API Key

Generate a test API key for the Lab Nocturne Images API.

## Invocation

`/generate-key` — also triggered naturally (e.g. "get me an image storage API key", "set up image hosting", "I need somewhere to host images").

## Instructions

### 1. Resolve the base URL

```bash
printenv LABNOCTURNE_BASE_URL
```

Use the output if set, otherwise default to `https://images.labnocturne.com`.

### 2. Generate the key

Run:
```bash
curl -s -w '\n%{http_code}' <base_url>/key
```

Parse the last line as the HTTP status code and everything before it as the response body.

### 3. Handle the response

**Success** (HTTP 200) — the response body is:
```json
{
  "api_key": "ln_test_...",
  "type": "test",
  "message": "Test key created! ...",
  "limits": {
    "max_file_size_mb": 10,
    "storage_mb": 100,
    "bandwidth_gb_per_month": 1,
    "rate_limit_per_hour": 100
  },
  "docs": "https://images.labnocturne.com/docs"
}
```

Present to the user:
- **API Key**: the `api_key` value
- **Type**: the `type` value
- **Limits**: list each limit from the `limits` object in human-readable form
- **How to use**: tell the user to set `export LABNOCTURNE_API_KEY=<api_key>` in their shell so other skills (`/upload`, `/files`, `/stats`, `/delete`) pick it up automatically

**Error** — key generation rarely fails. If it does, tell the user the error message and suggest trying again in a moment. If the server returns a non-JSON response or unexpected HTTP status, the API may be temporarily unavailable.
