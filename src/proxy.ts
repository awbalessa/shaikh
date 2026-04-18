import { NextRequest, NextResponse } from "next/server";

const LOCALES = ["ar", "en"];
const DEFAULT_LOCALE = "ar";

export function proxy(request: NextRequest) {
  const pathname = request.nextUrl.pathname;

  const localeInPath = LOCALES.some(
    (locale) => pathname.startsWith(`/${locale}/`) || pathname === `/${locale}`,
  );

  if (localeInPath) {
    const locale = pathname.split("/")[1];

    if (locale === DEFAULT_LOCALE) {
      const response = NextResponse.redirect(new URL("/", request.url));
      response.cookies.set("locale", locale, {
        maxAge: 60 * 60 * 24 * 365,
        path: "/",
      });
      return response;
    }

    const response = NextResponse.next();
    response.cookies.set("locale", locale, {
      maxAge: 60 * 60 * 24 * 365,
      path: "/",
    });
    return response;
  }

  if (pathname === "/") {
    const cookieLocale = detectLocaleFromCookie(request);

    if (cookieLocale) {
      const rewrite = request.nextUrl.clone();
      rewrite.pathname = `/${cookieLocale}/`;
      return NextResponse.rewrite(rewrite);
    }

    const detectedLocale = detectLocaleFromHeader(request) || DEFAULT_LOCALE;

    if (detectedLocale === DEFAULT_LOCALE) {
      const rewrite = request.nextUrl.clone();
      rewrite.pathname = `/${DEFAULT_LOCALE}/`;
      return NextResponse.rewrite(rewrite);
    } else {
      return NextResponse.redirect(new URL(`/${detectedLocale}/`, request.url));
    }
  }

  return NextResponse.next();
}

function detectLocaleFromCookie(request: NextRequest): string | null {
  return request.cookies.get("locale")?.value || null;
}

function detectLocaleFromHeader(request: NextRequest): string | null {
  const acceptLanguage = request.headers.get("accept-language") || "";

  const localePreferences = acceptLanguage
    .split(",")
    .map((pref) => pref.split(";")[0].trim().toLowerCase());

  for (const pref of localePreferences) {
    if (LOCALES.includes(pref)) return pref;
    const langCode = pref.split("-")[0];
    if (LOCALES.includes(langCode)) return langCode;
  }

  return null;
}

export const config = {
  matcher: ["/((?!api|_next/static|_next/image|favicon.ico).*)"],
};
