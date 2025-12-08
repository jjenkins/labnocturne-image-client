# PHP Examples - Lab Nocturne Images API

Complete PHP examples for integrating with the Lab Nocturne Images API.

## Prerequisites

- PHP 7.4+ (8.0+ recommended)
- cURL extension (usually enabled by default)

## Quick Start

### Basic Upload (cURL)

```php
<?php
// upload.php

const API_BASE = 'https://images.labnocturne.com';

function uploadImage($apiKey, $filePath) {
    $ch = curl_init(API_BASE . '/upload');

    $file = new CURLFile($filePath);
    $data = ['file' => $file];

    curl_setopt_array($ch, [
        CURLOPT_POST => true,
        CURLOPT_POSTFIELDS => $data,
        CURLOPT_RETURNTRANSFER => true,
        CURLOPT_HTTPHEADER => [
            'Authorization: Bearer ' . $apiKey
        ]
    ]);

    $response = curl_exec($ch);
    $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
    curl_close($ch);

    if ($httpCode !== 200) {
        throw new Exception("Upload failed with status: $httpCode");
    }

    return json_decode($response, true);
}

// Usage
$apiKey = 'ln_test_01jcd8x9k2...';
$result = uploadImage($apiKey, 'photo.jpg');

echo "Image URL: " . $result['url'] . "\n";
echo "Image ID: " . $result['id'] . "\n";
```

## Complete Client Class

```php
<?php
// LabNocturneClient.php

class LabNocturneClient {
    private $apiKey;
    private $baseUrl;

    public function __construct($apiKey, $baseUrl = 'https://images.labnocturne.com') {
        $this->apiKey = $apiKey;
        $this->baseUrl = $baseUrl;
    }

    /**
     * Upload an image file
     *
     * @param string $filePath Path to the image file
     * @return array Response data
     * @throws Exception
     */
    public function upload($filePath) {
        if (!file_exists($filePath)) {
            throw new Exception("File not found: $filePath");
        }

        $ch = curl_init($this->baseUrl . '/upload');

        $file = new CURLFile($filePath);
        $data = ['file' => $file];

        curl_setopt_array($ch, [
            CURLOPT_POST => true,
            CURLOPT_POSTFIELDS => $data,
            CURLOPT_RETURNTRANSFER => true,
            CURLOPT_HTTPHEADER => [
                'Authorization: Bearer ' . $this->apiKey
            ]
        ]);

        $response = curl_exec($ch);
        $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
        $error = curl_error($ch);
        curl_close($ch);

        if ($error) {
            throw new Exception("cURL error: $error");
        }

        if ($httpCode !== 200) {
            $errorData = json_decode($response, true);
            $errorMsg = $errorData['error']['message'] ?? 'Unknown error';
            throw new Exception("Upload failed: $errorMsg");
        }

        return json_decode($response, true);
    }

    /**
     * List uploaded files
     *
     * @param int $page Page number
     * @param int $limit Files per page
     * @param string $sort Sort order
     * @return array Response data
     * @throws Exception
     */
    public function listFiles($page = 1, $limit = 50, $sort = 'created_desc') {
        $query = http_build_query([
            'page' => $page,
            'limit' => $limit,
            'sort' => $sort
        ]);

        return $this->request('GET', "/files?$query");
    }

    /**
     * Get usage statistics
     *
     * @return array Response data
     * @throws Exception
     */
    public function getStats() {
        return $this->request('GET', '/stats');
    }

    /**
     * Delete an image (soft delete)
     *
     * @param string $imageId The image ID
     * @return array Response data
     * @throws Exception
     */
    public function deleteFile($imageId) {
        return $this->request('DELETE', "/i/$imageId");
    }

    /**
     * Generate a test API key
     *
     * @param string $baseUrl Base URL
     * @return string API key
     * @throws Exception
     */
    public static function generateTestKey($baseUrl = 'https://images.labnocturne.com') {
        $ch = curl_init($baseUrl . '/key');

        curl_setopt_array($ch, [
            CURLOPT_RETURNTRANSFER => true
        ]);

        $response = curl_exec($ch);
        $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
        curl_close($ch);

        if ($httpCode !== 200) {
            throw new Exception("Failed to generate API key");
        }

        $data = json_decode($response, true);
        return $data['api_key'];
    }

    /**
     * Make an API request
     *
     * @param string $method HTTP method
     * @param string $endpoint API endpoint
     * @return array Response data
     * @throws Exception
     */
    private function request($method, $endpoint) {
        $ch = curl_init($this->baseUrl . $endpoint);

        curl_setopt_array($ch, [
            CURLOPT_CUSTOMREQUEST => $method,
            CURLOPT_RETURNTRANSFER => true,
            CURLOPT_HTTPHEADER => [
                'Authorization: Bearer ' . $this->apiKey
            ]
        ]);

        $response = curl_exec($ch);
        $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
        $error = curl_error($ch);
        curl_close($ch);

        if ($error) {
            throw new Exception("cURL error: $error");
        }

        if ($httpCode < 200 || $httpCode >= 300) {
            $errorData = json_decode($response, true);
            $errorMsg = $errorData['error']['message'] ?? 'Unknown error';
            throw new Exception("API Error: $errorMsg");
        }

        return json_decode($response, true);
    }
}
```

## Usage Example

```php
<?php
// example.php
require_once 'LabNocturneClient.php';

try {
    // Generate a test API key
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

Run it:
```bash
php example.php
```

## Batch Upload Script

```php
<?php
// batch_upload.php
require_once 'LabNocturneClient.php';

function uploadDirectory($apiKey, $directory) {
    $client = new LabNocturneClient($apiKey);
    $uploaded = [];

    $imageExtensions = ['jpg', 'jpeg', 'png', 'gif', 'webp'];

    $files = scandir($directory);
    foreach ($files as $file) {
        $filePath = $directory . '/' . $file;

        if (!is_file($filePath)) {
            continue;
        }

        $ext = strtolower(pathinfo($file, PATHINFO_EXTENSION));
        if (!in_array($ext, $imageExtensions)) {
            continue;
        }

        try {
            echo "Uploading $file...\n";
            $result = $client->upload($filePath);
            $uploaded[] = $result;
            echo "  ✓ {$result['url']}\n";
        } catch (Exception $e) {
            echo "  ✗ Failed: {$e->getMessage()}\n";
        }
    }

    echo "\nUploaded " . count($uploaded) . " files\n";

    // Show stats
    $stats = $client->getStats();
    echo "Total storage: " . round($stats['storage_used_mb'], 2) . " MB\n";

    return $uploaded;
}

if ($argc < 3) {
    echo "Usage: php batch_upload.php <api_key> <directory>\n";
    exit(1);
}

$apiKey = $argv[1];
$directory = $argv[2];
uploadDirectory($apiKey, $directory);
```

## Using Guzzle (Optional)

If you prefer Guzzle HTTP client:

```bash
composer require guzzlehttp/guzzle
```

```php
<?php
// LabNocturneGuzzle.php
require 'vendor/autoload.php';

use GuzzleHttp\Client;
use GuzzleHttp\Exception\RequestException;

class LabNocturneClient {
    private $client;
    private $apiKey;

    public function __construct($apiKey, $baseUrl = 'https://images.labnocturne.com') {
        $this->apiKey = $apiKey;
        $this->client = new Client([
            'base_uri' => $baseUrl,
            'headers' => [
                'Authorization' => 'Bearer ' . $apiKey
            ]
        ]);
    }

    public function upload($filePath) {
        try {
            $response = $this->client->post('/upload', [
                'multipart' => [
                    [
                        'name' => 'file',
                        'contents' => fopen($filePath, 'r'),
                        'filename' => basename($filePath)
                    ]
                ]
            ]);

            return json_decode($response->getBody(), true);
        } catch (RequestException $e) {
            $this->handleError($e);
        }
    }

    public function listFiles($page = 1, $limit = 50, $sort = 'created_desc') {
        try {
            $response = $this->client->get('/files', [
                'query' => [
                    'page' => $page,
                    'limit' => $limit,
                    'sort' => $sort
                ]
            ]);

            return json_decode($response->getBody(), true);
        } catch (RequestException $e) {
            $this->handleError($e);
        }
    }

    public function getStats() {
        try {
            $response = $this->client->get('/stats');
            return json_decode($response->getBody(), true);
        } catch (RequestException $e) {
            $this->handleError($e);
        }
    }

    public function deleteFile($imageId) {
        try {
            $response = $this->client->delete("/i/$imageId");
            return json_decode($response->getBody(), true);
        } catch (RequestException $e) {
            $this->handleError($e);
        }
    }

    public static function generateTestKey($baseUrl = 'https://images.labnocturne.com') {
        $client = new Client(['base_uri' => $baseUrl]);
        $response = $client->get('/key');
        $data = json_decode($response->getBody(), true);
        return $data['api_key'];
    }

    private function handleError(RequestException $e) {
        if ($e->hasResponse()) {
            $error = json_decode($e->getResponse()->getBody(), true);
            $message = $error['error']['message'] ?? 'Unknown error';
            throw new Exception("API Error: $message");
        }
        throw new Exception("Request failed: " . $e->getMessage());
    }
}
```

## Laravel Integration

```php
<?php
// app/Services/ImageService.php
namespace App\Services;

use Illuminate\Support\Facades\Http;
use Illuminate\Http\UploadedFile;

class ImageService {
    private $apiKey;
    private $baseUrl = 'https://images.labnocturne.com';

    public function __construct() {
        $this->apiKey = config('services.labnocturne.api_key');
    }

    public function upload(UploadedFile $file) {
        $response = Http::withHeaders([
            'Authorization' => 'Bearer ' . $this->apiKey
        ])->attach(
            'file',
            file_get_contents($file->getRealPath()),
            $file->getClientOriginalName()
        )->post($this->baseUrl . '/upload');

        if (!$response->successful()) {
            throw new \Exception('Upload failed: ' . $response->json()['error']['message']);
        }

        return $response->json();
    }

    public function listFiles($page = 1, $limit = 50) {
        $response = Http::withHeaders([
            'Authorization' => 'Bearer ' . $this->apiKey
        ])->get($this->baseUrl . '/files', [
            'page' => $page,
            'limit' => $limit
        ]);

        return $response->json();
    }

    public function deleteFile($imageId) {
        $response = Http::withHeaders([
            'Authorization' => 'Bearer ' . $this->apiKey
        ])->delete($this->baseUrl . '/i/' . $imageId);

        return $response->successful();
    }
}
```

Laravel controller:

```php
<?php
// app/Http/Controllers/ImageController.php
namespace App\Http\Controllers;

use App\Services\ImageService;
use Illuminate\Http\Request;

class ImageController extends Controller {
    private $imageService;

    public function __construct(ImageService $imageService) {
        $this->imageService = $imageService;
    }

    public function upload(Request $request) {
        $request->validate([
            'image' => 'required|image|max:10240' // 10MB max
        ]);

        try {
            $result = $this->imageService->upload($request->file('image'));

            return response()->json([
                'success' => true,
                'url' => $result['url'],
                'id' => $result['id']
            ]);
        } catch (\Exception $e) {
            return response()->json([
                'success' => false,
                'message' => $e->getMessage()
            ], 500);
        }
    }

    public function list() {
        $files = $this->imageService->listFiles();
        return view('images.list', compact('files'));
    }
}
```

## WordPress Plugin Example

```php
<?php
/*
Plugin Name: Lab Nocturne Images
Description: Upload images to Lab Nocturne CDN
Version: 1.0
*/

class LabNocturne_Plugin {
    private $api_key;

    public function __construct() {
        $this->api_key = get_option('labnocturne_api_key');
        add_action('admin_menu', [$this, 'add_admin_menu']);
        add_action('admin_init', [$this, 'settings_init']);
    }

    public function add_admin_menu() {
        add_options_page(
            'Lab Nocturne Settings',
            'Lab Nocturne',
            'manage_options',
            'labnocturne',
            [$this, 'settings_page']
        );
    }

    public function settings_init() {
        register_setting('labnocturne_settings', 'labnocturne_api_key');

        add_settings_section(
            'labnocturne_section',
            'API Settings',
            null,
            'labnocturne_settings'
        );

        add_settings_field(
            'labnocturne_api_key',
            'API Key',
            [$this, 'api_key_render'],
            'labnocturne_settings',
            'labnocturne_section'
        );
    }

    public function api_key_render() {
        $value = get_option('labnocturne_api_key');
        echo '<input type="text" name="labnocturne_api_key" value="' . esc_attr($value) . '" style="width: 100%; max-width: 400px;">';
    }

    public function settings_page() {
        ?>
        <div class="wrap">
            <h1>Lab Nocturne Settings</h1>
            <form method="post" action="options.php">
                <?php
                settings_fields('labnocturne_settings');
                do_settings_sections('labnocturne_settings');
                submit_button();
                ?>
            </form>
        </div>
        <?php
    }

    public function upload_image($file_path) {
        $ch = curl_init('https://images.labnocturne.com/upload');

        $file = new CURLFile($file_path);
        $data = ['file' => $file];

        curl_setopt_array($ch, [
            CURLOPT_POST => true,
            CURLOPT_POSTFIELDS => $data,
            CURLOPT_RETURNTRANSFER => true,
            CURLOPT_HTTPHEADER => [
                'Authorization: Bearer ' . $this->api_key
            ]
        ]);

        $response = curl_exec($ch);
        curl_close($ch);

        return json_decode($response, true);
    }
}

new LabNocturne_Plugin();
```

## CLI Script

```php
#!/usr/bin/env php
<?php
// ln-images
require_once 'LabNocturneClient.php';

function printUsage() {
    echo "Lab Nocturne Images CLI\n\n";
    echo "Usage:\n";
    echo "  ln-images key                    Generate a test API key\n";
    echo "  ln-images upload <file>          Upload an image\n";
    echo "  ln-images list                   List files\n";
    echo "  ln-images stats                  Show usage statistics\n";
    echo "  ln-images delete <image_id>      Delete an image\n";
    echo "\n";
    echo "Environment Variables:\n";
    echo "  LABNOCTURNE_API_KEY    API key for authentication\n";
}

function getApiKey() {
    $apiKey = getenv('LABNOCTURNE_API_KEY');
    if (empty($apiKey)) {
        echo "Error: LABNOCTURNE_API_KEY environment variable not set\n";
        exit(1);
    }
    return $apiKey;
}

if ($argc < 2) {
    printUsage();
    exit(1);
}

$command = $argv[1];

try {
    switch ($command) {
        case 'key':
            $apiKey = LabNocturneClient::generateTestKey();
            echo "Generated API key: $apiKey\n\n";
            echo "Save this key:\n";
            echo "  export LABNOCTURNE_API_KEY='$apiKey'\n";
            break;

        case 'upload':
            if ($argc < 3) {
                echo "Usage: ln-images upload <file>\n";
                exit(1);
            }
            $apiKey = getApiKey();
            $client = new LabNocturneClient($apiKey);
            $result = $client->upload($argv[2]);
            echo "Uploaded: {$result['url']}\n";
            echo "ID: {$result['id']}\n";
            echo "Size: " . round($result['size'] / 1024, 2) . " KB\n";
            break;

        case 'list':
            $apiKey = getApiKey();
            $client = new LabNocturneClient($apiKey);
            $files = $client->listFiles();
            echo "Files (page {$files['pagination']['page']} of {$files['pagination']['total_pages']}):\n";
            foreach ($files['files'] as $file) {
                echo "\n{$file['id']}\n";
                echo "  URL: {$file['url']}\n";
                echo "  Size: " . round($file['size'] / 1024, 2) . " KB\n";
                echo "  Created: {$file['created_at']}\n";
            }
            break;

        case 'stats':
            $apiKey = getApiKey();
            $client = new LabNocturneClient($apiKey);
            $stats = $client->getStats();
            echo "Usage Statistics:\n";
            echo "  Storage: " . round($stats['storage_used_mb'], 2) . " MB / {$stats['quota_mb']} MB\n";
            echo "  Files: {$stats['file_count']}\n";
            echo "  Usage: " . round($stats['usage_percent'], 2) . "%\n";
            break;

        case 'delete':
            if ($argc < 3) {
                echo "Usage: ln-images delete <image_id>\n";
                exit(1);
            }
            $apiKey = getApiKey();
            $client = new LabNocturneClient($apiKey);
            $client->deleteFile($argv[2]);
            echo "Deleted: {$argv[2]}\n";
            break;

        default:
            printUsage();
            exit(1);
    }
} catch (Exception $e) {
    echo "Error: " . $e->getMessage() . "\n";
    exit(1);
}
```

Make it executable:
```bash
chmod +x ln-images

# Usage
./ln-images key
export LABNOCTURNE_API_KEY='ln_test_...'
./ln-images upload photo.jpg
./ln-images list
./ln-images stats
./ln-images delete img_01jcd...
```

## Composer.json (Optional)

```json
{
    "name": "labnocturne/image-client",
    "description": "PHP client for Lab Nocturne Images API",
    "require": {
        "php": ">=7.4",
        "ext-curl": "*",
        "ext-json": "*"
    },
    "require-dev": {
        "guzzlehttp/guzzle": "^7.5",
        "phpunit/phpunit": "^9.5"
    },
    "autoload": {
        "psr-4": {
            "LabNocturne\\": "src/"
        }
    }
}
```

## Next Steps

- Check out other language examples: [Python](../python/), [JavaScript](../javascript/), [Go](../go/)
- Read the [curl examples](../curl/) for raw HTTP requests
- Read the [API documentation](https://images.labnocturne.com/docs)
