import { NextRequest, NextResponse, type NextFetchEvent } from "next/server";

import {
  getInstallerKind,
  installIdRequestHeader,
  installIdResponseHeader,
  installLoggedRequestHeader,
  isUuid,
  queryParamsFromUrl,
  recordInstallEvent,
} from "./lib/install-telemetry";

const repoURL = "https://github.com/elvisun/newsjack";

export function proxy(request: NextRequest, event: NextFetchEvent) {
  const userAgent = request.headers.get("user-agent") ?? "";
  const installerKind = getInstallerKind(userAgent);

  if (installerKind) {
    const url = request.nextUrl.clone();
    const existingInstallId = request.headers.get(installIdRequestHeader);
    const installId = isUuid(existingInstallId)
      ? existingInstallId
      : crypto.randomUUID();
    const requestHeaders = new Headers(request.headers);

    url.pathname = "/install.sh";
    requestHeaders.set(installIdRequestHeader, installId);
    requestHeaders.set(installLoggedRequestHeader, "1");

    if (request.headers.get(installLoggedRequestHeader) !== "1") {
      event.waitUntil(
        recordInstallEvent({
          eventType: "curl_hit",
          headers: request.headers,
          id: installId,
          installId,
          installerKind,
          metadata: {
            host: request.headers.get("host"),
            method: request.method,
            path: request.nextUrl.pathname,
            vercel_id: request.headers.get("x-vercel-id"),
          },
          queryParams: queryParamsFromUrl(request.url),
          url: request.url,
        }).catch((error: unknown) => {
          console.error("Failed to record install curl_hit", error);
        }),
      );
    }

    const response = NextResponse.rewrite(url, {
      request: {
        headers: requestHeaders,
      },
    });
    response.headers.set(installIdResponseHeader, installId);
    return response;
  }

  return NextResponse.redirect(repoURL, 308);
}

export const config = {
  matcher: "/",
};
