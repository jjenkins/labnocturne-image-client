---
name: image-files
description: "List and browse uploaded images stored on Lab Nocturne. View file names, sizes, types, and CDN URLs with pagination and sorting. Image asset management for AI agents — find previously uploaded screenshots, photos, and files. Use this skill whenever the user wants to see what images they have, find an image ID, look up a previously uploaded file, or browse their hosted assets. Triggered by /files or naturally."
metadata:
  tags: "image-list file-management image-storage asset-management image-browser"
---

# List Files

List uploaded files from your Lab Nocturne Images account with pagination and sorting.

## Invocation

`/files [options]` — also triggered naturally (e.g. "list my hosted images", "show my image uploads", "what images have I stored", "find the image I uploaded earlier").

Options can be specified naturally (e.g. `/files newest first`, `/files limit 10`, `/files page 2`).

## Instructions

### 1. Parse options

Extract optional parameters from the user's input:
- **limit**: number of files to return (1-100, default 100)
- **offset**: number of files to skip for pagination (default 0)
- **sort**: one of `uploaded_at_desc` (default), `uploaded_at_asc`, `size_desc`, `size_asc`

Map natural language to sort values:
- "newest first", "most recent" → `uploaded_at_desc`
- "oldest first" → `uploaded_at_asc`
- "largest first", "biggest" → `size_desc`
- "smallest first" → `size_asc`

If the user asks for "page N", calculate offset as `(N - 1) * limit`.

### 2. Authenticate

Follow the steps in `references/auth.md` to resolve the API key and base URL.

### 3. Fetch the file list

Run:
```bash
curl -s -w '\n%{http_code}' \
  -H "Authorization: Bearer <api_key>" \
  "<base_url>/files?limit=<limit>&offset=<offset>&sort=<sort>"
```

Parse the last line as the HTTP status code and everything before it as the response body.

### 4. Handle the response

**Success** (HTTP 200) — the response body is:
```json
{
  "files": [
    {
      "id": "img_...",
      "url": "https://cdn.labnocturne.com/...",
      "filename": "photo.jpg",
      "size": 245678,
      "mime_type": "image/jpeg",
      "uploaded_at": "2025-01-01T00:00:00Z"
    }
  ],
  "pagination": {
    "total": 50,
    "limit": 100,
    "offset": 0,
    "next": "https://..."
  }
}
```

Present to the user as a formatted list:
- For each file show: **ID**, **filename**, **size** (human-readable, e.g. "240 KB"), **type** (`mime_type`), **uploaded** (relative time, e.g. "2 days ago")
- After the list, show pagination info: "Showing X-Y of Z files"
- If there are more files (i.e. `pagination.next` is present), tell the user how to see the next page (e.g. "Use `/files page 2` to see more")
- If the list is empty, tell the user they have no uploaded files yet. Suggest using `/upload <path>` to get started

**Error** — see `references/auth.md` for common error codes. If the server returns a non-JSON response or an unexpected HTTP status, tell the user the API may be temporarily unavailable.
