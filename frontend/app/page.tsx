import Link from "next/link";

export default function HomePage() {
  return (
    <section className="space-y-8">
      <div className="space-y-3">
        <p className="text-sm font-medium uppercase tracking-widest text-accent">
          StreamForge
        </p>
        <h1 className="text-4xl font-bold tracking-tight text-white">
          Upload. Transcode. Stream.
        </h1>
        <p className="max-w-2xl text-lg text-zinc-400">
          Upload an MP4, let FFmpeg generate adaptive HLS renditions, and play
          the stream directly in your browser.
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
    </section>
  );
}
