# Overview

Provides progress updates to users during long-running OpenCode operations to improve user experience.

# Details

- Adds progress callback parameter to SendMessage and pollForResponse
- Removes timeout, polls indefinitely until response is ready
- Every 10 minutes sends "Still processing..." message to user
- Quick operations return immediate response
- Final complete message always sent when ready

# File Paths

- internal/opencode/client.go
- internal/bot/handlers.go
