import { t, type Dictionary } from "intlayer";

interface Content {
  fromMessageAttempt: {
    entry: {
      title: string;
      subtext: string;
      google: string;
      apple: string;
      inputLabels: {
        email: string;
      };
      continue: string;
      close: string;
    };
    signup: {
      title: string;
      subtext: string;
      google: string;
      apple: string;
      inputLabels: {
        email: string;
        name: string;
        password: string;
      };
      signup: string;
      close: string;
    };
    login: {
      title: string;
      subtext: string;
      google: string;
      apple: string;
      inputLabels: {
        email: string;
        password: string;
      };
      forgotPassword: string;
      login: string;
      close: string;
    };
  };
}

export default {
  key: "auth-modal",
  content: {
    fromMessageAttempt: {
      entry: {
        title: t({
          en: "Log in to talk to Shaikh",
          ar: "سجّل الدخول للتحدث مع شيخ",
        }),
        subtext: t({
          en: "Get answers grounded in centuries of Islamic scholarship. Explore Tafsir, linguistics, rulings, and more—in your language.",
          ar: "احصل على إجابات مبنية على قرون من التراث العلمي الإسلامي. استكشف التفسير، والبلاغة، والأحكام، والمزيد — بلغتك.",
        }),
        google: t({
          en: "Continue with Google",
          ar: "تابع باستخدام Google",
        }),
        apple: t({
          en: "Continue with Apple",
          ar: "تابع باستخدام Apple",
        }),
        inputLabels: {
          email: t({
            en: "Email",
            ar: "البريد الإلكتروني",
          }),
        },
        continue: t({
          en: "Continue",
          ar: "تابع",
        }),
        close: t({
          en: "Close",
          ar: "أغلق",
        }),
      },

      signup: {
        title: t({
          en: "Sign up to talk to Shaikh",
          ar: "أنشئ حساب للتحدث مع شيخ",
        }),
        subtext: t({
          en: "Get answers grounded in centuries of Islamic scholarship. Explore Tafsir, linguistics, rulings, and more—in your language.",
          ar: "احصل على إجابات مبنية على قرون من التراث العلمي الإسلامي. استكشف التفسير، والبلاغة، والأحكام، والمزيد — بلغتك.",
        }),
        google: t({
          en: "Continue with Google",
          ar: "تابع باستخدام Google",
        }),
        apple: t({
          en: "Continue with Apple",
          ar: "تابع باستخدام Apple",
        }),
        inputLabels: {
          email: t({
            en: "Email",
            ar: "البريد الإلكتروني",
          }),
          name: t({
            en: "Enter your name",
            ar: "أدخل اسمك",
          }),
          password: t({
            en: "Set a password",
            ar: "أنشئ كلمة مرور",
          }),
        },
        signup: t({
          en: "Sign up",
          ar: "أنشئ حساب",
        }),
        close: t({
          en: "Close",
          ar: "أغلق",
        }),
      },

      login: {
        title: t({
          en: "Log in to talk to Shaikh",
          ar: "سجّل الدخول للتحدث مع شيخ",
        }),
        subtext: t({
          en: "Get answers grounded in centuries of Islamic scholarship. Explore Tafsir, linguistics, rulings, and more—in your language.",
          ar: "احصل على إجابات مبنية على قرون من التراث العلمي الإسلامي. استكشف التفسير، والبلاغة، والأحكام، والمزيد — بلغتك.",
        }),
        google: t({
          en: "Continue with Google",
          ar: "تابع باستخدام Google",
        }),
        apple: t({
          en: "Continue with Apple",
          ar: "تابع باستخدام Apple",
        }),
        inputLabels: {
          email: t({
            en: "Email",
            ar: "البريد الإلكتروني",
          }),
          password: t({
            en: "Password",
            ar: "كلمة المرور",
          }),
        },
        forgotPassword: t({
          en: "Forgot your password?",
          ar: "نسيت كلمة المرور؟",
        }),
        login: t({
          en: "Log in",
          ar: "سجّل الدخول",
        }),
        close: t({
          en: "Close",
          ar: "أغلق",
        }),
      },
    },
  },
} satisfies Dictionary<Content>;
