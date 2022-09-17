import { MenuItem, Select, SelectChangeEvent, useTheme } from "@mui/material";
import { useCallback } from "react";
import { networkOptions, useNetworkContext } from "../contexts/NetworkContext";

function NetworkSelector() {
  const theme = useTheme();
  const { currentNetwork, setCurrentNetwork } = useNetworkContext();
  const handleChange = useCallback(
    (e: SelectChangeEvent) => {
      setCurrentNetwork(networkOptions[Number(e.target.value)]);
    },
    [setCurrentNetwork]
  );
  return (
    <Select
      onChange={handleChange}
      value={(networkOptions.indexOf(currentNetwork) || 0).toString()}
      margin="dense"
      size="small"
      sx={{
        minWidth: 130,
        // theme fixes
        "& img": { filter: "invert(0)!important" },
        "& .MuiOutlinedInput-notchedOutline": {
          borderColor:
            theme.palette.mode === "light" ? "rgba(255,255,255,.6)" : null,
        },
        "&:hover .MuiOutlinedInput-notchedOutline": {
          borderColor:
            theme.palette.mode === "light" ? "rgba(255,255,255,.8)" : null,
        },
        "&.Mui-focused .MuiOutlinedInput-notchedOutline": {
          borderColor:
            theme.palette.mode === "light" ? "rgba(255,255,255,.8)" : null,
        },
        "& .MuiSvgIcon-root": {
          fill: theme.palette.mode === "light" ? "white" : null,
        },
      }}
      SelectDisplayProps={{
        style: { paddingTop: 4, paddingBottom: 4 },
      }}
    >
      {networkOptions.map((network, idx) => (
        <MenuItem key={network.endpoint} value={idx}>
          {network.logo !== "" ? (
            <img
              src={network.logo}
              alt={network.name}
              style={{
                height: 20,
                maxHeight: 20,
                verticalAlign: "middle",
                // theme fixes
                ...(theme.palette.mode === "light"
                  ? { filter: "invert(1)" }
                  : {}),
              }}
            />
          ) : (
            network.name
          )}
        </MenuItem>
      ))}
    </Select>
  );
}
export default NetworkSelector;
