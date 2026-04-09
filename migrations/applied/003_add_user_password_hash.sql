alter table public.users
add column if not exists password_hash text;

create index if not exists idx_users_email_lower
on public.users ((lower(email)));
