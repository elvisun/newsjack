#!/usr/bin/env node
import { spawnSync } from "node:child_process";
import { existsSync, readFileSync } from "node:fs";
import { resolve } from "node:path";

loadEnvFile(process.env.NEWSJACK_ENV_FILE ?? ".env");

const databaseEnvNames = [
  "NEWSJACK_DATABASE_URL",
  "DATABASE_URL",
  "POSTGRES_URL",
  "POSTGRES_PRISMA_URL",
];
const databaseUrl = databaseUrlFromEnv();

if (!databaseUrl) {
  console.error(`${databaseEnvNames.join(" or ")} is required.`);
  process.exit(1);
}

const query = `
  SELECT
    event_type,
    COALESCE(client_kind, 'unknown') AS client_kind,
    COALESCE(country, 'unknown') AS country,
    count(*)::int AS events,
    count(DISTINCT ip_hash) FILTER (WHERE ip_hash IS NOT NULL)::int AS daily_unique_ips
  FROM install_events
  WHERE created_at >= now() - interval '24 hours'
  GROUP BY event_type, client_kind, country
  ORDER BY event_type ASC, events DESC, client_kind ASC, country ASC;
`;

const result = spawnSync("psql", [
  databaseUrl,
  "-X",
  "-A",
  "-F",
  "\t",
  "-q",
  "-t",
  "-c",
  query,
], {
  encoding: "utf8",
});

if (result.error) {
  if (result.error.code === "ENOENT") {
    console.error("psql is required to run scripts/funnel-stats.mjs.");
  } else {
    console.error(result.error.message);
  }
  process.exit(1);
}

if (result.status !== 0) {
  process.stderr.write(result.stderr);
  process.exit(result.status ?? 1);
}

const rows = result.stdout
  .trim()
  .split("\n")
  .filter(Boolean)
  .map((line) => {
    const [eventType, clientKind, country, events, dailyUniqueIps] =
      line.split("\t");
    return {
      event_type: eventType,
      client_kind: clientKind,
      country,
      events: Number(events),
      daily_unique_ips: Number(dailyUniqueIps),
    };
  });

if (rows.length === 0) {
  console.log("No site traffic events in the last 24h.");
} else {
  console.log("newsjack.sh traffic, last 24h:");
  console.table(rows);
}

function databaseUrlFromEnv() {
  for (const name of databaseEnvNames) {
    if (process.env[name]) {
      return process.env[name];
    }
  }

  return undefined;
}

function loadEnvFile(path) {
  const envPath = resolve(path);
  if (!existsSync(envPath)) {
    return;
  }

  for (const line of readFileSync(envPath, "utf8").split(/\r?\n/)) {
    const trimmed = line.trim();
    if (!trimmed || trimmed.startsWith("#")) {
      continue;
    }

    const match = /^([A-Za-z_][A-Za-z0-9_]*)=(.*)$/.exec(trimmed);
    if (!match) {
      continue;
    }

    const [, key, rawValue] = match;
    if (process.env[key] !== undefined) {
      continue;
    }

    process.env[key] = parseEnvValue(rawValue);
  }
}

function parseEnvValue(rawValue) {
  const value = rawValue.trim();
  if (
    (value.startsWith('"') && value.endsWith('"')) ||
    (value.startsWith("'") && value.endsWith("'"))
  ) {
    return value.slice(1, -1);
  }

  return value;
}
