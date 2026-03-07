2026-03-07-14-30 | Fixed JSONL parsing issue where embedded newlines in JSON string values caused "No response from OpenCode" errors
2026-03-05-21-40 | Fixed config.toml not being created when running `opencode-telegram new` command
2026-03-05-22-00 | Fixed config set command overwriting entire config file instead of merging
2026-03-05-22-00 | Fixed allowed_users config array parsing issue with JSON array strings
2026-03-06-14-00 | Fixed empty response handling in OpenCode client to return meaningful error
2026-03-06-14-00 | Fixed model object format to use providerID and modelID fields
2026-03-06-14-00 | Fixed message response polling to retrieve responses correctly
2026-03-06-14-00 | Fixed excessive polling by implementing smarter polling intervals
2026-03-06-14-00 | Fixed OpenCode server communication with proper session management
