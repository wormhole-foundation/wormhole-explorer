import axios from "axios";
import { useEffect, useState } from "react";

// https://docs.github.com/en/rest/releases/releases#get-the-latest-release
function useLatestRelease(): string | null {
  const [latestRelease, setLatestRelease] = useState<string | null>(null);
  useEffect(() => {
    let cancelled = false;
    (async () => {
      while (!cancelled) {
        const response = await axios.get(
          "https://api.github.com/repos/wormhole-foundation/wormhole/releases/latest"
        );
        if (!cancelled) {
          setLatestRelease(response.data?.tag_name || null);
          await new Promise((resolve) => setTimeout(resolve, 60000));
        }
      }
    })();
    return () => {
      cancelled = true;
    };
  }, []);
  return latestRelease;
}
export default useLatestRelease;
