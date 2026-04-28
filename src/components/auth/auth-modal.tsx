"use client";

import { useState } from "react";
import { useAuthModal, type AuthModalVariant } from "@/hooks/use-auth-modal";
import { AnimatePresence } from "motion/react";
import { FaApple, FaGoogle } from "react-icons/fa";
import { IconX } from "@tabler/icons-react";
import {
  Dialog,
  DialogContent,
  DialogOverlay,
  DialogPortal,
  DialogTitle,
  DialogClose,
} from "../ui/dialog";
import { Input } from "../ui/input";
import { Separator } from "../ui/separator";
import { useIntlayer } from "next-intlayer/server";

export function AuthModal() {
  const { open, view, variant, closeModal } = useAuthModal();

  return (
    <AnimatePresence>
      {open && (
        <Dialog open={open} onOpenChange={closeModal}>
          <DialogPortal>
            <DialogOverlay />
            <DialogContent
              className="flex flex-col items-center
  pt-50"
            >
              <DialogClose asChild>
                <button className="absolute top-6 end-8 rounded-full opacity-75 hover:opacity-100 transition-opacity">
                  <IconX className="size-5" />
                </button>
              </DialogClose>

              {view === "entry" && <EntryView variant={variant} />}
              {view === "login" && <LoginView variant={variant} />}
              {view === "signup" && <SignupView variant={variant} />}
            </DialogContent>
          </DialogPortal>
        </Dialog>
      )}
    </AnimatePresence>
  );
}

function EntryView({ variant }: { variant: AuthModalVariant }) {
  const intlayerContent = useIntlayer("auth-modal");
  const variantContent = intlayerContent.variants[variant].entry;
  const shared = intlayerContent.shared.entry;
  const { setView, setEmail } = useAuthModal();
  const [email, setLocalEmail] = useState("");

  const handleContinue = async () => {
    // Check if email exists (stub for now)
    setEmail(email);
    setView("signup"); // or 'login' based on check-email response
  };

  return (
    <div className="w-full max-w-xl flex flex-col gap-9">
      <AuthModalHeader
        title={variantContent.title}
        subtitle={variantContent.subtext}
      />

      <div className="w-full max-w-80 mx-auto flex flex-col gap-5">
        <OAuthButtons />

        <Separator className="h-[0.5px]" />

        <form
          className="flex flex-col gap-2"
          onSubmit={(e) => {
            e.preventDefault();
            handleContinue();
          }}
        >
          <Input
            type="email"
            label={shared.inputLabels.email}
            value={email}
            onChange={(e) => setLocalEmail(e.target.value)}
            autoFocus
            className="w-full h-12"
          />
          <button
            type="submit"
            disabled={true}
            className="w-full h-12 bg-surface-inverse disabled:dark:bg-surface-inverse/50 text-text-inverse rounded-lg transition-colors"
          >
            <span className="in-disabled:opacity-25">{shared.continue}</span>
          </button>
        </form>
      </div>

      <CloseButton />
    </div>
  );
}

function LoginView({ variant }: { variant: AuthModalVariant }) {
  const intlayerContent = useIntlayer("auth-modal");
  const variantContent = intlayerContent.variants[variant].login;
  const shared = intlayerContent.shared.login;
  const { email, closeModal } = useAuthModal();
  const [password, setPassword] = useState("");

  const handleLogin = async () => {
    // Call login API (stub for now)
    closeModal();
  };

  return (
    <div className="w-full max-w-xl flex flex-col gap-9">
      <AuthModalHeader
        title={variantContent.title}
        subtitle={variantContent.subtext}
      />

      <OAuthButtons />

      <Separator className="h-[0.5px]" />

      <form
        className="flex flex-col gap-4"
        onSubmit={(e) => {
          e.preventDefault();
          handleLogin();
        }}
      >
        <Input
          type="email"
          label={shared.inputLabels.email}
          value={email}
          disabled
        />
        <Input
          type="password"
          label={shared.inputLabels.password}
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          autoFocus
        />
        <button
          type="button"
          className="text-text-link hover:text-text-link-hover text-sm
   text-right"
        >
          {shared.forgotPassword}
        </button>
        <button
          type="submit"
          className="w-full h-12 bg-primary hover:bg-primary-hover
  text-on-primary rounded-lg font-medium transition-colors"
        >
          {shared.login}
        </button>
      </form>

      <CloseButton />
    </div>
  );
}

function SignupView({ variant }: { variant: AuthModalVariant }) {
  const intlayerContent = useIntlayer("auth-modal");
  const variantContent = intlayerContent.variants[variant].signup;
  const shared = intlayerContent.shared.signup;
  const { email, closeModal } = useAuthModal();
  const [name, setName] = useState("");
  const [password, setPassword] = useState("");

  const handleSignup = async () => {
    // Call signup API (stub for now)
    closeModal();
  };

  return (
    <div className="w-full max-w-xl flex flex-col gap-9">
      <AuthModalHeader
        title={variantContent.title}
        subtitle={variantContent.subtext}
      />

      <OAuthButtons />

      <Separator className="h-[0.5px]" />

      <form
        className="flex flex-col gap-4"
        onSubmit={(e) => {
          e.preventDefault();
          handleSignup();
        }}
      >
        <Input
          type="email"
          label={shared.inputLabels.email}
          value={email}
          disabled
        />
        <Input
          type="text"
          label={shared.inputLabels.name}
          value={name}
          onChange={(e) => setName(e.target.value)}
          autoFocus
        />
        <Input
          type="password"
          label={shared.inputLabels.password}
          value={password}
          onChange={(e) => setPassword(e.target.value)}
        />
        <button
          type="submit"
          className="w-full h-12 bg-primary hover:bg-primary-hover
  text-on-primary rounded-lg font-medium transition-colors"
        >
          {shared.signup}
        </button>
      </form>

      <CloseButton />
    </div>
  );
}

function AuthModalHeader({
  title,
  subtitle,
}: {
  title: string;
  subtitle: string;
}) {
  return (
    <div className="flex flex-col gap-2 items-center text-center">
      <DialogTitle className="text-4xl">{title}</DialogTitle>
      <p>{subtitle}</p>
    </div>
  );
}

function OAuthButtons() {
  const intlayerContent = useIntlayer("auth-modal");
  const shared = intlayerContent.shared.entry;

  return (
    <div className="flex flex-col gap-2">
      <button
        className="flex items-center justify-center gap-2 w-full
   h-12 bg-surface-inverse hover:bg-surface-inverse-raised
  text-text-inverse rounded-lg font-medium transition-colors"
      >
        <FaGoogle className="size-5" />
        {shared.google}
      </button>
      <button
        className="flex items-center justify-center gap-2 w-full
   h-12 bg-surface hover:bg-surface-raised text-text-primary rounded-lg
  font-medium transition-colors border border-border"
      >
        <FaApple className="size-5" />
        {shared.apple}
      </button>
    </div>
  );
}

function CloseButton() {
  const intlayerContent = useIntlayer("auth-modal");
  const shared = intlayerContent.shared.entry;
  const { closeModal } = useAuthModal();

  return (
    <div className="text-center">
      <button
        onClick={closeModal}
        className="opacity-75 hover:opacity-100 transition-opacity"
      >
        {shared.close}
      </button>
    </div>
  );
}
