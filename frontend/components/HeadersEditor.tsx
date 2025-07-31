interface Header {
  id: string;
  key: string;
  value: string;
}

interface HeadersEditorProps {
  headers: Header[];
  updateHeader: (index: number, field: "key" | "value", value: string) => void;
  removeHeader: (index: number) => void;
  addHeader: () => void;
}

export default function HeadersEditor({ headers, updateHeader, removeHeader, addHeader }: HeadersEditorProps) {
  return (
    <div>
      {headers.map((header, index) => (
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
  );
}
