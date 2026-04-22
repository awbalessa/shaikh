"use client";

import { cn } from "@/lib/utils";
import { Tooltip } from "radix-ui";

const TooltipProvider = Tooltip.Provider;
const TooltipRoot = Tooltip.Root;
const TooltipTrigger = Tooltip.Trigger;

function TooltipContent({
  sideOffset = 4,
  children,
  className,
  ...props
}: Tooltip.TooltipContentProps) {
  return (
    <Tooltip.Portal>
      <Tooltip.Content
        sideOffset={sideOffset}
        className={cn(
          "z-50 rounded-md bg-surface px-2 py-1 text-xs text-text-primary shadow-md",
          className,
        )}
        {...props}
      >
        {children}
      </Tooltip.Content>
    </Tooltip.Portal>
  );
}

export {
  TooltipProvider,
  TooltipRoot as Tooltip,
  TooltipTrigger,
  TooltipContent,
};
