import { createContext, ReactNode, useContext, useState } from "react";

type AuthModalState = {
  open: boolean;
  view: "entry" | "signup" | "login";
  variant: "messageAttempt" | "cta";
  email: string;
  pendingMessage?: string;
};

const AuthModalContext = createContext<
  AuthModalState & {
    openModal: (variant: AuthModalState["variant"], message?: string) => void;
    closeModal: () => void;
    setView: (view: AuthModalState["view"]) => void;
    setEmail: (email: string) => void;
  }
>(null!);

export function AuthModalProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<AuthModalState>({
    open: false,
    view: "entry",
    variant: "cta",
    email: "",
    pendingMessage: "",
  });

  const openModal = (variant: AuthModalState["variant"], message?: string) =>
    setState((s) => ({
      ...s,
      open: true,
      view: "entry",
      variant: variant,
      pendingMessage: message,
    }));

  const closeModal = () =>
    setState((s) => ({ ...s, open: false, pendingMessage: "" }));

  const setView = (view: AuthModalState["view"]) =>
    setState((s) => ({ ...s, view: view }));

  const setEmail = (email: string) => setState((s) => ({ ...s, email: email }));

  return (
    <AuthModalContext.Provider
      value={{
        ...state,
        openModal,
        closeModal,
        setView,
        setEmail,
      }}
    >
      {children}
    </AuthModalContext.Provider>
  );
}

export function useAuthModal() {
  const ctx = useContext(AuthModalContext);
  if (!ctx)
    throw new Error("useAuthModal must be used within AuthModalProvider");
  return ctx;
}
