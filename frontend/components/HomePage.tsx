import useSWR, { mutate } from "swr";
import { API_ENDPOINTS, fetcher, proxyApi, requestsApi } from "@/api";
import RequestPanel from "@/components/RequestPanel";
import ResponsePanel from "@/components/ResponsePanel";
import Sidebar from "@/components/Sidebar";
import { useAppReducer } from "@/hooks/useAppReducer";
import type { Request } from "@/types";

export default function HomePage() {
  const { data: requests = [], error: requestsError } = useSWR<Request[]>(API_ENDPOINTS.REQUESTS, fetcher);
  const { state, setSelectedRequest, clearRequestAndResponse, startRequest, requestSuccess } = useAppReducer();

  if (requestsError) {
    console.error("Failed to fetch requests:", requestsError);
  }

  const handleSendRequest = async (request: Request) => {
    startRequest();
    const response = await proxyApi.sendRequest(request);
    requestSuccess(response);
  };

  const handleSaveRequest = async (request: Request) => {
    await requestsApi.save(request);
    await mutate(API_ENDPOINTS.REQUESTS);
  };

  const handleDeleteRequest = async (id: number) => {
    await requestsApi.delete(id);
    await mutate(API_ENDPOINTS.REQUESTS);
    if (state.selectedRequest?.id === id) {
      clearRequestAndResponse();
    }
  };

  return (
    <div className="flex h-screen bg-base-200">
      <Sidebar
        requests={requests}
        selectedRequest={state.selectedRequest}
        onSelectRequest={setSelectedRequest}
        onDeleteRequest={handleDeleteRequest}
        onRefresh={() => {
          mutate(API_ENDPOINTS.REQUESTS);
        }}
      />
      <div className="flex-1 flex overflow-hidden">
        <div className="w-1/2 border-r border-base-300 overflow-hidden">
          <RequestPanel
            request={state.selectedRequest}
            onSend={handleSendRequest}
            onSave={handleSaveRequest}
            loading={state.loading}
          />
        </div>
        <div className="w-1/2 overflow-hidden">
          <ResponsePanel response={state.response} loading={state.loading} />
        </div>
      </div>
    </div>
  );
}
