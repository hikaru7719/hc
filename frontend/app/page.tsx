'use client';

import { useState, useEffect } from 'react';
import RequestPanel from '@/components/RequestPanel';
import ResponsePanel from '@/components/ResponsePanel';
import Sidebar from '@/components/Sidebar';
import { Request, Response, Folder } from '@/types';

export default function Home() {
  const [requests, setRequests] = useState<Request[]>([]);
  const [folders, setFolders] = useState<Folder[]>([]);
  const [selectedRequest, setSelectedRequest] = useState<Request | null>(null);
  const [response, setResponse] = useState<Response | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    fetchRequests();
    fetchFolders();
  }, []);

  const fetchRequests = async () => {
    try {
      const res = await fetch('http://localhost:8080/api/requests');
      const data = await res.json();
      setRequests(data);
    } catch (error) {
      console.error('Failed to fetch requests:', error);
    }
  };

  const fetchFolders = async () => {
    try {
      const res = await fetch('http://localhost:8080/api/folders');
      const data = await res.json();
      setFolders(data);
    } catch (error) {
      console.error('Failed to fetch folders:', error);
    }
  };

  const handleSendRequest = async (request: Request) => {
    setLoading(true);
    try {
      const res = await fetch('http://localhost:8080/api/request', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
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
      console.error('Failed to send request:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleSaveRequest = async (request: Request) => {
    try {
      if (request.id) {
        await fetch(`http://localhost:8080/api/requests/${request.id}`, {
          method: 'PUT',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(request),
        });
      } else {
        await fetch('http://localhost:8080/api/requests', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(request),
        });
      }
      await fetchRequests();
    } catch (error) {
      console.error('Failed to save request:', error);
    }
  };

  const handleDeleteRequest = async (id: number) => {
    try {
      await fetch(`http://localhost:8080/api/requests/${id}`, {
        method: 'DELETE',
      });
      await fetchRequests();
      if (selectedRequest?.id === id) {
        setSelectedRequest(null);
        setResponse(null);
      }
    } catch (error) {
      console.error('Failed to delete request:', error);
    }
  };

  return (
    <div className="flex h-screen bg-base-200">
      <Sidebar
        requests={requests}
        folders={folders}
        selectedRequest={selectedRequest}
        onSelectRequest={setSelectedRequest}
        onDeleteRequest={handleDeleteRequest}
        onRefresh={() => {
          fetchRequests();
          fetchFolders();
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
