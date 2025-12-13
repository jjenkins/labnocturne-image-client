# Lab Nocturne Images - JavaScript/Node.js Client

JavaScript/Node.js client library for the [Lab Nocturne Images API](https://images.labnocturne.com).

## Installation

Due to npm's subdirectory limitations, install locally:

```bash
# Clone the repository
git clone https://github.com/jjenkins/labnocturne-image-client.git
cd labnocturne-image-client/javascript

# Install dependencies and link globally
npm install
npm link

# In your project directory
npm link labnocturne
```

**Alternative:** Use path dependency in your `package.json`:

```json
{
  "dependencies": {
    "labnocturne": "file:../labnocturne-image-client/javascript"
  }
}
```

## Requirements

- Node.js 18.0.0+
- form-data package (automatically installed)

## Quick Start

```javascript
import LabNocturneClient from 'labnocturne';

// Generate a test API key
const apiKey = await LabNocturneClient.generateTestKey();
console.log('API Key:', apiKey);

// Create client
const client = new LabNocturneClient(apiKey);

// Upload an image
const result = await client.upload('photo.jpg');
console.log('Image URL:', result.url);
console.log('Image ID:', result.id);

// List files
const files = await client.listFiles({ limit: 10 });
console.log('Total files:', files.pagination.total);

// Get usage stats
const stats = await client.getStats();
console.log(`Storage: ${stats.storage_used_mb.toFixed(2)} MB / ${stats.quota_mb} MB`);

// Delete a file
await client.deleteFile(result.id);
console.log('File deleted');
```

## API Reference

### `new LabNocturneClient(apiKey, baseUrl)`

Create a new client instance.

**Parameters:**
- `apiKey` (string): Your API key
- `baseUrl` (string, optional): Base URL for the API (default: `https://images.labnocturne.com`)

### Methods

#### `static async generateTestKey(baseUrl)`

Generate a test API key for development.

**Parameters:**
- `baseUrl` (string, optional): Base URL for the API

**Returns:** Promise<string> - API key

**Example:**
```javascript
const apiKey = await LabNocturneClient.generateTestKey();
```

#### `async upload(filePath)`

Upload an image file.

**Parameters:**
- `filePath` (string): Path to the image file

**Returns:** Promise<UploadResponse> - Object with `id`, `url`, `size`, `mime_type`, `created_at`

**Example:**
```javascript
const result = await client.upload('photo.jpg');
console.log(result.url);
```

#### `async listFiles(options)`

List uploaded files with pagination.

**Parameters:**
- `options` (object, optional):
  - `page` (number): Page number (default: 1)
  - `limit` (number): Files per page (default: 50)
  - `sort` (string): Sort order - `created_desc`, `created_asc`, `size_desc`, `size_asc`, `name_asc`, `name_desc`

**Returns:** Promise<ListFilesResponse> - Object with `files` array and `pagination` info

**Example:**
```javascript
const files = await client.listFiles({ page: 1, limit: 10, sort: 'size_desc' });
files.files.forEach(file => {
  console.log(`${file.id}: ${file.size} bytes`);
});
```

#### `async getStats()`

Get usage statistics for your account.

**Returns:** Promise<StatsResponse> - Object with `storage_used_bytes`, `storage_used_mb`, `file_count`, `quota_bytes`, `quota_mb`, `usage_percent`

**Example:**
```javascript
const stats = await client.getStats();
console.log(`Using ${stats.usage_percent.toFixed(1)}% of quota`);
```

#### `async deleteFile(imageId)`

Delete an image (soft delete).

**Parameters:**
- `imageId` (string): The image ID

**Returns:** Promise<void>

**Example:**
```javascript
await client.deleteFile('img_01jcd8x9k2n...');
```

## TypeScript Support

This package includes TypeScript definitions. No additional types package needed.

```typescript
import LabNocturneClient, { UploadResponse, StatsResponse } from 'labnocturne';

const client = new LabNocturneClient(apiKey);
const result: UploadResponse = await client.upload('photo.jpg');
const stats: StatsResponse = await client.getStats();
```

## Error Handling

```javascript
try {
  const result = await client.upload('photo.jpg');
  console.log('Success:', result.url);
} catch (error) {
  if (error.code === 'ENOENT') {
    console.error('File not found');
  } else {
    console.error('Upload failed:', error.message);
  }
}
```

## Complete Example

```javascript
import LabNocturneClient from 'labnocturne';

async function main() {
  try {
    // Generate test API key
    console.log('Generating test API key...');
    const apiKey = await LabNocturneClient.generateTestKey();
    console.log('API Key:', apiKey);
    console.log();

    // Create client
    const client = new LabNocturneClient(apiKey);

    // Upload an image
    console.log('Uploading image...');
    const upload = await client.upload('photo.jpg');
    console.log('Uploaded:', upload.url);
    console.log('Image ID:', upload.id);
    console.log('Size:', (upload.size / 1024).toFixed(2), 'KB');
    console.log();

    // List all files
    console.log('Listing files...');
    const files = await client.listFiles({ limit: 10 });
    console.log('Total files:', files.pagination.total);
    files.files.forEach(file => {
      console.log(`  - ${file.id}: ${(file.size / 1024).toFixed(2)} KB`);
    });
    console.log();

    // Get usage stats
    console.log('Usage statistics:');
    const stats = await client.getStats();
    console.log(`  Storage: ${stats.storage_used_mb.toFixed(2)} MB / ${stats.quota_mb} MB`);
    console.log(`  Files: ${stats.file_count}`);
    console.log(`  Usage: ${stats.usage_percent.toFixed(2)}%`);
    console.log();

    // Delete the uploaded file
    console.log('Deleting image...');
    await client.deleteFile(upload.id);
    console.log('Deleted successfully');

  } catch (error) {
    console.error('Error:', error.message);
  }
}

main();
```

Save as `example.js` and run:

```bash
node example.js
```

## Browser Usage

For browser environments, you'll need to adapt the upload method as `fs` is not available:

```javascript
// Browser-compatible upload using File object
async uploadFromBrowser(file) {
  const formData = new FormData();
  formData.append('file', file);

  const response = await fetch(`${this.baseUrl}/upload`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${this.apiKey}`
    },
    body: formData
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(`Upload failed: ${error.error.message}`);
  }

  return await response.json();
}
```

## License

MIT License - See [LICENSE](../LICENSE) for details.

## Links

- [Main Repository](https://github.com/jjenkins/labnocturne-image-client)
- [API Documentation](https://images.labnocturne.com/docs)
- [Other Language Clients](https://github.com/jjenkins/labnocturne-image-client#readme)
