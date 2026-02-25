"use client";
import { useEffect } from "react";

export function ThemeToggle() {
  useEffect(() => {
    const toggle = () => {
      document.documentElement.classList.toggle("dark");
    };

    const onKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === "j") {
        e.preventDefault();
        toggle();
      }
    };

    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, []);

  return null;
}
