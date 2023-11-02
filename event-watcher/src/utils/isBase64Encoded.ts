// This function uses a regex string to check if the input could
// possibly be base64 encoded.
//
// WARNING:  There are clear text strings that are NOT base64 encoded
//           that will pass this check.
export function isBase64Encoded(input: string): boolean {
  const b64Regex = new RegExp('^(?:[A-Za-z0-9+/]{4})*(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=)?$');
  const match = b64Regex.exec(input);
  if (match) {
    return true;
  }
  return false;
}
