# whisper_test.go

Hermetic tests for Whisper transcription helpers.

## Related files
- `whisper.go` — production transcription code under test.

## Important details
- Tests `transcribe` oversized-file guard, `postWhisper` success/error/invalid-JSON paths, and `transcribeGroq` / `transcribeOpenAI` key checks.
- Uses `httptest` servers for the HTTP path and temporarily maps `http.DefaultTransport` to route `api.groq.com` / `api.openai.com` to those servers.

## Caveats
- `http.DefaultTransport` is replaced per-test and restored via `defer`; tests must not call `t.Parallel()`.
- The fake audio files are not valid WAV data; they only need to exist and be readable for the multipart upload.
