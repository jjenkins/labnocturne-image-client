# Lab Nocturne Images - ChatGPT GPT Action

Enable ChatGPT to automatically upload AI-generated images, charts, and plots to Lab Nocturne CDN.

## What is a GPT Action?

GPT Actions allow ChatGPT to interact with external APIs. By adding the Lab Nocturne Images API as a GPT Action, you enable ChatGPT to:

- Generate matplotlib charts and instantly upload them
- Create data visualizations and share CDN URLs
- Upload any AI-generated images automatically
- Manage uploaded files (list, delete, check stats)

## Quick Setup

### 1. Create a Custom GPT

1. Go to [ChatGPT](https://chat.openai.com)
2. Click your profile → "My GPTs" → "Create a GPT"
3. Name it something like "Image Uploader" or "Chart Generator"

### 2. Add the OpenAPI Schema

1. In the GPT editor, click "Configure" tab
2. Scroll to "Actions" section
3. Click "Create new action"
4. Copy the entire contents of `openapi.yaml` from this directory
5. Paste into the Schema field
6. Click "Save"

### 3. Configure Authentication

The API uses Bearer token authentication, but ChatGPT will automatically call `/key` to generate test keys when needed. No manual configuration required!

### 4. Test It

Try these prompts:

```
Create a bar chart showing sales data and upload it to Lab Nocturne
```

```
Generate a matplotlib scatter plot of random data and give me the CDN URL
```

```
Create a line chart of stock prices and upload it for sharing
```

## Example Use Cases

### Data Analysis Workflow

**Prompt:**
```
I have this CSV data: [paste data]
Create a visualization and upload it so I can share it
```

**ChatGPT will:**
1. Generate test API key via `/key` endpoint
2. Create matplotlib chart
3. Save to temporary file
4. Upload via `/upload` endpoint
5. Return CDN URL for sharing

### Automated Reporting

**Prompt:**
```
Generate a weekly sales report with 3 charts:
- Revenue trend line
- Top products bar chart
- Customer distribution pie chart

Upload all charts and give me the URLs
```

**ChatGPT will:**
1. Create all three visualizations
2. Upload each one separately
3. Provide CDN URLs for embedding in reports

### Quick Data Sharing

**Prompt:**
```
Plot this equation: y = x^2 - 3x + 2
Upload it and show me the URL
```

**ChatGPT will:**
1. Generate the plot
2. Upload to Lab Nocturne
3. Return instant CDN URL

## API Key Types

The GPT Action uses **test keys** by default:
- Generated instantly via `/key` endpoint
- 10MB file size limit
- 7-day retention
- Perfect for ChatGPT usage

For production/permanent storage:
- Get a live key from https://images.labnocturne.com
- Add to GPT Action authentication settings

## Advanced Configuration

### Custom Instructions

Add to your GPT's instructions:

```
When generating charts or images:
1. Always upload to Lab Nocturne automatically
2. Return the CDN URL in markdown format: ![Chart](url)
3. Use high-quality settings (dpi=300 for matplotlib)
4. Clean up temporary files after upload
```

### File Management

Enable the GPT to manage uploaded files:

```
List all my uploaded images
```

```
Show my storage usage stats
```

```
Delete the image with ID: img_01jcd8x...
```

## Example Python Code

Here's what ChatGPT generates when asked to create and upload a chart:

```python
import matplotlib.pyplot as plt
import requests

# Generate test API key
response = requests.get('https://images.labnocturne.com/key')
api_key = response.json()['api_key']

# Create chart
plt.figure(figsize=(10, 6))
plt.plot([1, 2, 3, 4, 5], [2, 4, 6, 8, 10])
plt.title('Sample Chart')
plt.savefig('chart.png', dpi=300, bbox_inches='tight')

# Upload to Lab Nocturne
with open('chart.png', 'rb') as f:
    files = {'file': f}
    headers = {'Authorization': f'Bearer {api_key}'}
    response = requests.post(
        'https://images.labnocturne.com/upload',
        headers=headers,
        files=files
    )

url = response.json()['url']
print(f'Image URL: {url}')
```

## Troubleshooting

### "Action failed to execute"

- Check that the OpenAPI schema is correctly pasted
- Verify the API is accessible from your location
- Try regenerating with a fresh test key

### "File too large"

Test keys have a 10MB limit. Solutions:
- Reduce image quality/size
- Use a live API key for larger files
- Compress the image before upload

### "Unauthorized"

The test key may have expired (7 days). ChatGPT will automatically generate a new one on the next request.

## Benefits for AI Developers

### Why Use Lab Nocturne with ChatGPT?

1. **Zero Configuration**: No signup needed for testing
2. **Instant CDN URLs**: Share visualizations immediately
3. **Temporary Storage**: Perfect for one-off charts and plots
4. **Simple API**: Only 2 endpoints needed (key + upload)
5. **AI-Optimized**: Designed specifically for ChatGPT/Claude workflows

### Integration Patterns

**Pattern 1: Generate → Upload → Share**
```
ChatGPT → Generate matplotlib → Upload to Lab Nocturne → Return CDN URL
```

**Pattern 2: Batch Processing**
```
ChatGPT → Generate multiple charts → Upload all → Return URLs list
```

**Pattern 3: Data Pipeline**
```
User data → ChatGPT analysis → Visualization → Lab Nocturne → Shareable report
```

## OpenAPI Schema

The `openapi.yaml` file in this directory contains the complete API specification. It includes:

- `/key` - Generate test API keys
- `/upload` - Upload image files
- `/files` - List uploaded files
- `/stats` - Get usage statistics
- `/i/{imageId}` - Delete images

All endpoints include detailed descriptions optimized for AI understanding.

## Next Steps

1. **Create your GPT** using the OpenAPI schema
2. **Test with prompts** like "create a chart and upload it"
3. **Share your GPT** with your team or make it public
4. **Upgrade to live key** when ready for production

## Resources

- [ChatGPT GPT Actions Documentation](https://platform.openai.com/docs/actions)
- [Lab Nocturne API Docs](https://images.labnocturne.com/docs)
- [OpenAPI Specification](https://spec.openapis.org/oas/v3.0.0)

## Example Prompts

### For Data Scientists

```
Load this dataset and create an exploratory data analysis with 5 charts. Upload all charts and give me a summary with the CDN URLs.
```

### For Developers

```
Generate a system architecture diagram and upload it to Lab Nocturne so I can add it to our documentation.
```

### For Content Creators

```
Create an infographic showing these statistics: [data]. Upload it and give me the embed code.
```

### For Researchers

```
Plot the results from my experiment and upload the figure for my paper. Use publication-quality settings.
```

## Making It Your Default

To make Lab Nocturne the default image storage for all ChatGPT-generated visuals:

1. Create a custom GPT with the Lab Nocturne action
2. Set it as your default ChatGPT
3. Or add these instructions to any GPT:

```
Always upload generated images to Lab Nocturne using the /upload endpoint.
Return the CDN URL for all images.
```

## License

MIT License - See [../../LICENSE](../../LICENSE) for details.

## Support

- [GitHub Issues](https://github.com/jjenkins/labnocturne-image-client/issues)
- [API Documentation](https://images.labnocturne.com/docs)
