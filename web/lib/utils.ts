export function cn(...v: Array<string | undefined | false>) {
  return v.filter(Boolean).join(" ");
}

export function getIconStroke(size: number) {
  return size / 12;
}
