# whisper.go

Audio transcription via Groq or OpenAI Whisper.

## Related files
- `video.go` — extracts audio from videos and calls `transcribe`.
- `engines_test.go` — tests unsupported backend/file-not-found cases.

## Important details
- Supports `groq` (whisper-large-v3) and `openai` (whisper-1) backends.
- Rejects files larger than 25 MB.
- Uses multipart/form-data upload.

## Caveats
- Requires `GROQ_API_KEY` or `OPENAI_API_KEY`.
- 120-second HTTP timeout.
- Audio file path is caller-controlled (annotated with `#nosec`).
