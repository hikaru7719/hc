import type { Request, Response } from "@/types";
import { API_ENDPOINTS } from "./constants";

export interface ProxyRequest {
  method: string;
  url: string;
  headers: Record<string, string>;
  body: string;
}

export const proxyApi = {
  // Send a request through the proxy
  async sendRequest(request: Request): Promise<Response> {
    const proxyRequest: ProxyRequest = {
      method: request.method,
      url: request.url,
      headers: request.headers,
      body: request.body,
    };

    const res = await fetch(API_ENDPOINTS.PROXY, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(proxyRequest),
    });

    if (!res.ok) {
      throw new Error("Failed to send request");
    }

    return res.json();
  },
};
