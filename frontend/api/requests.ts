import type { Request } from "@/types";
import { API_ENDPOINTS } from "./constants";

export const requestsApi = {
  // Get all requests
  async getAll(): Promise<Request[]> {
    const res = await fetch(API_ENDPOINTS.REQUESTS);
    if (!res.ok) {
      throw new Error("Failed to fetch requests");
    }
    return res.json();
  },

  // Create a new request
  async create(request: Request): Promise<Request> {
    const res = await fetch(API_ENDPOINTS.REQUESTS, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(request),
    });
    if (!res.ok) {
      throw new Error("Failed to create request");
    }
    return res.json();
  },

  // Update an existing request
  async update(id: number, request: Request): Promise<Request> {
    const res = await fetch(API_ENDPOINTS.REQUEST_BY_ID(id), {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(request),
    });
    if (!res.ok) {
      throw new Error("Failed to update request");
    }
    return res.json();
  },

  // Save request (create or update)
  async save(request: Request): Promise<Request> {
    if (request.id) {
      return this.update(request.id, request);
    }
    return this.create(request);
  },

  // Delete a request
  async delete(id: number): Promise<void> {
    const res = await fetch(API_ENDPOINTS.REQUEST_BY_ID(id), {
      method: "DELETE",
    });
    if (!res.ok) {
      throw new Error("Failed to delete request");
    }
  },
};