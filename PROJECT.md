# Tubely - Project Documentation

This file documents the project for any LLM agent working in this repo.

## Overview
Starter code for the boot.dev course "Learn File Servers and CDNs with S3 and CloudFront" (`README.md`). It's a Go HTTP server for a toy video-hosting app ("Tubely") backed by SQLite, with JWT auth and (eventually, per the course) S3/CloudFront-backed video and thumbnail storage. Several handlers are intentionally unfinished stubs — see Incomplete/Stub Code below — since the course's lessons fill them in incrementally.

## Architecture
`main.go` builds an `http.ServeMux` directly in `main()` (no separate `makeHandler()`/test harness exists yet) and registers:
- `/app/`: static fileserver rooted at `FILEPATH_ROOT` (the `app/` dir: `index.html`, `app.js`, `styles.css`)
- `/assets/`: static fileserver rooted at `ASSETS_ROOT` (`assets/`, gitignored), wrapped in `cacheMiddleware` (`cache.go`) which sets `Cache-Control: max-age=3600`
- `POST /api/login`, `POST /api/refresh`, `POST /api/revoke`: auth (see below)
- `POST /api/users`: create a user
- `POST /api/videos`, `GET /api/videos`, `GET /api/videos/{videoID}`, `DELETE /api/videos/{videoID}`: video metadata CRUD, JWT-protected except `GET /api/videos/{videoID}`
- `POST /api/thumbnail_upload/{videoID}`: JWT-protected, see Incomplete/Stub Code
- `POST /api/video_upload/{videoID}`: empty stub, see Incomplete/Stub Code
- `POST /admin/reset`: wipes all DB tables, restricted to `PLATFORM=dev` (`403` otherwise)

`apiConfig` (`main.go`) holds `db` (`database.Client`), `jwtSecret`, `platform`, `filepathRoot`, `assetsRoot`, `s3Bucket`, `s3Region`, `s3CfDistribution`, `port` — all loaded from env vars in `main()`, each `log.Fatal`s if unset. `godotenv.Load(".env")` loads them locally.

`respondWithJSON`/`respondWithError` (`json.go`) centralize JSON responses.

## Database (`internal/database`)
No goose, no SQLC — this project uses raw `database/sql` with `github.com/mattn/go-sqlite3` against a local file (`DB_PATH`, default `./tubely.db`, gitignored).
- `database.NewClient(pathToDB)` opens the DB and runs `autoMigrate()` (`database.go`), which `CREATE TABLE IF NOT EXISTS` for `users`, `refresh_tokens`, `videos` — there's no migration history/versioning, just idempotent DDL run on every startup.
- `videos.video_url TEXT TEXT` (note the doubled type token in `database.go`) — harmless in SQLite (column affinity, not a real type), but worth fixing if anyone touches that schema.
- `users`: `id` (TEXT/UUID PK), `created_at`, `updated_at`, `password`, `email` (unique).
- `refresh_tokens`: `token` (PK), `created_at`, `updated_at`, `revoked_at` (nullable), `user_id`, `expires_at`.
- `videos`: `id`, `created_at`, `updated_at`, `title`, `description`, `thumbnail_url`, `video_url`, `user_id` — `thumbnail_url` is now a `data:<media-type>;base64,<data>` URL written directly to the column by the thumbnail upload handler; `video_url` is unwritten (video upload is a no-op stub).
- `Client.Reset()` deletes all rows from `refresh_tokens`, `users`, `videos` in that order (FK-safe), used by `POST /admin/reset`.
- Queries (`users.go`, `videos.go`, `refresh_tokens.go`) are plain `?`-placeholder SQL strings, one Go method per query — no generated code.

## Auth (`internal/auth/auth.go`)
- Password hashing via `argon2id` (`HashPassword`/`CheckPasswordHash`).
- JWTs via `golang-jwt/jwt/v5`: `MakeJWT(userID, secret, expiresIn)` signs HS256 with `Issuer: "tubely-access"` (`TokenTypeAccess` const). `ValidateJWT` checks signature, expiry, **and** issuer — returns `"invalid issuer"` error if mismatched.
- `GetBearerToken`/`GetAPIKey` extract `Authorization: Bearer <token>` / `Authorization: ApiKey <key>` respectively, each its own small function, both erroring on a missing/malformed header.
- `MakeRefreshToken()` returns a 256-bit random value hex-encoded, stored opaquely in `refresh_tokens` (not a JWT).
- Login issues a 30-day access token (`handler_login.go`); refresh tokens last 60 days. `POST /api/refresh` (`handler_refresh.go`) re-issues a 1-hour access token from a valid refresh token.

## Incomplete/Stub Code
This is starter code with course exercises left for the student — do not "fix" these without checking with the user first, they may be intentionally unfinished lesson placeholders:
- `handler_upload_video.go`: `handlerUploadVideo` is a literally empty function body — registered at `POST /api/video_upload/{videoID}` but does nothing (router still returns `200` with an empty response).
- `handler_upload_thumbnail.go`: `handlerUploadThumbnail` is implemented — parses the multipart form (10MB max memory), reads the `thumbnail` form file, checks the authenticated user owns the video (`401` if not), base64-encodes the bytes (`encoding/base64`) into a `data:<media-type>;base64,<data>` URL, sets `thumbnail_url` via `cfg.db.UpdateVideo`, and responds with the updated `database.Video`. No separate GET route is needed since the data URL is served directly by the browser; expected to move to S3 later in the course.
- `s3Bucket`/`s3Region`/`s3CfDistribution` are read from env and stored on `apiConfig` but nothing in the code currently calls AWS S3/CloudFront — wiring is expected to land as the course progresses.

## Building and Running
```bash
go mod download
cp .env.example .env       # then fill in real values per course instructions
./samplesdownload.sh        # downloads sample images/videos into samples/
go run .
```
- Requires `ffmpeg`/`ffprobe` on `PATH` (not yet used by any handler, but required per `README.md`/course).
- `sqlite3` CLI is optional, only for manually inspecting `tubely.db`.
- AWS CLI + `~/.aws/credentials` (via `aws configure`) needed once S3 features are implemented; not required to run the server today.
- On startup: creates `tubely.db` (gitignored) and `ASSETS_ROOT` (`assets/`, gitignored) if missing (`cfg.ensureAssetsDir()`, `assets.go`).
- Server listens on `PORT` (`.env.example` default `8091`); logs `http://localhost:<port>/app/` on startup.

## Required Env Vars (`.env.example`)
`DB_PATH`, `JWT_SECRET`, `PLATFORM`, `FILEPATH_ROOT`, `ASSETS_ROOT`, `S3_BUCKET`, `S3_REGION`, `S3_CF_DISTRO`, `PORT` — all fatal-on-missing in `main()`.

## Project Structure
```
tubely/
├── app/                       # Static frontend served at /app/ (index.html, app.js, styles.css)
├── assets/                    # Uploaded asset storage, served at /assets/ (gitignored, created at startup)
├── internal/
│   ├── auth/                  # Password hashing, JWT, bearer/API-key header parsing
│   └── database/              # SQLite client: raw SQL queries, auto-migrate DDL, Reset()
├── main.go                    # apiConfig, route registration, server startup
├── assets.go                  # ensureAssetsDir
├── cache.go                   # cacheMiddleware (Cache-Control on /assets/)
├── json.go                    # respondWithJSON / respondWithError
├── reset.go                   # POST /admin/reset handler
├── handler_users.go           # POST /api/users
├── handler_login.go           # POST /api/login
├── handler_refresh.go         # POST /api/refresh, POST /api/revoke
├── handler_video_meta.go      # /api/videos CRUD
├── handler_upload_thumbnail.go # POST /api/thumbnail_upload/{videoID}
├── handler_upload_video.go    # POST /api/video_upload/{videoID} (stub, see Incomplete/Stub Code)
├── samplesdownload.sh         # Downloads sample images/videos into samples/ (gitignored)
├── .env.example               # Template for .env (gitignored)
├── tubely.db                  # SQLite DB file (gitignored, created at startup)
└── PROJECT.md                  # This documentation file
```

---
**Note**: Update this file whenever significant changes are made to the server implementation. Keep it under ~200 lines: fold updates into the existing section they belong to rather than appending a new dated section, and trim content that's directly derivable from the code.
