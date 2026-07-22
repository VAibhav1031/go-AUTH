# go-AUTH — Containerized CLI Login System with Optional 2FA

A command-line authentication system built in Go, supporting user registration, login, account lockout, session expiry, and optional TOTP-based two-factor authentication (Google Authenticator compatible) ,QR code Generated for ease. Runs fully containerized with persistent SQLite storage.

## Features

- User registration with bcrypt password hashing (salted)
- Login with account lockout after repeated failed attempts
- Optional TOTP-based 2FA — enable/disable per account, compatible with Google Authenticator or any standard authenticator app
- Session management with configurable expiry
- Interactive CLI with command history and Tab-completion (via `chzyer/readline`)
- Context-aware command set: available commands change automatically depending on whether you're logged in
- SQLite storage, persisted across container restarts via a Docker volume

## Requirements

- Docker and Docker Compose
- (Optional, for local dev without Docker) Go 1.26+

## Getting Started

Clone the repo and build/run with Docker Compose:

```bash
git clone https://github.com/VAibhav1031/go-AUTH.git
cd go-AUTH
docker compose build
docker compose run --rm -it app_cli
```

> **Note:** Use `docker compose run --rm -it app_cli`, not `docker compose up`. This is an interactive CLI, and `up` doesn't attach your terminal's stdin the way an interactive session needs. `run` gives you a proper attached terminal, and `--rm` cleans up the container when you exit. Your data isn't affected either way — it's persisted separately in the `auth_db` volume.

> `docker compose up` only provide stderr/stdout  no stdin , tty which is needed for the interactive cli usage

Your database lives in a named Docker volume (`auth_db`, mounted at `/app/data` inside the container), so registered users and session state survive restarts and rebuilds. To wipe all data and start fresh:

```bash
docker compose down -v
```

## Running Locally (without Docker)

```bash
go mod download
go run ./cmd/go_auth_cli/
```

This will create a local SQLite file wherever the DB path resolves outside the container — adjust the connection path in `internal/cli/initiator.go` if you want it somewhere specific for local testing.

## Usage

On start, you'll see a prompt. Available commands depend on whether you're logged in:

**Before login:**

| Command    | Description                              |
|------------|-------------------------------------------|
| `register` | Create a new user account                 |
| `login`    | Log in with username/password (+ TOTP if enabled) |
| `help`     | Show available commands                   |
| `exit`     | Quit the program                          |

**After login:**

| Command       | Description                        |
|---------------|-------------------------------------|
| `whoami`      | Show current user details           |
| `enable-2fa`  | Enable TOTP-based 2FA               |
| `disable-2fa` | Disable 2FA                         |
| `logout`      | End the current session             |
| `help`        | Show available commands             |

Press **Tab** to auto-complete commands, and use the **Up/Down arrows** to recall command history.

### Registering

```
> register
Username: alice
Password: ********
Confirm Password: ********
Registered user with Id 1
```

Passwords are hashed with bcrypt before storage — the plaintext password is never written to disk.

### Logging in

```
> login
Username: alice
Password: ********
── Logged in ──
Username:      alice
Registered on: 2026-07-22 10:03:11
MFA status:    disabled
Session expires at: 2026-07-22 10:08:11
Last login:    (first login)
```

If 2FA is enabled on the account, you'll be prompted for your 6-digit authenticator code after the password check succeeds.

### Enabling 2FA

```
> enable-2fa
Secret (add this to your authenticator app): JBSWY3DPEHPK3PXP
Current 2FA code: 123456
2FA enabled.
```

Add the printed secret to an authenticator app (Google Authenticator, Authy, etc.), then confirm with the current code it generates to finish enabling 2FA.

### Account lockout

After 3 failed login attempts, the account is locked for 15 minutes. Attempts reset to 0 on a successful login.

### Session expiry

Sessions expire after a fixed timeout (default: 5 minutes) from login. Once expired, the CLI automatically returns you to the pre-login command set on your next input.

## Project Structure

```
.
├── cmd/go_auth_cli/       # Entrypoint
├── internal/
│   ├── cli/               # REPL loop, command handlers, DB models
│   └── logger/            # Structured (slog) logging setup
├── Dockerfile
├── docker-compose.yml
└── go.mod
```

## Database Schema

```sql
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    created_at DATETIME,
    mfa_enabled BOOL,
    mfa_secret TEXT,
    last_login DATETIME,
    attempts INTEGER,
    blocked_time DATETIME
);
```

Created automatically on first run if it doesn't already exist.

## Security Notes

- Passwords are hashed with bcrypt (default cost), never stored or logged in plaintext.
- Login credentials and 2FA codes are entered via interactive follow-up prompts rather than command-line arguments, so they never end up in shell history.
- Password input is masked (no echo) at the terminal.
- TOTP is implemented using [`pquerna/otp`](https://github.com/pquerna/otp), a well-tested RFC 6238 implementation, rather than a hand-rolled version.
- You can easily Scan the secret key with google-Authenticator APP for ease and use the 2FA code to verify

## Tech Stack

- **Go** 1.26
- **modernc.org/sqlite** — pure-Go SQLite driver (no CGO required)
- **jmoiron/sqlx** — thin wrapper over `database/sql` for struct scanning
- **chzyer/readline** — interactive prompt, history, and Tab-completion
- **golang.org/x/crypto/bcrypt** — password hashing
- **pquerna/otp** — TOTP generation/validation
- **mdp/qrterminal/v3** - QR Code Generator using the TOTP-secret URL
