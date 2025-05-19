# Technical Design Document: Proto-to-C4 Architecture Visualization System

## 1. Overview

This document outlines a system to maintain architectural documentation that is synchronized with service implementation. The system creates a bidirectional relationship between protocol buffer definitions (the implementation) and C4 architectural models (the documentation), ensuring architectural documentation remains accurate and useful.

### 1.1 Problem Statement

Current architectural documentation faces several challenges:
- Manual updates lead to documentation drift
- No connection between service implementation and architecture visualization
- Difficulty in tracking architectural impact of code changes
- Inconsistent representation of services across documents

### 1.2 Solution Approach

We propose a parallel build system that:
1. Extracts service definitions from Proto files to create architectural objects
2. Allows manual design of connections between these objects
3. Integrates with IcePanel for visualization and sharing

The key innovation is separating object creation from connection design, enabling validation and preventing configuration of non-existent services.

## 2. Architecture

### 2.1 Parallel Build Architecture

![Proto-to-C4 Architecture](https://example.com/architecture-diagram.png)

Rather than a linear multiphase build that would require editing generated files, we implement two parallel processes:

1. **Proto-to-IcePanel Generator**: Creates and updates objects in IcePanel based on Proto definitions
2. **Mermaid-to-IcePanel Connection Creator**: Validates objects exist and creates connections between them

This approach maintains clear boundaries between tools and enforces architectural integrity.

### 2.2 Component Breakdown

#### 2.2.1 Proto-to-IcePanel Generator

```
ProtoCompiler → Object Extractor → IcePanel Importer
     ↑                                     ↓
 BSR Services                       IcePanel Objects
     ↑
TDD Proto Specs (tdd/protos/)
```

- **Input Sources**:
  - Existing service definitions from BSR
  - Speculative service definitions from `tdd/protos/` directory
- **Processing**:
  - Extracts service metadata (names, descriptions, types)
  - Identifies internal vs external systems
  - Maps to C4 model object types
- **Output**:
  - Service objects in IcePanel with appropriate metadata

#### 2.2.2 Mermaid-to-IcePanel Connection Creator

```
Mermaid Parser → Object Validator → Connection Creator
                        ↑                  ↓
                 IcePanel Objects   IcePanel Connections
```

- **Input Sources**:
  - Mermaid C4 diagrams defining connections
- **Processing**:
  - Validates referenced objects exist in IcePanel
  - Fails fast if objects don't exist
  - Maps connection metadata (labels, types)
- **Output**:
  - Connections between services in IcePanel

## 3. Implementation Details

### 3.1 Proto Source Management

#### 3.1.1 Existing Services
- Source directly from BSR (Buf Schema Registry)
- Version pinning to specific production releases

#### 3.1.2 Speculative Services
- Located in `tdd/protos/` directory within design documents
- Follow standard Proto formatting conventions
- Must be valid Proto files that could compile

### 3.2 Scoping and Isolation

All changes will be scoped to a single Technical Design Document:
- Each TDD gets a dedicated IcePanel landscape or version
- Prevents conflicts between concurrent design efforts
- Allows parallel work on multiple designs

### 3.3 Object Identification and Mapping

| Proto Element | C4 Model Element | Notes |
|---------------|------------------|-------|
| Service       | System           | Internal system |
| External Service Reference | System_Ext | Services outside our boundary |
| Database Service | SystemDb | Services with persistence |
| Package/Namespace | System_Boundary | Logical grouping |

### 3.4 Validation Rules

The Mermaid-to-IcePanel tool will enforce these rules:
1. All referenced objects must exist in IcePanel before connections can be created
2. Connection syntax must be valid Mermaid C4 format
3. Connection endpoints must be valid object identifiers
4. Bidirectional relationships use the BiRel syntax

## 4. Workflow

### 4.1 Service Definition Workflow

1. Developer retrieves current service definitions from BSR
2. For new services, developer creates Proto files in `tdd/protos/`
3. Developer runs Proto-to-IcePanel tool to update objects
4. IcePanel is updated with the latest service objects

### 4.2 Architecture Design Workflow

1. Designer views existing objects in IcePanel
2. Designer creates Mermaid C4 file defining connections
3. Designer runs Mermaid-to-IcePanel tool to create connections
4. If validation fails, designer corrects Mermaid file
5. IcePanel is updated with the connections

### 4.3 CI Integration

The process can be integrated into CI pipelines:
1. On Proto changes, update objects automatically
2. On Mermaid changes, validate and update connections
3. Generate reports of architectural changes

## 5. Future Enhancements

### 5.1 Metadata Propagation
- Carry Proto comments, deprecation info, etc. into IcePanel objects
- Track service owners and team information

### 5.2 Difference Detection
- Flag architectural changes when BSR-sourced Protos change
- Alert on breaking changes to interfaces

### 5.3 TDD-to-Production Tracking
- Track which speculative Proto designs are implemented
- Provide metrics on design-to-implementation accuracy

### 5.4 Bidirectional Updates
- Allow updates from IcePanel to flow back to Proto definitions
- Enable visual design with code generation

## 6. Appendix

### 6.1 IcePanel API Requirements

The IcePanel API must support:
- Object creation and updates
- Connection creation and updates
- Landscape/version/diagram isolation
- Object lookup for validation

### 6.2 Tools and Dependencies

- Buf and BSR for Proto management
- Protoc for Proto compilation
- Mermaid for C4 diagram syntax
- mermaid-icepanel tool (modified) for IcePanel integration

## 7. Linear Issues Implementation Plan

This section outlines the sequential implementation tasks as a series of issues, with checklists for tracking progress. The issues are ordered for logical development flow, with parallel tracks identified where possible.

### Issue 1: Infrastructure Setup and Core Libraries

**Description**: Establish the basic project structure and shared components.

**Checklist**:
- [x] Create project repository structure
- [x] Set up Go modules and dependencies
- [x] Implement shared data models for both tools
- [x] Create IcePanel API client library
- [x] Set up unit testing framework
- [x] Implement logging and error handling frameworks
- [x] Create configuration management framework

### Issue 2: Proto Processing Pipeline

**Description**: Develop the system to extract service information from Proto files.

**Checklist**:
- [x] Create protoc plugin structure (`cmd/protoc-gen-icepanel`)
- [x] Implement proto descriptor processor in generator package
- [x] Create C4 model mapping logic from Proto to IcePanel objects 
- [x] Implement service classification logic (internal/external/database)
- [ ] Extract metadata from Proto comments
- [ ] Add support for `tdd/protos/` directory for speculative services
- [ ] Add proto validation to ensure correctness
- [x] Create tests with sample Proto files
- [ ] Support BSR integration for retrieving existing service definitions

### Issue 3: IcePanel Object Management (Can run in parallel with Issue 2)

**Description**: Develop the components for creating and managing objects in IcePanel.

**Checklist**:
- [ ] Implement IcePanel object creation
- [ ] Create object update/merge logic
- [ ] Develop object lookup and retrieval
- [ ] Implement landscape/version management
- [ ] Add support for object metadata
- [ ] Create object comparison for change detection
- [ ] Implement dry-run mode for validation

### Issue 4: Proto-to-IcePanel Generator Integration

**Description**: Integrate the Proto processing and IcePanel object management components.

**Checklist**:
- [ ] Connect proto descriptor processor to IcePanel API client
- [ ] Implement command-line interface for the plugin
- [ ] Create configuration handling for plugin options
- [ ] Add detailed logging and error reporting
- [ ] Implement object creation transaction handling
- [ ] Add incremental update support for existing IcePanel objects
- [ ] Create comprehensive test suite with sample Proto files
- [ ] Add plugin usage documentation

### Issue 5: Mermaid C4 Parser (Can run in parallel with Issues 2-4)

**Description**: Develop the parser for Mermaid C4 diagram syntax.

**Checklist**:
- [x] Implement Mermaid C4 lexer and parser
- [x] Create data models for C4 connections
- [x] Add support for bidirectional relationships
- [x] Implement connection metadata extraction
- [x] Develop validation for C4 syntax
- [x] Create test suite with sample Mermaid files
- [ ] Add error reporting with line/position information

### Issue 6: IcePanel Connection Management (Depends on Issue 3)

**Description**: Develop the components for creating and managing connections in IcePanel.

**Checklist**:
- [x] Implement connection creation in IcePanel
- [ ] Create connection update/merge logic
- [x] Develop support for connection metadata
- [ ] Implement connection validation
- [x] Add support for bidirectional connections
- [x] Create test suite for connection operations

### Issue 7: Mermaid-to-IcePanel Connection Creator Integration

**Description**: Integrate the Mermaid parser with IcePanel connection management.

**Checklist**:
- [x] Connect Mermaid parser to connection creation pipeline
- [ ] Implement object validation against IcePanel
- [x] Create command-line interface
- [x] Add configuration file support
- [x] Implement logging and error reporting
- [ ] Develop transaction handling for bulk operations
- [x] Create comprehensive test suite
- [x] Add documentation

### Issue 8: End-to-End Workflow Integration

**Description**: Ensure both tools work together seamlessly in the complete workflow.

**Checklist**:
- [ ] Create end-to-end test scenarios
- [ ] Implement shared configuration options
- [ ] Develop workflow documentation
- [ ] Create example projects and templates
- [ ] Add user guides
- [ ] Implement error handling and recovery
- [ ] Create integration tests with real IcePanel instances

### Issue 9: CI/CD Integration

**Description**: Integrate the tools into CI/CD pipelines.

**Checklist**:
- [ ] Create GitHub Actions workflow
- [ ] Implement GitLab CI pipeline
- [ ] Develop Docker containers for both tools
- [ ] Create sample CI configuration files
- [ ] Add automated testing in CI
- [ ] Implement versioning strategy
- [ ] Create release management process

### Issue 10: Metrics and Reporting

**Description**: Add capabilities for tracking and reporting on architectural changes.

**Checklist**:
- [ ] Implement change detection logic
- [ ] Create reporting formats (HTML, Markdown, JSON)
- [ ] Add metrics collection
- [ ] Develop visualization for architectural changes
- [ ] Implement alerting for significant changes
- [ ] Create dashboard for tracking design implementation
- [ ] Add historical trend analysis
