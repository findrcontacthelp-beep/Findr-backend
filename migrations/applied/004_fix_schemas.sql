-- ============================================================
-- 004_fix_schemas.sql
-- Fix schema discrepancies and add missing tables
-- ============================================================

-- 1. Rename projects to posts
ALTER TABLE IF EXISTS public.projects RENAME TO posts;

-- 2. Rename columns in users table
ALTER TABLE IF EXISTS public.users RENAME COLUMN firebase_uid TO user_uuid;

-- 3. Rename columns in posts (formerly projects) table
ALTER TABLE IF EXISTS public.posts RENAME COLUMN author_uid TO author_uuid;

-- 4. Create campus_buzz table
CREATE TABLE IF NOT EXISTS public.campus_buzz (
    id          uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    text        text NOT NULL,
    created_at  timestamptz DEFAULT now()
);

-- 5. Create index for campus_buzz
CREATE INDEX IF NOT EXISTS idx_campus_buzz_created ON public.campus_buzz(created_at DESC);

-- 6. Add some sample data for testing
INSERT INTO public.campus_buzz (text) VALUES ('Welcome to Findr! Check out the new features.');
INSERT INTO public.campus_buzz (text) VALUES ('Hackathon registration is now open!');
