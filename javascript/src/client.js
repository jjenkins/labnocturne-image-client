/**
 * Lab Nocturne Images API Client
 */

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
