import { Box } from "@mui/material";
import useChainHeartbeats from "../hooks/useChainHeartbeats";
import useHeartbeats from "../hooks/useHeartbeats";
import Chains from "./Chains";
import Guardians from "./Guardians";
import LatestVAAs from "./LatestVAAs";

function Main() {
  const heartbeats = useHeartbeats();
  const chainIdsToHeartbeats = useChainHeartbeats(heartbeats);
  return (
    <Box mt={2}>
      <Chains chainIdsToHeartbeats={chainIdsToHeartbeats} />
      <Guardians heartbeats={heartbeats} />
      <LatestVAAs />
    </Box>
  );
}
export default Main;
