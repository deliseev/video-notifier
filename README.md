# Video Notifier

A clean, modular, and extensible service to monitor video playlists across various platforms. Built with Go, following **Domain-Driven Design (DDD)** and **Hexagonal Architecture (Ports & Adapters)** principles.

## Features
- **Platform Agnostic**: Currently supports YouTube, designed to easily extend to RuTube, Okko, and others.
- **TDD-First**: Core business logic is fully covered by unit tests with interface-based mocking.
- **Concurrent**: Handles multiple playlists simultaneously using Go's goroutines.
- **Configuration-as-Code**: Easy playlist management via YAML configuration with hot-reloading support (via `fsnotify`).
- **Graceful Shutdown**: Ensures the service stops cleanly, preventing data corruption.

## Architecture
This project follows the **Ports and Adapters** pattern:
- `internal/domain`: Pure business entities (no dependencies).
- `internal/usecase`: The "checker" service that implements orchestration logic.
- `internal/infrastructure`: Adapters for external systems (YouTube RSS, Memory/SQLite storage, Logging).

## Getting Started

### Prerequisites
- Go 1.22+

### Configuration
Create a `config.yaml` file in the root directory:

```yaml
database_path: "videos.db"
telegram_token: "your_bot_token"
playlists:
  - id: "PLwKZ4_gthFJ0uavGK_RQy_l6DiDl_l6jo"
    source: "youtube"
```

### Running the service
```bash
go run cmd/notifier/main.go
```

## Development
- **Testing**: Run core logic tests:
  ```bash
  go test -v ./internal/usecase/...
  ```
- **Adding new platforms**: To add a new source, simply implement the `VideoFetcher` interface in `internal/infrastructure/` and register it in the `getFetcher` factory in `main.go`.

## Roadmap
- [ ] Implement SQLite persistence for cross-process state.
- [ ] Add Telegram Bot adapter for notifications.
- [ ] Integrate `fsnotify` for live config reloading.
