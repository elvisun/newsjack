CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS install_events (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  event_type text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  ip_hash text,
  country text,
  region text,
  user_agent text,
  client_kind text,
  referer text,
  accept_language text,
  query_params jsonb NOT NULL DEFAULT '{}'::jsonb,
  metadata jsonb NOT NULL DEFAULT '{}'::jsonb
);

ALTER TABLE install_events
  ADD COLUMN IF NOT EXISTS client_kind text;

DO $$
DECLARE
  constraint_name text;
BEGIN
  FOR constraint_name IN
    SELECT conname
    FROM pg_constraint
    WHERE conrelid = 'install_events'::regclass
      AND contype = 'c'
      AND pg_get_constraintdef(oid) LIKE '%event_type%'
  LOOP
    EXECUTE format('ALTER TABLE install_events DROP CONSTRAINT %I', constraint_name);
  END LOOP;
END $$;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_name = 'install_events'
      AND column_name = 'installer_kind'
  ) THEN
    EXECUTE
      'UPDATE install_events
       SET client_kind = COALESCE(client_kind, installer_kind)
       WHERE event_type <> ''site_visit''';
  END IF;
END $$;

UPDATE install_events
SET event_type = 'install_request'
WHERE event_type <> 'site_visit'
  AND event_type <> 'install_request';

ALTER TABLE install_events
  ADD CONSTRAINT install_events_event_type_check
  CHECK (event_type IN ('site_visit', 'install_request'));

CREATE INDEX IF NOT EXISTS install_events_event_type_created_at_idx
  ON install_events (event_type, created_at DESC);

CREATE INDEX IF NOT EXISTS install_events_client_kind_created_at_idx
  ON install_events (client_kind, created_at DESC);
