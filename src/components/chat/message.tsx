"use client";

import { cn } from "@/lib/utils";
import { isTextUIPart, UIMessage } from "ai";
import { Streamdown } from "streamdown";
import { memo, useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
  IconCheck,
  IconCopy,
  IconEdit,
  IconInfoCircle,
} from "@tabler/icons-react";
import { AnimatePresence, motion } from "motion/react";
import { useIntlayer } from "next-intlayer";
import { Tooltip, TooltipContent, TooltipTrigger } from "../ui/tooltip";

type MessageProps = {
  message: UIMessage;
  editingMessageID: string | null;
  onStartEditing: (id: string) => void;
  onStopEditing: () => void;
  onEditMessage: (id: string, text: string) => void;
};

const Message = memo(function Message({
  message,
  editingMessageID,
  onStartEditing,
  onStopEditing,
  onEditMessage,
}: MessageProps) {
  if (message.role === "user") {
    return (
      <UserMessage
        message={message}
        editingMessageID={editingMessageID}
        onStartEditing={onStartEditing}
        onStopEditing={onStopEditing}
        onEditMessage={onEditMessage}
      />
    );
  }
  return <AssistantMessage message={message} />;
});

export default Message;

type UserMessageProps = {
  message: UIMessage;
  editingMessageID: string | null;
  onStartEditing: (id: string) => void;
  onStopEditing: () => void;
  onEditMessage: (id: string, text: string) => void;
};

function UserMessage({
  message,
  editingMessageID,
  onStartEditing,
  onStopEditing,
  onEditMessage,
}: UserMessageProps) {
  const parts = useMemo(
    () => message.parts.filter(isTextUIPart),
    [message.parts],
  );
  const getText = useCallback(() => parts.map((p) => p.text).join(""), [parts]);
  const [editValue, setEditValue] = useState(getText());
  const isEditing = editingMessageID === message.id;

  const handleEdit = () => {
    setEditValue(getText());
    onStartEditing(message.id);
  };

  const handleEditSend = () => {
    const trimmed = editValue.trim();
    if (trimmed !== getText()) {
      onEditMessage(message.id, trimmed);
    }
    onStopEditing();
  };

  const handleEditCancel = () => {
    onStopEditing();
  };

  if (isEditing) {
    return (
      <UserMessageEditor
        originalValue={getText()}
        value={editValue}
        onChange={setEditValue}
        onSend={handleEditSend}
        onCancel={handleEditCancel}
      />
    );
  }

  return (
    <div dir="ltr" className="group flex flex-col w-full items-end">
      <div dir="auto" className="bg-surface rounded-xl px-4 py-2 max-w-[80%]">
        {parts.map((part, i) => (
          <span key={i}>{part.text}</span>
        ))}
      </div>
      <UserMessageActions
        getText={getText}
        onEdit={handleEdit}
        className="opacity-0 group-hover:opacity-100 transition-opacity duration-100"
      />
    </div>
  );
}

type AssistantMessageProps = {
  message: UIMessage;
};

function AssistantMessage({ message }: AssistantMessageProps) {
  return (
    <div dir="auto" className="flex flex-col w-full">
      {message.parts.filter(isTextUIPart).map((part, i) => (
        <Streamdown key={i}>{part.text}</Streamdown>
      ))}
    </div>
  );
}

type UserMessageActionsProps = React.ComponentPropsWithoutRef<"div"> & {
  getText: () => string;
  onEdit: () => void;
};

function UserMessageActions({
  getText,
  onEdit,
  className,
}: UserMessageActionsProps) {
  return (
    <div className={cn("flex items-center py-1", className)}>
      <MessageCopyAction getText={getText} />
      <MessageEditAction onEdit={onEdit} />
    </div>
  );
}

type MessageCopyActionProps = {
  getText: () => string;
};

function MessageCopyAction({ getText }: MessageCopyActionProps) {
  const content = useIntlayer("chat-message").user.tooltips;
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
          className="rounded-md p-1 hover:bg-surface text-text-tertiary hover:text-text-primary transition-colors"
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
                <IconCopy className="size-4" />
              </motion.span>
            )}
          </AnimatePresence>
        </button>
      </TooltipTrigger>
      <TooltipContent side="bottom">
        {copied ? content.copied : content.copy}
      </TooltipContent>
    </Tooltip>
  );
}

type MessageEditActionProps = {
  onEdit: () => void;
};

function MessageEditAction({ onEdit }: MessageEditActionProps) {
  const content = useIntlayer("chat-message").user.tooltips;
  const [hovered, setHovered] = useState(false);
  return (
    <Tooltip open={hovered} onOpenChange={setHovered}>
      <TooltipTrigger asChild>
        <button
          type="button"
          onClick={onEdit}
          tabIndex={-1}
          className="rounded-md p-1 text-text-tertiary hover:text-text-primary hover:bg-surface dark:hover:bg-surface-raised transition-colors"
        >
          <IconEdit className="size-4" />
        </button>
      </TooltipTrigger>
      <TooltipContent side="bottom">{content.edit}</TooltipContent>
    </Tooltip>
  );
}

type UserMessageEditorProps = {
  originalValue: string;
  value: string;
  onChange: (value: string) => void;
  onSend: () => void;
  onCancel: () => void;
};

function UserMessageEditor({
  originalValue,
  value,
  onChange,
  onSend,
  onCancel,
}: UserMessageEditorProps) {
  const content = useIntlayer("chat-message").user.editor;
  const isDirty = value.trim() !== originalValue.trim();
  const textAreaRef = useRef<HTMLTextAreaElement | null>(null);

  useEffect(() => {
    const el = textAreaRef.current;
    if (!el) return;
    el.style.height = "auto";
    el.style.height = `${el.scrollHeight}px`;
  }, [value]);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") onCancel();
    };
    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [onCancel]);

  return (
    <div className="flex flex-col w-full p-3 gap-3 bg-surface rounded-lg">
      <textarea
        ref={textAreaRef}
        dir="auto"
        autoFocus
        className="w-full p-3 bg-background dark:bg-surface-raised rounded-lg outline-none resize-none text-text-primary focus-visible:ring-1 focus-visible:ring-primary"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        onKeyDown={(e) => {
          if (e.key === "Enter" && !e.shiftKey) {
            e.preventDefault();
            onSend();
          }
        }}
        rows={2}
      />
      <div className="flex flex-row justify-between items-center">
        <div className="flex flex-row gap-1 items-center">
          <IconInfoCircle className="size-3" />
          <p className="text-xs">{content.warning}</p>
        </div>
        <div className="flex flex-row gap-2">
          <button
            onClick={onCancel}
            className="px-4 py-2 w-21.25 rounded-md border border-border bg-transparent hover:bg-surface-raised transition-colors"
          >
            {content.cancel}
          </button>
          <button
            disabled={!isDirty}
            onClick={onSend}
            className="px-4 py-2 w-21.25 rounded-md disabled:bg-foreground/65 disabled:text-text-inverse/65 bg-foreground text-text-inverse transition-opacity"
          >
            {content.send}
          </button>
        </div>
      </div>
    </div>
  );
}
