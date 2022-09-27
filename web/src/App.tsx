import { GitHub } from "@mui/icons-material";
import {
  AppBar,
  Box,
  Button,
  CssBaseline,
  IconButton,
  Toolbar,
  Typography,
} from "@mui/material";
import { ErrorBoundary } from "react-error-boundary";
import {
  HashRouter as Router,
  Link,
  Redirect,
  Route,
  Switch,
} from "react-router-dom";
import CustomThemeProvider from "./components/CustomThemeProvider";
import ErrorFallback from "./components/ErrorFallback";
import Governance from "./components/Governance";
import Main from "./components/Main";
import { NetworkContextProvider } from "./contexts/NetworkContext";
import { SettingsContextProvider } from "./contexts/SettingsContext";
import WormholeStatsIcon from "./icons/WormholeStatsIcon";

function App() {
  return (
    <ErrorBoundary FallbackComponent={ErrorFallback}>
      <SettingsContextProvider>
        <CustomThemeProvider>
          <CssBaseline />
          <Router>
            <NetworkContextProvider>
              <AppBar position="static" color="transparent" elevation={0}>
                <Toolbar variant="dense">
                  <Button
                    component={Link}
                    to="/"
                    color="inherit"
                    sx={{ textTransform: "none" }}
                    size="small"
                  >
                    <Box pr={1} display="flex" alignItems="center">
                      <WormholeStatsIcon />
                    </Box>
                    <Typography variant="h6">Wormscan</Typography>
                  </Button>
                  <Box px={2}>
                    <Button component={Link} to="/governance">
                      Governance
                    </Button>
                  </Box>
                  <Box flexGrow={1} />
                  <Box>
                    <IconButton
                      sx={{ ml: 1 }}
                      href="https://github.com/certusone/wormhole-explorer"
                      target="_blank"
                      rel="noopener noreferrer"
                      color="inherit"
                    >
                      <GitHub />
                    </IconButton>
                  </Box>
                </Toolbar>
              </AppBar>
              <Switch>
                <Route exact path="/governance">
                  <Governance />
                </Route>
                <Route exact path="/">
                  <Main />
                </Route>
                <Route>
                  <Redirect to="/" />
                </Route>
              </Switch>
            </NetworkContextProvider>
          </Router>
        </CustomThemeProvider>
      </SettingsContextProvider>
    </ErrorBoundary>
  );
}

export default App;
