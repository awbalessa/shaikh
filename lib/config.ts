export type AppLang = "ar" | "en";
export const APP_LANG = "ar";
export type Dir = "rtl" | "ltr";
export const BASE_DIR: Dir = APP_LANG === "ar" ? "rtl" : "ltr";
