import { gateway } from "ai";

export const models = {
  low: gateway("google/gemini-3.1-flash-lite-preview"),
  medium: gateway("google/gemini-3-flash"),
  high: gateway("google/gemini-3.1-pro-preview"),
} as const;

export type ModelTier = keyof typeof models;
