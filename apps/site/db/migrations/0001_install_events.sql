CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS install_events (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  install_id uuid,
  event_type text NOT NULL CHECK (
    event_type IN (
      'curl_hit',
      'install_started',
      'install_completed',
      'install_failed'
    )
  ),
  created_at timestamptz NOT NULL DEFAULT now(),
  ip_hash text,
  country text,
  region text,
  user_agent text,
  referer text,
  accept_language text,
  query_params jsonb NOT NULL DEFAULT '{}'::jsonb,
  installer_kind text,
  metadata jsonb NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX IF NOT EXISTS install_events_event_type_created_at_idx
  ON install_events (event_type, created_at DESC);

CREATE INDEX IF NOT EXISTS install_events_install_id_idx
  ON install_events (install_id)
  WHERE install_id IS NOT NULL;
