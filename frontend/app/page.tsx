import Link from "next/link";

export default function HomePage() {
  return (
    <section className="space-y-8">
      <div className="space-y-3">
        <p className="text-sm font-medium uppercase tracking-widest text-accent">
          Phase 1
        </p>
        <h1 className="text-4xl font-bold tracking-tight text-white">
          Upload. Store. Track.
        </h1>
        <p className="max-w-2xl text-lg text-zinc-400">
          StreamForge accepts video uploads, persists metadata in PostgreSQL, and
          stores files locally. Processing and HLS come in later phases.
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
