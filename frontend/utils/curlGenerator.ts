import { HTTP_METHODS, METHODS_WITH_BODY } from "@/constants/http";
import type { Request } from "@/types";

export function generateCurlCommand(request: Request): string {
  const parts: string[] = ["curl"];

  // Method (default is GET, so only add if not GET)
  if (request.method && request.method !== HTTP_METHODS.GET) {
    parts.push(`-X ${request.method}`);
  }

  // Headers
  if (request.headers) {
    for (const [key, value] of Object.entries(request.headers)) {
      if (key && value) {
        // Escape quotes in header values
        const escapedValue = String(value).replace(/"/g, '\\"');
        parts.push(`-H "${key}: ${escapedValue}"`);
      }
    }
  }

  // Body
  if (request.body && request.method && (METHODS_WITH_BODY as readonly string[]).includes(request.method)) {
    // Escape quotes and newlines in body
    const escapedBody = request.body
      .replace(/\\/g, "\\\\")
      .replace(/"/g, '\\"')
      .replace(/\n/g, "\\n")
      .replace(/\r/g, "\\r")
      .replace(/\t/g, "\\t");
    parts.push(`-d "${escapedBody}"`);
  }

  // URL (should be last)
  if (request.url) {
    // Wrap URL in quotes if it contains special characters
    if (request.url.match(/[&?<>|]/)) {
      parts.push(`"${request.url}"`);
    } else {
      parts.push(request.url);
    }
  }

  return parts.join(" ");
}
