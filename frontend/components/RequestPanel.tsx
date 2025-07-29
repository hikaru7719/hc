import { useState, useEffect } from 'react';
import { Request } from '@/types';

interface RequestPanelProps {
  request: Request | null;
  onSend: (request: Request) => void;
  onSave: (request: Request) => void;
  loading: boolean;
}

export default function RequestPanel({
  request,
  onSend,
  onSave,
  loading,
}: RequestPanelProps) {
  const [name, setName] = useState('');
  const [method, setMethod] = useState('GET');
  const [url, setUrl] = useState('');
  const [headers, setHeaders] = useState<Array<{ key: string; value: string }>>([
    { key: '', value: '' },
  ]);
  const [body, setBody] = useState('');
  const [activeTab, setActiveTab] = useState<'headers' | 'body'>('headers');

  useEffect(() => {
    if (request) {
      setName(request.name);
      setMethod(request.method);
      setUrl(request.url);
      setHeaders(
        Object.entries(request.headers).length > 0
          ? Object.entries(request.headers).map(([key, value]) => ({ key, value }))
          : [{ key: '', value: '' }]
      );
      setBody(request.body);
    } else {
      setName('New Request');
      setMethod('GET');
      setUrl('');
      setHeaders([{ key: '', value: '' }]);
      setBody('');
    }
  }, [request]);

  const handleSend = () => {
    const headersObj = headers.reduce((acc, { key, value }) => {
      if (key) acc[key] = value;
      return acc;
    }, {} as Record<string, string>);

    onSend({
      id: request?.id,
      name,
      folder_id: request?.folder_id || null,
      method,
      url,
      headers: headersObj,
      body,
    });
  };

  const handleSave = () => {
    const headersObj = headers.reduce((acc, { key, value }) => {
      if (key) acc[key] = value;
      return acc;
    }, {} as Record<string, string>);

    onSave({
      id: request?.id,
      name,
      folder_id: request?.folder_id || null,
      method,
      url,
      headers: headersObj,
      body,
    });
  };

  const addHeader = () => {
    setHeaders([...headers, { key: '', value: '' }]);
  };

  const updateHeader = (index: number, field: 'key' | 'value', value: string) => {
    const newHeaders = [...headers];
    newHeaders[index][field] = value;
    setHeaders(newHeaders);
  };

  const removeHeader = (index: number) => {
    setHeaders(headers.filter((_, i) => i !== index));
  };

  return (
    <div className="h-full flex flex-col">
      <div className="p-4 border-b border-base-300">
        <div className="flex gap-2 mb-3">
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="input input-bordered input-sm flex-1"
            placeholder="Request name"
          />
          <button onClick={handleSave} className="btn btn-secondary btn-sm">
            Save
          </button>
          <button
            onClick={handleSend}
            disabled={loading || !url}
            className="btn btn-primary btn-sm"
          >
            {loading ? (
              <span className="loading loading-spinner loading-xs"></span>
            ) : (
              'Send'
            )}
          </button>
        </div>
        <div className="flex gap-2">
          <select
            value={method}
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
            value={url}
            onChange={(e) => setUrl(e.target.value)}
            className="input input-bordered input-sm flex-1"
            placeholder="https://api.example.com/endpoint"
          />
        </div>
      </div>

      <div className="flex-1 flex flex-col">
        <div className="tabs tabs-boxed p-4">
          <button
            className={`tab ${activeTab === 'headers' ? 'tab-active' : ''}`}
            onClick={() => setActiveTab('headers')}
          >
            Headers
          </button>
          <button
            className={`tab ${activeTab === 'body' ? 'tab-active' : ''}`}
            onClick={() => setActiveTab('body')}
          >
            Body
          </button>
        </div>

        <div className="flex-1 p-4 overflow-y-auto">
          {activeTab === 'headers' ? (
            <div>
              {headers.map((header, index) => (
                <div key={index} className="flex gap-2 mb-2">
                  <input
                    type="text"
                    value={header.key}
                    onChange={(e) => updateHeader(index, 'key', e.target.value)}
                    className="input input-bordered input-sm flex-1"
                    placeholder="Header name"
                  />
                  <input
                    type="text"
                    value={header.value}
                    onChange={(e) => updateHeader(index, 'value', e.target.value)}
                    className="input input-bordered input-sm flex-1"
                    placeholder="Header value"
                  />
                  <button
                    onClick={() => removeHeader(index)}
                    className="btn btn-ghost btn-sm"
                  >
                    <svg
                      xmlns="http://www.w3.org/2000/svg"
                      className="h-4 w-4"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M6 18L18 6M6 6l12 12"
                      />
                    </svg>
                  </button>
                </div>
              ))}
              <button onClick={addHeader} className="btn btn-ghost btn-sm">
                + Add Header
              </button>
            </div>
          ) : (
            <textarea
              value={body}
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