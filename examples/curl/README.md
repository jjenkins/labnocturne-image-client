# curl Examples - Lab Nocturne Images API

The simplest way to interact with the Lab Nocturne Images API. Perfect for testing and scripting.

## Prerequisites

Just `curl` - that's it! (Available on macOS, Linux, and Windows with Git Bash)

## Quick Start

### 1. Generate a Test API Key

```bash
curl https://images.labnocturne.com/key
```

Response:
```json
{
  "api_key": "ln_test_01jcd8x9k2...",
  "type": "test",
  "quota": {
    "max_file_size_mb": 10,
    "retention_days": 7
  },
  "created_at": "2025-12-08T10:30:00Z"
}
```

Save your API key for the next steps:
```bash
export API_KEY="ln_test_01jcd8x9k2..."
```

### 2. Upload an Image

```bash
curl -X POST https://images.labnocturne.com/upload \
  -H "Authorization: Bearer $API_KEY" \
  -F "file=@photo.jpg"
```

Response:
```json
{
  "id": "img_01jcd8x9k2n...",
  "url": "https://cdn.labnocturne.com/i/01jcd8x9k2n...jpg",
  "size": 245678,
  "mime_type": "image/jpeg",
  "created_at": "2025-12-08T10:31:00Z"
}
```

### 3. Retrieve Image Info

```bash
curl https://cdn.labnocturne.com/i/01jcd8x9k2n...jpg
```

This returns the actual image file. Use `-I` to see headers only:

```bash
curl -I https://cdn.labnocturne.com/i/01jcd8x9k2n...jpg
```

### 4. List Your Files

```bash
curl -H "Authorization: Bearer $API_KEY" \
  "https://images.labnocturne.com/files?page=1&limit=10"
```

Response:
```json
{
  "files": [
    {
      "id": "img_01jcd8x9k2n...",
      "url": "https://cdn.labnocturne.com/i/01jcd8x9k2n...jpg",
      "size": 245678,
      "mime_type": "image/jpeg",
      "created_at": "2025-12-08T10:31:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 10,
    "total": 1,
    "total_pages": 1
  }
}
```

### 5. Get Usage Statistics

```bash
curl -H "Authorization: Bearer $API_KEY" \
  https://images.labnocturne.com/stats
```

Response:
```json
{
  "storage_used_bytes": 245678,
  "storage_used_mb": 0.23,
  "file_count": 1,
  "quota_bytes": 10485760,
  "quota_mb": 10,
  "usage_percent": 2.34
}
```

### 6. Delete an Image

```bash
curl -X DELETE \
  -H "Authorization: Bearer $API_KEY" \
  https://images.labnocturne.com/i/img_01jcd8x9k2n...
```

Response:
```json
{
  "success": true,
  "message": "File deleted successfully",
  "id": "img_01jcd8x9k2n..."
}
```

## Advanced Examples

### Upload with Pretty JSON Output

```bash
curl -X POST https://images.labnocturne.com/upload \
  -H "Authorization: Bearer $API_KEY" \
  -F "file=@photo.jpg" | jq '.'
```

Note: Requires `jq` for JSON formatting

### Upload Multiple Files in a Script

```bash
#!/bin/bash
API_KEY="ln_test_01jcd8x9k2..."

for file in *.jpg; do
  echo "Uploading $file..."
  curl -X POST https://images.labnocturne.com/upload \
    -H "Authorization: Bearer $API_KEY" \
    -F "file=@$file"
  echo ""
done
```

### List Files with Sorting

```bash
# Sort by creation date (newest first)
curl -H "Authorization: Bearer $API_KEY" \
  "https://images.labnocturne.com/files?sort=created_desc"

# Sort by size (largest first)
curl -H "Authorization: Bearer $API_KEY" \
  "https://images.labnocturne.com/files?sort=size_desc"

# Sort by filename
curl -H "Authorization: Bearer $API_KEY" \
  "https://images.labnocturne.com/files?sort=name_asc"
```

### Download an Image

```bash
# Download to specific filename
curl -o my-image.jpg https://cdn.labnocturne.com/i/01jcd8x9k2n...jpg

# Download preserving original filename
curl -O -J https://cdn.labnocturne.com/i/01jcd8x9k2n...jpg
```

### Check API Health

```bash
curl https://images.labnocturne.com/health
```

Response:
```json
{
  "status": "ok",
  "timestamp": "2025-12-08T10:35:00Z"
}
```

## Error Handling

### Invalid API Key

```bash
curl -X POST https://images.labnocturne.com/upload \
  -H "Authorization: Bearer invalid_key" \
  -F "file=@photo.jpg"
```

Response (401):
```json
{
  "error": {
    "message": "Invalid API key",
    "type": "unauthorized",
    "code": "invalid_api_key"
  }
}
```

### File Too Large

```bash
curl -X POST https://images.labnocturne.com/upload \
  -H "Authorization: Bearer $API_KEY" \
  -F "file=@huge-file.jpg"
```

Response (413):
```json
{
  "error": {
    "message": "File size exceeds limit for test keys (10MB)",
    "type": "file_too_large",
    "code": "file_size_exceeded"
  }
}
```

### File Not Found

```bash
curl -X DELETE \
  -H "Authorization: Bearer $API_KEY" \
  https://images.labnocturne.com/i/img_invalid
```

Response (404):
```json
{
  "error": {
    "message": "File not found",
    "type": "not_found",
    "code": "file_not_found"
  }
}
```

## Complete Workflow Script

Save this as `image-upload.sh`:

```bash
#!/bin/bash

# Configuration
API_BASE="https://images.labnocturne.com"
IMAGE_FILE="${1:-photo.jpg}"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}Lab Nocturne Image Upload Script${NC}\n"

# Step 1: Get API Key
echo -e "${GREEN}1. Generating test API key...${NC}"
KEY_RESPONSE=$(curl -s "$API_BASE/key")
API_KEY=$(echo $KEY_RESPONSE | jq -r '.api_key')

if [ -z "$API_KEY" ] || [ "$API_KEY" = "null" ]; then
  echo -e "${RED}Failed to generate API key${NC}"
  exit 1
fi

echo "API Key: $API_KEY"

# Step 2: Upload Image
echo -e "\n${GREEN}2. Uploading image: $IMAGE_FILE${NC}"
UPLOAD_RESPONSE=$(curl -s -X POST "$API_BASE/upload" \
  -H "Authorization: Bearer $API_KEY" \
  -F "file=@$IMAGE_FILE")

IMAGE_URL=$(echo $UPLOAD_RESPONSE | jq -r '.url')
IMAGE_ID=$(echo $UPLOAD_RESPONSE | jq -r '.id')

if [ -z "$IMAGE_URL" ] || [ "$IMAGE_URL" = "null" ]; then
  echo -e "${RED}Upload failed${NC}"
  echo $UPLOAD_RESPONSE | jq '.'
  exit 1
fi

echo "Image uploaded successfully!"
echo "URL: $IMAGE_URL"
echo "ID: $IMAGE_ID"

# Step 3: Get Stats
echo -e "\n${GREEN}3. Current usage statistics:${NC}"
curl -s -H "Authorization: Bearer $API_KEY" "$API_BASE/stats" | jq '.'

echo -e "\n${GREEN}Done!${NC}"
```

Make it executable and run:
```bash
chmod +x image-upload.sh
./image-upload.sh my-photo.jpg
```

## Tips and Best Practices

1. **Store API Keys Securely**: Use environment variables, not hardcoded strings
   ```bash
   echo 'export LABNOCTURNE_API_KEY="ln_test_..."' >> ~/.bashrc
   source ~/.bashrc
   ```

2. **Use `-f` for Error Detection**: Add `-f` flag to fail on HTTP errors
   ```bash
   curl -f -X POST ... || echo "Upload failed!"
   ```

3. **Progress Bar for Large Files**: Use `--progress-bar`
   ```bash
   curl --progress-bar -X POST ... -F "file=@large.jpg"
   ```

4. **Verbose Output for Debugging**: Use `-v`
   ```bash
   curl -v -X POST ...
   ```

5. **Save Response Headers**: Use `-D` to save headers
   ```bash
   curl -D headers.txt -X POST ...
   ```

## Next Steps

- Check out other language examples: [JavaScript](../javascript/), [Python](../python/), [Go](../go/)
- Read the full [API Documentation](https://images.labnocturne.com/docs)
- Explore the [main repository](https://github.com/jjenkins/labnocturne)
