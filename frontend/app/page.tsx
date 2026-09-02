"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { getStats, listVideos, type SystemStats, type Video } from "@/lib/api";
import { SystemStatsPanel, SystemStatsSkeleton } from "@/components/system-stats";
import { VideoCard, VideoCardSkeleton } from "@/components/video-card";

export default function HomePage() {
  const [stats, setStats] = useState<SystemStats | null>(null);
  const [recent, setRecent] = useState<Video[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let active = true;

    async function load() {
      try {
        const [statsData, videos] = await Promise.all([getStats(), listVideos()]);
        if (!active) return;
        setStats(statsData);
        setRecent(videos.slice(0, 3));
      } catch {
        // stats panel is optional on failure
      } finally {
        if (active) setLoading(false);
      }
    }

    load();
    const interval = setInterval(async () => {
      try {
        const statsData = await getStats();
        if (active) setStats(statsData);
      } catch {
        // ignore transient errors
      }
    }, 5000);

    return () => {
      active = false;
      clearInterval(interval);
    };
  }, []);

  return (
    <section className="space-y-10">
      <div className="space-y-3">
        <p className="text-sm font-medium uppercase tracking-widest text-accent">
          StreamForge
        </p>
        <h1 className="text-4xl font-bold tracking-tight text-white">
          Upload. Transcode. Stream.
        </h1>
        <p className="max-w-2xl text-lg text-zinc-400">
          A distributed video platform — Kafka job queue, FFmpeg HLS transcoding,
          Redis progress, and WebSocket live updates.
        </p>
      </div>

      <div className="flex flex-wrap gap-4">
        <Link
          href="/upload"
          className="rounded-lg bg-accent px-5 py-2.5 text-sm font-medium text-white transition hover:bg-accent-hover"
        >
          Upload a video
        </Link>
        <Link
          href="/videos"
          className="rounded-lg border border-surface-border px-5 py-2.5 text-sm font-medium text-zinc-200 transition hover:border-zinc-500 hover:text-white"
        >
          View library
        </Link>
      </div>

      <div className="space-y-4">
        <h2 className="text-lg font-semibold text-white">System overview</h2>
        {loading && !stats ? <SystemStatsSkeleton /> : stats ? <SystemStatsPanel stats={stats} /> : null}
      </div>

      <div className="rounded-xl border border-surface-border bg-surface-raised p-6">
        <h2 className="text-lg font-semibold text-white">Architecture</h2>
        <pre className="mt-4 overflow-x-auto text-sm leading-relaxed text-zinc-400">
{`Upload → API → Postgres (QUEUED) → Kafka → Worker → FFmpeg → HLS
                    ↓                              ↓
                 WebSocket  ←── Redis pub/sub ←── progress`}
        </pre>
      </div>

      {!loading && recent.length > 0 && (
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold text-white">Recent videos</h2>
            <Link href="/videos" className="text-sm text-accent hover:underline">
              View all
            </Link>
          </div>
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {recent.map((video) => (
              <VideoCard key={video.id} video={video} />
            ))}
          </div>
        </div>
      )}

      {loading && (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {Array.from({ length: 3 }).map((_, i) => (
            <VideoCardSkeleton key={i} />
          ))}
        </div>
      )}
    </section>
  );
}
