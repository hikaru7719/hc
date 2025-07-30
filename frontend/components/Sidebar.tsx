import type { Request } from "@/types";

interface SidebarProps {
  requests: Request[];
  selectedRequest: Request | null;
  onSelectRequest: (request: Request) => void;
  onDeleteRequest: (id: number) => void;
  onRefresh: () => void;
}

export default function Sidebar({
  requests,
  selectedRequest,
  onSelectRequest,
  onDeleteRequest,
  onRefresh,
}: SidebarProps) {
  return (
    <div className="w-64 bg-base-100 border-r border-base-300 flex flex-col">
      <div className="p-4 border-b border-base-300">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold">Requests</h2>
          <button type="button" onClick={onRefresh} className="btn btn-sm btn-ghost" title="Refresh">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              className="h-4 w-4"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <title>Refresh</title>
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
              />
            </svg>
          </button>
        </div>
      </div>
      <div className="flex-1 overflow-y-auto">
        <ul className="menu p-2">
          {requests.map((request) => (
            <li key={request.id}>
              <div
                className={`flex items-center justify-between ${selectedRequest?.id === request.id ? "active" : ""}`}
              >
                <button type="button" onClick={() => onSelectRequest(request)} className="flex-1 text-left">
                  <span className="badge badge-sm mr-2">{request.method}</span>
                  <span className="text-sm">{request.name}</span>
                </button>
                {request.id && (
                  <button
                    type="button"
                    onClick={(e) => {
                      e.stopPropagation();
                      if (request.id) onDeleteRequest(request.id);
                    }}
                    className="btn btn-ghost btn-xs"
                  >
                    <svg
                      xmlns="http://www.w3.org/2000/svg"
                      className="h-4 w-4"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                    >
                      <title>Delete</title>
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                      />
                    </svg>
                  </button>
                )}
              </div>
            </li>
          ))}
        </ul>
      </div>
    </div>
  );
}
