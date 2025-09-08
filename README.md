# Learn S3 File Storage in Go (Video & Image Handling)

A small, production-leaning Go app that demonstrates **file storage done right** on AWS:
- Uploading & downloading **large videos and images**
- **Private vs. public** delivery models
- **Presigned URLs**, **CloudFront** acceleration, and **least-privilege IAM**
- **Video streaming optimization** (`-movflags +faststart`)
- Pragmatic **error handling, timeouts, and retries**

---

## Highlights — What I Practiced & Applied

- **Amazon S3**  
  - Object storage layout with logical prefixes (e.g., `landscape/…`, `portrait/…`)  
  - Safe uploads with content type validation and server-side cleanup of temp files  
  - Private object access using **presigned GET** URLs (time-bound, least privilege)

- **CloudFront (CDN)**  
  - edge delivery for public assets to reduce latency & offload S3  
  - Cache-control and content-type considerations for fast loads

- **IAM, Permissions & Policies**  
  - **Least-privilege** policy design for Put/Get on a single bucket & prefix set  
  - Separation of duties: the API signs access; clients never see AWS credentials

- **Optimizing for Video Streaming**  
  - `ffprobe` to determine **aspect ratio** (bucket organization & UX)  
  - `ffmpeg -movflags +faststart` to move the `moov` atom to the start of MP4 for **instant playback** on the web

- **Private & Public Cloud Operations**  
  - Private video objects served via **presigned URLs**  
  - Public assets (e.g., thumbnails) optionally via **CloudFront** for global performance

---
Forked From: https://github.com/bootdotdev/learn-file-storage-s3-golang-starter

