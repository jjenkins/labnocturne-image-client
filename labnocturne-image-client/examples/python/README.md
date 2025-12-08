# Python Examples - Lab Nocturne Images API

Complete Python examples for integrating with the Lab Nocturne Images API using the `requests` library.

## Prerequisites

- Python 3.7+
- `requests` library

## Installation

```bash
pip install requests
```

## Quick Start

### Basic Upload

```python
# upload.py
import requests

API_BASE = "https://images.labnocturne.com"

def upload_image(api_key, file_path):
    """Upload an image file"""
    with open(file_path, 'rb') as f:
        files = {'file': f}
        headers = {'Authorization': f'Bearer {api_key}'}

        response = requests.post(
            f'{API_BASE}/upload',
            headers=headers,
            files=files
        )
        response.raise_for_status()
        return response.json()

# Usage
api_key = 'ln_test_01jcd8x9k2...'
result = upload_image(api_key, 'photo.jpg')
print(f"Image URL: {result['url']}")
print(f"Image ID: {result['id']}")
```

## Complete Client Class

```python
# labnocturne_client.py
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
```

## Usage Examples

### Complete Workflow

```python
# example.py
from labnocturne_client import LabNocturneClient

def main():
    # Generate a test API key
    print("Generating test API key...")
    api_key = LabNocturneClient.generate_test_key()
    print(f"API Key: {api_key}\n")

    # Create client (use context manager for automatic cleanup)
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

Run it:
```bash
python example.py
```

### Upload Multiple Files

```python
# batch_upload.py
from pathlib import Path
from labnocturne_client import LabNocturneClient
import sys

def upload_directory(api_key: str, directory: str):
    """Upload all images in a directory"""
    client = LabNocturneClient(api_key)

    image_extensions = {'.jpg', '.jpeg', '.png', '.gif', '.webp'}
    directory_path = Path(directory)

    uploaded_files = []

    for file_path in directory_path.iterdir():
        if file_path.suffix.lower() in image_extensions:
            try:
                print(f"Uploading {file_path.name}...")
                result = client.upload(str(file_path))
                uploaded_files.append(result)
                print(f"  ✓ {result['url']}")
            except Exception as e:
                print(f"  ✗ Failed: {e}")

    print(f"\nUploaded {len(uploaded_files)} files")

    # Show stats
    stats = client.get_stats()
    print(f"Total storage: {stats['storage_used_mb']:.2f} MB")

    client.close()
    return uploaded_files

if __name__ == '__main__':
    if len(sys.argv) < 3:
        print("Usage: python batch_upload.py <api_key> <directory>")
        sys.exit(1)

    api_key = sys.argv[1]
    directory = sys.argv[2]
    upload_directory(api_key, directory)
```

### Download Images

```python
# download.py
import requests
from labnocturne_client import LabNocturneClient
from pathlib import Path

def download_all_files(api_key: str, output_dir: str = './downloads'):
    """Download all uploaded files"""
    client = LabNocturneClient(api_key)
    output_path = Path(output_dir)
    output_path.mkdir(exist_ok=True)

    # Get all files
    files = client.list_files(limit=1000)

    print(f"Downloading {len(files['files'])} files...")

    for file_info in files['files']:
        # Extract filename from URL
        filename = file_info['url'].split('/')[-1]
        output_file = output_path / filename

        print(f"Downloading {filename}...")
        response = requests.get(file_info['url'])
        response.raise_for_status()

        with open(output_file, 'wb') as f:
            f.write(response.content)

        print(f"  ✓ Saved to {output_file}")

    print(f"\nAll files downloaded to {output_dir}")
    client.close()

if __name__ == '__main__':
    import sys
    if len(sys.argv) < 2:
        print("Usage: python download.py <api_key> [output_dir]")
        sys.exit(1)

    api_key = sys.argv[1]
    output_dir = sys.argv[2] if len(sys.argv) > 2 else './downloads'
    download_all_files(api_key, output_dir)
```

### With Progress Bar

```python
# upload_progress.py
import requests
from tqdm import tqdm
from pathlib import Path

API_BASE = "https://images.labnocturne.com"

def upload_with_progress(api_key: str, file_path: str):
    """Upload file with progress bar"""
    file_size = Path(file_path).stat().st_size

    with open(file_path, 'rb') as f:
        # Wrap file object with tqdm
        with tqdm(total=file_size, unit='B', unit_scale=True, desc=file_path) as pbar:
            def read_callback(data):
                pbar.update(len(data))
                return data

            files = {'file': (Path(file_path).name, f)}
            headers = {'Authorization': f'Bearer {api_key}'}

            response = requests.post(
                f'{API_BASE}/upload',
                headers=headers,
                files=files
            )

            response.raise_for_status()
            return response.json()

# Install tqdm first: pip install tqdm
```

### Async/Await Support

```python
# async_client.py
import aiohttp
import asyncio
from typing import Dict

class AsyncLabNocturneClient:
    """Async client for Lab Nocturne Images API"""

    def __init__(self, api_key: str, base_url: str = "https://images.labnocturne.com"):
        self.api_key = api_key
        self.base_url = base_url

    async def upload(self, file_path: str) -> Dict:
        """Upload an image file asynchronously"""
        async with aiohttp.ClientSession() as session:
            with open(file_path, 'rb') as f:
                data = aiohttp.FormData()
                data.add_field('file', f)

                headers = {'Authorization': f'Bearer {self.api_key}'}

                async with session.post(
                    f'{self.base_url}/upload',
                    headers=headers,
                    data=data
                ) as response:
                    response.raise_for_status()
                    return await response.json()

    async def list_files(self, page: int = 1, limit: int = 50) -> Dict:
        """List files asynchronously"""
        async with aiohttp.ClientSession() as session:
            headers = {'Authorization': f'Bearer {self.api_key}'}
            params = {'page': page, 'limit': limit}

            async with session.get(
                f'{self.base_url}/files',
                headers=headers,
                params=params
            ) as response:
                response.raise_for_status()
                return await response.json()

    async def get_stats(self) -> Dict:
        """Get stats asynchronously"""
        async with aiohttp.ClientSession() as session:
            headers = {'Authorization': f'Bearer {self.api_key}'}

            async with session.get(
                f'{self.base_url}/stats',
                headers=headers
            ) as response:
                response.raise_for_status()
                return await response.json()

# Usage
async def main():
    client = AsyncLabNocturneClient('ln_test_abc123...')

    # Upload multiple files concurrently
    files = ['photo1.jpg', 'photo2.jpg', 'photo3.jpg']
    results = await asyncio.gather(*[client.upload(f) for f in files])

    for result in results:
        print(f"Uploaded: {result['url']}")

# Run
# asyncio.run(main())

# Install aiohttp first: pip install aiohttp
```

### Error Handling

```python
# error_handling.py
from labnocturne_client import LabNocturneClient
import requests

def safe_upload(api_key: str, file_path: str):
    """Upload with comprehensive error handling"""
    client = LabNocturneClient(api_key)

    try:
        result = client.upload(file_path)
        print(f"Success: {result['url']}")
        return result

    except requests.exceptions.HTTPError as e:
        if e.response.status_code == 401:
            print("Error: Invalid API key")
        elif e.response.status_code == 413:
            print("Error: File too large for your account tier")
        elif e.response.status_code == 400:
            error_data = e.response.json()
            print(f"Error: {error_data['error']['message']}")
        else:
            print(f"HTTP Error: {e}")

    except FileNotFoundError:
        print(f"Error: File not found: {file_path}")

    except Exception as e:
        print(f"Unexpected error: {e}")

    finally:
        client.close()

    return None
```

### CLI Tool

```python
#!/usr/bin/env python3
# ln-images.py - Command-line tool for Lab Nocturne Images API

import sys
import argparse
from labnocturne_client import LabNocturneClient

def cmd_generate_key(args):
    """Generate a test API key"""
    api_key = LabNocturneClient.generate_test_key()
    print(f"Generated API key: {api_key}")
    print("\nSave this key:")
    print(f"  export LABNOCTURNE_API_KEY='{api_key}'")

def cmd_upload(args):
    """Upload a file"""
    client = LabNocturneClient(args.api_key)
    result = client.upload(args.file)
    print(f"Uploaded: {result['url']}")
    print(f"ID: {result['id']}")
    print(f"Size: {result['size'] / 1024:.2f} KB")
    client.close()

def cmd_list(args):
    """List files"""
    client = LabNocturneClient(args.api_key)
    files = client.list_files(page=args.page, limit=args.limit, sort=args.sort)

    print(f"Files (page {files['pagination']['page']} of {files['pagination']['total_pages']}):")
    for file in files['files']:
        print(f"  {file['id']}")
        print(f"    URL: {file['url']}")
        print(f"    Size: {file['size'] / 1024:.2f} KB")
        print(f"    Created: {file['created_at']}")

    client.close()

def cmd_stats(args):
    """Show usage stats"""
    client = LabNocturneClient(args.api_key)
    stats = client.get_stats()

    print("Usage Statistics:")
    print(f"  Storage: {stats['storage_used_mb']:.2f} MB / {stats['quota_mb']} MB")
    print(f"  Files: {stats['file_count']}")
    print(f"  Usage: {stats['usage_percent']:.2f}%")

    client.close()

def cmd_delete(args):
    """Delete a file"""
    client = LabNocturneClient(args.api_key)
    client.delete_file(args.image_id)
    print(f"Deleted: {args.image_id}")
    client.close()

def main():
    parser = argparse.ArgumentParser(description='Lab Nocturne Images CLI')
    subparsers = parser.add_subparsers(dest='command', help='Commands')

    # Generate key command
    subparsers.add_parser('key', help='Generate a test API key')

    # Upload command
    upload_parser = subparsers.add_parser('upload', help='Upload an image')
    upload_parser.add_argument('file', help='File to upload')
    upload_parser.add_argument('--api-key', required=True, help='API key')

    # List command
    list_parser = subparsers.add_parser('list', help='List files')
    list_parser.add_argument('--api-key', required=True, help='API key')
    list_parser.add_argument('--page', type=int, default=1, help='Page number')
    list_parser.add_argument('--limit', type=int, default=50, help='Files per page')
    list_parser.add_argument('--sort', default='created_desc', help='Sort order')

    # Stats command
    stats_parser = subparsers.add_parser('stats', help='Show usage statistics')
    stats_parser.add_argument('--api-key', required=True, help='API key')

    # Delete command
    delete_parser = subparsers.add_parser('delete', help='Delete a file')
    delete_parser.add_argument('image_id', help='Image ID to delete')
    delete_parser.add_argument('--api-key', required=True, help='API key')

    args = parser.parse_args()

    if args.command == 'key':
        cmd_generate_key(args)
    elif args.command == 'upload':
        cmd_upload(args)
    elif args.command == 'list':
        cmd_list(args)
    elif args.command == 'stats':
        cmd_stats(args)
    elif args.command == 'delete':
        cmd_delete(args)
    else:
        parser.print_help()

if __name__ == '__main__':
    main()
```

Make it executable:
```bash
chmod +x ln-images.py

# Usage
./ln-images.py key
./ln-images.py upload photo.jpg --api-key ln_test_...
./ln-images.py list --api-key ln_test_...
./ln-images.py stats --api-key ln_test_...
./ln-images.py delete img_01jcd... --api-key ln_test_...
```

## Requirements.txt

```txt
requests>=2.31.0
```

Optional dependencies:
```txt
# For progress bars
tqdm>=4.66.0

# For async support
aiohttp>=3.9.0
```

## Next Steps

- Try the [Go examples](../go/)
- Try the [JavaScript examples](../javascript/)
- Check out [curl examples](../curl/) for raw HTTP requests
- Read the [API documentation](https://images.labnocturne.com/docs)
