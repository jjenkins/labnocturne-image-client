# frozen_string_literal: true

require_relative 'lib/labnocturne/version'

Gem::Specification.new do |spec|
  spec.name          = 'labnocturne'
  spec.version       = LabNocturne::VERSION
  spec.authors       = ['Lab Nocturne']
  spec.email         = ['support@labnocturne.com']

  spec.summary       = 'Ruby client for Lab Nocturne Images API'
  spec.description   = 'A simple Ruby client library for uploading and managing images with the Lab Nocturne Images API'
  spec.homepage      = 'https://github.com/jjenkins/labnocturne-image-client'
  spec.license       = 'MIT'
  spec.required_ruby_version = '>= 2.7.0'

  spec.metadata['homepage_uri'] = spec.homepage
  spec.metadata['source_code_uri'] = 'https://github.com/jjenkins/labnocturne-image-client/tree/main/ruby'
  spec.metadata['bug_tracker_uri'] = 'https://github.com/jjenkins/labnocturne-image-client/issues'

  spec.files = Dir.glob('lib/**/*') + %w[README.md]
  spec.require_paths = ['lib']

  # No runtime dependencies - uses standard library only

  # Development dependencies
  spec.add_development_dependency 'rspec', '~> 3.12'
  spec.add_development_dependency 'rubocop', '~> 1.50'
end
