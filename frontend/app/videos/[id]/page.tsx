"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useParams } from "next/navigation";
import { HLSPlayer } from "@/components/hls-player";
import {
  availableResolutions,
  getVideo,
  playbackUrl,
  type Video,
} from "@/lib/api";
import { StatusBadge, formatDate } from "@/components/status-badge";

function DetailRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-start justify-between gap-4 border-b border-surface-border py-3 last:border-b-0">
      <dt className="text-sm text-zinc-400">{label}</dt>
      <dd className="text-right text-sm text-white">{value}</dd>
    </div>
  );
}

function statusMessage(status: Video["status"]): string {
  switch (status) {
    case "UPLOADED":
      return "Upload complete. Waiting to be queued.";
    case "QUEUED":
      return "Queued in Kafka. Waiting for a worker to consume the job.";
    case "PROCESSING":
      return "A background worker is generating adaptive HLS renditions.";
    case "READY":
      return "Adaptive HLS stream is ready to play.";
    case "FAILED":
      return "Processing failed. Check the error details below.";
    default:
      return "";
  }
}

export default function VideoDetailPage() {
  const params = useParams<{ id: string }>();
  const [video, setVideo] = useState<Video | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let active = true;
    let interval: ReturnType<typeof setInterval> | undefined;

    async function load() {
      try {
        const data = await getVideo(params.id);
        if (!active) return;
        setVideo(data);
        setError(null);

        if (data.status === "UPLOADED" || data.status === "QUEUED" || data.status === "PROCESSING") {
          interval = setInterval(async () => {
            try {
              const updated = await getVideo(params.id);
              if (!active) return;
              setVideo(updated);
              if (updated.status === "READY" || updated.status === "FAILED") {
                clearInterval(interval);
              }
            } catch {
              // keep polling on transient errors
            }
          }, 2000);
        }
      } catch (err) {
        if (active) {
          setError(err instanceof Error ? err.message : "Failed to load video");
        }
      } finally {
        if (active) setLoading(false);
      }
    }

    if (params.id) load();

    return () => {
      active = false;
      if (interval) clearInterval(interval);
    };
  }, [params.id]);

  if (loading) {
    return <p className="text-zinc-400">Loading video...</p>;
  }

  if (error || !video) {
    return (
      <section className="space-y-4">
        <p className="rounded-lg border border-rose-500/30 bg-rose-500/10 px-4 py-3 text-sm text-rose-300">
          {error ?? "Video not found"}
        </p>
        <Link href="/videos" className="text-sm text-accent hover:underline">
          Back to videos
        </Link>
      </section>
    );
  }

  const streamUrl = playbackUrl(video);
  const resolutions = availableResolutions(video);

  return (
    <section className="space-y-6">
      <div className="flex items-start justify-between gap-4">
        <div>
          <Link href="/videos" className="text-sm text-zinc-400 transition hover:text-white">
            ← Back to videos
          </Link>
          <h1 className="mt-2 text-3xl font-bold text-white">{video.filename}</h1>
          <div className="mt-3">
            <StatusBadge status={video.status} />
          </div>
        </div>
      </div>

      <div className="grid gap-6 lg:grid-cols-[2fr_1fr]">
        <div className="space-y-4">
          {streamUrl ? (
            <div className="overflow-hidden rounded-xl border border-surface-border bg-black">
              <HLSPlayer src={streamUrl} />
            </div>
          ) : (
            <div className="flex aspect-video items-center justify-center rounded-xl border border-dashed border-surface-border bg-surface-raised">
              <p className="text-sm text-zinc-400">
                {video.status === "QUEUED" || video.status === "PROCESSING" || video.status === "UPLOADED"
                  ? "Preparing HLS stream..."
                  : "Playback unavailable"}
              </p>
            </div>
          )}

          <div className="rounded-xl border border-surface-border bg-surface-raised p-6">
            <h2 className="text-lg font-semibold text-white">Processing status</h2>
            <p className="mt-2 text-sm text-zinc-400">{statusMessage(video.status)}</p>
            {(video.status === "UPLOADED" || video.status === "QUEUED" || video.status === "PROCESSING") && (
              <p className="mt-4 animate-pulse text-sm text-amber-300">
                Refreshing automatically every 2 seconds...
              </p>
            )}
          </div>
        </div>

        <dl className="rounded-xl border border-surface-border bg-surface-raised p-6">
          <h2 className="mb-2 text-lg font-semibold text-white">Details</h2>
          <DetailRow label="Video ID" value={video.id} />
          <DetailRow label="Filename" value={video.filename} />
          <DetailRow label="Status" value={video.status} />
          <DetailRow label="Codec" value={video.codec ?? "—"} />
          <DetailRow
            label="Source resolution"
            value={
              video.width && video.height ? `${video.width}×${video.height}` : "—"
            }
          />
          <DetailRow
            label="HLS renditions"
            value={resolutions.length > 0 ? resolutions.join(", ") : "—"}
          />
          <DetailRow
            label="Duration"
            value={video.duration ? `${video.duration.toFixed(1)}s` : "—"}
          />
          <DetailRow label="Created" value={formatDate(video.created_at)} />
          <DetailRow label="Updated" value={formatDate(video.updated_at)} />
          {video.error_message && (
            <DetailRow label="Error" value={video.error_message} />
          )}
        </dl>
      </div>
    </section>
  );
}
