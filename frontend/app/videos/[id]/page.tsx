"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useParams } from "next/navigation";
import { getVideo, type Video } from "@/lib/api";
import { StatusBadge, formatDate } from "@/components/status-badge";

function DetailRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-start justify-between gap-4 border-b border-surface-border py-3 last:border-b-0">
      <dt className="text-sm text-zinc-400">{label}</dt>
      <dd className="text-right text-sm text-white">{value}</dd>
    </div>
  );
}

export default function VideoDetailPage() {
  const params = useParams<{ id: string }>();
  const [video, setVideo] = useState<Video | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let active = true;

    async function load() {
      try {
        const data = await getVideo(params.id);
        if (active) setVideo(data);
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
        <div className="rounded-xl border border-surface-border bg-surface-raised p-6">
          <h2 className="text-lg font-semibold text-white">Processing status</h2>
          <p className="mt-2 text-sm text-zinc-400">
            Phase 1 stores the upload and marks it as <span className="text-zinc-200">UPLOADED</span>.
            FFmpeg transcoding and HLS playback arrive in Phase 2.
          </p>
          {video.status === "UPLOADED" && (
            <p className="mt-4 rounded-lg border border-sky-500/20 bg-sky-500/10 px-4 py-3 text-sm text-sky-200">
              File saved successfully. Waiting for processing pipeline.
            </p>
          )}
        </div>

        <dl className="rounded-xl border border-surface-border bg-surface-raised p-6">
          <h2 className="mb-2 text-lg font-semibold text-white">Details</h2>
          <DetailRow label="Video ID" value={video.id} />
          <DetailRow label="Filename" value={video.filename} />
          <DetailRow label="Status" value={video.status} />
          <DetailRow label="Storage path" value={video.source_path} />
          <DetailRow label="Duration" value={video.duration ? `${video.duration}s` : "—"} />
          <DetailRow
            label="Resolution"
            value={
              video.width && video.height ? `${video.width}×${video.height}` : "—"
            }
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
