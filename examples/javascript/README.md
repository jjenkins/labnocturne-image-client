# JavaScript/Node.js Examples - Lab Nocturne Images API

Complete examples for using the Lab Nocturne Images API in Node.js and browser environments.

## Prerequisites

- Node.js 18+ (for Node.js examples)
- Modern browser (for browser examples)

## Installation

No SDK required! Just use native `fetch` API (Node.js 18+) or any HTTP client.

Optional: Install form-data for multipart uploads in Node.js:
```bash
npm install form-data
```

## Node.js Examples

### Basic Upload Example

```javascript
// upload.js
import fs from 'fs';
import FormData from 'form-data';

const API_BASE = 'https://images.labnocturne.com';

async function uploadImage(apiKey, filePath) {
  const form = new FormData();
  form.append('file', fs.createReadStream(filePath));

  const response = await fetch(`${API_BASE}/upload`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${apiKey}`,
      ...form.getHeaders()
    },
    body: form
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error.message);
  }

  return await response.json();
}

// Usage
const apiKey = 'ln_test_01jcd8x9k2...';
const result = await uploadImage(apiKey, './photo.jpg');
console.log('Image URL:', result.url);
console.log('Image ID:', result.id);
```

### Complete Client Class

```javascript
// labnocturne-client.js
import fs from 'fs';
import FormData from 'form-data';

class LabNocturneClient {
  constructor(apiKey, baseUrl = 'https://images.labnocturne.com') {
    this.apiKey = apiKey;
    this.baseUrl = baseUrl;
  }

  async request(endpoint, options = {}) {
    const url = `${this.baseUrl}${endpoint}`;
    const headers = {
      'Authorization': `Bearer ${this.apiKey}`,
      ...options.headers
    };

    const response = await fetch(url, {
      ...options,
      headers
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(`API Error: ${error.error.message} (${error.error.code})`);
    }

    return await response.json();
  }

  async upload(filePath) {
    const form = new FormData();
    form.append('file', fs.createReadStream(filePath));

    const response = await fetch(`${this.baseUrl}/upload`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${this.apiKey}`,
        ...form.getHeaders()
      },
      body: form
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(`Upload failed: ${error.error.message}`);
    }

    return await response.json();
  }

  async listFiles(options = {}) {
    const params = new URLSearchParams({
      page: options.page || 1,
      limit: options.limit || 50,
      sort: options.sort || 'created_desc'
    });

    return await this.request(`/files?${params}`);
  }

  async getStats() {
    return await this.request('/stats');
  }

  async deleteFile(imageId) {
    return await this.request(`/i/${imageId}`, {
      method: 'DELETE'
    });
  }

  static async generateTestKey(baseUrl = 'https://images.labnocturne.com') {
    const response = await fetch(`${baseUrl}/key`);
    if (!response.ok) {
      throw new Error('Failed to generate API key');
    }
    const data = await response.json();
    return data.api_key;
  }
}

export default LabNocturneClient;
```

### Usage Example

```javascript
// example.js
import LabNocturneClient from './labnocturne-client.js';

async function main() {
  // Generate a test API key
  const apiKey = await LabNocturneClient.generateTestKey();
  console.log('Generated API key:', apiKey);

  // Create client
  const client = new LabNocturneClient(apiKey);

  // Upload an image
  console.log('\nUploading image...');
  const upload = await client.upload('./photo.jpg');
  console.log('Uploaded:', upload.url);
  console.log('Image ID:', upload.id);
  console.log('Size:', (upload.size / 1024).toFixed(2), 'KB');

  // List all files
  console.log('\nListing files...');
  const files = await client.listFiles({ limit: 10 });
  console.log('Total files:', files.pagination.total);
  files.files.forEach(file => {
    console.log(`- ${file.id}: ${(file.size / 1024).toFixed(2)} KB`);
  });

  // Get usage stats
  console.log('\nUsage statistics:');
  const stats = await client.getStats();
  console.log(`Storage: ${stats.storage_used_mb.toFixed(2)} MB / ${stats.quota_mb} MB`);
  console.log(`Files: ${stats.file_count}`);
  console.log(`Usage: ${stats.usage_percent.toFixed(2)}%`);

  // Delete a file
  console.log('\nDeleting image...');
  await client.deleteFile(upload.id);
  console.log('Deleted successfully');
}

main().catch(console.error);
```

Run it:
```bash
node example.js
```

### Using Native Fetch (Node.js 18+)

```javascript
// Simple upload without dependencies
import fs from 'fs';
import { Blob } from 'buffer';

async function uploadWithFetch(apiKey, filePath) {
  const fileBuffer = fs.readFileSync(filePath);
  const blob = new Blob([fileBuffer]);

  const formData = new FormData();
  formData.append('file', blob, filePath);

  const response = await fetch('https://images.labnocturne.com/upload', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${apiKey}`
    },
    body: formData
  });

  return await response.json();
}
```

## Browser Examples

### Upload with File Input

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Lab Nocturne Image Upload</title>
</head>
<body>
  <h1>Image Upload</h1>

  <input type="text" id="apiKey" placeholder="Enter API key" style="width: 300px;">
  <button onclick="generateKey()">Generate Test Key</button>

  <br><br>

  <input type="file" id="fileInput" accept="image/*">
  <button onclick="uploadImage()">Upload</button>

  <div id="result"></div>
  <div id="preview"></div>

  <script>
    const API_BASE = 'https://images.labnocturne.com';

    async function generateKey() {
      try {
        const response = await fetch(`${API_BASE}/key`);
        const data = await response.json();
        document.getElementById('apiKey').value = data.api_key;
        showResult('Test key generated: ' + data.api_key);
      } catch (error) {
        showResult('Error: ' + error.message, true);
      }
    }

    async function uploadImage() {
      const apiKey = document.getElementById('apiKey').value;
      const fileInput = document.getElementById('fileInput');
      const file = fileInput.files[0];

      if (!apiKey) {
        showResult('Please enter an API key', true);
        return;
      }

      if (!file) {
        showResult('Please select a file', true);
        return;
      }

      const formData = new FormData();
      formData.append('file', file);

      try {
        showResult('Uploading...');

        const response = await fetch(`${API_BASE}/upload`, {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${apiKey}`
          },
          body: formData
        });

        if (!response.ok) {
          const error = await response.json();
          throw new Error(error.error.message);
        }

        const result = await response.json();
        showResult(`Upload successful! Image ID: ${result.id}`);
        showPreview(result.url);
      } catch (error) {
        showResult('Error: ' + error.message, true);
      }
    }

    function showResult(message, isError = false) {
      const resultDiv = document.getElementById('result');
      resultDiv.textContent = message;
      resultDiv.style.color = isError ? 'red' : 'green';
      resultDiv.style.marginTop = '10px';
    }

    function showPreview(url) {
      const previewDiv = document.getElementById('preview');
      previewDiv.innerHTML = `
        <h3>Uploaded Image:</h3>
        <img src="${url}" style="max-width: 500px; margin-top: 10px;">
        <p><a href="${url}" target="_blank">${url}</a></p>
      `;
    }
  </script>
</body>
</html>
```

### Drag and Drop Upload

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Drag & Drop Upload</title>
  <style>
    #dropZone {
      border: 3px dashed #ccc;
      border-radius: 10px;
      padding: 50px;
      text-align: center;
      font-size: 18px;
      color: #666;
      cursor: pointer;
      transition: all 0.3s;
    }
    #dropZone.dragover {
      border-color: #4CAF50;
      background-color: #f0f8f0;
    }
    #gallery img {
      max-width: 200px;
      margin: 10px;
      border-radius: 5px;
      box-shadow: 0 2px 5px rgba(0,0,0,0.2);
    }
  </style>
</head>
<body>
  <h1>Drag & Drop Image Upload</h1>

  <input type="text" id="apiKey" placeholder="API key">
  <button onclick="generateKey()">Generate Test Key</button>

  <div id="dropZone">
    Drag and drop images here, or click to select
  </div>

  <div id="status"></div>
  <div id="gallery"></div>

  <script>
    const API_BASE = 'https://images.labnocturne.com';
    const dropZone = document.getElementById('dropZone');

    async function generateKey() {
      const response = await fetch(`${API_BASE}/key`);
      const data = await response.json();
      document.getElementById('apiKey').value = data.api_key;
    }

    dropZone.addEventListener('dragover', (e) => {
      e.preventDefault();
      dropZone.classList.add('dragover');
    });

    dropZone.addEventListener('dragleave', () => {
      dropZone.classList.remove('dragover');
    });

    dropZone.addEventListener('drop', async (e) => {
      e.preventDefault();
      dropZone.classList.remove('dragover');

      const files = Array.from(e.dataTransfer.files);
      for (const file of files) {
        if (file.type.startsWith('image/')) {
          await uploadFile(file);
        }
      }
    });

    dropZone.addEventListener('click', () => {
      const input = document.createElement('input');
      input.type = 'file';
      input.accept = 'image/*';
      input.multiple = true;
      input.onchange = async (e) => {
        for (const file of e.target.files) {
          await uploadFile(file);
        }
      };
      input.click();
    });

    async function uploadFile(file) {
      const apiKey = document.getElementById('apiKey').value;
      if (!apiKey) {
        alert('Please generate an API key first');
        return;
      }

      const statusDiv = document.getElementById('status');
      statusDiv.textContent = `Uploading ${file.name}...`;

      const formData = new FormData();
      formData.append('file', file);

      try {
        const response = await fetch(`${API_BASE}/upload`, {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${apiKey}`
          },
          body: formData
        });

        const result = await response.json();

        if (!response.ok) {
          throw new Error(result.error.message);
        }

        statusDiv.textContent = `Uploaded: ${file.name}`;
        addToGallery(result.url, result.id);
      } catch (error) {
        statusDiv.textContent = `Error: ${error.message}`;
      }
    }

    function addToGallery(url, id) {
      const gallery = document.getElementById('gallery');
      const img = document.createElement('img');
      img.src = url;
      img.title = id;
      gallery.appendChild(img);
    }
  </script>
</body>
</html>
```

### React Component

```jsx
// ImageUploader.jsx
import { useState } from 'react';

const API_BASE = 'https://images.labnocturne.com';

export default function ImageUploader() {
  const [apiKey, setApiKey] = useState('');
  const [uploading, setUploading] = useState(false);
  const [images, setImages] = useState([]);
  const [error, setError] = useState('');

  const generateKey = async () => {
    try {
      const response = await fetch(`${API_BASE}/key`);
      const data = await response.json();
      setApiKey(data.api_key);
      setError('');
    } catch (err) {
      setError('Failed to generate key');
    }
  };

  const uploadImage = async (file) => {
    if (!apiKey) {
      setError('Please generate an API key first');
      return;
    }

    setUploading(true);
    setError('');

    const formData = new FormData();
    formData.append('file', file);

    try {
      const response = await fetch(`${API_BASE}/upload`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${apiKey}`
        },
        body: formData
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error.message);
      }

      const result = await response.json();
      setImages([...images, result]);
    } catch (err) {
      setError(err.message);
    } finally {
      setUploading(false);
    }
  };

  const handleFileChange = (e) => {
    const file = e.target.files[0];
    if (file) {
      uploadImage(file);
    }
  };

  return (
    <div className="image-uploader">
      <h2>Lab Nocturne Image Uploader</h2>

      <div>
        <input
          type="text"
          value={apiKey}
          onChange={(e) => setApiKey(e.target.value)}
          placeholder="API Key"
          style={{ width: '300px', marginRight: '10px' }}
        />
        <button onClick={generateKey}>Generate Test Key</button>
      </div>

      <div style={{ marginTop: '20px' }}>
        <input
          type="file"
          accept="image/*"
          onChange={handleFileChange}
          disabled={!apiKey || uploading}
        />
        {uploading && <span> Uploading...</span>}
      </div>

      {error && <div style={{ color: 'red', marginTop: '10px' }}>{error}</div>}

      <div style={{ marginTop: '20px' }}>
        {images.map((img) => (
          <div key={img.id} style={{ marginBottom: '20px' }}>
            <img src={img.url} alt={img.id} style={{ maxWidth: '400px' }} />
            <p>ID: {img.id}</p>
            <p>Size: {(img.size / 1024).toFixed(2)} KB</p>
          </div>
        ))}
      </div>
    </div>
  );
}
```

## TypeScript Support

```typescript
// types.ts
export interface UploadResponse {
  id: string;
  url: string;
  size: number;
  mime_type: string;
  created_at: string;
}

export interface FileInfo {
  id: string;
  url: string;
  size: number;
  mime_type: string;
  created_at: string;
}

export interface ListFilesResponse {
  files: FileInfo[];
  pagination: {
    page: number;
    limit: number;
    total: number;
    total_pages: number;
  };
}

export interface StatsResponse {
  storage_used_bytes: number;
  storage_used_mb: number;
  file_count: number;
  quota_bytes: number;
  quota_mb: number;
  usage_percent: number;
}

export interface ApiError {
  error: {
    message: string;
    type: string;
    code: string;
  };
}
```

```typescript
// client.ts
import FormData from 'form-data';
import fs from 'fs';
import type { UploadResponse, ListFilesResponse, StatsResponse } from './types';

export class LabNocturneClient {
  constructor(
    private apiKey: string,
    private baseUrl: string = 'https://images.labnocturne.com'
  ) {}

  async upload(filePath: string): Promise<UploadResponse> {
    const form = new FormData();
    form.append('file', fs.createReadStream(filePath));

    const response = await fetch(`${this.baseUrl}/upload`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${this.apiKey}`,
        ...form.getHeaders()
      },
      body: form
    });

    if (!response.ok) {
      throw new Error(`Upload failed: ${response.statusText}`);
    }

    return await response.json();
  }

  async listFiles(options?: {
    page?: number;
    limit?: number;
    sort?: string;
  }): Promise<ListFilesResponse> {
    const params = new URLSearchParams({
      page: String(options?.page || 1),
      limit: String(options?.limit || 50),
      sort: options?.sort || 'created_desc'
    });

    const response = await fetch(`${this.baseUrl}/files?${params}`, {
      headers: {
        'Authorization': `Bearer ${this.apiKey}`
      }
    });

    return await response.json();
  }

  async getStats(): Promise<StatsResponse> {
    const response = await fetch(`${this.baseUrl}/stats`, {
      headers: {
        'Authorization': `Bearer ${this.apiKey}`
      }
    });

    return await response.json();
  }

  async deleteFile(imageId: string): Promise<void> {
    await fetch(`${this.baseUrl}/i/${imageId}`, {
      method: 'DELETE',
      headers: {
        'Authorization': `Bearer ${this.apiKey}`
      }
    });
  }

  static async generateTestKey(
    baseUrl: string = 'https://images.labnocturne.com'
  ): Promise<string> {
    const response = await fetch(`${baseUrl}/key`);
    const data = await response.json();
    return data.api_key;
  }
}
```

## Error Handling

```javascript
try {
  const result = await client.upload('./photo.jpg');
  console.log('Success:', result.url);
} catch (error) {
  if (error.message.includes('file_too_large')) {
    console.error('File is too large for your account tier');
  } else if (error.message.includes('unauthorized')) {
    console.error('Invalid API key');
  } else {
    console.error('Upload failed:', error.message);
  }
}
```

## Package.json Example

```json
{
  "name": "labnocturne-example",
  "version": "1.0.0",
  "type": "module",
  "dependencies": {
    "form-data": "^4.0.0"
  },
  "scripts": {
    "upload": "node upload.js"
  }
}
```

## Next Steps

- Try the [Python examples](../python/)
- Try the [Go examples](../go/)
- Read the [curl examples](../curl/) for raw HTTP requests
- Check the [API documentation](https://images.labnocturne.com/docs)
