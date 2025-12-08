# Ruby Examples - Lab Nocturne Images API

Complete Ruby examples for integrating with the Lab Nocturne Images API.

## Prerequisites

- Ruby 2.7+
- No external gems required (uses standard library)

## Quick Start

### Basic Upload (Standard Library)

```ruby
# upload.rb
require 'net/http'
require 'uri'
require 'json'

API_BASE = 'https://images.labnocturne.com'

def upload_image(api_key, file_path)
  uri = URI.parse("#{API_BASE}/upload")

  request = Net::HTTP::Post.new(uri)
  request['Authorization'] = "Bearer #{api_key}"

  form_data = [['file', File.open(file_path)]]
  request.set_form form_data, 'multipart/form-data'

  response = Net::HTTP.start(uri.hostname, uri.port, use_ssl: true) do |http|
    http.request(request)
  end

  raise "Upload failed: #{response.code}" unless response.code == '200'

  JSON.parse(response.body)
end

# Usage
api_key = 'ln_test_01jcd8x9k2...'
result = upload_image(api_key, 'photo.jpg')
puts "Image URL: #{result['url']}"
puts "Image ID: #{result['id']}"
```

## Complete Client Class

```ruby
# labnocturne_client.rb
require 'net/http'
require 'uri'
require 'json'

class LabNocturneClient
  attr_reader :api_key, :base_url

  def initialize(api_key, base_url = 'https://images.labnocturne.com')
    @api_key = api_key
    @base_url = base_url
  end

  # Upload an image file
  def upload(file_path)
    uri = URI.parse("#{@base_url}/upload")

    request = Net::HTTP::Post.new(uri)
    request['Authorization'] = "Bearer #{@api_key}"

    form_data = [['file', File.open(file_path)]]
    request.set_form form_data, 'multipart/form-data'

    response = make_request(uri, request)
    handle_response(response)
  end

  # List uploaded files
  def list_files(page: 1, limit: 50, sort: 'created_desc')
    uri = URI.parse("#{@base_url}/files")
    uri.query = URI.encode_www_form(page: page, limit: limit, sort: sort)

    request = Net::HTTP::Get.new(uri)
    request['Authorization'] = "Bearer #{@api_key}"

    response = make_request(uri, request)
    handle_response(response)
  end

  # Get usage statistics
  def get_stats
    uri = URI.parse("#{@base_url}/stats")

    request = Net::HTTP::Get.new(uri)
    request['Authorization'] = "Bearer #{@api_key}"

    response = make_request(uri, request)
    handle_response(response)
  end

  # Delete an image (soft delete)
  def delete_file(image_id)
    uri = URI.parse("#{@base_url}/i/#{image_id}")

    request = Net::HTTP::Delete.new(uri)
    request['Authorization'] = "Bearer #{@api_key}"

    response = make_request(uri, request)
    handle_response(response)
  end

  # Generate a test API key
  def self.generate_test_key(base_url = 'https://images.labnocturne.com')
    uri = URI.parse("#{base_url}/key")
    response = Net::HTTP.get_response(uri)

    raise "Failed to generate key: #{response.code}" unless response.code == '200'

    result = JSON.parse(response.body)
    result['api_key']
  end

  private

  def make_request(uri, request)
    Net::HTTP.start(uri.hostname, uri.port, use_ssl: uri.scheme == 'https') do |http|
      http.request(request)
    end
  end

  def handle_response(response)
    unless response.code.start_with?('2')
      error_data = JSON.parse(response.body) rescue {}
      error_msg = error_data.dig('error', 'message') || 'Unknown error'
      raise "API Error: #{error_msg}"
    end

    JSON.parse(response.body)
  end
end
```

## Usage Example

```ruby
# example.rb
require_relative 'labnocturne_client'

# Generate a test API key
puts "Generating test API key..."
api_key = LabNocturneClient.generate_test_key
puts "API Key: #{api_key}\n\n"

# Create client
client = LabNocturneClient.new(api_key)

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

Run it:
```bash
ruby example.rb
```

## Batch Upload Script

```ruby
# batch_upload.rb
require_relative 'labnocturne_client'
require 'pathname'

def upload_directory(api_key, directory)
  client = LabNocturneClient.new(api_key)
  uploaded_files = []

  image_extensions = %w[.jpg .jpeg .png .gif .webp]
  dir_path = Pathname.new(directory)

  dir_path.children.each do |file_path|
    next unless file_path.file?
    next unless image_extensions.include?(file_path.extname.downcase)

    begin
      puts "Uploading #{file_path.basename}..."
      result = client.upload(file_path.to_s)
      uploaded_files << result
      puts "  ✓ #{result['url']}"
    rescue StandardError => e
      puts "  ✗ Failed: #{e.message}"
    end
  end

  puts "\nUploaded #{uploaded_files.length} files"

  # Show stats
  stats = client.get_stats
  puts "Total storage: #{stats['storage_used_mb'].round(2)} MB"

  uploaded_files
end

if __FILE__ == $0
  if ARGV.length < 2
    puts "Usage: ruby batch_upload.rb <api_key> <directory>"
    exit 1
  end

  api_key = ARGV[0]
  directory = ARGV[1]
  upload_directory(api_key, directory)
end
```

## Using HTTParty Gem (Optional)

If you prefer a cleaner HTTP client:

```bash
gem install httparty
```

```ruby
# labnocturne_httparty.rb
require 'httparty'

class LabNocturneClient
  include HTTParty
  base_uri 'https://images.labnocturne.com'

  def initialize(api_key)
    @api_key = api_key
    @headers = { 'Authorization' => "Bearer #{api_key}" }
  end

  def upload(file_path)
    response = self.class.post('/upload',
      headers: @headers,
      body: { file: File.new(file_path) }
    )
    handle_response(response)
  end

  def list_files(page: 1, limit: 50, sort: 'created_desc')
    response = self.class.get('/files',
      headers: @headers,
      query: { page: page, limit: limit, sort: sort }
    )
    handle_response(response)
  end

  def get_stats
    response = self.class.get('/stats', headers: @headers)
    handle_response(response)
  end

  def delete_file(image_id)
    response = self.class.delete("/i/#{image_id}", headers: @headers)
    handle_response(response)
  end

  def self.generate_test_key
    response = get('/key')
    raise "Failed to generate key" unless response.success?
    response['api_key']
  end

  private

  def handle_response(response)
    unless response.success?
      error_msg = response.dig('error', 'message') || 'Unknown error'
      raise "API Error: #{error_msg}"
    end
    response.parsed_response
  end
end
```

## CLI Tool

```ruby
#!/usr/bin/env ruby
# ln-images
require_relative 'labnocturne_client'
require 'optparse'

def print_usage
  puts "Lab Nocturne Images CLI"
  puts
  puts "Usage:"
  puts "  ln-images key                    Generate a test API key"
  puts "  ln-images upload <file>          Upload an image"
  puts "  ln-images list [options]         List files"
  puts "  ln-images stats                  Show usage statistics"
  puts "  ln-images delete <image_id>      Delete an image"
  puts
  puts "Environment Variables:"
  puts "  LABNOCTURNE_API_KEY    API key for authentication"
end

def get_api_key
  api_key = ENV['LABNOCTURNE_API_KEY']
  if api_key.nil? || api_key.empty?
    puts "Error: LABNOCTURNE_API_KEY environment variable not set"
    exit 1
  end
  api_key
end

def cmd_generate_key
  api_key = LabNocturneClient.generate_test_key
  puts "Generated API key: #{api_key}"
  puts
  puts "Save this key:"
  puts "  export LABNOCTURNE_API_KEY='#{api_key}'"
end

def cmd_upload(file_path)
  api_key = get_api_key
  client = LabNocturneClient.new(api_key)

  result = client.upload(file_path)
  puts "Uploaded: #{result['url']}"
  puts "ID: #{result['id']}"
  puts "Size: #{(result['size'] / 1024.0).round(2)} KB"
end

def cmd_list(options = {})
  api_key = get_api_key
  client = LabNocturneClient.new(api_key)

  files = client.list_files(**options)

  puts "Files (page #{files['pagination']['page']} of #{files['pagination']['total_pages']}):"
  files['files'].each do |file|
    puts "\n#{file['id']}"
    puts "  URL: #{file['url']}"
    puts "  Size: #{(file['size'] / 1024.0).round(2)} KB"
    puts "  Created: #{file['created_at']}"
  end
end

def cmd_stats
  api_key = get_api_key
  client = LabNocturneClient.new(api_key)

  stats = client.get_stats
  puts "Usage Statistics:"
  puts "  Storage: #{stats['storage_used_mb'].round(2)} MB / #{stats['quota_mb']} MB"
  puts "  Files: #{stats['file_count']}"
  puts "  Usage: #{stats['usage_percent'].round(2)}%"
end

def cmd_delete(image_id)
  api_key = get_api_key
  client = LabNocturneClient.new(api_key)

  client.delete_file(image_id)
  puts "Deleted: #{image_id}"
end

# Main
if ARGV.empty?
  print_usage
  exit 1
end

command = ARGV[0]

begin
  case command
  when 'key'
    cmd_generate_key
  when 'upload'
    if ARGV.length < 2
      puts "Usage: ln-images upload <file>"
      exit 1
    end
    cmd_upload(ARGV[1])
  when 'list'
    options = {}
    OptionParser.new do |opts|
      opts.on('--page PAGE', Integer, 'Page number') { |v| options[:page] = v }
      opts.on('--limit LIMIT', Integer, 'Files per page') { |v| options[:limit] = v }
      opts.on('--sort SORT', 'Sort order') { |v| options[:sort] = v }
    end.parse!(ARGV[1..-1])
    cmd_list(options)
  when 'stats'
    cmd_stats
  when 'delete'
    if ARGV.length < 2
      puts "Usage: ln-images delete <image_id>"
      exit 1
    end
    cmd_delete(ARGV[1])
  else
    print_usage
    exit 1
  end
rescue StandardError => e
  puts "Error: #{e.message}"
  exit 1
end
```

Make it executable:
```bash
chmod +x ln-images

# Usage
./ln-images key
export LABNOCTURNE_API_KEY='ln_test_...'
./ln-images upload photo.jpg
./ln-images list --page 1 --limit 10
./ln-images stats
./ln-images delete img_01jcd...
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

## Gemfile (Optional)

```ruby
# Gemfile
source 'https://rubygems.org'

gem 'httparty', '~> 0.21'  # Optional, for cleaner HTTP client
gem 'rspec', '~> 3.12'     # For testing
```

## Testing with RSpec

```ruby
# spec/labnocturne_client_spec.rb
require 'rspec'
require_relative '../labnocturne_client'

RSpec.describe LabNocturneClient do
  describe '.generate_test_key' do
    it 'generates a test API key' do
      api_key = LabNocturneClient.generate_test_key
      expect(api_key).not_to be_empty
      expect(api_key).to start_with('ln_test_')
    end
  end

  describe '#upload' do
    it 'uploads a file successfully' do
      api_key = LabNocturneClient.generate_test_key
      client = LabNocturneClient.new(api_key)

      # Create a test file
      File.write('test.jpg', 'fake image content')

      result = client.upload('test.jpg')
      expect(result['id']).not_to be_empty
      expect(result['url']).not_to be_empty

      File.delete('test.jpg')
    end
  end
end
```

Run tests:
```bash
bundle install
rspec spec/
```

## Next Steps

- Try the [PHP examples](../php/)
- Check out [Python examples](../python/) or [JavaScript examples](../javascript/)
- Read the [API documentation](https://images.labnocturne.com/docs)
