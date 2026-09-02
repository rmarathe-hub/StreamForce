import type { Video, VideoStatus } from "@/lib/api";

const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8081";
const WS_BASE = API_BASE.replace(/^http/, "ws");

export interface VideoEvent {
  video_id: string;
  status: VideoStatus;
  progress_percent?: number | null;
  claimed_by?: string | null;
  hls_path?: string | null;
  thumbnail_path?: string | null;
  codec?: string | null;
  duration?: number | null;
  width?: number | null;
  height?: number | null;
  error_message?: string | null;
}

export function subscribeVideoUpdates(
  videoId: string,
  onUpdate: (event: VideoEvent) => void,
  onConnectionChange?: (connected: boolean) => void,
): () => void {
  let active = true;
  let socket: WebSocket | null = null;
  let reconnectTimer: ReturnType<typeof setTimeout> | undefined;

  function connect() {
    if (!active) return;

    socket = new WebSocket(`${WS_BASE}/api/ws/videos/${videoId}`);

    socket.onopen = () => {
      onConnectionChange?.(true);
    };

    socket.onmessage = (message) => {
      try {
        const event = JSON.parse(message.data as string) as VideoEvent;
        onUpdate(event);
      } catch {
        // ignore malformed messages
      }
    };

    socket.onclose = () => {
      onConnectionChange?.(false);
      if (!active) return;
      reconnectTimer = setTimeout(connect, 2000);
    };

    socket.onerror = () => {
      socket?.close();
    };
  }

  connect();

  return () => {
    active = false;
    if (reconnectTimer) clearTimeout(reconnectTimer);
    socket?.close();
  };
}

export function mergeVideoEvent(current: Video, event: VideoEvent): Video {
  return {
    ...current,
    status: event.status,
    progress_percent: event.progress_percent ?? current.progress_percent,
    claimed_by: event.claimed_by ?? current.claimed_by,
    hls_path: event.hls_path ?? current.hls_path,
    thumbnail_path: event.thumbnail_path ?? current.thumbnail_path,
    codec: event.codec ?? current.codec,
    duration: event.duration ?? current.duration,
    width: event.width ?? current.width,
    height: event.height ?? current.height,
    error_message: event.error_message ?? current.error_message,
  };
}

export function videoFromEvent(event: VideoEvent, current: Video): Video {
  return mergeVideoEvent(current, event);
}
