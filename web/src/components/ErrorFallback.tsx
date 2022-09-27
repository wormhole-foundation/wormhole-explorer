import { Typography } from "@mui/material";

function ErrorFallback({ error }: { error: { message: string } }) {
  return (
    <div role="alert">
      <Typography>Something went wrong:</Typography>
      <pre>{error.message}</pre>
    </div>
  );
}

export default ErrorFallback;
