import type { SystemStats } from "@/lib/api";

function StatCard({
  label,
  value,
  accent,
}: {
  label: string;
  value: number | string;
  accent?: string;
}) {
  return (
    <div className="rounded-xl border border-surface-border bg-surface-raised p-4">
      <p className="text-xs font-medium uppercase tracking-wider text-zinc-500">{label}</p>
      <p className={`mt-2 text-2xl font-bold ${accent ?? "text-white"}`}>{value}</p>
    </div>
  );
}

export function SystemStatsPanel({ stats }: { stats: SystemStats }) {
  return (
    <div className="space-y-4">
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard label="Total videos" value={stats.total_videos} />
        <StatCard label="Ready" value={stats.ready_count} accent="text-emerald-400" />
        <StatCard label="Processing" value={stats.processing_count} accent="text-amber-400" />
        <StatCard label="Queued" value={stats.queued_count} accent="text-violet-400" />
      </div>

      <div className="rounded-xl border border-surface-border bg-surface-raised p-4">
        <p className="text-xs font-medium uppercase tracking-wider text-zinc-500">
          Active workers
        </p>
        {stats.active_workers.length > 0 ? (
          <div className="mt-3 flex flex-wrap gap-2">
            {stats.active_workers.map((worker) => (
              <span
                key={worker}
                className="rounded-full border border-emerald-500/30 bg-emerald-500/10 px-3 py-1 text-xs font-medium text-emerald-300"
              >
                {worker}
              </span>
            ))}
          </div>
        ) : (
          <p className="mt-2 text-sm text-zinc-500">No workers processing right now</p>
        )}
      </div>
    </div>
  );
}

export function SystemStatsSkeleton() {
  return (
    <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
      {Array.from({ length: 4 }).map((_, i) => (
        <div
          key={i}
          className="h-20 animate-pulse rounded-xl border border-surface-border bg-surface-raised"
        />
      ))}
    </div>
  );
}
