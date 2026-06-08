#!/usr/bin/env node
import { spawnSync } from "node:child_process";

const databaseUrl = process.env.NEWSJACK_DATABASE_URL;

if (!databaseUrl) {
  console.error("NEWSJACK_DATABASE_URL is required.");
  process.exit(1);
}

const query = `
  SELECT
    event_type,
    COALESCE(country, 'unknown') AS country,
    count(*)::int AS count
  FROM install_events
  WHERE created_at >= now() - interval '24 hours'
  GROUP BY event_type, country
  ORDER BY event_type ASC, count DESC, country ASC;
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
    const [eventType, country, count] = line.split("\t");
    return {
      event_type: eventType,
      country,
      count: Number(count),
    };
  });

if (rows.length === 0) {
  console.log("No install events in the last 24h.");
} else {
  console.log("Install funnel events, last 24h:");
  console.table(rows);
}
