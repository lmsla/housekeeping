# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

BiMAP Housekeeping is an Elasticsearch index lifecycle management service written in Go. It automates index operations like deletion, allocation (hot/warm/cold tiering), force merge, rollover, and more based on configurable filters (age, pattern, space, water level).

**Key characteristics:**
- Distributed as RPM/DEB packages for production deployment
- Runs as a systemd service or single-shot execution
- Supports test mode (dry-run) to preview operations without execution
- Logs operations to both local files and back to Elasticsearch

## Build and Development Commands

### Building the Application

```bash
# Navigate to the main source directory
cd es-curator

# Build the binary
go build -o bimap-housekeeping main.go

# The binary will be created at: es-curator/bimap-housekeeping
```

### Running Tests

```bash
cd es-curator
go test ./... -v
```

### Running the Application Locally

**Important:** The application expects configuration files (`setting.yml` and `config.yml`) in the same directory as the binary.

```bash
# Copy sample configs (first time only)
cp setting.yml.sample setting.yml
cp config.yml.sample config.yml

# Edit configs with your ES connection and actions
vim setting.yml
vim config.yml

# Run in test mode (dry-run, no actual changes)
./bimap-housekeeping  # Ensure test_mode: true in setting.yml

# Run in execution mode
# Set test_mode: false and execute_cron: false in setting.yml, then:
./bimap-housekeeping
```

### Running as a Systemd Service

After package installation (RPM/DEB), configs are located at `/etc/bimap-housekeeping/`:

```bash
# Start service
systemctl start bimap-housekeeping

# Stop service
systemctl stop bimap-housekeeping

# Restart service
systemctl restart bimap-housekeeping

# Enable auto-start on boot
systemctl enable bimap-housekeeping

# Check service status
systemctl status bimap-housekeeping

# View logs
tail -f /var/log/bimap-housekeeping/housekeeping_*.log
```

## Architecture Overview

### Code Structure

```
es-curator/
├── main.go              # Entry point: config loading, ES client init, cron scheduling
├── global/              # Global variables and shared state
├── structs/             # Data structures for config and actions
│   └── env.go          # EnviromentModel, ActionStruct, Option, Filter definitions
├── utils/              # Configuration loading and cron management
├── log_record/         # Structured logging (JSON format) with ES writeback support
├── metrics/            # Local metrics collection (operations, resource usage, ES health)
└── job/                # Core business logic
    ├── job.go          # ES client initialization
    ├── actions.go      # Action control flow (Action_controll, ActionExecutor)
    ├── indices.go      # Index operations (delete, close, open, rollover)
    ├── filters.go      # Filter processing (age, pattern, space, water_level)
    ├── filters_node.go # Node role filtering
    └── tools.go        # ES API helpers (allocation, forcemerge, etc.)
```

### Configuration Architecture

**Two-file configuration system:**

1. **setting.yml** - Environment and service settings:
   - ES connection (url, credentials)
   - Execution mode (test_mode, execute_cron, period)
   - Logging (path, rotation, ES writeback)

2. **config.yml** - Action definitions:
   - List of actions to execute
   - Each action has: type, description, options, filters
   - Actions execute sequentially with optional delays

### Action Execution Flow

1. **Action_controll()** in [actions.go](es-curator/job/actions.go:231) orchestrates execution
2. For each enabled action:
   - Generate unique execution ID (UUID)
   - Process filters to get matching indices
   - Execute action via ActionExecutor
   - Apply delay if specified
3. **ActionExecutor** handles test vs. execution mode branching and metrics collection

### Supported Actions

| Action | Description | Key Options | Typical Use Case |
|--------|-------------|-------------|------------------|
| `delete_indices` | Delete matching indices | - | Cleanup old logs |
| `allocation` | Move indices to tier | key, value, allocation_type | Hot→Warm→Cold tiering |
| `forcemerge` | Reduce segment count | max_num_segment | Optimize read-heavy indices |
| `close` | Close indices | - | Archive rarely-used data |
| `open` | Reopen closed indices | - | Restore access |
| `rollover` | Alias-based index rotation | rollover_alias, max_size, max_age, max_docs | Time-series data management |

**Rollover specifics:**
- Unlike other actions, rollover doesn't use filters (operates on alias directly)
- Requires pre-configured alias with `is_write_index: true` on one index
- See [docs/ROLLOVER_GUIDE.md](docs/ROLLOVER_GUIDE.md) for detailed setup

### Filter System

**Filter combinations:**
- **pattern** (prefix/suffix/regex): Required for all non-rollover actions to avoid system index accidents
- **age** (older/younger/range): Time-based filtering
- **space**: Keep newest N GB, delete older indices
- **water_level**: Delete oldest indices when cluster disk usage exceeds threshold
- **node_role** (h/w/c): Used with allocation to target specific tiers

**Important constraints:**
- Cannot mix time filters (age) with space filters (space/water_level)
- Cannot combine space and water_level in same action
- Always include pattern filter as safety measure

### Test Mode

**Purpose:** Preview operations without making changes

**Behavior:**
- Set `test_mode: true` in setting.yml
- Actions log matched indices but don't execute operations
- Logs clearly marked with "Test mode" label
- Essential for validating configs before production runs

**How it works:**
- [ActionExecutor.ExecuteOnIndices](es-curator/job/actions.go:27) branches on `global.EnvConfig.INFORMATION.TestMode`
- Test mode: logs index details via `logTestMode()`
- Execution mode: calls actual operation function (DeleteIndex, Allocation, etc.)

### Metrics and Monitoring

**Local metrics system** (metrics/metrics.go):
- Operation counts (total, success, failure) by action type
- Average execution time per operation
- ES health status (connection, cluster health, response time)
- Resource usage (memory, heap, goroutine count)

**Metrics output:**
- Single-run mode: prints summary to log after completion (see [printMetricsSummary](es-curator/main.go:96))
- Cron mode: updates resource metrics every 30 seconds
- All metrics logged with `logType: "Metrics"` for ES writeback filtering

### Logging Strategy

**Three log streams:**
1. **Standard logger** (global.Logger): Procedures and high-level flow
2. **Detail logger** (global.Detail_Logger): Per-index operation details
3. **Stderr logger** (global.Stderr_logger): Critical startup errors

**Log types** (logType field):
- `Procedures`: Action execution flow
- `Detail`: Individual index processing
- `Metrics`: Performance and health data

**ES writeback:**
- Enabled via `log.toes: true` in setting.yml
- Writes logs to ES for Kibana visualization
- Health check interval configurable (log.health_check_interval)

## Development Guidelines

### Adding a New Action

1. **Define option fields** in [structs/env.go](es-curator/structs/env.go) Option struct
2. **Implement handler function** in actions.go (e.g., `handleNewAction`)
3. **Add ES operation function** in appropriate job/*.go file
4. **Register in switch statement** in [executeAction](es-curator/job/actions.go:192)
5. **Update metrics** tracking in [ActionExecutor.ExecuteOnIndices](es-curator/job/actions.go:27)

Example handler pattern:
```go
func handleNewAction(uuid string, indexList []string, action structs.Actiond) {
    executor := &ActionExecutor{UUID: uuid, Action: "new_action"}
    if indexList != nil {
        executor.LogActionSummary("NewAction", indexList)
        executor.ExecuteOnIndices(indexList, NewActionOperation)
    }
}
```

### Adding a New Filter Type

1. **Add filter struct fields** in [structs/env.go](es-curator/structs/env.go) Filter struct
2. **Implement filter logic** in job/filters.go
3. **Register in processFilters** in [actions.go](es-curator/job/actions.go:117)
4. **Update documentation** in README.md filter sections

### Error Handling Patterns

- **Panic recovery**: ActionExecutor uses defer/recover to catch panics and mark operations as failed
- **Retry with backoff**: ES operations (e.g., ClusterHealthWithRetry in [indices.go](es-curator/job/indices.go:74)) use exponential backoff
- **Safety checks**: Action_controll validates config before execution ([actions.go:234](es-curator/job/actions.go:234))

### Testing Practices

**Before production:**
1. Enable test_mode and run with production-like configs
2. Verify matched indices in logs match expectations
3. Check that pattern filters prevent system index selection
4. Disable test_mode and run once without execute_cron
5. Monitor logs and metrics summary
6. Enable execute_cron for scheduled operation

**Common pitfall:** Forgetting to set test_mode: false can lead to belief that actions executed when they didn't.

## Important Constraints and Safety

### Configuration Validation
- YAML indentation errors cause parse failures
- Filter type mixing (age + space) causes undefined behavior
- Missing pattern filter risks system index deletion

### Production Safety Checklist
- [ ] Pattern filters on all non-rollover actions
- [ ] Test mode validation completed
- [ ] Log path writable and has sufficient disk space
- [ ] ES credentials correct and have necessary permissions
- [ ] Delay values appropriate for cluster size (large operations may need longer delays)
- [ ] Rollover aliases pre-configured with write index

### ES API Permissions Required
- `indices:admin/delete` - delete_indices
- `indices:admin/close` - close
- `indices:admin/open` - open
- `indices:admin/forcemerge` - forcemerge
- `indices:admin/settings/update` - allocation
- `indices:admin/rollover` - rollover
- `cluster:monitor/health` - health checks
- `indices:admin/create` (write to log index if log.toes enabled)

## Deployment Notes

### Package Installation
- RPM/DEB packages install to standard locations
- Binary: `/usr/local/bin/bimap-housekeeping` or similar
- Configs: `/etc/bimap-housekeeping/`
- Logs: Configurable (default `/var/log/bimap-housekeeping`)
- Service file: `/etc/systemd/system/bimap-housekeeping.service`

### Service Configuration
See [docs/ubuntu Service 啟用步驟.md](docs/ubuntu Service 啟用步驟.md) for systemd setup details.

Key service parameters:
- `WorkingDirectory`: Must contain config files
- `ExecStart`: Path to binary
- `Restart=always`: Ensures resilience

### Common Deployment Patterns

**Pattern 1: Time-based cleanup**
```yaml
- action: delete_indices
  filters:
    - filtertype: pattern
      kind: prefix
      value: logs-
    - filtertype: age
      direction: older
      unit: days
      unit_count: 30
```

**Pattern 2: Hot→Warm→Cold tiering**
```yaml
# Move 1-day-old indices to warm
- action: allocation
  options:
    key: _tier_preference
    value: data_warm
    allocation_type: include
  filters:
    - filtertype: age
      direction: older
      unit: days
      unit_count: 1
    - filtertype: pattern
      kind: prefix
      value: logs-

# Move 7-day-old indices to cold
- action: allocation
  options:
    key: _tier_preference
    value: data_cold
    allocation_type: include
  filters:
    - filtertype: age
      direction: older
      unit: days
      unit_count: 7
```

**Pattern 3: Rollover + lifecycle**
See [docs/ROLLOVER_GUIDE.md](docs/ROLLOVER_GUIDE.md) for comprehensive examples.

## Additional Documentation

- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) - Detailed architecture diagrams and component descriptions
- [docs/ROLLOVER_GUIDE.md](docs/ROLLOVER_GUIDE.md) - Complete rollover setup and troubleshooting
- [docs/RELEASE.md](docs/RELEASE.md) - Version history and changelog
- [README.md](README.md) - User manual (Chinese) with config reference

## Troubleshooting Quick Reference

**Service won't start:**
- Check config file syntax: `go run main.go` shows parse errors
- Verify ES connectivity: curl ES URL with credentials
- Check file permissions on configs and log path

**Actions not executing:**
- Confirm test_mode: false in setting.yml
- Check disable_action: false on each action
- Verify filters match indices: enable test_mode to preview

**Rollover fails:**
- Ensure alias exists: `GET /_aliases/your-alias`
- Verify write index set: alias must have `is_write_index: true` on one index
- Check index naming follows rollover convention (e.g., logs-000001)

**Memory issues:**
- Reduce number of indices processed per run (use stricter filters)
- Increase delay between actions to allow GC
- Monitor via resource metrics in log output
