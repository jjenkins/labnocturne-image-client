<?php

namespace LabNocturne;

use Exception;

/**
 * Lab Nocturne Images API Client
 */
class LabNocturneClient
{
    private $apiKey;
    private $baseUrl;

    public function __construct($apiKey, $baseUrl = 'https://images.labnocturne.com')
    {
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
    public function upload($filePath)
    {
        if (!file_exists($filePath)) {
            throw new Exception("File not found: $filePath");
        }

        $ch = curl_init($this->baseUrl . '/upload');

        $file = new \CURLFile($filePath);
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
    public function listFiles($page = 1, $limit = 50, $sort = 'created_desc')
    {
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
    public function getStats()
    {
        return $this->request('GET', '/stats');
    }

    /**
     * Delete an image (soft delete)
     *
     * @param string $imageId The image ID
     * @return array Response data
     * @throws Exception
     */
    public function deleteFile($imageId)
    {
        return $this->request('DELETE', "/i/$imageId");
    }

    /**
     * Generate a test API key
     *
     * @param string $baseUrl Base URL
     * @return string API key
     * @throws Exception
     */
    public static function generateTestKey($baseUrl = 'https://images.labnocturne.com')
    {
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
    private function request($method, $endpoint)
    {
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
