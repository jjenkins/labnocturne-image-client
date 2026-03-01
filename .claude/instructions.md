# Lab Nocturne Images API - Client Examples

This repository contains client examples for the Lab Nocturne Images API across multiple programming languages.

## Project Overview

Lab Nocturne Images is a simple, curl-first image storage service. This repo provides working code examples in:
- curl (command-line)
- JavaScript/Node.js
- Python
- Go
- Ruby
- PHP

## Key Concepts

### API Keys
- **Test Keys**: `ln_test_*` - Free, 10MB limit, 7-day retention
- **Live Keys**: `ln_live_*` - Paid, 100MB limit, permanent storage

### Main Endpoints
- `GET /key` - Generate test API key
- `POST /upload` - Upload image (multipart/form-data)
- `GET /i/:id` - Retrieve image (redirects to CDN)
- `GET /files` - List uploaded files
- `GET /stats` - Usage statistics
- `DELETE /i/:id` - Soft delete image

### Supported Formats
JPEG, PNG, GIF, WebP

## Working with this Repository

When adding or modifying examples:
1. Keep code simple and beginner-friendly
2. Include error handling
3. Show the complete request/response flow
4. Test examples before committing
5. Follow each language's conventions and best practices
6. Include comments explaining key API concepts

Each language example should demonstrate:
- Getting a test API key
- Uploading an image
- Listing files
- Getting usage stats
- Deleting an image
- Proper error handling

## Style Guidelines

- Use real, working code (not pseudocode)
- Include setup/installation instructions in each language's README
- Show actual API responses in comments or output
- Keep dependencies minimal
- Prioritize clarity over cleverness
