import { useState } from 'react';
import { Response } from '@/types';

interface ResponsePanelProps {
  response: Response | null;
  loading: boolean;
}

export default function ResponsePanel({ response, loading }: ResponsePanelProps) {
  const [activeTab, setActiveTab] = useState<'body' | 'headers'>('body');

  const formatBody = (body: string) => {
    try {
      const parsed = JSON.parse(body);
      return JSON.stringify(parsed, null, 2);
    } catch {
      return body;
    }
  };

  const getStatusBadgeClass = (statusCode: number) => {
    if (statusCode >= 200 && statusCode < 300) return 'badge-success';
    if (statusCode >= 300 && statusCode < 400) return 'badge-info';
    if (statusCode >= 400 && statusCode < 500) return 'badge-warning';
    if (statusCode >= 500) return 'badge-error';
    return 'badge-ghost';
  };

  return (
    <div className="h-full flex flex-col">
      <div className="p-4 border-b border-base-300">
        <h3 className="text-lg font-semibold mb-2">Response</h3>
        {response && (
          <div className="flex items-center gap-4">
            <span className={`badge ${getStatusBadgeClass(response.status_code)}`}>
              {response.status_code}
            </span>
            <span className="text-sm text-base-content/70">
              {response.duration}ms
            </span>
          </div>
        )}
      </div>

      {loading ? (
        <div className="flex-1 flex items-center justify-center">
          <span className="loading loading-spinner loading-lg"></span>
        </div>
      ) : response ? (
        <>
          <div className="tabs tabs-boxed p-4">
            <button
              className={`tab ${activeTab === 'body' ? 'tab-active' : ''}`}
              onClick={() => setActiveTab('body')}
            >
              Body
            </button>
            <button
              className={`tab ${activeTab === 'headers' ? 'tab-active' : ''}`}
              onClick={() => setActiveTab('headers')}
            >
              Headers
            </button>
          </div>

          <div className="flex-1 p-4 overflow-hidden">
            {activeTab === 'body' ? (
              <div className="h-full overflow-auto">
                <pre className="text-sm bg-base-300 p-4 rounded-lg min-w-0">
                  <code className="block whitespace-pre-wrap break-words">{formatBody(response.body)}</code>
                </pre>
              </div>
            ) : (
              <table className="table table-sm">
                <thead>
                  <tr>
                    <th>Header</th>
                    <th>Value</th>
                  </tr>
                </thead>
                <tbody>
                  {Object.entries(response.headers).map(([key, value]) => (
                    <tr key={key}>
                      <td className="font-medium">{key}</td>
                      <td className="text-sm">{value}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        </>
      ) : (
        <div className="flex-1 flex items-center justify-center text-base-content/50">
          <p>Send a request to see the response</p>
        </div>
      )}
    </div>
  );
}