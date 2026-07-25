## ADDED Requirements

### Requirement: MCP server configuration via YAML
The system SHALL support configuring MCP (Model Context Protocol) servers in the YAML configuration file, with support for stdio and HTTP transports.

#### Scenario: Stdio transport configuration
- **WHEN** an MCP server is configured with `transport: stdio`, a `command`, and optional `args`
- **THEN** the system SHALL launch the command as a subprocess and communicate over stdin/stdout using the MCP protocol

#### Scenario: HTTP transport configuration
- **WHEN** an MCP server is configured with `transport: http` and a `url`
- **THEN** the system SHALL connect to the URL using HTTP streaming transport

#### Scenario: Environment variables for MCP server process
- **WHEN** an MCP server configuration includes `env` key-value pairs
- **THEN** the system SHALL set those environment variables on the server subprocess

### Requirement: MCP tools integrate with ToolRegistry
The system SHALL register MCP-provided tools into the existing ToolRegistry, making them available to the LLM alongside built-in tools.

#### Scenario: MCP tools appear in LLM tool list
- **WHEN** an MCP server is connected and its tools are registered
- **THEN** the `ToLLMFormats()` method SHALL include both built-in and MCP tools

#### Scenario: MCP tool execution delegates to MCP server
- **WHEN** the LLM calls an MCP-provided tool
- **THEN** the system SHALL route the call to the appropriate MCP server and return the response as a tool result observation

#### Scenario: MCP tool name collision with built-in
- **WHEN** an MCP server provides a tool with the same name as a built-in tool
- **THEN** the system SHALL log a warning and skip the duplicate MCP tool registration

### Requirement: Lazy MCP server connection
The system SHALL defer MCP server connection until the first tool from that server is actually invoked, to minimize startup time.

#### Scenario: Server not connected on startup
- **WHEN** the agent starts and MCP servers are configured
- **THEN** the system SHALL cache the server configurations but SHALL NOT establish connections until a tool from that server is called

#### Scenario: Connection established on first tool call
- **WHEN** the LLM requests a tool from an MCP server that has not been connected yet
- **THEN** the system SHALL establish the connection, retrieve the tool list, and execute the requested tool

#### Scenario: Connection failure produces error observation
- **WHEN** an MCP server fails to connect or the subprocess exits abnormally
- **THEN** the system SHALL return an error observation to the LLM, allowing it to try an alternative approach
