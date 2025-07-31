import { useState } from "react";
import { useRequestPanelReducer } from "@/hooks/useRequestPanelReducer";
import HeadersEditor from "@/components/HeadersEditor";
import type { Request } from "@/types";
import { generateCurlCommand } from "@/utils/curlGenerator";
import { HTTP_METHOD_LIST, COPY_FEEDBACK_DURATION } from "@/constants/http";

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
    getRequestObject,
  } = useRequestPanelReducer(request);

  const [copiedCurl, setCopiedCurl] = useState(false);

  const handleSend = () => {
    onSend(getRequestObject(request));
  };

  const handleSave = () => {
    onSave(getRequestObject(request));
  };

  const handleCopyAsCurl = async () => {
    const requestObj = getRequestObject(request);
    const curlCommand = generateCurlCommand(requestObj);

    await navigator.clipboard.writeText(curlCommand);
    setCopiedCurl(true);
    setTimeout(() => setCopiedCurl(false), COPY_FEEDBACK_DURATION);
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
            {HTTP_METHOD_LIST.map((method) => (
              <option key={method} value={method}>
                {method}
              </option>
            ))}
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
            <HeadersEditor
              headers={state.headers}
              updateHeader={updateHeader}
              removeHeader={removeHeader}
              addHeader={addHeader}
            />
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
