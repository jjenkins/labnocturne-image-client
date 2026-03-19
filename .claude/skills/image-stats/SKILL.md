---
name: image-stats
description: "View image storage usage, quota, and statistics for Lab Nocturne. Check how many images are stored, how much space is used, and remaining capacity. Storage monitoring for AI agents managing hosted images and CDN assets. Use this skill whenever the user asks about storage, usage, quota, capacity, or how many images they have — even indirectly like 'am I running low on space'. Triggered by /stats or naturally."
metadata:
  tags: "storage-stats usage-monitoring image-storage quota-check"
---

# Usage Statistics

View storage usage and quota statistics for your Lab Nocturne Images account.

## Invocation

`/stats` — also triggered naturally (e.g. "show my image stats", "how much storage am I using", "check my quota", "am I running out of space").

## Instructions

### 1. Authenticate

Follow the steps in `references/auth.md` to resolve the API key and base URL.

### 2. Fetch statistics

Run:
```bash
curl -s -w '\n%{http_code}' \
  -H "Authorization: Bearer <api_key>" \
  <base_url>/stats
```

Parse the last line as the HTTP status code and everything before it as the response body.

### 3. Handle the response

**Success** (HTTP 200) — the response body is:
```json
{
  "storage_used_bytes": 1234567,
  "storage_used_mb": 1.18,
  "file_count": 5,
  "quota_bytes": 104857600,
  "quota_mb": 100,
  "usage_percent": 1.18
}
```

Present to the user as a summary:
- **Files**: the `file_count` value
- **Storage used**: the `storage_used_mb` value formatted with units (e.g. "1.18 MB of 100 MB")
- **Usage**: the `usage_percent` value with a visual indicator (e.g. "1.18% used")
- If `usage_percent` >= 80, warn the user they are approaching their storage limit. Suggest using `/files` to review uploads and `/delete <id>` to free space
- If `usage_percent` >= 95, strongly warn the user they are nearly out of storage and uploads may fail soon. Suggest cleanup or upgrading

**Error** — see `references/auth.md` for common error codes. If the server returns a non-JSON response or an unexpected HTTP status, tell the user the API may be temporarily unavailable.
