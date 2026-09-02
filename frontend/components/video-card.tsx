import Link from "next/link";
import { VideoThumbnail } from "@/components/video-thumbnail";
import { formatDate } from "@/components/status-badge";
import type { Video } from "@/lib/api";

export function VideoCard({ video }: { video: Video }) {
  return (
    <Link
      href={`/videos/${video.id}`}
      className="group block overflow-hidden rounded-xl border border-surface-border bg-surface-raised transition hover:border-zinc-600 hover:bg-surface-raised/80"
    >
      <VideoThumbnail video={video} className="rounded-none border-0" />
      <div className="space-y-1 p-4">
        <h3 className="truncate font-medium text-white group-hover:text-accent">
          {video.filename}
        </h3>
        <p className="text-xs text-zinc-500">{formatDate(video.created_at)}</p>
        {video.claimed_by && video.status === "PROCESSING" && (
          <p className="text-xs text-zinc-400">Worker: {video.claimed_by}</p>
        )}
      </div>
    </Link>
  );
}

export function VideoCardSkeleton() {
  return (
    <div className="overflow-hidden rounded-xl border border-surface-border bg-surface-raised">
      <div className="aspect-video animate-pulse bg-zinc-800" />
      <div className="space-y-2 p-4">
        <div className="h-4 w-3/4 animate-pulse rounded bg-zinc-800" />
        <div className="h-3 w-1/2 animate-pulse rounded bg-zinc-800" />
      </div>
    </div>
  );
}
