"use client";

import { useCallback, useState } from "react";
import { useRouter } from "next/navigation";
import { uploadVideo } from "@/lib/api";
import { formatBytes } from "@/components/status-badge";

export default function UploadPage() {
  const router = useRouter();
  const [dragging, setDragging] = useState(false);
  const [file, setFile] = useState<File | null>(null);
  const [progress, setProgress] = useState(0);
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const selectFile = useCallback((selected: File | null) => {
    setError(null);
    setProgress(0);
    setFile(selected);
  }, []);

  const onDrop = useCallback(
    (event: React.DragEvent<HTMLDivElement>) => {
      event.preventDefault();
      setDragging(false);
      const dropped = event.dataTransfer.files?.[0];
      if (dropped) selectFile(dropped);
    },
    [selectFile],
  );

  async function handleUpload() {
    if (!file || uploading) return;

    setUploading(true);
    setError(null);

    try {
      const video = await uploadVideo(file, setProgress);
      router.push(`/videos/${video.id}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Upload failed");
      setUploading(false);
    }
  }

  return (
    <section className="mx-auto max-w-2xl space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-white">Upload video</h1>
        <p className="mt-2 text-zinc-400">
          Drag an MP4 into the browser. FFmpeg will transcode it into adaptive HLS
          renditions after upload.
        </p>
      </div>

      <div
        onDragOver={(event) => {
          event.preventDefault();
          setDragging(true);
        }}
        onDragLeave={() => setDragging(false)}
        onDrop={onDrop}
        className={`rounded-2xl border-2 border-dashed p-10 text-center transition ${
          dragging
            ? "border-accent bg-accent/10"
            : "border-surface-border bg-surface-raised"
        }`}
      >
        <p className="text-lg font-medium text-white">Drop your video here</p>
        <p className="mt-2 text-sm text-zinc-400">or choose a file below</p>
        <label className="mt-6 inline-block cursor-pointer rounded-lg border border-surface-border px-4 py-2 text-sm text-zinc-200 transition hover:border-zinc-500">
          Browse files
          <input
            type="file"
            accept="video/mp4,video/quicktime,video/webm,video/x-matroska,.mp4,.mov,.webm,.mkv"
            className="hidden"
            onChange={(event) => selectFile(event.target.files?.[0] ?? null)}
          />
        </label>
      </div>

      {file && (
        <div className="rounded-xl border border-surface-border bg-surface-raised p-5 space-y-4">
          <div className="flex items-start justify-between gap-4">
            <div>
              <p className="font-medium text-white">{file.name}</p>
              <p className="text-sm text-zinc-400">{formatBytes(file.size)}</p>
            </div>
            <button
              type="button"
              onClick={() => selectFile(null)}
              className="text-sm text-zinc-400 transition hover:text-white"
              disabled={uploading}
            >
              Clear
            </button>
          </div>

          {uploading && (
            <div className="space-y-2">
              <div className="flex justify-between text-sm text-zinc-400">
                <span>Uploading</span>
                <span>{progress}%</span>
              </div>
              <div className="h-2 overflow-hidden rounded-full bg-zinc-800">
                <div
                  className="h-full rounded-full bg-accent transition-all"
                  style={{ width: `${progress}%` }}
                />
              </div>
            </div>
          )}

          <button
            type="button"
            onClick={handleUpload}
            disabled={uploading}
            className="w-full rounded-lg bg-accent px-4 py-2.5 text-sm font-medium text-white transition hover:bg-accent-hover disabled:cursor-not-allowed disabled:opacity-60"
          >
            {uploading ? "Uploading..." : "Upload video"}
          </button>
        </div>
      )}

      {error && (
        <p className="rounded-lg border border-rose-500/30 bg-rose-500/10 px-4 py-3 text-sm text-rose-300">
          {error}
        </p>
      )}
    </section>
  );
}
