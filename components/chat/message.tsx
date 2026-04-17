"use client";

import { cn } from "@/lib/utils";
import React, {
  createContext,
  useCallback,
  useContext,
  useMemo,
  useState,
} from "react";
import { Tooltip, TooltipContent, TooltipTrigger } from "../ui/tooltip";
import { motion, AnimatePresence } from "motion/react";
import { IconCheck } from "@tabler/icons-react";

Message.Content = MessageContent;
Message.Actions = MessageActions;
Message.Action = MessageAction;
Message.CopyAction = MessageCopyAction;

type MessageContextValue = { from: "system" | "user" | "assistant" };

const MessageContext = createContext<MessageContextValue | null>(null);

function useMessage() {
  const ctx = useContext(MessageContext);
  if (!ctx) throw new Error("useMessage must be within <Message>");
  return ctx;
}

type MessageProps = React.ComponentPropsWithoutRef<"div"> & {
  from: "system" | "user" | "assistant";
};

export default function Message({
  from,
  children,
  className,
  ...props
}: MessageProps) {
  const contextValue = useMemo(() => ({ from }), [from]);

  return (
    <MessageContext.Provider value={contextValue}>
      <div
        dir={from === "user" ? "ltr" : "auto"}
        className={cn(
          "group flex flex-col w-full",
          from === "user" && "items-end",
          className,
        )}
        {...props}
      >
        {children}
      </div>
    </MessageContext.Provider>
  );
}

function MessageContent({
  children,
  className,
}: React.ComponentPropsWithoutRef<"div">) {
  const { from } = useMessage();
  return (
    <div
      dir="auto"
      className={cn(
        from === "user" && "bg-muted rounded-2xl px-4 py-2 max-w-[80%]",
        className,
      )}
    >
      {children}
    </div>
  );
}

function MessageActions({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex items-center gap-0.5 py-1 opacity-0 group-hover:opacity-100 transition-opacity duration-100">
      {children}
    </div>
  );
}

type MessageActionProps = {
  label: React.ReactNode;
  onClick?: () => void;
  children: React.ReactNode;
};

function MessageAction({ label, onClick, children }: MessageActionProps) {
  const [hovered, setHovered] = useState(false);
  return (
    <Tooltip open={hovered} onOpenChange={setHovered}>
      <TooltipTrigger asChild>
        <button
          type="button"
          onClick={onClick}
          tabIndex={-1}
          className="rounded-md p-1 text-muted-foreground hover:text-foreground hover:bg-muted transition-colors"
        >
          {children}
        </button>
      </TooltipTrigger>
      <TooltipContent side="bottom">{label}</TooltipContent>
    </Tooltip>
  );
}

type MessageCopyActionProps = {
  label: React.ReactNode;
  copiedLabel: React.ReactNode;
  getText: () => string;
  children: React.ReactNode;
};

function MessageCopyAction({
  label,
  copiedLabel,
  getText,
  children,
}: MessageCopyActionProps) {
  const [copied, setCopied] = useState(false);
  const [hovered, setHovered] = useState(false);

  const handleClick = useCallback(() => {
    navigator.clipboard.writeText(getText());
    setCopied(true);
    setTimeout(() => setCopied(false), 1000);
  }, [getText]);

  return (
    <Tooltip open={hovered || copied} onOpenChange={setHovered}>
      <TooltipTrigger asChild>
        <button
          type="button"
          onClick={handleClick}
          tabIndex={-1}
          className="rounded-md p-1 text-muted-foreground hover:text-foreground hover:bg-muted transition-colors"
        >
          <AnimatePresence mode="wait" initial={false}>
            {copied ? (
              <motion.span
                key="check"
                initial={{ scale: 0 }}
                animate={{ scale: 1 }}
                exit={{ scale: 0 }}
                transition={{ duration: 0.1 }}
                className="block"
              >
                <IconCheck className="size-4" />
              </motion.span>
            ) : (
              <motion.span
                key="icon"
                initial={{ scale: 0 }}
                animate={{ scale: 1 }}
                exit={{ scale: 0 }}
                transition={{ duration: 0.1 }}
                className="block"
              >
                {children}
              </motion.span>
            )}
          </AnimatePresence>
        </button>
      </TooltipTrigger>
      <TooltipContent side="bottom">
        {copied ? copiedLabel : label}
      </TooltipContent>
    </Tooltip>
  );
}
