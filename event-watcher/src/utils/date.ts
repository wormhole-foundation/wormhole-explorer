export const checkIfDateIsInMilliseconds = (timestamp: string | number): boolean => {
  const date = +new Date(timestamp);

  if (Math.abs(Date.now() - date) < Math.abs(Date.now() - date * 1000)) {
    return true;
  } else {
    return false;
  }
};
