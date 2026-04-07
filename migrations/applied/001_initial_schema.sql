-- ============================================================
-- Findr: Firebase Firestore → Supabase Migration Schema
-- Run this ENTIRE file in Supabase SQL Editor BEFORE importing
-- ============================================================

create extension if not exists "uuid-ossp";

-- ============================================================
-- 1. USERS
-- ============================================================
create table public.users (
  id              uuid primary key default uuid_generate_v4(),
  firebase_uid    text unique not null,
  name            text not null default '',
  email           text,
  profile_picture text default '',
  fcm_token       text,
  -- profileHeader fields (flattened)
  headline        text default '',
  role_title      text default '',
  is_student      boolean default true,
  college_name    text default '',
  company_name    text default '',
  experience      text default '',
  ctc             text default '',
  location        text default '',
  lat             real default 0,
  lng             real default 0,
  profile_image_url text default '',
  banner_image_url  text default '',
  skills          text[] default '{}',
  social_links    jsonb default '{}',
  -- collegeInfo fields
  college_year    text default '',
  college_stream  text default '',
  college_grade   text default '',
  college_start   text default '',
  college_end     text default '',
  college_institute text default '',
  -- experienceInfo fields
  exp_title       text default '',
  exp_company     text default '',
  exp_type        text default '',
  exp_location    text default '',
  exp_description text default '',
  exp_ctc         text default '',
  exp_start       text default '',
  exp_end         text default '',
  exp_currently_working boolean default false,
  -- about
  about_text      text default '',
  -- activities stored as JSON array
  activities      jsonb default '[]',
  -- other
  user_list       text[] default '{}',
  stability       int default 0,
  created_at      timestamptz default now(),
  updated_at      timestamptz default now()
);

-- ============================================================
-- 2. USER PROFILE VIEWERS
-- ============================================================
create table public.user_viewers (
  id          uuid primary key default uuid_generate_v4(),
  user_id     uuid references public.users(id) on delete cascade,
  viewer_uid  text not null,
  viewed_at   timestamptz default now()
);

-- ============================================================
-- 3. USER RATINGS
-- ============================================================
create table public.user_ratings (
  id        uuid primary key default uuid_generate_v4(),
  user_id   uuid references public.users(id) on delete cascade,
  rater_uid text not null,
  rating    int not null default 0,
  unique(user_id, rater_uid)
);

-- ============================================================
-- 4. USER VERIFICATIONS
-- ============================================================
create table public.user_verifications (
  id          uuid primary key default uuid_generate_v4(),
  user_id     uuid references public.users(id) on delete cascade,
  verifier_uid text not null,
  verified_at timestamptz default now(),
  unique(user_id, verifier_uid)
);

-- ============================================================
-- 5. USER EXTRA ACTIVITIES
-- ============================================================
create table public.user_extra_activities (
  id          uuid primary key default uuid_generate_v4(),
  firebase_id text,
  user_id     uuid references public.users(id) on delete cascade,
  name        text not null default '',
  domain      text default '',
  link        text default '',
  description text default '',
  media       text default ''
);

-- ============================================================
-- 6. USER NOTIFICATIONS
-- ============================================================
create table public.user_notifications (
  id              uuid primary key default uuid_generate_v4(),
  user_id         uuid references public.users(id) on delete cascade,
  firebase_key    text,
  last_notified_at timestamptz
);

-- ============================================================
-- 7. TOPICS
-- ============================================================
create table public.topics (
  id          uuid primary key default uuid_generate_v4(),
  firebase_id text unique,
  topic       text not null,
  enabled     boolean default true
);

-- ============================================================
-- 8. METADATA (app config key-value)
-- ============================================================
create table public.metadata (
  key        text primary key,
  value      jsonb not null default '{}',
  updated_at timestamptz default now()
);

-- ============================================================
-- 9. AVAILABLE ROLES (global role catalog)
-- ============================================================
create table public.available_roles (
  id          uuid primary key default uuid_generate_v4(),
  firebase_id text unique,
  name        text not null
);

-- ============================================================
-- 10. ROLE REQUESTS (user-submitted role suggestions)
-- ============================================================
create table public.role_requests (
  id          uuid primary key default uuid_generate_v4(),
  firebase_id text unique,
  name        text not null
);

-- ============================================================
-- 11. PROJECTS / POSTS / EVENTS (unified table)
-- ============================================================
create table public.projects (
  id              uuid primary key default uuid_generate_v4(),
  firebase_id     text unique,
  author_uid      text,
  author_id       uuid references public.users(id) on delete set null,
  author_name     text default '',
  type            text not null default 'Post',
  title           text not null,
  title_lower     text default '',
  description     text default '',
  tags            text[] default '{}',
  image_urls      text[] default '{}',
  file_urls       text[] default '{}',
  video_url       text default '',
  links           text[] default '{}',
  roles_needed    text[] default '{}',
  -- project roles (embedded as JSONB array)
  project_roles   jsonb default '[]',
  -- event details (null for non-events)
  event_details   jsonb,
  -- enrolled/requested persons (maps of uid→name)
  enrolled_persons  jsonb default '{}',
  requested_persons jsonb default '{}',
  -- stats
  likes           text[] default '{}',
  likes_count     int default 0,
  comments_count  int default 0,
  views_count     int default 0,
  post_views_count int default 0,
  created_at      timestamptz default now(),
  updated_at      timestamptz default now()
);

-- ============================================================
-- 12. PROJECT COMMENTS (from subcollection + inline)
-- ============================================================
create table public.project_comments (
  id                uuid primary key default uuid_generate_v4(),
  firebase_id       text,
  project_id        uuid references public.projects(id) on delete cascade,
  sender_uid        text,
  sender_id         uuid references public.users(id) on delete set null,
  sender_name       text default '',
  sender_image_url  text default '',
  post_id           text,
  text              text not null default '',
  nesting_level     int default 0,
  parent_comment_id text,
  root_comment_id   text,
  is_top_level      boolean default true,
  reply_count       int default 0,
  likes             text[] default '{}',
  likes_count       int default 0,
  created_at        timestamptz default now()
);

-- ============================================================
-- 13. PROJECT POST VIEWS
-- ============================================================
create table public.project_post_views (
  id         uuid primary key default uuid_generate_v4(),
  project_id uuid references public.projects(id) on delete cascade,
  viewer_uid text not null,
  viewed_at  timestamptz default now(),
  unique(project_id, viewer_uid)
);

-- ============================================================
-- 14. CHATS
-- ============================================================
create table public.chats (
  id                uuid primary key default uuid_generate_v4(),
  firebase_id       text unique,
  last_message      text,
  last_message_at   timestamptz,
  unread_messages   jsonb default '{}',
  created_at        timestamptz default now()
);

-- ============================================================
-- 15. CHAT PARTICIPANTS
-- ============================================================
create table public.chat_participants (
  chat_id    uuid references public.chats(id) on delete cascade,
  user_uid   text not null,
  user_id    uuid references public.users(id) on delete cascade,
  primary key (chat_id, user_uid)
);

-- ============================================================
-- 16. CHAT MESSAGES
-- ============================================================
create table public.chat_messages (
  id          uuid primary key default uuid_generate_v4(),
  firebase_id text,
  message_id  text,
  chat_id     uuid references public.chats(id) on delete cascade,
  sender_uid  text,
  sender_id   uuid references public.users(id) on delete set null,
  receiver_uid text,
  message     text not null default '',
  status      text default 'sent',
  reply_to    jsonb,
  media       jsonb,
  created_at  timestamptz default now()
);

-- ============================================================
-- 17. ENROLLMENTS (project role applications)
-- ============================================================
create table public.enrollments (
  id              uuid primary key default uuid_generate_v4(),
  firebase_id     text unique,
  user_uid        text,
  user_id         uuid references public.users(id) on delete cascade,
  user_name       text default '',
  user_profile_pic text default '',
  post_id         text,
  project_id      uuid references public.projects(id) on delete set null,
  role_id         text,
  role_name       text default '',
  message         text,
  status          text default 'PENDING',
  pending         boolean default true,
  accepted        boolean default false,
  rejected        boolean default false,
  requested_at    timestamptz default now(),
  responded_at    timestamptz
);

-- ============================================================
-- 18. REGISTRATIONS (event registrations)
-- ============================================================
create table public.registrations (
  id              uuid primary key default uuid_generate_v4(),
  firebase_id     text unique,
  user_uid        text,
  user_id         uuid references public.users(id) on delete cascade,
  user_name       text default '',
  user_profile_pic text default '',
  post_id         text,
  project_id      uuid references public.projects(id) on delete set null,
  status          text default 'REGISTERED',
  registered      boolean default true,
  attended        boolean default false,
  attendance_confirmed boolean default false,
  cancelled       boolean default false,
  registered_at   timestamptz default now()
);

-- ============================================================
-- 19. PLACEMENT REVIEWS
-- ============================================================
create table public.placement_reviews (
  id                    uuid primary key default uuid_generate_v4(),
  firebase_id           text unique,
  submitted_by_uid      text,
  submitted_by_id       uuid references public.users(id) on delete set null,
  submitted_by_name     text default '',
  submitted_at          timestamptz default now(),
  company_name          text not null,
  company_logo          text,
  year                  int,
  month                 text,
  academic_year         text,
  visit_date            timestamptz,
  difficulty            text,
  overall_experience    text,
  package_type          text,
  package_min           real,
  package_max           real,
  package_list          jsonb,
  students_shortlisted  int default 0,
  students_selected     int default 0,
  eligibility_branches  text[] default '{}',
  eligibility_cgpa      real,
  eligibility_max_backlogs int,
  eligibility_other     text,
  tips                  text[] default '{}',
  rounds                jsonb default '[]',
  verification_status   text default 'NOT_VERIFIED',
  verified_at           timestamptz,
  upvotes               int default 0
);

-- ============================================================
-- 20. TEST TABLES (raw JSONB)
-- ============================================================
create table public.test_projects (
  id          uuid primary key default uuid_generate_v4(),
  firebase_id text unique,
  raw_data    jsonb not null default '{}',
  created_at  timestamptz default now()
);

create table public.test_posts (
  id          uuid primary key default uuid_generate_v4(),
  firebase_id text unique,
  raw_data    jsonb not null default '{}',
  created_at  timestamptz default now()
);

-- ============================================================
-- INDEXES
-- ============================================================
create index idx_users_firebase_uid         on public.users(firebase_uid);
create index idx_users_email                on public.users(email);
create index idx_user_viewers_user          on public.user_viewers(user_id);
create index idx_user_ratings_user          on public.user_ratings(user_id);
create index idx_user_verifications_user    on public.user_verifications(user_id);
create index idx_user_extra_activities_user on public.user_extra_activities(user_id);
create index idx_projects_author            on public.projects(author_id);
create index idx_projects_type              on public.projects(type);
create index idx_projects_created           on public.projects(created_at desc);
create index idx_projects_firebase          on public.projects(firebase_id);
create index idx_project_comments_project   on public.project_comments(project_id);
create index idx_project_post_views_project on public.project_post_views(project_id);
create index idx_chats_firebase             on public.chats(firebase_id);
create index idx_chat_messages_chat         on public.chat_messages(chat_id);
create index idx_chat_messages_created      on public.chat_messages(created_at desc);
create index idx_enrollments_user           on public.enrollments(user_uid);
create index idx_enrollments_post           on public.enrollments(post_id);
create index idx_registrations_post         on public.registrations(post_id);
create index idx_placement_reviews_company  on public.placement_reviews(company_name);

-- ============================================================
-- UPDATED_AT TRIGGER
-- ============================================================
create or replace function public.handle_updated_at()
returns trigger as $$
begin
  new.updated_at = now();
  return new;
end;
$$ language plpgsql;

create trigger set_users_updated_at
  before update on public.users
  for each row execute function public.handle_updated_at();

create trigger set_projects_updated_at
  before update on public.projects
  for each row execute function public.handle_updated_at();

-- ============================================================
-- ROW LEVEL SECURITY
-- ============================================================

alter table public.users enable row level security;
create policy "Anyone can read user profiles"
  on public.users for select using (true);
create policy "Users can update own profile"
  on public.users for update using (auth.uid()::text = firebase_uid);

alter table public.projects enable row level security;
create policy "Anyone can read projects"
  on public.projects for select using (true);
create policy "Authenticated users can create projects"
  on public.projects for insert with check (auth.role() = 'authenticated');
create policy "Authors can update own projects"
  on public.projects for update
  using (author_id = (select id from public.users where firebase_uid = auth.uid()::text));

alter table public.project_comments enable row level security;
create policy "Anyone can read comments"
  on public.project_comments for select using (true);
create policy "Authenticated users can add comments"
  on public.project_comments for insert with check (auth.role() = 'authenticated');

alter table public.chats enable row level security;
create policy "Participants can read their chats"
  on public.chats for select
  using (id in (
    select chat_id from public.chat_participants
    where user_id = (select id from public.users where firebase_uid = auth.uid()::text)
  ));

alter table public.chat_messages enable row level security;
create policy "Participants can read chat messages"
  on public.chat_messages for select
  using (chat_id in (
    select chat_id from public.chat_participants
    where user_id = (select id from public.users where firebase_uid = auth.uid()::text)
  ));
create policy "Participants can send messages"
  on public.chat_messages for insert
  with check (chat_id in (
    select chat_id from public.chat_participants
    where user_id = (select id from public.users where firebase_uid = auth.uid()::text)
  ));

alter table public.enrollments enable row level security;
create policy "Users can read own enrollments"
  on public.enrollments for select
  using (user_id = (select id from public.users where firebase_uid = auth.uid()::text));

alter table public.placement_reviews enable row level security;
create policy "Anyone can read placement reviews"
  on public.placement_reviews for select using (true);

-- ============================================================
-- ENABLE REALTIME
-- ============================================================
alter publication supabase_realtime add table public.chats;
alter publication supabase_realtime add table public.chat_messages;
alter publication supabase_realtime add table public.projects;
alter publication supabase_realtime add table public.project_comments;

-- ============================================================
-- DONE!
-- ============================================================
