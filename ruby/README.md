# Lab Nocturne Images - Ruby Client

Ruby client library for the [Lab Nocturne Images API](https://images.labnocturne.com).

## Installation

Add to your Gemfile:

```ruby
gem 'labnocturne',
    git: 'https://github.com/jjenkins/labnocturne-image-client',
    glob: 'ruby/*.gemspec'
```

Then run:

```bash
bundle install
```

## Requirements

- Ruby 2.7+
- No external dependencies (uses standard library only)

## Quick Start

```ruby
require 'labnocturne'

# Generate a test API key
api_key = LabNocturne::Client.generate_test_key
puts "API Key: #{api_key}"

# Create client
client = LabNocturne::Client.new(api_key)

# Upload an image
result = client.upload('photo.jpg')
puts "Image URL: #{result['url']}"
puts "Image ID: #{result['id']}"

# List files
files = client.list_files(limit: 10)
puts "Total files: #{files['pagination']['total']}"

# Get usage stats
stats = client.get_stats
puts "Storage: #{stats['storage_used_mb'].round(2)} MB / #{stats['quota_mb']} MB"

# Delete a file
client.delete_file(result['id'])
puts "File deleted"
```

## API Reference

### `LabNocturne::Client.new(api_key, base_url = 'https://images.labnocturne.com')`

Create a new client instance.

**Parameters:**
- `api_key` (String): Your API key
- `base_url` (String, optional): Base URL for the API

### Class Methods

#### `.generate_test_key(base_url = 'https://images.labnocturne.com')`

Generate a test API key for development.

**Returns:** API key string

**Example:**
```ruby
api_key = LabNocturne::Client.generate_test_key
```

### Instance Methods

#### `#upload(file_path)`

Upload an image file.

**Parameters:**
- `file_path` (String): Path to the image file

**Returns:** Hash with `:id`, `:url`, `:size`, `:mime_type`, `:created_at`

**Example:**
```ruby
result = client.upload('photo.jpg')
puts result['url']
```

#### `#list_files(page: 1, limit: 50, sort: 'created_desc')`

List uploaded files with pagination.

**Parameters:**
- `page` (Integer): Page number (default: 1)
- `limit` (Integer): Files per page (default: 50)
- `sort` (String): Sort order - `created_desc`, `created_asc`, `size_desc`, `size_asc`, `name_asc`, `name_desc`

**Returns:** Hash with `files` array and `pagination` info

**Example:**
```ruby
files = client.list_files(page: 1, limit: 10, sort: 'size_desc')
files['files'].each do |file|
  puts "#{file['id']}: #{file['size']} bytes"
end
```

#### `#get_stats`

Get usage statistics for your account.

**Returns:** Hash with `storage_used_bytes`, `storage_used_mb`, `file_count`, `quota_bytes`, `quota_mb`, `usage_percent`

**Example:**
```ruby
stats = client.get_stats
puts "Using #{stats['usage_percent'].round(1)}% of quota"
```

#### `#delete_file(image_id)`

Delete an image (soft delete).

**Parameters:**
- `image_id` (String): The image ID

**Returns:** Hash with success status

**Example:**
```ruby
client.delete_file('img_01jcd8x9k2n...')
```

## Error Handling

```ruby
begin
  result = client.upload('photo.jpg')
  puts "Success: #{result['url']}"
rescue Errno::ENOENT
  puts "Error: File not found"
rescue StandardError => e
  case e.message
  when /file_too_large/
    puts "Error: File is too large for your account tier"
  when /unauthorized/
    puts "Error: Invalid API key"
  else
    puts "Error: #{e.message}"
  end
end
```

## Complete Example

```ruby
require 'labnocturne'

# Generate test API key
puts "Generating test API key..."
api_key = LabNocturne::Client.generate_test_key
puts "API Key: #{api_key}\n\n"

# Create client
client = LabNocturne::Client.new(api_key)

begin
  # Upload an image
  puts "Uploading image..."
  upload_result = client.upload('photo.jpg')
  puts "Uploaded: #{upload_result['url']}"
  puts "Image ID: #{upload_result['id']}"
  puts "Size: #{(upload_result['size'] / 1024.0).round(2)} KB\n\n"

  # List all files
  puts "Listing files..."
  files_result = client.list_files(limit: 10)
  puts "Total files: #{files_result['pagination']['total']}"
  files_result['files'].each do |file|
    puts "  - #{file['id']}: #{(file['size'] / 1024.0).round(2)} KB"
  end
  puts

  # Get usage stats
  puts "Usage statistics:"
  stats = client.get_stats
  puts "  Storage: #{stats['storage_used_mb'].round(2)} MB / #{stats['quota_mb']} MB"
  puts "  Files: #{stats['file_count']}"
  puts "  Usage: #{stats['usage_percent'].round(2)}%\n\n"

  # Delete the uploaded file
  puts "Deleting image..."
  client.delete_file(upload_result['id'])
  puts "Deleted successfully"

rescue StandardError => e
  puts "Error: #{e.message}"
end
```

## Development

Install development dependencies:

```bash
bundle install
```

Run tests:

```bash
rspec
```

Lint code:

```bash
rubocop
```

## License

MIT License - See [LICENSE](../LICENSE) for details.

## Links

- [Main Repository](https://github.com/jjenkins/labnocturne-image-client)
- [API Documentation](https://images.labnocturne.com/docs)
- [Other Language Clients](https://github.com/jjenkins/labnocturne-image-client#readme)
