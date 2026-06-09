import { NextRequest, NextResponse, type NextFetchEvent } from "next/server";

import {
  getClientKind,
  getInstallerKind,
  queryParamsFromUrl,
  recordTrafficEvent,
  type TrafficEventType,
} from "./lib/install-telemetry";

const repoURL = "https://github.com/elvisun/newsjack";

export function proxy(request: NextRequest, event: NextFetchEvent) {
  const userAgent = request.headers.get("user-agent") ?? "";
  const installerKind = getInstallerKind(userAgent);

  if (installerKind) {
    const url = request.nextUrl.clone();
    url.pathname = "/install.sh";
    recordRequest(event, request, "install_request", installerKind);
    return NextResponse.rewrite(url);
  }

  recordRequest(event, request, "site_visit", getClientKind(userAgent));
  return NextResponse.redirect(repoURL, 308);
}

function recordRequest(
  event: NextFetchEvent,
  request: NextRequest,
  eventType: TrafficEventType,
  clientKind: string,
) {
  event.waitUntil(
    recordTrafficEvent({
      eventType,
      headers: request.headers,
      clientKind,
      metadata: {
        host: request.headers.get("host"),
        method: request.method,
        path: request.nextUrl.pathname,
        vercel_id: request.headers.get("x-vercel-id"),
      },
      queryParams: queryParamsFromUrl(request.url),
      url: request.url,
    }).catch((error: unknown) => {
      console.error(`Failed to record ${eventType}`, error);
    }),
  );
}

export const config = {
  matcher: "/",
};
