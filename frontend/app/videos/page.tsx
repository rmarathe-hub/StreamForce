"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { listVideos, type Video } from "@/lib/api";
import { VideoCard, VideoCardSkeleton } from "@/components/video-card";

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
            {loading ? "Loading library..." : `${videos.length} video${videos.length === 1 ? "" : "s"} in your library`}
          </p>
        </div>
        <Link
          href="/upload"
          className="rounded-lg bg-accent px-4 py-2 text-sm font-medium text-white transition hover:bg-accent-hover"
        >
          Upload
        </Link>
      </div>

      {error && (
        <p className="rounded-lg border border-rose-500/30 bg-rose-500/10 px-4 py-3 text-sm text-rose-300">
          {error}
        </p>
      )}

      {loading && (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {Array.from({ length: 6 }).map((_, i) => (
            <VideoCardSkeleton key={i} />
          ))}
        </div>
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
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {videos.map((video) => (
            <VideoCard key={video.id} video={video} />
          ))}
        </div>
      )}
    </section>
  );
}
