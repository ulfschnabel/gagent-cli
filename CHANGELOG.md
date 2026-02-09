# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.3.0] - 2026-02-09

### Added
- **Rich Text Formatting**: Support for bold, italic, underline, strikethrough, colors, and font styles
  - New `append-formatted` command with formatting flags
  - Support for named styles: heading1-4, title, subtitle, normal
  - Font size (8-72pt) and font family customization
  - Text and background color support (#rrggbb format)

- **List Support**: Create bullet, numbered, lettered, roman, and checklist lists
  - New `insert-list` command
  - Support for nested lists with indent levels (0-9)
  - Multiple list styles: bullet, numbered, lettered, roman, checklist

- **Paragraph Formatting**: Advanced paragraph styling capabilities
  - New `format-paragraph` command
  - Alignment options: left, center, right, justify
  - Indentation control: start, end, first-line
  - Line spacing and paragraph spacing controls

- **Table Support**: Create and populate tables with data
  - New `insert-table` command
  - Support for CSV data import
  - Header row support
  - Configurable rows and columns

- **Document Structure**: Enhanced document organization
  - New `insert-pagebreak` command for page breaks
  - New `insert-hr` command for horizontal rules
  - New `insert-toc` command for table of contents placeholder

- **Template-based Formatting**: Batch operations via JSON templates
  - New `format-template` command
  - Support for complex document structures
  - JSON-based template definition
  - Combine multiple formatting operations in one call

### Changed
- Enhanced `docs` command group with rich formatting capabilities
- Updated README.md with comprehensive formatting examples
- Improved documentation structure

### Technical
- Added comprehensive test coverage for all formatting features
- Implemented validation functions for colors, fonts, alignments, etc.
- Added helper functions for style mapping and parsing
- All features follow test-driven development (TDD) approach

## [0.2.0] - Previous Release

Previous version with basic document operations.

