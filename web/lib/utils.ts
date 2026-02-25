export function cn(...v: Array<string | undefined | false>) {
  return v.filter(Boolean).join(" ");
}
