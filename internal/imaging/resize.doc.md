# resize.go

Resizes video frames using pure Go image processing.

## Related files
- `engines/video.go` — extracts frames and may call resize.
- `briefing/video_html.go` — embeds frame images.

## Important details
- Preserves aspect ratio using Catmull-Rom scaling.
- Supports JPEG and PNG output.
- Default `MaxWidth` is 1024 and JPEG quality is 75.

## Caveats
- `#nosec` annotations for file paths because the caller controls input/output paths.
- Does not handle exotic color spaces or animated images.
