## ADDED Requirements

### Requirement: Full-screen terminal UI with header, viewport, and input
The system SHALL provide a bubbletea-based full-screen TUI with a header bar, scrollable output viewport, multi-line input area, and a help bar.

#### Scenario: TUI layout on startup
- **WHEN** the application starts
- **THEN** the TUI SHALL render a header showing model name, permission level, and token count; a scrollable viewport for conversation history; a multi-line text input area; and a bottom help bar showing keyboard shortcuts

#### Scenario: Window resize adaptation
- **WHEN** the terminal window is resized
- **THEN** the TUI SHALL recalculate component sizes and re-render the layout to fit the new dimensions

### Requirement: Multi-line task input with Ctrl+Enter submission
The system SHALL accept multi-line input in the textarea component and submit the task when the user presses Ctrl+Enter.

#### Scenario: Single-line task submission
- **WHEN** the user types a task and presses Ctrl+Enter
- **THEN** the system SHALL clear the input area, append the task to the conversation viewport as a user message, and begin agent execution

#### Scenario: Multi-line task input
- **WHEN** the user types multiple lines in the textarea and presses Ctrl+Enter
- **THEN** the system SHALL submit the entire multi-line content as a single task

### Requirement: Real-time agent progress display
The system SHALL display agent execution progress in real time, including the current step number, active tool name, and conversation token count.

#### Scenario: Progress updates during execution
- **WHEN** the ReActLoop emits an AgentState update
- **THEN** the TUI SHALL update the status line to show the current step, tool name, and token count within the same rendering frame

#### Scenario: Tool call started notification
- **WHEN** a tool execution begins
- **THEN** the TUI SHALL append a tool-call entry to the viewport with the tool name and a running status indicator

#### Scenario: Tool call completed notification
- **WHEN** a tool execution finishes
- **THEN** the TUI SHALL update the corresponding tool-call entry in the viewport with success/failure status

#### Scenario: Spinner animation during execution
- **WHEN** the agent is running (between progress updates)
- **THEN** the TUI SHALL display an animated spinner in the status bar

### Requirement: Permission prompt as blocking modal dialog
The system SHALL display a modal dialog when a tool requires permission escalation, blocking the ReActLoop until the user responds.

#### Scenario: Permission modal appears on PROMPT verdict
- **WHEN** the PolicyEngine returns a PROMPT verdict for a tool call
- **THEN** the TUI SHALL render a modal overlay showing the tool name, required permission level, and reason; and SHALL block the ReActLoop goroutine until the user selects Allow or Deny

#### Scenario: User allows permission once
- **WHEN** the user selects "Allow" in the permission modal
- **THEN** the system SHALL send an ALLOW response to the ReActLoop, close the modal, and the tool SHALL execute

#### Scenario: User denies permission
- **WHEN** the user selects "Deny" in the permission modal
- **THEN** the system SHALL send a DENY response to the ReActLoop, close the modal, and an error observation SHALL be appended

#### Scenario: User permanently allows a tool
- **WHEN** the user selects "Always Allow" in the permission modal
- **THEN** the system SHALL register a permanent override for that tool in the PolicyEngine, send an ALLOW response, and close the modal

#### Scenario: Permission modal timeout
- **WHEN** the permission modal has been displayed for 120 seconds without user response
- **THEN** the system SHALL auto-deny the request, close the modal, and unblock the ReActLoop

### Requirement: Slash commands for REPL control
The system SHALL support slash commands (`/`) for model switching, permission changes, help display, history viewing, and exit.

#### Scenario: Model listing and switching
- **WHEN** the user types `/model` without arguments
- **THEN** the system SHALL display the current model and list all built-in presets
- **WHEN** the user types `/model <name>`
- **THEN** the system SHALL switch to the specified model and auto-detect its endpoint

#### Scenario: Permission level switching
- **WHEN** the user types `/permission execute`
- **THEN** the system SHALL update the current permission level to EXECUTE and reflect the change in the header

#### Scenario: Help display
- **WHEN** the user types `/help`
- **THEN** the system SHALL display available commands and keyboard shortcuts in the viewport

#### Scenario: Exit REPL
- **WHEN** the user types `/exit` or presses Ctrl+C while the agent is idle
- **THEN** the system SHALL quit the TUI with exit code 0

### Requirement: Conversation history scrolling and navigation
The system SHALL maintain conversation history in a scrollable viewport, with Tab toggling focus between the input area and the viewport.

#### Scenario: Scroll through conversation history
- **WHEN** the viewport has focus and the user presses Up/Down or PageUp/PageDown
- **THEN** the viewport SHALL scroll through the conversation history

#### Scenario: Tab toggles focus
- **WHEN** the user presses Tab
- **THEN** focus SHALL toggle between the input area and the viewport

#### Scenario: Ctrl+C cancels running agent
- **WHEN** the agent is executing and the user presses Ctrl+C
- **THEN** the system SHALL cancel the current agent execution via context cancellation and return to idle state
