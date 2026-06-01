import { NextRequest, NextResponse } from "next/server";

const installerUserAgent =
  /\b(curl|wget|httpie|python-requests|go-http-client|libwww-perl|powershell|aria2)\b/i;

const repoURL = "https://github.com/elvisun/newsjack";

export function proxy(request: NextRequest) {
  const userAgent = request.headers.get("user-agent") ?? "";

  if (installerUserAgent.test(userAgent)) {
    const url = request.nextUrl.clone();
    url.pathname = "/install.sh";
    return NextResponse.rewrite(url);
  }

  return NextResponse.redirect(repoURL, 308);
}

export const config = {
  matcher: "/",
};
