import { cn } from "@/lib/utils";
import { useState } from "react";
import { motion } from "motion/react";

type InputProps = React.InputHTMLAttributes<HTMLInputElement> & {
  label?: string;
  error?: string;
};

export function Input({
  value,
  onChange,
  type = "text",
  label,
  className,
  ...props
}: InputProps) {
  const [isFocused, setIsFocused] = useState(false);
  const hasValue = !!value;

  return (
    <div className="relative">
      <input
        dir={hasValue ? "auto" : undefined}
        type={type}
        value={value}
        onChange={onChange}
        onFocus={() => setIsFocused(true)}
        onBlur={() => setIsFocused(false)}
        className={cn(
          "bg-surface-input outline-none border border-border focus:border-border-strong transition-colors rounded-md px-4 pt-6 pb-2",
          className,
        )}
        {...props}
      />
      {label && (
        <motion.label
          animate={{
            y: isFocused || hasValue ? -7 : 0,
            x: isFocused || hasValue ? -3 : 0,
            scale: isFocused || hasValue ? 9 / 14 : 1,
          }}
          transition={{ duration: 0.15 }}
          className="absolute start-3 top-2.5 text-text-tertiary pointer-events-none"
        >
          {label}
        </motion.label>
      )}
    </div>
  );
}
