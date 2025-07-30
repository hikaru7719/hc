export { requestsApi } from "./requests";
export { proxyApi } from "./proxy";
export { API_ENDPOINTS } from "./constants";

// SWR fetcher function
export const fetcher = async (url: string) => {
  const res = await fetch(url);
  if (!res.ok) {
    throw new Error("Failed to fetch");
  }
  return res.json();
};
