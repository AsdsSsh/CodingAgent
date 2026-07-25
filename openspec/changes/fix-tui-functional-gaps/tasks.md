## 1. Fix listenProgress timeout

- [x] 1.1 Remove 5s timeout from `listenProgress`, use pure blocking channel read

## 2. Fix emitProgress missing CurrentTool

- [x] 2.1 Add tool name emission in ReActLoop's tool execution loop

## 3. Fix permission modal dead-end

- [x] 3.1 Add keyboard handling for permission modal: y/Enter=Allow, n/Esc=Deny, a=Always Allow

## 4. Improve API key error visibility

- [x] 4.1 Make initialization errors more visible with red banner styling
