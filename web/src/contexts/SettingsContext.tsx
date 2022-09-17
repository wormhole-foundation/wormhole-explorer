import React, {
  ReactNode,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";

const STORAGE_KEY = "settings";

export type Theme = "light" | "dark" | "auto";

type Settings = {
  backgroundUrl?: string;
  backgroundOpacity?: number;
  defaultEndpoint?: string;
  theme: Theme;
};

type SettingsContextValue = {
  settings: Settings;
  updateBackgroundOpacity(value: number): void;
  updateBackgroundUrl(value: string): void;
  updateDefaultEndpoint(value: string): void;
  updateTheme(value: Theme): void;
};

const isTheme = (arg: any): arg is Theme => {
  return arg && (arg === "light" || arg === "dark" || arg === "auto");
};

const isSettings = (arg: any): arg is Settings => {
  return arg && arg.theme && isTheme(arg.theme);
};

let localStorageSettings: Settings | null = null;
try {
  const value = localStorage.getItem(STORAGE_KEY);
  if (value) {
    const parsedValue = JSON.parse(value);
    if (isSettings(parsedValue)) {
      localStorageSettings = parsedValue;
    }
  }
} catch (e) {}

const initialSettings: Settings = localStorageSettings || { theme: "auto" };

const saveSettings = (settings: Settings) => {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(settings));
  } catch (e) {}
};

const SettingsContext = React.createContext<SettingsContextValue>({
  settings: initialSettings,
  updateBackgroundOpacity: (value: number) => {},
  updateBackgroundUrl: (value: string) => {},
  updateDefaultEndpoint: (value: string) => {},
  updateTheme: (value: Theme) => {},
});

export const SettingsContextProvider = ({
  children,
}: {
  children: ReactNode;
}) => {
  const [settings, setSettings] = useState<Settings>(initialSettings);
  const updateBackgroundOpacity = useCallback((value: number) => {
    setSettings((settings) => ({ ...settings, backgroundOpacity: value }));
  }, []);
  const updateBackgroundUrl = useCallback((value: string) => {
    setSettings((settings) => ({ ...settings, backgroundUrl: value }));
  }, []);
  const updateDefaultEndpoint = useCallback((value: string) => {
    setSettings((settings) => ({ ...settings, defaultEndpoint: value }));
  }, []);
  const updateTheme = useCallback((value: Theme) => {
    setSettings((settings) => ({ ...settings, theme: value }));
  }, []);
  // sync settings to state
  useEffect(() => {
    saveSettings(settings);
  }, [settings]);
  const value = useMemo(
    () => ({
      settings,
      updateBackgroundOpacity,
      updateBackgroundUrl,
      updateDefaultEndpoint,
      updateTheme,
    }),
    [
      settings,
      updateBackgroundOpacity,
      updateBackgroundUrl,
      updateDefaultEndpoint,
      updateTheme,
    ]
  );
  return (
    <SettingsContext.Provider value={value}>
      {children}
    </SettingsContext.Provider>
  );
};

export const useSettingsContext = () => {
  return useContext(SettingsContext);
};
