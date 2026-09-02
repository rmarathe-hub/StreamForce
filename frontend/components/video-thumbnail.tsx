import Image from "next/image";
import { thumbnailUrl, type Video } from "@/lib/api";
import { StatusBadge } from "@/components/status-badge";

function Placeholder({ status }: { status: Video["status"] }) {
  const labels: Record<Video["status"], string> = {
    UPLOADED: "Uploaded",
    QUEUED: "Queued",
    PROCESSING: "Processing",
    READY: "Ready",
    FAILED: "Failed",
  };

  return (
    <div className="flex h-full w-full flex-col items-center justify-center gap-2 bg-gradient-to-br from-zinc-800 to-zinc-900 text-zinc-500">
      <svg className="h-10 w-10 opacity-40" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth={1.5}
          d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z"
        />
      </svg>
      <span className="text-xs font-medium">{labels[status]}</span>
    </div>
  );
}

export function VideoThumbnail({
  video,
  className = "",
}: {
  video: Video;
  className?: string;
}) {
  const src = thumbnailUrl(video);

  return (
    <div
      className={`relative aspect-video overflow-hidden rounded-lg border border-surface-border bg-zinc-900 ${className}`}
    >
      {src ? (
        <Image
          src={src}
          alt={video.filename}
          fill
          className="object-cover"
          sizes="(max-width: 768px) 100vw, 320px"
          unoptimized
        />
      ) : (
        <Placeholder status={video.status} />
      )}
      <div className="absolute left-2 top-2">
        <StatusBadge status={video.status} />
      </div>
      {video.duration != null && (
        <span className="absolute bottom-2 right-2 rounded bg-black/70 px-1.5 py-0.5 text-xs text-white">
          {video.duration.toFixed(1)}s
        </span>
      )}
    </div>
  );
}
