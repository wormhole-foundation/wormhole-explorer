import { createTheme, responsiveFontSizes, ThemeProvider } from "@mui/material";
import { grey } from "@mui/material/colors";
import { ReactNode, useMemo } from "react";

function CustomThemeProvider({ children }: { children: ReactNode }) {
  const mode = "light";
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
                  scrollbarColor: `${grey[400]} rgb(255,255,255)`,
                },
                "*::-webkit-scrollbar": {
                  width: "8px",
                  height: "8px",
                  backgroundColor: "rgb(255,255,255)",
                },
                "*::-webkit-scrollbar-thumb": {
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
  return <ThemeProvider theme={theme}>{children}</ThemeProvider>;
}

export default CustomThemeProvider;
