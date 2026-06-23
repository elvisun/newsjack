export const HSTS_HEADER_NAME = "Strict-Transport-Security";
export const HSTS_HEADER_VALUE = "max-age=300";

export function applySecurityHeaders(headers: Headers) {
  headers.set(HSTS_HEADER_NAME, HSTS_HEADER_VALUE);
  return headers;
}
