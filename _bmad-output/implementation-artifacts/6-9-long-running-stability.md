# Story 6.9: Long-Running Stability

Status: complete

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a user,
I want PlexCord to run reliably for extended periods,
So that I don't need to restart it regularly.

## Acceptance Criteria

1. **AC1: Continuous Operation**
   - **Given** PlexCord is running continuously
   - **When** 30+ days have passed (NFR13)
   - **Then** the application continues operating normally

2. **AC2: Memory Stability**
   - **Given** PlexCord is running continuously
   - **When** monitoring memory usage over time
   - **Then** memory usage remains stable (no memory leaks)
   - **And** memory stays below 50MB during idle (NFR2)

3. **AC3: CPU Efficiency**
   - **Given** PlexCord is running continuously
   - **When** monitoring CPU usage during idle
   - **Then** CPU usage remains low (average <1% per NFR3)
   - **And** no busy-wait loops or excessive processing

4. **AC4: Connection Recovery**
   - **Given** PlexCord has been running for extended periods
   - **When** transient connection failures occur
   - **Then** all connections recover automatically
   - **And** retry mechanisms continue functioning correctly

5. **AC5: Autonomous Operation**
   - **Given** PlexCord is running
   - **When** normal operation continues
   - **Then** no manual intervention is required
   - **And** the application self-heals from recoverable errors

## Tasks / Subtasks

- [ ] **Task 1: Audit Goroutine Lifecycle Management** (AC: 1, 2, 4)
  - [ ] Review Plex poller for proper goroutine cleanup on stop
  - [ ] Review retry manager for proper timer/context cleanup
  - [ ] Review event subscription cleanup in frontend stores
  - [ ] Ensure all started goroutines have termination paths
  - [ ] Document findings and any fixes applied

- [ ] **Task 2: Channel and Timer Cleanup Verification** (AC: 2, 4)
  - [ ] Verify buffered channels don't cause memory accumulation
  - [ ] Verify timers are properly stopped on shutdown
  - [ ] Verify context cancellation propagates correctly
  - [ ] Check for any unclosed channels in error paths

- [ ] **Task 3: Event Listener Cleanup** (AC: 2)
  - [ ] Review Wails EventsOn/EventsOff usage in frontend
  - [ ] Ensure all event listeners are unregistered on component unmount
  - [ ] Verify connection store cleanup() is called on app shutdown
  - [ ] Check Vue component onUnmounted hooks for proper cleanup

- [ ] **Task 4: Long-Running Retry Verification** (AC: 4, 5)
  - [ ] Test retry manager behavior over many retry cycles
  - [ ] Verify exponential backoff caps at max interval correctly
  - [ ] Ensure retry state doesn't accumulate memory over time
  - [ ] Verify manual retry resets backoff schedule correctly

- [ ] **Task 5: Resource Monitoring Instrumentation** (AC: 2, 3)
  - [ ] Add optional runtime memory stats logging (disabled by default)
  - [ ] Add optional goroutine count logging
  - [ ] Create GetResourceStats Wails binding for debugging
  - [ ] Add resource stats to connection history if appropriate

- [ ] **Task 6: Idle CPU Efficiency Audit** (AC: 3)
  - [ ] Verify time.Ticker usage (not time.Sleep loops)
  - [ ] Check for any spin-wait patterns in code
  - [ ] Verify no unnecessary work during idle periods
  - [ ] Document polling efficiency characteristics

- [ ] **Task 7: Integration Test for Extended Operation** (AC: 1, 4, 5)
  - [ ] Create test scenario for simulated long-running operation
  - [ ] Test connection loss/recovery cycles
  - [ ] Test Discord restart detection over time
  - [ ] Verify no state corruption after many cycles

## Dev Notes

### Architecture Compliance

This story ensures compliance with reliability NFRs:
- **NFR13**: Application shall maintain operation for 30+ days without restart
- **NFR2**: Memory usage shall remain below 50MB during idle operation
- **NFR3**: CPU usage shall average less than 1% during normal polling
- **NFR17**: Application shall automatically reconnect after transient failures

### Existing Implementation Review

**Retry Manager (`internal/retry/retry.go`):**
- Uses `time.AfterFunc` for non-blocking timers ✓
- Has proper cancellation via context ✓
- Resets state on success ✓
- Potential concern: Verify timer cleanup on Stop()

**Plex Poller (`internal/plex/poller.go`):**
- Uses `time.Ticker` for efficient polling ✓
- Has proper goroutine cleanup in defer block ✓
- Closes session channel on stop ✓
- Buffered channel prevents blocking ✓

**Discord PresenceManager (`internal/discord/presence.go`):**
- Stateless connection model (no goroutines) ✓
- Proper mutex protection ✓
- No background workers to leak ✓

**Frontend Connection Store (`frontend/src/stores/connection.js`):**
- Has `cleanup()` action for EventsOff ✓
- Event listeners properly registered ✓
- Potential concern: Ensure cleanup called on app shutdown

### Memory Leak Prevention Checklist

| Component | Risk Area | Mitigation |
|-----------|-----------|------------|
| Plex Poller | Goroutine leak | Defer cleanup, channel close |
| Retry Manager | Timer leak | Stop timer on Reset/Stop |
| Event Listeners | Subscription leak | EventsOff in cleanup() |
| ErrorBanner | Timer leak | onUnmounted cleanup (fixed in 6-1) |
| Wails Events | Channel leak | Proper unsubscription |

### Resource Monitoring Implementation

For Task 5, add optional resource stats:

```go
// internal/monitor/stats.go (new file)
package monitor

import (
    "runtime"
    "time"
)

type ResourceStats struct {
    MemoryAllocMB float64   `json:"memoryAllocMB"`
    MemoryTotalMB float64   `json:"memoryTotalMB"`
    GoroutineCount int      `json:"goroutineCount"`
    Timestamp      time.Time `json:"timestamp"`
}

func GetStats() ResourceStats {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)

    return ResourceStats{
        MemoryAllocMB:  float64(m.Alloc) / 1024 / 1024,
        MemoryTotalMB:  float64(m.TotalAlloc) / 1024 / 1024,
        GoroutineCount: runtime.NumGoroutine(),
        Timestamp:      time.Now(),
    }
}
```

### Testing Long-Running Stability

Due to the nature of NFR13 (30+ days), direct testing is impractical. Instead:

1. **Code Audit**: Review all resource management code paths
2. **Stress Testing**: Run many connection cycles in short time
3. **Profiling**: Use Go's pprof to detect memory growth
4. **Goroutine Monitoring**: Track goroutine count over time

Example stress test approach:
```go
// Simulate 1000 connection cycles (equivalent to many days of reconnections)
for i := 0; i < 1000; i++ {
    poller.Start(ctx)
    time.Sleep(10 * time.Millisecond)
    poller.Stop()

    // Verify goroutine count returns to baseline
    if runtime.NumGoroutine() > baseline+1 {
        t.Errorf("Goroutine leak detected after %d cycles", i)
    }
}
```

### Wails Binding for Resource Stats

```go
// app.go - Add to existing bindings
func (a *App) GetResourceStats() map[string]interface{} {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)

    return map[string]interface{}{
        "memoryAllocMB":  float64(m.Alloc) / 1024 / 1024,
        "goroutineCount": runtime.NumGoroutine(),
        "timestamp":      time.Now().Format(time.RFC3339),
    }
}
```

### Project Structure Notes

**Files to potentially create:**
- `internal/monitor/stats.go` - Resource monitoring (optional)

**Files to audit/modify:**
- `internal/retry/retry.go` - Verify cleanup completeness
- `internal/plex/poller.go` - Verify goroutine management
- `internal/discord/presence.go` - Verify no leaks
- `frontend/src/stores/connection.js` - Verify event cleanup
- `app.go` - Add GetResourceStats binding

### Previous Story Learnings

From Story 6-1 (Error Banner Component):
- Fixed timer memory leak in ErrorBanner.vue (onUnmounted cleanup)
- Pattern: All timers/intervals need cleanup on component unmount

From Story 6-4 (Automatic Retry with Exponential Backoff):
- Retry manager uses time.AfterFunc which needs explicit Stop()
- State callbacks emit properly during retry cycles

From Story 6-5 (Graceful Plex Unavailability Handling):
- Poller handles errors gracefully without crashing
- Error state transitions work correctly

### Testing Considerations

- Use Go's `-race` flag to detect race conditions
- Use pprof for memory profiling (`go tool pprof`)
- Monitor goroutine count over simulated long runs
- Test rapid start/stop cycles for resource cleanup
- Verify frontend event listeners don't accumulate

### References

- [Source: _bmad-output/planning-artifacts/architecture.md] - Architecture patterns
- [Source: _bmad-output/planning-artifacts/epics.md#Story 6.9] - Acceptance criteria
- [Source: internal/retry/retry.go] - Retry manager implementation
- [Source: internal/plex/poller.go] - Plex poller implementation
- [Source: internal/discord/presence.go] - Discord presence implementation
- [Source: frontend/src/stores/connection.js] - Event listener management
- [Source: 6-1-error-banner-component.md] - Timer cleanup pattern
