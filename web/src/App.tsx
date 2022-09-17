import { GitHub } from "@mui/icons-material";
import {
  AppBar,
  Box,
  CssBaseline,
  IconButton,
  Toolbar,
  Typography,
} from "@mui/material";
import CustomThemeProvider from "./components/CustomThemeProvider";
import Main from "./components/Main";
import NetworkSelector from "./components/NetworkSelector";
import Settings from "./components/Settings";
import { NetworkContextProvider } from "./contexts/NetworkContext";
import { SettingsContextProvider } from "./contexts/SettingsContext";
import WormholeStatsIcon from "./icons/WormholeStatsIcon";

function App() {
  return (
    <SettingsContextProvider>
      <CustomThemeProvider>
        <CssBaseline />
        <NetworkContextProvider>
          <AppBar position="static">
            <Toolbar variant="dense">
              <Box pr={1} display="flex" alignItems="center">
                <WormholeStatsIcon />
              </Box>
              <Typography variant="h6">Explorer</Typography>
              <Box flexGrow={1} />
            </Toolbar>
          </AppBar>
          <Main />
        </NetworkContextProvider>
      </CustomThemeProvider>
    </SettingsContextProvider>
  );
}

export default App;
