import { useAuthModal } from "@/hooks/use-auth-modal";
import { AnimatePresence } from "motion/react";
import {
  Dialog,
  DialogContent,
  DialogOverlay,
  DialogPortal,
  DialogTitle,
} from "../ui/dialog";
import { useIntlayer } from "next-intlayer/server";

export function AuthModal() {
  const content = useIntlayer("auth-modal");
  const { open, view, variant, closeModal } = useAuthModal();

  return (
    <AnimatePresence>
      {open && (
        <Dialog open={open} onOpenChange={closeModal}>
          <DialogPortal>
            <DialogOverlay />
            <DialogContent>
              {variant === "messageAttempt" && (
                <>
                  <DialogTitle>
                    {view === "entry" && content.fromMessageAttempt.entry.title}
                    {view === "signup" &&
                      content.fromMessageAttempt.signup.title}
                    {view === "login" && content.fromMessageAttempt.login.title}
                  </DialogTitle>
                </>
              )}
            </DialogContent>
          </DialogPortal>
        </Dialog>
      )}
    </AnimatePresence>
  );
}
