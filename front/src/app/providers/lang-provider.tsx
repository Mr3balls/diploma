import {
  createContext,
  PropsWithChildren,
  useCallback,
  useContext,
  useMemo,
  useState,
} from "react";
import { translations, type Lang } from "@/shared/lib/i18n";

const LANG_KEY = "ace:lang";

type LangContextValue = {
  lang: Lang;
  setLang: (lang: Lang) => void;
  t: (key: string, vars?: Record<string, string | number>) => string;
};

const LangContext = createContext<LangContextValue | undefined>(undefined);

function getInitialLang(): Lang {
  const stored = localStorage.getItem(LANG_KEY);
  if (stored === "ru" || stored === "en" || stored === "kk") return stored;
  return "ru";
}

export function LangProvider({ children }: PropsWithChildren) {
  const [lang, setLangState] = useState<Lang>(getInitialLang);

  const setLang = useCallback((l: Lang) => {
    localStorage.setItem(LANG_KEY, l);
    setLangState(l);
  }, []);

  const t = useCallback(
    (key: string, vars?: Record<string, string | number>): string => {
      const dict = translations[lang];
      let str = (dict as Record<string, string>)[key] ?? (translations.ru as Record<string, string>)[key] ?? key;
      if (vars) {
        for (const [k, v] of Object.entries(vars)) {
          str = str.replace(`{${k}}`, String(v));
        }
      }
      return str;
    },
    [lang],
  );

  const value = useMemo<LangContextValue>(
    () => ({ lang, setLang, t }),
    [lang, setLang, t],
  );

  return <LangContext.Provider value={value}>{children}</LangContext.Provider>;
}

export function useLang() {
  const value = useContext(LangContext);
  if (!value) throw new Error("useLang must be used within LangProvider");
  return value;
}
