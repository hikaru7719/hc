import { useEffect, useState } from "react";
import { useRequestPanelReducer } from "@/hooks/useRequestPanelReducer";
import type { Request } from "@/types";
import { generateCurlCommand } from "@/utils/curlGenerator";

interface RequestPanelProps {
  request: Request | null;
  onSend: (request: Request) => void;
  onSave: (request: Request) => void;
  loading: boolean;
}

export default function RequestPanel({ request, onSend, onSave, loading }: RequestPanelProps) {
  const {
    state,
    setName,
    setMethod,
    setUrl,
    setBody,
    setActiveTab,
    addHeader,
    updateHeader,
    removeHeader,
    resetFromRequest,
    getRequestObject,
  } = useRequestPanelReducer();

  const [copiedCurl, setCopiedCurl] = useState(false);

  useEffect(() => {
    resetFromRequest(request);
  }, [request, resetFromRequest]);

  const handleSend = () => {
    onSend(getRequestObject(request));
  };

  const handleSave = () => {
    onSave(getRequestObject(request));
  };

  const handleCopyAsCurl = async () => {
    const requestObj = getRequestObject(request);
    const curlCommand = generateCurlCommand(requestObj);

    try {
      await navigator.clipboard.writeText(curlCommand);
      setCopiedCurl(true);
      setTimeout(() => setCopiedCurl(false), 2000);
    } catch (err) {
      console.error("Failed to copy to clipboard:", err);
    }
  };

  return (
    <div className="h-full flex flex-col">
      <div className="p-4 border-b border-base-300">
        <div className="flex gap-2 mb-3">
          <input
            type="text"
            value={state.name}
            onChange={(e) => setName(e.target.value)}
            className="input input-bordered input-sm flex-1"
            placeholder="Request name"
          />
          <button type="button" onClick={handleSave} className="btn btn-secondary btn-sm">
            Save
          </button>
          <button
            type="button"
            onClick={handleSend}
            disabled={loading || !state.url}
            className="btn btn-primary btn-sm"
          >
            {loading ? <span className="loading loading-spinner loading-xs"></span> : "Send"}
          </button>
          <button type="button" onClick={handleCopyAsCurl} className="btn btn-ghost btn-sm" disabled={!state.url}>
            {copiedCurl ? "Copied!" : "Copy as cURL"}
          </button>
        </div>
        <div className="flex gap-2">
          <select
            value={state.method}
            onChange={(e) => setMethod(e.target.value)}
            className="select select-bordered select-sm w-32"
          >
            <option value="GET">GET</option>
            <option value="POST">POST</option>
            <option value="PUT">PUT</option>
            <option value="DELETE">DELETE</option>
            <option value="PATCH">PATCH</option>
            <option value="HEAD">HEAD</option>
            <option value="OPTIONS">OPTIONS</option>
          </select>
          <input
            type="text"
            value={state.url}
            onChange={(e) => setUrl(e.target.value)}
            className="input input-bordered input-sm flex-1"
            placeholder="https://api.example.com/endpoint"
          />
        </div>
      </div>

      <div className="flex-1 flex flex-col">
        <div className="tabs tabs-boxed p-4">
          <button
            type="button"
            className={`tab ${state.activeTab === "headers" ? "tab-active" : ""}`}
            onClick={() => setActiveTab("headers")}
          >
            Headers
          </button>
          <button
            type="button"
            className={`tab ${state.activeTab === "body" ? "tab-active" : ""}`}
            onClick={() => setActiveTab("body")}
          >
            Body
          </button>
        </div>

        <div className="flex-1 p-4 overflow-y-auto">
          {state.activeTab === "headers" ? (
            <div>
              {state.headers.map((header, index) => (
                <div key={header.id} className="flex gap-2 mb-2">
                  <input
                    type="text"
                    value={header.key}
                    onChange={(e) => updateHeader(index, "key", e.target.value)}
                    className="input input-bordered input-sm flex-1"
                    placeholder="Header name"
                  />
                  <input
                    type="text"
                    value={header.value}
                    onChange={(e) => updateHeader(index, "value", e.target.value)}
                    className="input input-bordered input-sm flex-1"
                    placeholder="Header value"
                  />
                  <button type="button" onClick={() => removeHeader(index)} className="btn btn-ghost btn-sm">
                    <svg
                      xmlns="http://www.w3.org/2000/svg"
                      className="h-4 w-4"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                    >
                      <title>Remove</title>
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                    </svg>
                  </button>
                </div>
              ))}
              <button type="button" onClick={addHeader} className="btn btn-ghost btn-sm">
                + Add Header
              </button>
            </div>
          ) : (
            <textarea
              value={state.body}
              onChange={(e) => setBody(e.target.value)}
              className="textarea textarea-bordered w-full h-full font-mono text-sm"
              placeholder="Request body (JSON, XML, etc.)"
            />
          )}
        </div>
      </div>
    </div>
  );
}
