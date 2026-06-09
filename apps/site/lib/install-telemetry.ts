import { neon } from "@neondatabase/serverless";

export const trafficEventTypes = ["site_visit", "install_request"] as const;

export type TrafficEventType = (typeof trafficEventTypes)[number];

export type JsonValue =
  | null
  | string
  | number
  | boolean
  | JsonObject
  | JsonValue[];

export type JsonObject = {
  [key: string]: JsonValue;
};

type RecordTrafficEventInput = {
  eventType: TrafficEventType;
  headers: Headers;
  clientKind?: string | null;
  metadata?: JsonObject;
  queryParams?: JsonObject;
  url?: string | URL;
};

type RecordTrafficEventResult = {
  id: string;
  stored: boolean;
  reason?: "missing_database_url";
};

const databaseEnvNames = [
  "NEWSJACK_DATABASE_URL",
  "DATABASE_URL",
  "POSTGRES_URL",
  "POSTGRES_PRISMA_URL",
] as const;

const installerMatchers = [
  { kind: "curl", pattern: /\bcurl(?:\/|\b)/i },
  { kind: "wget", pattern: /\bwget(?:\/|\b)/i },
  { kind: "httpie", pattern: /\bhttps?ie(?:\/|\b)|\bhttpie(?:\/|\b)/i },
  { kind: "python-requests", pattern: /\bpython-requests(?:\/|\b)/i },
  { kind: "go-http-client", pattern: /\bgo-http-client(?:\/|\b)/i },
  { kind: "libwww-perl", pattern: /\blibwww-perl(?:\/|\b)/i },
  { kind: "powershell", pattern: /\bpowershell(?:\/|\b)/i },
  { kind: "aria2", pattern: /\baria2(?:\/|\b)/i },
] as const;

const agentMatchers = [
  { kind: "codex", pattern: /\bcodex(?:\/|\b)/i },
  { kind: "claude", pattern: /\bclaude(?:\/|\b)/i },
  { kind: "openai", pattern: /\bopenai(?:\/|\b)/i },
  { kind: "anthropic", pattern: /\banthropic(?:\/|\b)/i },
  { kind: "googlebot", pattern: /\bgooglebot(?:\/|\b)/i },
  { kind: "bingbot", pattern: /\bbingbot(?:\/|\b)/i },
  { kind: "slackbot", pattern: /\bslackbot(?:\/|\b)/i },
] as const;

let warnedMissingDatabaseUrl = false;
let warnedMissingIpHashSalt = false;

export function getInstallerKind(userAgent: string): string | null {
  for (const matcher of installerMatchers) {
    if (matcher.pattern.test(userAgent)) {
      return matcher.kind;
    }
  }

  return null;
}

export function getClientKind(userAgent: string): string {
  const installerKind = getInstallerKind(userAgent);
  if (installerKind) {
    return installerKind;
  }

  for (const matcher of agentMatchers) {
    if (matcher.pattern.test(userAgent)) {
      return matcher.kind;
    }
  }

  if (/\b(bot|crawler|spider)\b/i.test(userAgent)) {
    return "bot";
  }

  if (/\b(chrome|firefox|safari|edg|opera|vivaldi)\b/i.test(userAgent)) {
    return "browser";
  }

  return userAgent.trim() ? "unknown" : "missing";
}

export function queryParamsFromUrl(url: string | URL): JsonObject {
  const parsedUrl = typeof url === "string" ? new URL(url) : url;
  const queryParams: JsonObject = {};

  parsedUrl.searchParams.forEach((value, key) => {
    const current = queryParams[key];
    if (typeof current === "string") {
      queryParams[key] = [current, value];
    } else if (Array.isArray(current)) {
      current.push(value);
    } else {
      queryParams[key] = value;
    }
  });

  return queryParams;
}

export async function recordTrafficEvent(
  input: RecordTrafficEventInput,
): Promise<RecordTrafficEventResult> {
  const id = crypto.randomUUID();
  const databaseUrl = databaseUrlFromEnv();

  if (!databaseUrl) {
    if (!warnedMissingDatabaseUrl) {
      warnedMissingDatabaseUrl = true;
      console.warn(
        `${databaseEnvNames.join(" or ")} is not set; skipping traffic telemetry`,
      );
    }
    return { id, stored: false, reason: "missing_database_url" };
  }

  const sql = neon(databaseUrl);
  const metadata = input.metadata ?? {};
  const queryParams = input.queryParams ?? (
    input.url ? queryParamsFromUrl(input.url) : {}
  );
  const userAgent = headerOrNull(input.headers, "user-agent") ?? "";
  const clientKind = input.clientKind ?? getClientKind(userAgent);

  await sql`
    INSERT INTO install_events (
      id,
      event_type,
      ip_hash,
      country,
      region,
      user_agent,
      client_kind,
      referer,
      accept_language,
      query_params,
      metadata
    )
    VALUES (
      ${id},
      ${input.eventType},
      ${await hashIpFromHeaders(input.headers, databaseUrl)},
      ${headerOrNull(input.headers, "x-vercel-ip-country")},
      ${headerOrNull(input.headers, "x-vercel-ip-country-region")},
      ${userAgent || null},
      ${clientKind},
      ${headerOrNull(input.headers, "referer", "referrer")},
      ${headerOrNull(input.headers, "accept-language")},
      ${JSON.stringify(queryParams)}::jsonb,
      ${JSON.stringify(metadata)}::jsonb
    )
  `;

  return { id, stored: true };
}

function databaseUrlFromEnv(): string | undefined {
  for (const name of databaseEnvNames) {
    const value = process.env[name];
    if (value) {
      return value;
    }
  }

  return undefined;
}

function headerOrNull(headers: Headers, ...names: string[]): string | null {
  for (const name of names) {
    const value = headers.get(name);
    if (value && value.trim()) {
      return value.trim();
    }
  }

  return null;
}

function clientIpFromHeaders(headers: Headers): string | null {
  const forwardedFor = headerOrNull(
    headers,
    "x-forwarded-for",
    "x-vercel-forwarded-for",
  );
  if (forwardedFor) {
    return forwardedFor.split(",")[0]?.trim() || null;
  }

  return headerOrNull(headers, "x-real-ip", "cf-connecting-ip", "x-client-ip");
}

async function hashIpFromHeaders(
  headers: Headers,
  fallbackSalt: string,
): Promise<string | null> {
  const ip = clientIpFromHeaders(headers);
  if (!ip) {
    return null;
  }

  const salt = process.env.NEWSJACK_IP_HASH_SALT || fallbackSalt;
  if (!salt) {
    if (!warnedMissingIpHashSalt) {
      warnedMissingIpHashSalt = true;
      console.warn("IP hash salt is not set; skipping IP hash");
    }
    return null;
  }

  const day = new Date().toISOString().slice(0, 10);
  const digest = await crypto.subtle.digest(
    "SHA-256",
    new TextEncoder().encode(`${salt}:${day}:${ip}`),
  );

  return Array.from(new Uint8Array(digest), (byte) =>
    byte.toString(16).padStart(2, "0"),
  ).join("");
}
