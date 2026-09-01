"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { listVideos, type Video } from "@/lib/api";
import { StatusBadge, formatDate } from "@/components/status-badge";

export default function VideosPage() {
  const [videos, setVideos] = useState<Video[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let active = true;

    async function load() {
      try {
        const data = await listVideos();
        if (active) setVideos(data);
      } catch (err) {
        if (active) {
          setError(err instanceof Error ? err.message : "Failed to load videos");
        }
      } finally {
        if (active) setLoading(false);
      }
    }

    load();
    return () => {
      active = false;
    };
  }, []);

  return (
    <section className="space-y-6">
      <div className="flex items-end justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold text-white">Videos</h1>
          <p className="mt-2 text-zinc-400">
            Uploaded videos persist across refreshes.
          </p>
        </div>
        <Link
          href="/upload"
          className="rounded-lg bg-accent px-4 py-2 text-sm font-medium text-white transition hover:bg-accent-hover"
        >
          Upload
        </Link>
      </div>

      {loading && <p className="text-zinc-400">Loading videos...</p>}
      {error && (
        <p className="rounded-lg border border-rose-500/30 bg-rose-500/10 px-4 py-3 text-sm text-rose-300">
          {error}
        </p>
      )}

      {!loading && !error && videos.length === 0 && (
        <div className="rounded-xl border border-dashed border-surface-border bg-surface-raised p-10 text-center">
          <p className="text-zinc-300">No videos yet.</p>
          <Link href="/upload" className="mt-3 inline-block text-sm text-accent hover:underline">
            Upload your first video
          </Link>
        </div>
      )}

      {!loading && videos.length > 0 && (
        <div className="overflow-hidden rounded-xl border border-surface-border">
          <table className="min-w-full divide-y divide-surface-border text-sm">
            <thead className="bg-surface-raised text-left text-zinc-400">
              <tr>
                <th className="px-4 py-3 font-medium">Filename</th>
                <th className="px-4 py-3 font-medium">Status</th>
                <th className="px-4 py-3 font-medium">Duration</th>
                <th className="px-4 py-3 font-medium">Created</th>
                <th className="px-4 py-3 font-medium" />
              </tr>
            </thead>
            <tbody className="divide-y divide-surface-border bg-surface/40">
              {videos.map((video) => (
                <tr key={video.id} className="hover:bg-surface-raised/60">
                  <td className="px-4 py-3 font-medium text-white">{video.filename}</td>
                  <td className="px-4 py-3">
                    <StatusBadge status={video.status} />
                  </td>
                  <td className="px-4 py-3 text-zinc-400">
                    {video.duration ? `${video.duration.toFixed(1)}s` : "—"}
                  </td>
                  <td className="px-4 py-3 text-zinc-400">{formatDate(video.created_at)}</td>
                  <td className="px-4 py-3 text-right">
                    <Link
                      href={`/videos/${video.id}`}
                      className="text-accent transition hover:text-white"
                    >
                      Details
                    </Link>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </section>
  );
}
