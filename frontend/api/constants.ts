// API endpoints
export const API_ENDPOINTS = {
  // Base URL
  BASE: "/api",

  // Request endpoints
  REQUESTS: "/api/requests",
  REQUEST_BY_ID: (id: number) => `/api/requests/${id}`,

  // Proxy endpoint
  PROXY: "/api/request",

  // Folder endpoints (for future use)
  FOLDERS: "/api/folders",
  FOLDER_BY_ID: (id: number) => `/api/folders/${id}`,
} as const;
