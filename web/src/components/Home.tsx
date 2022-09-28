import { Box } from "@mui/material";
import useChainHeartbeats from "../hooks/useChainHeartbeats";
import useHeartbeats from "../hooks/useHeartbeats";
import Chains from "./Chains";
import LatestVAAs from "./LatestVAAs";

function Home() {
  const heartbeats = useHeartbeats();
  const chainIdsToHeartbeats = useChainHeartbeats(heartbeats);
  return (
    <Box>
      <Chains chainIdsToHeartbeats={chainIdsToHeartbeats} />
      <LatestVAAs />
    </Box>
  );
}
export default Home;
