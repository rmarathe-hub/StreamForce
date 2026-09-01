ALTER TABLE videos
    ADD COLUMN IF NOT EXISTS claimed_by TEXT,
    ADD COLUMN IF NOT EXISTS claimed_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_videos_claimed_at ON videos (claimed_at)
    WHERE status = 'PROCESSING';
