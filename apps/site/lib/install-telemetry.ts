import { neon } from "@neondatabase/serverless";

export const installIdRequestHeader = "x-newsjack-install-id";
export const installLoggedRequestHeader = "x-newsjack-install-logged";
export const installIdResponseHeader = "X-Newsjack-Install-Id";
export const installEventSecretHeader = "x-newsjack-install-event-secret";

export const installEventTypes = [
  "curl_hit",
  "install_started",
  "install_completed",
  "install_failed",
] as const;

export const installCallbackEventTypes = [
  "install_started",
  "install_completed",
  "install_failed",
] as const;

export type InstallEventType = (typeof installEventTypes)[number];
export type InstallCallbackEventType = (typeof installCallbackEventTypes)[number];

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

type RecordInstallEventInput = {
  eventType: InstallEventType;
  headers: Headers;
  id?: string;
  installId?: string | null;
  installerKind?: string | null;
  metadata?: JsonObject;
  queryParams?: JsonObject;
  url?: string | URL;
};

type RecordInstallEventResult = {
  id: string;
  stored: boolean;
  reason?: "missing_database_url";
};

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

export function isInstallCallbackEventType(
  value: unknown,
): value is InstallCallbackEventType {
  return (
    typeof value === "string" &&
    installCallbackEventTypes.includes(value as InstallCallbackEventType)
  );
}

export function isUuid(value: unknown): value is string {
  return (
    typeof value === "string" &&
    /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i.test(
      value,
    )
  );
}

export function isJsonObject(value: unknown): value is JsonObject {
  return (
    value !== null &&
    typeof value === "object" &&
    !Array.isArray(value)
  );
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

export async function recordInstallEvent(
  input: RecordInstallEventInput,
): Promise<RecordInstallEventResult> {
  const id = input.id ?? crypto.randomUUID();
  const databaseUrl = process.env.NEWSJACK_DATABASE_URL;

  if (!databaseUrl) {
    if (!warnedMissingDatabaseUrl) {
      warnedMissingDatabaseUrl = true;
      console.warn("NEWSJACK_DATABASE_URL is not set; skipping install telemetry");
    }
    return { id, stored: false, reason: "missing_database_url" };
  }

  const sql = neon(databaseUrl);
  const metadata = input.metadata ?? {};
  const queryParams = input.queryParams ?? (
    input.url ? queryParamsFromUrl(input.url) : {}
  );

  await sql`
    INSERT INTO install_events (
      id,
      install_id,
      event_type,
      ip_hash,
      country,
      region,
      user_agent,
      referer,
      accept_language,
      query_params,
      installer_kind,
      metadata
    )
    VALUES (
      ${id},
      ${input.installId ?? (input.eventType === "curl_hit" ? id : null)},
      ${input.eventType},
      ${await hashIpFromHeaders(input.headers)},
      ${headerOrNull(input.headers, "x-vercel-ip-country")},
      ${headerOrNull(input.headers, "x-vercel-ip-country-region")},
      ${headerOrNull(input.headers, "user-agent")},
      ${headerOrNull(input.headers, "referer", "referrer")},
      ${headerOrNull(input.headers, "accept-language")},
      ${JSON.stringify(queryParams)}::jsonb,
      ${input.installerKind ?? null},
      ${JSON.stringify(metadata)}::jsonb
    )
  `;

  return { id, stored: true };
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

async function hashIpFromHeaders(headers: Headers): Promise<string | null> {
  const ip = clientIpFromHeaders(headers);
  if (!ip) {
    return null;
  }

  const salt = process.env.NEWSJACK_IP_HASH_SALT;
  if (!salt) {
    if (!warnedMissingIpHashSalt) {
      warnedMissingIpHashSalt = true;
      console.warn("NEWSJACK_IP_HASH_SALT is not set; skipping IP hash");
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
