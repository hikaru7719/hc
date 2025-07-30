"use client";

import { useState } from "react";
import useSWR, { mutate } from "swr";
import RequestPanel from "@/components/RequestPanel";
import ResponsePanel from "@/components/ResponsePanel";
import Sidebar from "@/components/Sidebar";
import type { Request, Response } from "@/types";

const fetcher = async (url: string) => {
  const res = await fetch(url);
  if (!res.ok) {
    throw new Error("Failed to fetch");
  }
  return res.json();
};

export default function Home() {
  const { data: requests = [], error: requestsError } = useSWR<Request[]>("/api/requests", fetcher);
  const [selectedRequest, setSelectedRequest] = useState<Request | null>(null);
  const [response, setResponse] = useState<Response | null>(null);
  const [loading, setLoading] = useState(false);

  if (requestsError) {
    console.error("Failed to fetch requests:", requestsError);
  }

  const handleSendRequest = async (request: Request) => {
    setLoading(true);
    try {
      const res = await fetch("/api/request", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          method: request.method,
          url: request.url,
          headers: request.headers,
          body: request.body,
        }),
      });

      const data = await res.json();
      setResponse(data);
    } catch (error) {
      console.error("Failed to send request:", error);
    } finally {
      setLoading(false);
    }
  };

  const handleSaveRequest = async (request: Request) => {
    try {
      if (request.id) {
        await fetch(`/api/requests/${request.id}`, {
          method: "PUT",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify(request),
        });
      } else {
        await fetch("/api/requests", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify(request),
        });
      }
      await mutate("/api/requests");
    } catch (error) {
      console.error("Failed to save request:", error);
    }
  };

  const handleDeleteRequest = async (id: number) => {
    try {
      await fetch(`/api/requests/${id}`, {
        method: "DELETE",
      });
      await mutate("/api/requests");
      if (selectedRequest?.id === id) {
        setSelectedRequest(null);
        setResponse(null);
      }
    } catch (error) {
      console.error("Failed to delete request:", error);
    }
  };

  return (
    <div className="flex h-screen bg-base-200">
      <Sidebar
        requests={requests}
        selectedRequest={selectedRequest}
        onSelectRequest={setSelectedRequest}
        onDeleteRequest={handleDeleteRequest}
        onRefresh={() => {
          mutate("/api/requests");
        }}
      />
      <div className="flex-1 flex overflow-hidden">
        <div className="w-1/2 border-r border-base-300 overflow-hidden">
          <RequestPanel
            request={selectedRequest}
            onSend={handleSendRequest}
            onSave={handleSaveRequest}
            loading={loading}
          />
        </div>
        <div className="w-1/2 overflow-hidden">
          <ResponsePanel response={response} loading={loading} />
        </div>
      </div>
    </div>
  );
}
