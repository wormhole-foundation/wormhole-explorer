import {
  Box,
  createTheme,
  responsiveFontSizes,
  ThemeProvider,
} from "@mui/material";
import { grey } from "@mui/material/colors";
import { ReactNode, useCallback, useEffect, useMemo, useState } from "react";
import { useSettingsContext } from "../contexts/SettingsContext";

const mediaQueryList =
  window.matchMedia && window.matchMedia("(prefers-color-scheme: dark)");

function CustomThemeProvider({ children }: { children: ReactNode }) {
  const {
    settings: { theme: themePreference, backgroundOpacity, backgroundUrl },
  } = useSettingsContext();
  const [userPrefersDark, setUserPrefersDark] = useState<boolean>(
    mediaQueryList && mediaQueryList.matches ? true : false
  );
  const handleUserPreferenceChange = useCallback(
    (event: MediaQueryListEvent) => {
      setUserPrefersDark(event.matches ? true : false);
    },
    []
  );
  useEffect(() => {
    if (themePreference === "auto") {
      mediaQueryList.addEventListener("change", handleUserPreferenceChange);
      return () => {
        mediaQueryList.removeEventListener(
          "change",
          handleUserPreferenceChange
        );
      };
    }
  }, [themePreference, handleUserPreferenceChange]);
  const mode = "light";
  // themePreference === "dark" ||
  // (themePreference === "auto" && userPrefersDark)
  //   ? "dark"
  //   : "light";
  const theme = useMemo(
    () =>
      responsiveFontSizes(
        createTheme({
          palette: {
            mode,
          },
          components: {
            MuiCssBaseline: {
              styleOverrides: {
                body: {
                  overflowY: "scroll",
                },
                "*": {
                  scrollbarWidth: "thin",
                  scrollbarColor:
                    // mode === "dark"
                    // ? `${grey[700]} ${grey[900]}`
                    // :
                    `${grey[400]} rgb(255,255,255)`,
                },
                "*::-webkit-scrollbar": {
                  width: "8px",
                  height: "8px",
                  backgroundColor:
                    // mode === "dark" ? grey[900] :
                    "rgb(255,255,255)",
                },
                "*::-webkit-scrollbar-thumb": {
                  // mode === "dark" ? grey[700] :
                  backgroundColor: grey[400],
                  borderRadius: "4px",
                },
                "*::-webkit-scrollbar-corner": {
                  // this hides an annoying white box which appears when both scrollbars are present
                  backgroundColor: "transparent",
                },
              },
            },
          },
        })
      ),
    [mode]
  );
  return (
    <ThemeProvider theme={theme}>
      {children}
      {backgroundUrl && (
        <Box
          sx={{
            backgroundImage: `url(${backgroundUrl})`,
            backgroundPosition: "center",
            backgroundSize: "cover",
            opacity: backgroundOpacity || 0.1,
            position: "fixed",
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            zIndex: -1,
          }}
        />
      )}
    </ThemeProvider>
  );
}

export default CustomThemeProvider;
