# Lab Nocturne Images API - Client Examples

Developer-friendly code examples for integrating with the Lab Nocturne Images API. Get up and running in under 60 seconds.

## Quick Start

The Lab Nocturne Images API is a simple, curl-first image storage service. No dashboards, no configuration required.

### 1. Get a Test API Key

```bash
curl https://images.labnocturne.com/key
```

Response:
```json
{
  "api_key": "ln_test_abc123...",
  "type": "test",
  "expires_in": "7 days"
}
```

### 2. Upload an Image

```bash
curl -X POST https://images.labnocturne.com/upload \
  -H "Authorization: Bearer ln_test_abc123..." \
  -F "file=@photo.jpg"
```

Response:
```json
{
  "id": "img_01jcd8x9k2n...",
  "url": "https://cdn.labnocturne.com/i/01jcd8x9k2n...jpg",
  "size": 245678,
  "mime_type": "image/jpeg"
}
```

### 3. Use Your Image

The URL returned is a CloudFront CDN URL - just use it directly in your HTML, app, or anywhere you need it:

```html
<img src="https://cdn.labnocturne.com/i/01jcd8x9k2n...jpg" alt="My image">
```

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/key` | Generate a test API key |
| POST | `/upload` | Upload an image file |
| GET | `/i/:id` | Retrieve an image (redirects to CDN) |
| GET | `/files` | List your uploaded files |
| GET | `/stats` | Get usage statistics |
| DELETE | `/i/:id` | Delete an image (soft delete) |

## Language Examples

Choose your language to see complete working examples:

- **[curl](examples/curl/)** - Command-line examples (start here!)
- **[JavaScript/Node.js](examples/javascript/)** - Node.js and browser examples
- **[Python](examples/python/)** - Using the requests library
- **[Go](examples/go/)** - Native Go client
- **[Ruby](examples/ruby/)** - Using net/http
- **[PHP](examples/php/)** - Using cURL and Guzzle

## Key Features

### Test Keys
- Generate instantly without signup
- Prefix: `ln_test_*`
- 10MB file size limit
- Files expire after 7 days
- Perfect for development and testing

### Live Keys
- Require email + payment
- Prefix: `ln_live_*`
- 100MB file size limit
- Files stored permanently
- Production-ready

## Common Operations

### Upload an Image
```bash
POST /upload
Authorization: Bearer ln_test_abc123...
Content-Type: multipart/form-data

file=@photo.jpg
```

### List Images
```bash
GET /files?page=1&limit=50&sort=created_desc
Authorization: Bearer ln_test_abc123...
```

### Get Usage Stats
```bash
GET /stats
Authorization: Bearer ln_test_abc123...
```

Returns:
```json
{
  "storage_used_bytes": 1234567,
  "storage_used_mb": 1.18,
  "file_count": 42,
  "quota_bytes": 10485760,
  "quota_mb": 10
}
```

### Delete an Image
```bash
DELETE /i/img_01jcd8x9k2n...
Authorization: Bearer ln_test_abc123...
```

## Error Handling

All errors return JSON with helpful information:

```json
{
  "error": {
    "message": "File size exceeds limit for test keys (10MB)",
    "type": "file_too_large",
    "code": "file_size_exceeded"
  }
}
```

Common HTTP status codes:
- `200` - Success
- `400` - Bad request (invalid parameters)
- `401` - Unauthorized (invalid API key)
- `413` - File too large
- `500` - Server error

## File Formats

Supported image formats:
- JPEG (`.jpg`, `.jpeg`)
- PNG (`.png`)
- GIF (`.gif`)
- WebP (`.webp`)

## Limits

### Test Keys (`ln_test_*`)
- Max file size: 10MB
- File retention: 7 days
- No cost

### Live Keys (`ln_live_*`)
- Max file size: 100MB
- File retention: Permanent
- Pay as you go pricing

## Need Help?

- API Documentation: https://images.labnocturne.com/docs
- GitHub Issues: https://github.com/jjenkins/labnocturne-image-client/issues

## License

MIT License - See LICENSE file for details
