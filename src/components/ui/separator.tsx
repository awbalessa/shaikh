import { cn } from "@/lib/utils";
import { Separator as RadixSeparator } from "radix-ui";

export function Separator({
  className,
  ...props
}: RadixSeparator.SeparatorProps) {
  return (
    <RadixSeparator.Root
      className={cn("bg-border h-px w-full", className)}
      {...props}
    ></RadixSeparator.Root>
  );
}
