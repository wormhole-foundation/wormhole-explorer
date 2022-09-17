import {
  ContrastOutlined,
  DarkModeOutlined,
  LightModeOutlined,
  SettingsOutlined,
} from "@mui/icons-material";
import {
  Box,
  Dialog,
  IconButton,
  Slider,
  TextField,
  ToggleButton,
  ToggleButtonGroup,
  Typography,
} from "@mui/material";
import { useCallback, useState } from "react";
import { Theme, useSettingsContext } from "../contexts/SettingsContext";

function SettingsContent() {
  const {
    settings,
    updateBackgroundOpacity,
    updateBackgroundUrl,
    updateTheme,
  } = useSettingsContext();
  const handleThemeChange = useCallback(
    (event: any, newTheme: Theme) => {
      updateTheme(newTheme);
    },
    [updateTheme]
  );
  const handleBackgroundOpacityChange = useCallback(
    (event: any) => {
      updateBackgroundOpacity(event.target.value);
    },
    [updateBackgroundOpacity]
  );
  const handleBackgroundUrlChange = useCallback(
    (event: any) => {
      updateBackgroundUrl(event.target.value);
    },
    [updateBackgroundUrl]
  );
  return (
    <>
      <Box mt={2} mx={2} textAlign="center">
        <ToggleButtonGroup
          value={settings.theme}
          exclusive
          onChange={handleThemeChange}
        >
          <ToggleButton value="light">
            <LightModeOutlined />
          </ToggleButton>
          <ToggleButton value="dark">
            <DarkModeOutlined />
          </ToggleButton>
          <ToggleButton value="auto">
            <ContrastOutlined />
          </ToggleButton>
        </ToggleButtonGroup>
      </Box>
      <Box m={2}>
        <TextField
          value={settings.backgroundUrl || ""}
          onChange={handleBackgroundUrlChange}
          label="Background URL"
          margin="dense"
          fullWidth
        />
      </Box>
      <Box m={2}>
        <Typography variant="body2">Background Opacity</Typography>
        <Box pr={2} pt={2}>
          <Slider
            min={0.05}
            max={1}
            step={0.05}
            value={settings.backgroundOpacity || 0.1}
            onChange={handleBackgroundOpacityChange}
          />
        </Box>
      </Box>
    </>
  );
}

function Settings() {
  const [open, setOpen] = useState(false);
  const handleOpen = useCallback(() => {
    setOpen(true);
  }, []);
  const handleClose = useCallback(() => {
    setOpen(false);
  }, []);
  return (
    <>
      <IconButton color="inherit" onClick={handleOpen}>
        <SettingsOutlined />
      </IconButton>
      <Dialog open={open} onClose={handleClose} maxWidth="xs" fullWidth>
        <SettingsContent />
      </Dialog>
    </>
  );
}

export default Settings;
