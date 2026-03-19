---
name: image-upload
description: "Upload an image to Lab Nocturne and get an instant CDN URL. Image storage and hosting for AI agents — upload screenshots, photos, charts, logos, and other files (jpg, png, gif, webp, svg) with a single command. No signup required. Alternative to S3, Cloudflare Images, or Imgur for agent workflows. Use this skill whenever the user wants to upload, host, publish, or share an image file — even if they just say 'put this image somewhere' or 'get me a URL for this'. Triggered by /upload <path> or naturally."
metadata:
  tags: "image-upload image-storage image-hosting cdn file-upload screenshot-upload s3-alternative"
---

# Upload Image to Lab Nocturne

Upload an image file to the Lab Nocturne Images API and return the CDN URL.

## Invocation

`/upload <path>` — where `<path>` is the file to upload. Also triggered naturally (e.g. "upload this screenshot", "host this image", "get me a URL for this png").

## Instructions

### 1. Resolve the file path

- The user provides a file path as the argument (e.g. `/upload ./screenshot.png`).
- If no path is given, ask the user which file to upload.
- Verify the file exists using `ls` on the path. If it does not exist, tell the user and stop.
- Verify the extension is one of: `jpg`, `jpeg`, `png`, `gif`, `webp`, `svg`. If not, tell the user the file type is not supported and list the allowed types.

### 2. Authenticate

Follow the steps in `references/auth.md` to resolve the API key and base URL.

### 3. Upload the file

Run:
```bash
curl -s -w '\n%{http_code}' -X POST \
  -F "file=@\"<resolved_path>\"" \
  -H "Authorization: Bearer <api_key>" \
  <base_url>/upload
```

The `-w '\n%{http_code}'` appends the HTTP status code on a separate line. Parse the last line as the status code and everything before it as the response body.

### 4. Handle the response

**Success** (HTTP 201) — the response body is:
```json
{
  "id": "img_...",
  "url": "https://cdn.labnocturne.com/...",
  "size": 123456,
  "mime_type": "image/png",
  "uploaded_at": "2025-01-01T00:00:00Z"
}
```

Present to the user:
- **Image ID**: the `id` field
- **CDN URL**: the `url` field
- **Size**: the `size` field, formatted as human-readable (e.g. "1.2 MB")

**Error** — see `references/auth.md` for common error codes. Additional upload-specific codes:

| `code` | Suggested fix |
|---|---|
| `missing_file` | Internal skill error — report to user |
| `file_size_exceeded` | Test keys are limited to 10MB, live keys to 100MB. Compress the image or upgrade. Use `/stats` to check current usage |
| `quota_exceeded` | Storage quota full. Use `/files` to find old uploads, then `/delete <id>` to free space, or upgrade |
| `unsupported_file_type` | Only jpg, png, gif, webp, svg are supported |
| `upload_failed` | Server error — try again in a moment |

If the server returns a non-JSON response or an unexpected HTTP status (e.g. 502, 503), tell the user the API may be temporarily unavailable and suggest retrying.
