"""Lab Nocturne Images API Client"""

import requests
from typing import Optional, Dict, List
from pathlib import Path


class LabNocturneClient:
    """Client for Lab Nocturne Images API"""

    def __init__(self, api_key: str, base_url: str = "https://images.labnocturne.com"):
        self.api_key = api_key
        self.base_url = base_url
        self.session = requests.Session()
        self.session.headers.update({'Authorization': f'Bearer {api_key}'})

    def _request(self, method: str, endpoint: str, **kwargs) -> Dict:
        """Make an API request"""
        url = f"{self.base_url}{endpoint}"
        response = self.session.request(method, url, **kwargs)

        if not response.ok:
            error_data = response.json()
            error_msg = error_data.get('error', {}).get('message', 'Unknown error')
            raise Exception(f"API Error: {error_msg}")

        return response.json()

    def upload(self, file_path: str) -> Dict:
        """
        Upload an image file

        Args:
            file_path: Path to the image file

        Returns:
            Dict with id, url, size, mime_type, created_at
        """
        with open(file_path, 'rb') as f:
            files = {'file': f}
            response = requests.post(
                f'{self.base_url}/upload',
                headers={'Authorization': f'Bearer {self.api_key}'},
                files=files
            )
            response.raise_for_status()
            return response.json()

    def list_files(
        self,
        page: int = 1,
        limit: int = 50,
        sort: str = 'created_desc'
    ) -> Dict:
        """
        List uploaded files

        Args:
            page: Page number (default: 1)
            limit: Files per page (default: 50)
            sort: Sort order (created_desc, created_asc, size_desc, size_asc, name_asc, name_desc)

        Returns:
            Dict with files array and pagination info
        """
        params = {
            'page': page,
            'limit': limit,
            'sort': sort
        }
        return self._request('GET', '/files', params=params)

    def get_stats(self) -> Dict:
        """
        Get usage statistics

        Returns:
            Dict with storage_used_bytes, file_count, quota info
        """
        return self._request('GET', '/stats')

    def delete_file(self, image_id: str) -> Dict:
        """
        Delete an image (soft delete)

        Args:
            image_id: The image ID (e.g., 'img_01jcd8x9k2n...')

        Returns:
            Dict with success status
        """
        return self._request('DELETE', f'/i/{image_id}')

    @staticmethod
    def generate_test_key(base_url: str = "https://images.labnocturne.com") -> str:
        """
        Generate a test API key

        Returns:
            API key string
        """
        response = requests.get(f'{base_url}/key')
        response.raise_for_status()
        return response.json()['api_key']

    def close(self):
        """Close the session"""
        self.session.close()

    def __enter__(self):
        """Context manager entry"""
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        """Context manager exit"""
        self.close()
