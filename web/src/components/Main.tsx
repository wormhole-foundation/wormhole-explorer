import useHeartbeats from "../hooks/useHeartbeats";
import Guardians from "./Guardians";
import LatestVAAs from "./LatestVAAs";

function Main() {
  const heartbeats = useHeartbeats();
  return (
    <>
      <Guardians heartbeats={heartbeats} />
      <LatestVAAs />
    </>
  );
}
export default Main;
