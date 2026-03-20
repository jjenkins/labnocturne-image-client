# Lab Nocturne Images

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![API Status](https://img.shields.io/badge/API-Online-success)](https://images.labnocturne.com)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://github.com/jjenkins/agent-image-skills/pulls)

Image storage for AI agents. Upload, retrieve, and manage images with a single curl call — no dashboards, no signup, no configuration.

```bash
# Get a test key (no signup required)
curl https://images.labnocturne.com/key

# Upload an image
curl -X POST https://images.labnocturne.com/upload \
  -H "Authorization: Bearer ln_test_abc123..." \
  -F "file=@photo.jpg"

# Returns: {"id":"img_xyz","url":"https://cdn.labnocturne.com/i/01jcd8x9k2n...jpg"}
```

<!-- ![Demo](docs/demo.gif) -->

## The Problem

Agents like Claude Code, ChatGPT, and custom AI workflows generate images all day — charts, screenshots, mockups, visual diffs. But where do they go?

- **Cloudinary** is $99/month after the free tier runs out
- **Imgur** fired all US staff and the service is degrading
- **ImageKit** is $89/month and built for marketers, not automation
- **Self-hosting** S3 requires AWS expertise and infrastructure management

AI agents don't need dashboards. They need an API that just works.

## Why Lab Nocturne?

Built for automation, not humans.

| Feature | Lab Nocturne | Cloudinary | Imgur | ImageKit |
|---------|--------------|------------|-------|----------|
| **Test without signup** | ✅ Instant test keys | ❌ Email required | ⚠️ Sign-up flow | ❌ Email required |
| **Ephemeral storage** | ✅ 7-day auto-cleanup | ❌ Pay for old files | ❌ No control | ❌ Manual deletion |
| **Transparent pricing** | ✅ $5/mo → $20/mo | ⚠️ Free → $99/mo jump | ⚠️ Unstable | ⚠️ $89/mo → $249/mo |
| **Agent-first design** | ✅ Claude Code skills | ❌ Dashboard-only | ❌ No API docs | ❌ Marketing tools |
| **API reliability** | ✅ CloudFront CDN | ✅ Good | ⚠️ Degrading | ✅ Good |
| **Webhook callbacks** | 🔜 Coming soon | ✅ Yes | ❌ No | ✅ Yes |

## Why Agents Need Image Storage

- **Memory and recall** — A user sends a screenshot to an agent on Discord and asks about it days later. Agents need persistent storage to hold visual context across conversations.
- **Sharing and collaboration** — An agent generates a chart, uploads it, and hands back a CDN URL. The user drops the link in Slack, a doc, or an email — no manual download/re-upload step.
- **Asset management** — Coding agents like Claude Code working on web projects need a place to host images during development: logos, screenshots, mockups.
- **Transient artifacts** — Test keys give agents 7-day ephemeral storage. Perfect for one-off visualizations, debug screenshots, or CI artifacts that don't need to live forever.

## Agent Integrations

### Claude Code (Built-in Skills)

Install all five skills with the MCP skills CLI:

```bash
npx skills add jjenkins/agent-image-skills
```

Available commands:
- `/upload` - Upload an image and get a CDN URL
- `/files` - List all uploaded images
- `/stats` - Check storage usage
- `/delete` - Remove an image
- `/generate-key` - Create a new test API key

Or clone the repo and the skills are available automatically when Claude Code runs from the project directory:

```bash
git clone https://github.com/jjenkins/agent-image-skills.git
cd agent-image-skills
# /upload, /files, /stats, /delete, /generate-key are now available
```

Set `LABNOCTURNE_API_KEY` in your environment to use a specific key, or leave it unset and the skills will auto-generate a test key.

### ChatGPT (GPT Action)

Add Lab Nocturne as a GPT Action using the OpenAPI schema in [`integrations/chatgpt-action/`](integrations/chatgpt-action/).

### Any Agent (curl)

One API call is all it takes:

```bash
curl -X POST https://images.labnocturne.com/upload \
  -H "Authorization: Bearer $LABNOCTURNE_API_KEY" \
  -F "file=@screenshot.png"
```

## Quick Start Examples

### Python

```python
from labnocturne import LabNocturneClient

# Generate test API key (no signup required)
api_key = LabNocturneClient.generate_test_key()

# Create client and upload
client = LabNocturneClient(api_key)
result = client.upload('photo.jpg')
print(f"Image URL: {result['url']}")
# Output: https://cdn.labnocturne.com/i/01jcd8x9k2n...jpg
```

[Full Python Documentation →](python/README.md)

### Go

```go
import "github.com/jjenkins/agent-image-skills/go/labnocturne"

// Generate test API key
apiKey, _ := labnocturne.GenerateTestKey()

// Create client and upload
client := labnocturne.NewClient(apiKey)
result, _ := client.Upload("photo.jpg")
fmt.Println("Image URL:", result.URL)
```

[Full Go Documentation →](go/README.md)

### Ruby

```ruby
require 'labnocturne'

# Generate test API key
api_key = LabNocturne::Client.generate_test_key

# Create client and upload
client = LabNocturne::Client.new(api_key)
result = client.upload('photo.jpg')
puts "Image URL: #{result['url']}"
```

[Full Ruby Documentation →](ruby/README.md)

### JavaScript/Node.js

```javascript
import LabNocturneClient from 'labnocturne';

// Generate test API key
const apiKey = await LabNocturneClient.generateTestKey();

// Create client and upload
const client = new LabNocturneClient(apiKey);
const result = await client.upload('photo.jpg');
console.log('Image URL:', result.url);
```

[Full JavaScript Documentation →](javascript/README.md)

### PHP

```php
use LabNocturne\LabNocturneClient;

// Generate test API key
$apiKey = LabNocturneClient::generateTestKey();

// Create client and upload
$client = new LabNocturneClient($apiKey);
$result = $client->upload('photo.jpg');
echo "Image URL: {$result['url']}\n";
```

[Full PHP Documentation →](php/README.md)

## Installation

| Language | Method |
|----------|--------|
| **Python** | `pip install git+https://github.com/jjenkins/agent-image-skills.git#subdirectory=python` |
| **Go** | `go get github.com/jjenkins/agent-image-skills/go/labnocturne` |
| **Ruby** | Add to Gemfile: `gem 'labnocturne', git: 'https://github.com/jjenkins/agent-image-skills', glob: 'ruby/*.gemspec'` |
| **JavaScript** | Clone repo, then `cd javascript && npm install && npm link` |
| **PHP** | `composer require labnocturne/image-client` |

## API Overview

All client libraries implement the same core methods:

| Method | Description |
|--------|-------------|
| `generateTestKey()` | Generate a test API key (static method, no auth required) |
| `upload(filePath)` | Upload an image file, returns CDN URL |
| `listFiles(page, limit, sort)` | List uploaded files with pagination |
| `getStats()` | Get usage statistics (file count, storage used) |
| `deleteFile(imageId)` | Delete an image (soft delete, 30-day recovery) |

## How It Works

### 1. Get a Test API Key

No signup required for testing:

```bash
curl https://images.labnocturne.com/key
```

Returns a test key with:
- 10MB file size limit
- 7-day file retention (auto-cleanup)
- Perfect for development and CI/CD

### 2. Upload an Image

```bash
curl -X POST https://images.labnocturne.com/upload \
  -H "Authorization: Bearer ln_test_abc123..." \
  -F "file=@photo.jpg"
```

Returns:
```json
{
  "id": "img_01jcd8x9k2n3p4q5r6s7t8u9v0",
  "url": "https://cdn.labnocturne.com/i/01jcd8x9k2n...jpg",
  "size": 245678,
  "uploaded_at": "2026-03-19T17:22:00Z"
}
```

### 3. Use Your Image

```html
<img src="https://cdn.labnocturne.com/i/01jcd8x9k2n...jpg" alt="My image">
```

The URL is a CloudFront CDN URL - fast, global delivery with 99.9% uptime.

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/key` | Generate a test API key |
| POST | `/upload` | Upload an image file |
| GET | `/i/:id` | Retrieve an image (redirects to CDN) |
| GET | `/files` | List your uploaded files |
| GET | `/stats` | Get usage statistics |
| DELETE | `/i/:id` | Delete an image (soft delete) |

## Pricing

| Tier | Price | File Size Limit | Storage | Retention |
|------|-------|-----------------|---------|-----------|
| **Test** | Free | 10MB | 100MB | 7 days (auto-cleanup) |
| **Starter** | $5/mo | 100MB | 10GB | Permanent |
| **Pro** | $20/mo | 100MB | 100GB | Permanent |

No surprise charges. No usage-based pricing. No bandwidth fees.

## Key Features

### Test Keys (`ln_test_*`)
- Generate instantly without signup
- 10MB file size limit
- Files expire after 7 days
- Perfect for development and testing

### Live Keys (`ln_live_*`)
- Require email + payment
- 100MB file size limit
- Files stored permanently
- Production-ready

## Supported Image Formats

- JPEG (`.jpg`, `.jpeg`)
- PNG (`.png`)
- GIF (`.gif`)
- WebP (`.webp`)

## Error Handling

All APIs return JSON errors with helpful messages:

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

## Language-Specific Documentation

Detailed documentation for each language:

- **[Python](python/README.md)** - Full Python client documentation with examples
- **[Go](go/README.md)** - Full Go client documentation with examples
- **[Ruby](ruby/README.md)** - Full Ruby client documentation with examples
- **[JavaScript/Node.js](javascript/README.md)** - Full JavaScript client documentation with examples
- **[PHP](php/README.md)** - Full PHP client documentation with examples
- **[curl](examples/curl/README.md)** - Command-line examples for testing

## Roadmap

- [x] Test keys with auto-cleanup
- [x] Multi-language SDKs (Python, Go, Ruby, JS, PHP)
- [x] Claude Code MCP skills
- [x] ChatGPT GPT Action integration
- [ ] Webhook callbacks for upload completion
- [ ] Image transformations (resize, crop, optimize)
- [ ] Temporary URLs (signed, expiring links)
- [ ] Bulk upload API

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - See [LICENSE](LICENSE) for details.

## Links

- **Website**: https://images.labnocturne.com
- **API Documentation**: https://images.labnocturne.com/docs
- **GitHub Issues**: https://github.com/jjenkins/agent-image-skills/issues

## Need Help?

Check the language-specific README files for detailed examples and usage instructions, or open an issue on GitHub.
