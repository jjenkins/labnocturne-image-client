# Lab Nocturne Images - Python Client

Python client library for the [Lab Nocturne Images API](https://images.labnocturne.com).

## Installation

Install directly from GitHub:

```bash
pip install git+https://github.com/jjenkins/labnocturne-image-client.git#subdirectory=python
```

## Requirements

- Python 3.7+
- requests library (automatically installed)

## Quick Start

```python
from labnocturne import LabNocturneClient

# Generate a test API key
api_key = LabNocturneClient.generate_test_key()
print(f"API Key: {api_key}")

# Create client
client = LabNocturneClient(api_key)

# Upload an image
result = client.upload('photo.jpg')
print(f"Image URL: {result['url']}")
print(f"Image ID: {result['id']}")

# List files
files = client.list_files(limit=10)
print(f"Total files: {files['pagination']['total']}")

# Get usage stats
stats = client.get_stats()
print(f"Storage: {stats['storage_used_mb']:.2f} MB / {stats['quota_mb']} MB")

# Delete a file
client.delete_file(result['id'])
print("File deleted")
```

## API Reference

### `LabNocturneClient(api_key, base_url="https://images.labnocturne.com")`

Create a new client instance.

**Parameters:**
- `api_key` (str): Your API key
- `base_url` (str, optional): Base URL for the API

### Methods

#### `generate_test_key(base_url="https://images.labnocturne.com")` (static)

Generate a test API key for development.

**Returns:** API key string

**Example:**
```python
api_key = LabNocturneClient.generate_test_key()
```

#### `upload(file_path)`

Upload an image file.

**Parameters:**
- `file_path` (str): Path to the image file

**Returns:** Dict with `id`, `url`, `size`, `mime_type`, `created_at`

**Example:**
```python
result = client.upload('photo.jpg')
print(result['url'])
```

#### `list_files(page=1, limit=50, sort='created_desc')`

List uploaded files with pagination.

**Parameters:**
- `page` (int): Page number (default: 1)
- `limit` (int): Files per page (default: 50)
- `sort` (str): Sort order - `created_desc`, `created_asc`, `size_desc`, `size_asc`, `name_asc`, `name_desc`

**Returns:** Dict with `files` array and `pagination` info

**Example:**
```python
files = client.list_files(page=1, limit=10, sort='size_desc')
for file in files['files']:
    print(f"{file['id']}: {file['size']} bytes")
```

#### `get_stats()`

Get usage statistics for your account.

**Returns:** Dict with `storage_used_bytes`, `storage_used_mb`, `file_count`, `quota_bytes`, `quota_mb`, `usage_percent`

**Example:**
```python
stats = client.get_stats()
print(f"Using {stats['usage_percent']:.1f}% of quota")
```

#### `delete_file(image_id)`

Delete an image (soft delete).

**Parameters:**
- `image_id` (str): The image ID

**Returns:** Dict with success status

**Example:**
```python
client.delete_file('img_01jcd8x9k2n...')
```

## Context Manager Support

The client supports Python's context manager protocol:

```python
with LabNocturneClient(api_key) as client:
    result = client.upload('photo.jpg')
    print(result['url'])
# Session is automatically closed
```

## Error Handling

```python
try:
    result = client.upload('photo.jpg')
except FileNotFoundError:
    print("File not found")
except Exception as e:
    print(f"Upload failed: {e}")
```

## Complete Example

```python
from labnocturne import LabNocturneClient

def main():
    # Generate test API key
    print("Generating test API key...")
    api_key = LabNocturneClient.generate_test_key()
    print(f"API Key: {api_key}\n")

    # Use context manager for automatic cleanup
    with LabNocturneClient(api_key) as client:
        # Upload an image
        print("Uploading image...")
        upload_result = client.upload('photo.jpg')
        print(f"Uploaded: {upload_result['url']}")
        print(f"Image ID: {upload_result['id']}")
        print(f"Size: {upload_result['size'] / 1024:.2f} KB\n")

        # List all files
        print("Listing files...")
        files = client.list_files(limit=10)
        print(f"Total files: {files['pagination']['total']}")
        for file in files['files']:
            print(f"  - {file['id']}: {file['size'] / 1024:.2f} KB")
        print()

        # Get usage stats
        print("Usage statistics:")
        stats = client.get_stats()
        print(f"  Storage: {stats['storage_used_mb']:.2f} MB / {stats['quota_mb']} MB")
        print(f"  Files: {stats['file_count']}")
        print(f"  Usage: {stats['usage_percent']:.2f}%\n")

        # Delete the uploaded file
        print("Deleting image...")
        client.delete_file(upload_result['id'])
        print("Deleted successfully")

if __name__ == '__main__':
    main()
```

## Development

Install with development dependencies:

```bash
pip install git+https://github.com/jjenkins/labnocturne-image-client.git#subdirectory=python[dev]
```

Run tests:

```bash
pytest
```

Format code:

```bash
black labnocturne/
```

Type checking:

```bash
mypy labnocturne/
```

## License

MIT License - See [LICENSE](../LICENSE) for details.

## Links

- [Main Repository](https://github.com/jjenkins/labnocturne-image-client)
- [API Documentation](https://images.labnocturne.com/docs)
- [Other Language Clients](https://github.com/jjenkins/labnocturne-image-client#readme)
