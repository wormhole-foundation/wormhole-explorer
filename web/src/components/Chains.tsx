import {
  Box,
  CircularProgress,
  Grid,
  Tooltip,
  Typography,
} from "@mui/material";
import { useMemo } from "react";
import {
  ChainIdToHeartbeats,
  HeartbeatInfo,
} from "../hooks/useChainHeartbeats";
import chainIdToIcon from "../utils/chainIdToIcon";
import chainIdToName from "../utils/chainIdToName";
import { EXPECTED_GUARDIAN_COUNT } from "../utils/consts";
import { BEHIND_DIFF } from "./Alerts";

function Chain({
  chainId,
  heartbeats,
}: {
  chainId: string;
  heartbeats: HeartbeatInfo[];
}) {
  const highest = useMemo(() => {
    let highest = BigInt(0);
    heartbeats.forEach((heartbeat) => {
      const height = BigInt(heartbeat.network.height);
      if (height > highest) {
        highest = height;
      }
    });
    return highest;
  }, [heartbeats]);
  const upCount = useMemo(
    () =>
      heartbeats.reduce(
        (total, heartbeat) =>
          total +
          (heartbeat.network.height === 0 ||
          highest - BigInt(heartbeat.network.height) > BEHIND_DIFF
            ? 0
            : 1),
        0
      ),
    [heartbeats, highest]
  );
  const percentUp = (upCount / EXPECTED_GUARDIAN_COUNT) * 100;
  const icon = chainIdToIcon(Number(chainId));
  const name = chainIdToName(Number(chainId));
  return (
    <Grid key={chainId} item xs={2}>
      <Box p={1} textAlign="center">
        <Tooltip title={name}>
          <Box sx={{ position: "relative", display: "inline-flex" }}>
            <CircularProgress
              variant="determinate"
              value={percentUp || 100}
              color={
                upCount > 15 ? "success" : upCount > 13 ? "warning" : "error"
              }
            />
            <Box
              sx={{
                top: 0,
                left: 0,
                bottom: 0,
                right: 0,
                position: "absolute",
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
              }}
            >
              {icon ? (
                <img
                  src={icon}
                  alt={name}
                  style={{ height: 22, maxWidth: 22 }}
                />
              ) : null}
            </Box>
          </Box>
        </Tooltip>
        <Box sx={{ mt: -0.5 }}>
          <Typography variant="caption">
            {upCount}/{EXPECTED_GUARDIAN_COUNT}
          </Typography>
        </Box>
      </Box>
    </Grid>
  );
}

function Chains({
  chainIdsToHeartbeats,
}: {
  chainIdsToHeartbeats: ChainIdToHeartbeats;
}) {
  return (
    <Box mt={1} sx={{ backgroundColor: "#ebeef9" }}>
      <Grid container spacing={1}>
        {Object.keys(chainIdsToHeartbeats).map((chainId) => (
          <Chain
            key={chainId}
            chainId={chainId}
            heartbeats={chainIdsToHeartbeats[Number(chainId)]}
          />
        ))}
      </Grid>
    </Box>
  );
}

export default Chains;
