# frozen_string_literal: true

require 'net/http'
require 'uri'
require 'json'

module LabNocturne
  # Client for Lab Nocturne Images API
  class Client
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
end
