export type VideoStatus = "UPLOADED" | "QUEUED" | "PROCESSING" | "READY" | "FAILED";

export interface Video {
  id: string;
  filename: string;
  status: VideoStatus;
  source_path: string;
  hls_path: string | null;
  codec: string | null;
  duration: number | null;
  width: number | null;
  height: number | null;
  claimed_by?: string | null;
  progress_percent?: number | null;
  created_at: string;
  updated_at: string;
  error_message: string | null;
}

const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8081";

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE}${path}`, init);
  if (!response.ok) {
    let message = `Request failed (${response.status})`;
    try {
      const body = await response.json();
      if (body.error) message = body.error;
    } catch {
      // ignore parse errors
    }
    throw new Error(message);
  }
  return response.json() as Promise<T>;
}

export function listVideos(): Promise<Video[]> {
  return request<Video[]>("/api/videos");
}

export function getVideo(id: string): Promise<Video> {
  return request<Video>(`/api/videos/${id}`);
}

export function playbackUrl(video: Video): string | null {
  if (video.status !== "READY" || !video.hls_path) return null;
  return `${API_BASE}/media/${video.hls_path}`;
}

export function availableResolutions(video: Video): string[] {
  if (!video.height) return [];
  if (video.height >= 1080) return ["1080p", "720p", "480p"];
  if (video.height >= 720) return ["720p", "480p"];
  if (video.height >= 480) return ["480p"];
  return [`${video.height}p`];
}

export async function uploadVideo(
  file: File,
  onProgress?: (percent: number) => void,
): Promise<Video> {
  return new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest();
    xhr.open("POST", `${API_BASE}/api/videos`);

    xhr.upload.onprogress = (event) => {
      if (event.lengthComputable && onProgress) {
        onProgress(Math.round((event.loaded / event.total) * 100));
      }
    };

    xhr.onload = () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        try {
          resolve(JSON.parse(xhr.responseText) as Video);
        } catch {
          reject(new Error("Invalid response from server"));
        }
        return;
      }

      let message = `Upload failed (${xhr.status})`;
      try {
        const body = JSON.parse(xhr.responseText);
        if (body.error) message = body.error;
      } catch {
        // ignore
      }
      reject(new Error(message));
    };

    xhr.onerror = () => reject(new Error("Network error during upload"));

    const formData = new FormData();
    formData.append("file", file);
    xhr.send(formData);
  });
}
