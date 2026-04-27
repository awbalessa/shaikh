import { Dialog } from "radix-ui";
import { motion } from "motion/react";
import { cn } from "@/lib/utils";

const DialogRoot = Dialog.Root;
export { DialogRoot as Dialog };
export const DialogPortal = Dialog.Portal;
export const DialogClose = Dialog.Close;

export function DialogOverlay({ ...props }) {
  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      exit={{ opacity: 0 }}
      transition={{ duration: 0.15 }}
    >
      <Dialog.Overlay className="fixed inset-0 bg-background" {...props} />
    </motion.div>
  );
}

export function DialogContent({
  children,
  className,
  ...props
}: Dialog.DialogContentProps) {
  return (
    <Dialog.Content asChild {...props}>
      <motion.div
        initial={{ opacity: 0, scale: 0.95 }}
        animate={{ opacity: 1, scale: 1 }}
        exit={{ opacity: 0, scale: 0.95 }}
        transition={{ duration: 0.2 }}
        className={cn(
          "fixed inset-0 flex items-center justify-center",
          className,
        )}
      >
        {children}
      </motion.div>
    </Dialog.Content>
  );
}

export function DialogTitle({
  children,
  className,
  ...props
}: Dialog.DialogTitleProps) {
  return (
    <Dialog.Title className={cn(className)} {...props}>
      {children}
    </Dialog.Title>
  );
}
