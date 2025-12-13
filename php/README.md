# Lab Nocturne Images - PHP Client

PHP client library for the [Lab Nocturne Images API](https://images.labnocturne.com).

## Installation

### Via Packagist (Recommended)

```bash
composer require labnocturne/image-client
```

### Via GitHub (Alternative)

If you need to install directly from GitHub:

```json
{
    "repositories": [
        {
            "type": "vcs",
            "url": "https://github.com/jjenkins/labnocturne-image-client"
        }
    ],
    "require": {
        "labnocturne/image-client": "^1.0"
    }
}
```

Then run:

```bash
composer install
```

## Requirements

- PHP 7.4+
- cURL extension (usually enabled by default)
- JSON extension (usually enabled by default)

## Quick Start

```php
<?php

require 'vendor/autoload.php';

use LabNocturne\LabNocturneClient;

// Generate a test API key
$apiKey = LabNocturneClient::generateTestKey();
echo "API Key: $apiKey\n";

// Create client
$client = new LabNocturneClient($apiKey);

// Upload an image
$result = $client->upload('photo.jpg');
echo "Image URL: {$result['url']}\n";
echo "Image ID: {$result['id']}\n";

// List files
$files = $client->listFiles(1, 10);
echo "Total files: {$files['pagination']['total']}\n";

// Get usage stats
$stats = $client->getStats();
echo "Storage: " . round($stats['storage_used_mb'], 2) . " MB / {$stats['quota_mb']} MB\n";

// Delete a file
$client->deleteFile($result['id']);
echo "File deleted\n";
```

## API Reference

### `new LabNocturneClient($apiKey, $baseUrl = 'https://images.labnocturne.com')`

Create a new client instance.

**Parameters:**
- `$apiKey` (string): Your API key
- `$baseUrl` (string, optional): Base URL for the API

### Static Methods

#### `LabNocturneClient::generateTestKey($baseUrl = 'https://images.labnocturne.com')`

Generate a test API key for development.

**Parameters:**
- `$baseUrl` (string, optional): Base URL for the API

**Returns:** string - API key

**Throws:** Exception on failure

**Example:**
```php
$apiKey = LabNocturneClient::generateTestKey();
```

### Instance Methods

#### `upload($filePath)`

Upload an image file.

**Parameters:**
- `$filePath` (string): Path to the image file

**Returns:** array - Associative array with `id`, `url`, `size`, `mime_type`, `created_at`

**Throws:** Exception on failure

**Example:**
```php
$result = $client->upload('photo.jpg');
echo $result['url'];
```

#### `listFiles($page = 1, $limit = 50, $sort = 'created_desc')`

List uploaded files with pagination.

**Parameters:**
- `$page` (int): Page number (default: 1)
- `$limit` (int): Files per page (default: 50)
- `$sort` (string): Sort order - `created_desc`, `created_asc`, `size_desc`, `size_asc`, `name_asc`, `name_desc`

**Returns:** array - Associative array with `files` array and `pagination` info

**Throws:** Exception on failure

**Example:**
```php
$files = $client->listFiles(1, 10, 'size_desc');
foreach ($files['files'] as $file) {
    echo "{$file['id']}: {$file['size']} bytes\n";
}
```

#### `getStats()`

Get usage statistics for your account.

**Returns:** array - Associative array with `storage_used_bytes`, `storage_used_mb`, `file_count`, `quota_bytes`, `quota_mb`, `usage_percent`

**Throws:** Exception on failure

**Example:**
```php
$stats = $client->getStats();
echo "Using " . round($stats['usage_percent'], 1) . "% of quota\n";
```

#### `deleteFile($imageId)`

Delete an image (soft delete).

**Parameters:**
- `$imageId` (string): The image ID

**Returns:** array - Associative array with success status

**Throws:** Exception on failure

**Example:**
```php
$client->deleteFile('img_01jcd8x9k2n...');
```

## Error Handling

```php
try {
    $result = $client->upload('photo.jpg');
    echo "Success: {$result['url']}\n";
} catch (Exception $e) {
    if (strpos($e->getMessage(), 'File not found') !== false) {
        echo "Error: File not found\n";
    } elseif (strpos($e->getMessage(), 'file_too_large') !== false) {
        echo "Error: File is too large for your account tier\n";
    } elseif (strpos($e->getMessage(), 'unauthorized') !== false) {
        echo "Error: Invalid API key\n";
    } else {
        echo "Error: {$e->getMessage()}\n";
    }
}
```

## Complete Example

```php
<?php

require 'vendor/autoload.php';

use LabNocturne\LabNocturneClient;

try {
    // Generate test API key
    echo "Generating test API key...\n";
    $apiKey = LabNocturneClient::generateTestKey();
    echo "API Key: $apiKey\n\n";

    // Create client
    $client = new LabNocturneClient($apiKey);

    // Upload an image
    echo "Uploading image...\n";
    $upload = $client->upload('photo.jpg');
    echo "Uploaded: {$upload['url']}\n";
    echo "Image ID: {$upload['id']}\n";
    echo "Size: " . round($upload['size'] / 1024, 2) . " KB\n\n";

    // List all files
    echo "Listing files...\n";
    $files = $client->listFiles(1, 10);
    echo "Total files: {$files['pagination']['total']}\n";
    foreach ($files['files'] as $file) {
        $size = round($file['size'] / 1024, 2);
        echo "  - {$file['id']}: $size KB\n";
    }
    echo "\n";

    // Get usage stats
    echo "Usage statistics:\n";
    $stats = $client->getStats();
    echo "  Storage: " . round($stats['storage_used_mb'], 2) . " MB / {$stats['quota_mb']} MB\n";
    echo "  Files: {$stats['file_count']}\n";
    echo "  Usage: " . round($stats['usage_percent'], 2) . "%\n\n";

    // Delete the uploaded file
    echo "Deleting image...\n";
    $client->deleteFile($upload['id']);
    echo "Deleted successfully\n";

} catch (Exception $e) {
    echo "Error: " . $e->getMessage() . "\n";
}
```

Save as `example.php` and run:

```bash
php example.php
```

## Laravel Integration

You can use the client in Laravel applications:

```php
// config/services.php
return [
    'labnocturne' => [
        'api_key' => env('LABNOCTURNE_API_KEY'),
    ],
];

// app/Services/ImageService.php
namespace App\Services;

use LabNocturne\LabNocturneClient;

class ImageService
{
    private $client;

    public function __construct()
    {
        $this->client = new LabNocturneClient(
            config('services.labnocturne.api_key')
        );
    }

    public function upload($filePath)
    {
        return $this->client->upload($filePath);
    }

    // ... other methods
}
```

## Development

Install development dependencies:

```bash
composer install --dev
```

Run tests:

```bash
./vendor/bin/phpunit
```

## License

MIT License - See [LICENSE](../LICENSE) for details.

## Links

- [Main Repository](https://github.com/jjenkins/labnocturne-image-client)
- [API Documentation](https://images.labnocturne.com/docs)
- [Other Language Clients](https://github.com/jjenkins/labnocturne-image-client#readme)
