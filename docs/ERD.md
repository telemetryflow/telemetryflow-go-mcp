# TFO-MCP Entity Relationship Diagrams

> Entity Relationship Diagrams for TelemetryFlow MCP Server

---

## Table of Contents

- [Overview](#overview)
- [Complete Domain ERD](#complete-domain-erd)
- [Session Aggregate ERD](#session-aggregate-erd)
- [Conversation Aggregate ERD](#conversation-aggregate-erd)
- [Value Objects ERD](#value-objects-erd)
- [Database Schema](#database-schema)

---

## Overview

This document provides Entity Relationship Diagrams that describe the data structures and relationships within the TFO-MCP Server.

```mermaid
flowchart LR
    subgraph ERD["Entity Relationship Diagrams"]
        DOMAIN["Domain ERD"]
        SESSION["Session ERD"]
        CONV["Conversation ERD"]
        VO["Value Objects"]
        DB["Database Schema"]
    end

    DOMAIN --> |"Describes"| STRUCT["Domain Structure"]
    SESSION --> |"Details"| SAGG["Session Aggregate"]
    CONV --> |"Details"| CAGG["Conversation Aggregate"]
    VO --> |"Defines"| TYPES["Value Types"]
    DB --> |"Maps to"| TABLES["PostgreSQL Tables"]

    style DOMAIN fill:#e3f2fd,stroke:#2196f3
    style SESSION fill:#e8f5e9,stroke:#4caf50
    style CONV fill:#fff3e0,stroke:#ff9800
    style VO fill:#fce4ec,stroke:#e91e63
    style DB fill:#f3e5f5,stroke:#9c27b0
```

---

## Complete Domain ERD

```mermaid
erDiagram
    SESSION ||--o{ CONVERSATION : contains
    SESSION ||--o{ TOOL : registers
    SESSION ||--o{ RESOURCE : registers
    SESSION ||--o{ PROMPT : registers
    CONVERSATION ||--o{ MESSAGE : contains
    MESSAGE ||--o{ CONTENT_BLOCK : contains
    TOOL ||--o{ TOOL_RESULT : produces
    RESOURCE ||--o{ RESOURCE_CONTENT : provides

    SESSION {
        uuid id PK
        string client_name
        string client_version
        string protocol_version
        enum state
        datetime created_at
        datetime updated_at
    }

    CONVERSATION {
        uuid id PK
        uuid session_id FK
        string model
        string system_prompt
        enum status
        int max_tokens
        float temperature
        datetime created_at
    }

    MESSAGE {
        uuid id PK
        uuid conversation_id FK
        enum role
        datetime created_at
    }

    CONTENT_BLOCK {
        uuid id PK
        uuid message_id FK
        enum type
        text content
        string mime_type
    }

    TOOL {
        string name PK
        string description
        json input_schema
        boolean enabled
    }

    TOOL_RESULT {
        uuid id PK
        string tool_name FK
        json content
        boolean is_error
        datetime executed_at
    }

    RESOURCE {
        string uri PK
        string name
        string description
        string mime_type
        bigint size
    }

    RESOURCE_CONTENT {
        uuid id PK
        string uri FK
        text content
        string blob
    }

    PROMPT {
        string name PK
        string description
        json arguments
    }
```

---

## Session Aggregate ERD

```mermaid
erDiagram
    SESSION_AGGREGATE ||--|| SESSION_ID : has
    SESSION_AGGREGATE ||--|| CLIENT_INFO : has
    SESSION_AGGREGATE ||--|| CAPABILITIES : has
    SESSION_AGGREGATE ||--o{ TOOL_ENTITY : contains
    SESSION_AGGREGATE ||--o{ RESOURCE_ENTITY : contains
    SESSION_AGGREGATE ||--o{ PROMPT_ENTITY : contains
    SESSION_AGGREGATE ||--o{ CONVERSATION_AGGREGATE : manages

    SESSION_AGGREGATE {
        SessionID id
        SessionState state
        ClientInfo client
        Capabilities capabilities
        datetime created_at
    }

    SESSION_ID {
        uuid value
    }

    CLIENT_INFO {
        string name
        string version
    }

    CAPABILITIES {
        boolean tools
        boolean resources
        boolean prompts
        boolean logging
    }

    TOOL_ENTITY {
        string name
        string description
        JSONSchema input_schema
        ToolHandler handler
        boolean enabled
    }

    RESOURCE_ENTITY {
        ResourceURI uri
        string name
        MimeType mime_type
        ResourceReader reader
    }

    PROMPT_ENTITY {
        string name
        string description
        PromptArgument[] arguments
        PromptGenerator generator
    }

    CONVERSATION_AGGREGATE {
        ConversationID id
        Model model
        Message[] messages
        ConversationStatus status
    }
```

---

## Conversation Aggregate ERD

```mermaid
erDiagram
    CONVERSATION_AGGREGATE ||--|| CONVERSATION_ID : has
    CONVERSATION_AGGREGATE ||--|| MODEL : uses
    CONVERSATION_AGGREGATE ||--o{ MESSAGE_ENTITY : contains
    CONVERSATION_AGGREGATE ||--o{ TOOL_BINDING : uses
    MESSAGE_ENTITY ||--|| MESSAGE_ID : has
    MESSAGE_ENTITY ||--|| ROLE : has
    MESSAGE_ENTITY ||--o{ CONTENT_BLOCK : contains

    CONVERSATION_AGGREGATE {
        ConversationID id
        SessionID session_id
        Model model
        SystemPrompt system_prompt
        ConversationStatus status
        int max_tokens
        float temperature
    }

    CONVERSATION_ID {
        uuid value
    }

    MODEL {
        string value
    }

    MESSAGE_ENTITY {
        MessageID id
        Role role
        ContentBlock[] content
        datetime created_at
    }

    MESSAGE_ID {
        uuid value
    }

    ROLE {
        enum value "user|assistant"
    }

    CONTENT_BLOCK {
        ContentType type
        string text
        string data
        string mime_type
    }

    TOOL_BINDING {
        string tool_name
        boolean enabled
    }
```

---

## Value Objects ERD

```mermaid
erDiagram
    VALUE_OBJECTS ||--o{ IDENTIFIERS : contains
    VALUE_OBJECTS ||--o{ CONTENT_TYPES : contains
    VALUE_OBJECTS ||--o{ MCP_TYPES : contains

    IDENTIFIERS {
        SessionID session_id "uuid"
        ConversationID conversation_id "uuid"
        MessageID message_id "uuid"
        ToolID tool_id "uuid"
        ResourceID resource_id "uuid"
        PromptID prompt_id "uuid"
        RequestID request_id "uuid"
    }

    CONTENT_TYPES {
        ContentType type "text|image|resource"
        Role role "user|assistant"
        Model model "claude-*"
        MimeType mime_type "text/*"
        TextContent text "string"
        SystemPrompt system "string"
    }

    MCP_TYPES {
        JSONRPCVersion version "2.0"
        MCPMethod method "string"
        MCPCapability capability "string"
        MCPLogLevel log_level "enum"
        MCPProtocolVersion protocol "2024-11-05"
        MCPErrorCode error_code "int"
    }
```

---

## Database Schema

### PostgreSQL Schema

```mermaid
erDiagram
    sessions ||--o{ conversations : has
    sessions ||--o{ session_tools : has
    sessions ||--o{ session_resources : has
    sessions ||--o{ session_prompts : has
    conversations ||--o{ messages : contains
    messages ||--o{ content_blocks : contains

    sessions {
        uuid id PK
        string client_name
        string client_version
        string protocol_version
        string state
        jsonb capabilities
        timestamp created_at
        timestamp updated_at
        timestamp closed_at
    }

    conversations {
        uuid id PK
        uuid session_id FK
        string model
        text system_prompt
        string status
        int max_tokens
        decimal temperature
        timestamp created_at
        timestamp updated_at
    }

    messages {
        uuid id PK
        uuid conversation_id FK
        string role
        jsonb metadata
        timestamp created_at
    }

    content_blocks {
        uuid id PK
        uuid message_id FK
        string type
        text text_content
        bytea blob_content
        string mime_type
        int sequence
    }

    session_tools {
        uuid id PK
        uuid session_id FK
        string tool_name
        text description
        jsonb input_schema
        boolean enabled
        timestamp registered_at
    }

    session_resources {
        uuid id PK
        uuid session_id FK
        string uri
        string name
        text description
        string mime_type
        timestamp registered_at
    }

    session_prompts {
        uuid id PK
        uuid session_id FK
        string name
        text description
        jsonb arguments
        timestamp registered_at
    }
```

### ClickHouse Analytics Schema

```mermaid
erDiagram
    mcp_requests {
        uuid request_id PK
        uuid session_id
        string method
        datetime timestamp
        int duration_ms
        boolean success
        string error_message
    }

    tool_executions {
        uuid execution_id PK
        uuid session_id
        string tool_name
        datetime timestamp
        int duration_ms
        boolean success
        jsonb input_params
        jsonb result
    }

    claude_api_calls {
        uuid call_id PK
        uuid session_id
        uuid conversation_id
        string model
        datetime timestamp
        int duration_ms
        int input_tokens
        int output_tokens
        float cost
        boolean success
    }

    session_metrics {
        uuid session_id PK
        datetime start_time
        datetime end_time
        int total_requests
        int total_tools_called
        int total_api_calls
        int total_tokens
    }
```

---

## Related Documentation

- [Data Flow Diagrams](DFD.md)
- [Architecture Guide](ARCHITECTURE.md)
- [Development Guide](DEVELOPMENT.md)

---

<div align="center">

**[Back to Documentation Index](README.md)**

</div>
