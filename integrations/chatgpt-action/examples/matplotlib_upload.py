#!/usr/bin/env python3
"""
Lab Nocturne Images - Matplotlib Integration Example

This script demonstrates how to create a matplotlib chart and upload it
to Lab Nocturne CDN. Perfect for AI-generated visualizations.

Usage:
    python matplotlib_upload.py
"""

import requests
import matplotlib.pyplot as plt
import numpy as np
from datetime import datetime


def generate_test_key():
    """Generate a test API key from Lab Nocturne."""
    response = requests.get('https://images.labnocturne.com/key')
    response.raise_for_status()
    return response.json()['api_key']


def upload_image(api_key, file_path):
    """Upload an image file to Lab Nocturne."""
    with open(file_path, 'rb') as f:
        files = {'file': f}
        headers = {'Authorization': f'Bearer {api_key}'}
        response = requests.post(
            'https://images.labnocturne.com/upload',
            headers=headers,
            files=files
        )
        response.raise_for_status()
        return response.json()


def create_sample_chart():
    """Create a sample matplotlib chart."""
    # Generate sample data
    x = np.linspace(0, 10, 100)
    y1 = np.sin(x)
    y2 = np.cos(x)

    # Create figure with high quality settings
    plt.figure(figsize=(10, 6), dpi=300)

    # Plot data
    plt.plot(x, y1, label='sin(x)', linewidth=2)
    plt.plot(x, y2, label='cos(x)', linewidth=2)

    # Styling
    plt.title('Trigonometric Functions', fontsize=16, fontweight='bold')
    plt.xlabel('x', fontsize=12)
    plt.ylabel('y', fontsize=12)
    plt.legend(fontsize=10)
    plt.grid(True, alpha=0.3)

    # Save with tight layout
    output_file = 'chart.png'
    plt.savefig(output_file, bbox_inches='tight', dpi=300)
    plt.close()

    return output_file


def main():
    """Main workflow: Create chart → Upload → Get CDN URL."""
    print("Lab Nocturne Images - Matplotlib Example\n")

    # Step 1: Generate test API key
    print("1. Generating test API key...")
    api_key = generate_test_key()
    print(f"   API Key: {api_key[:20]}...")

    # Step 2: Create matplotlib chart
    print("\n2. Creating matplotlib chart...")
    chart_file = create_sample_chart()
    print(f"   Saved to: {chart_file}")

    # Step 3: Upload to Lab Nocturne
    print("\n3. Uploading to Lab Nocturne CDN...")
    result = upload_image(api_key, chart_file)

    # Step 4: Display results
    print("\n✓ Upload successful!")
    print(f"\nImage ID:  {result['id']}")
    print(f"CDN URL:   {result['url']}")
    print(f"Size:      {result['size']:,} bytes")
    print(f"Type:      {result['mime_type']}")
    print(f"Created:   {result['created_at']}")

    # Step 5: Show usage in markdown
    print(f"\nMarkdown embed code:")
    print(f"![Chart]({result['url']})")

    print(f"\nHTML embed code:")
    print(f'<img src="{result["url"]}" alt="Chart">')


if __name__ == '__main__':
    main()
