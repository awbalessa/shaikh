"use client";

import { cn } from "@/lib/utils";
import { isTextUIPart, UIMessage } from "ai";
import { Streamdown } from "streamdown";
import { memo, useCallback, useMemo, useState } from "react";
import { IconCheck, IconCopy, IconEdit } from "@tabler/icons-react";
import { AnimatePresence, motion } from "motion/react";
import { useIntlayer } from "next-intlayer";
import { Tooltip, TooltipContent, TooltipTrigger } from "../ui/tooltip";

type MessageProps = {
  message: UIMessage;
};

const Message = memo(function Message({ message }: MessageProps) {
  const content = useIntlayer("chat-message");
  const isUser = message.role === "user";
  const parts = useMemo(
    () => message.parts.filter(isTextUIPart),
    [message.parts],
  );
  const getText = useCallback(() => parts.map((p) => p.text).join(""), [parts]);

  return (
    <div
      dir={isUser ? "ltr" : "auto"}
      className={cn("group flex flex-col w-full", isUser && "items-end")}
    >
      <div
        dir="auto"
        className={cn(
          isUser &&
            "bg-surface dark:bg-surface-raised rounded-2xl px-4 py-2 max-w-[80%]",
        )}
      >
        {isUser
          ? parts.map((part, i) => <span key={i}>{part.text}</span>)
          : parts.map((part, i) => (
              <Streamdown key={i}>{part.text}</Streamdown>
            ))}
      </div>
      {isUser && (
        <div className="flex items-center gap-0.5 py-1 opacity-0 group-hover:opacity-100 transition-opacity duration-100">
          <MessageCopyAction
            label={content.actions.copy}
            copiedLabel={content.actions.copied}
            getText={getText}
          >
            <IconCopy className="size-4" />
          </MessageCopyAction>
          <MessageAction label={content.actions.edit} onClick={() => {}}>
            <IconEdit className="size-4" />
          </MessageAction>
        </div>
      )}
    </div>
  );
});

export default Message;

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
