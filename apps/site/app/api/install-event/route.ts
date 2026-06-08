import {
  installEventSecretHeader,
  isInstallCallbackEventType,
  isJsonObject,
  isUuid,
  recordInstallEvent,
  type JsonObject,
} from "@/lib/install-telemetry";

export const runtime = "nodejs";
export const dynamic = "force-dynamic";

type InstallEventBody = {
  event_type?: unknown;
  install_id?: unknown;
  metadata?: unknown;
};

export async function POST(request: Request) {
  if (!isAuthorized(request.headers)) {
    return json({ error: "unauthorized" }, 401);
  }

  let body: InstallEventBody;
  try {
    body = (await request.json()) as InstallEventBody;
  } catch {
    return json({ error: "invalid_json" }, 400);
  }

  if (!isInstallCallbackEventType(body.event_type)) {
    return json({ error: "invalid_event_type" }, 400);
  }

  if (!isUuid(body.install_id)) {
    return json({ error: "invalid_install_id" }, 400);
  }

  if (body.metadata !== undefined && !isJsonObject(body.metadata)) {
    return json({ error: "invalid_metadata" }, 400);
  }

  try {
    const result = await recordInstallEvent({
      eventType: body.event_type,
      headers: request.headers,
      installId: body.install_id,
      metadata: (body.metadata ?? {}) as JsonObject,
      url: request.url,
    });

    return json(
      {
        event_id: result.id,
        ok: true,
        stored: result.stored,
      },
      202,
    );
  } catch (error: unknown) {
    console.error("Failed to record install callback event", error);
    return json({ ok: true, stored: false }, 202);
  }
}

function isAuthorized(headers: Headers): boolean {
  const configuredSecret = process.env.NEWSJACK_INSTALL_EVENT_SECRET;
  if (!configuredSecret) {
    console.warn("NEWSJACK_INSTALL_EVENT_SECRET is not set");
    return false;
  }

  const authorization = headers.get("authorization");
  const bearerPrefix = "Bearer ";
  const bearerSecret = authorization?.startsWith(bearerPrefix)
    ? authorization.slice(bearerPrefix.length)
    : null;
  const headerSecret = headers.get(installEventSecretHeader);

  return bearerSecret === configuredSecret || headerSecret === configuredSecret;
}

function json(body: JsonObject, status: number) {
  return Response.json(body, { status });
}
