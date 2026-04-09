-- ============================================================
-- Add interests column to users table for registration flow
-- ============================================================

ALTER TABLE public.users ADD COLUMN IF NOT EXISTS interests text[] DEFAULT '{}';
