ALTER TABLE videos
    ADD COLUMN IF NOT EXISTS thumbnail_path TEXT;
