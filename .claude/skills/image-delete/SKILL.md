---
name: image-delete
description: "Delete an image from Lab Nocturne image storage. Remove uploaded screenshots, photos, or other hosted files by image ID. Soft-delete with confirmation — clean up old uploads, free storage space, remove unwanted images from CDN. Use this skill whenever the user wants to remove, delete, or clean up an image — even if they say 'get rid of that image' or 'I don't need that upload anymore'. Triggered by /delete <image_id> or naturally."
metadata:
  tags: "image-delete file-delete image-storage cleanup"
---

# Delete Image

Soft-delete an image from your Lab Nocturne Images account.

## Invocation

`/delete <image_id>` — also triggered naturally (e.g. "delete image img_abc...", "remove that image", "clean up my old uploads").

## Instructions

### 1. Resolve the image ID

- The user provides an image ID as the argument (e.g. `/delete img_01JXYZ...`).
- If no ID is given, ask the user which image to delete. Suggest using `/files` to find the ID.
- If the user provides a raw ULID without the `img_` prefix, prepend `img_` automatically.
- The ID should match the pattern `img_` followed by a ULID (26 alphanumeric characters).

### 2. Confirm with the user

**This is a destructive action.** Before executing, tell the user which image ID will be deleted and ask for confirmation. Only proceed if the user confirms.

### 3. Authenticate

Follow the steps in `references/auth.md` to resolve the API key and base URL.

### 4. Delete the image

Run:
```bash
curl -s -w '\n%{http_code}' -X DELETE \
  -H "Authorization: Bearer <api_key>" \
  <base_url>/i/<image_id>
```

Parse the last line as the HTTP status code and everything before it as the response body.

### 5. Handle the response

**Success** (HTTP 200) — the response body is:
```json
{
  "success": true,
  "message": "File deleted successfully",
  "id": "img_..."
}
```

Present to the user:
- Confirm the image was deleted successfully
- Show the deleted image ID
- If the user was deleting to free space, suggest running `/stats` to check updated usage

**Error** — see `references/auth.md` for common error codes. Additional delete-specific codes:

| `code` | Suggested fix |
|---|---|
| `file_not_found` | The image ID was not found. Use `/files` to list your images and find the correct ID |

If the server returns a non-JSON response or an unexpected HTTP status, tell the user the API may be temporarily unavailable.
