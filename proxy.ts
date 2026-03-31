import { NextRequest, NextResponse } from "next/server";
import { locales, defaultLocale, isLocale, Locale } from "./lib/i18n/locale";
import Negotiator from "negotiator";
import { match } from "@formatjs/intl-localematcher";

function getLocale(request: NextRequest): Locale {
  const cookieLocale = request.cookies.get("locale")?.value;
  if (isLocale(cookieLocale)) return cookieLocale;

  const acceptLanguage = request.headers.get("accept-language") ?? "";
  const languages = new Negotiator({
    headers: { "accept-language": acceptLanguage },
  }).languages();

  return match(languages, locales, defaultLocale) as Locale;
}

export function proxy(request: NextRequest) {
  const { pathname } = request.nextUrl;

  const pathLocale = pathname.split("/")[1];

  if (isLocale(pathLocale)) {
    if (pathLocale === defaultLocale && pathname === `/${defaultLocale}`) {
      const res = NextResponse.redirect(new URL("/", request.url));
      res.cookies.set("locale", defaultLocale);
      return res;
    }

    if (pathLocale === defaultLocale) {
      const newPath = pathname.replace(`/${defaultLocale}`, "") || "/";
      const res = NextResponse.redirect(new URL(newPath, request.url));
      res.cookies.set("locale", defaultLocale);
      return res;
    }

    const res = NextResponse.next();
    res.cookies.set("locale", pathLocale);
    return res;
  }

  const locale = getLocale(request);
  return NextResponse.rewrite(new URL(`/${locale}${pathname}`, request.url));
}

export const config = {
  matcher: ["/((?!_next|api|favicon.ico|.*\\..*).*)"],
};
