import { useReducer } from "react";
import type { Request } from "@/types";

interface RequestPanelState {
  name: string;
  method: string;
  url: string;
  headers: Array<{ id: string; key: string; value: string }>;
  body: string;
  activeTab: "headers" | "body";
}

type RequestPanelAction =
  | { type: "SET_NAME"; payload: string }
  | { type: "SET_METHOD"; payload: string }
  | { type: "SET_URL"; payload: string }
  | { type: "SET_HEADERS"; payload: Array<{ id: string; key: string; value: string }> }
  | { type: "SET_BODY"; payload: string }
  | { type: "SET_ACTIVE_TAB"; payload: "headers" | "body" }
  | { type: "ADD_HEADER" }
  | { type: "UPDATE_HEADER"; payload: { index: number; field: "key" | "value"; value: string } }
  | { type: "REMOVE_HEADER"; payload: number }
  | { type: "RESET_FROM_REQUEST"; payload: Request | null };

const initialState: RequestPanelState = {
  name: "New Request",
  method: "GET",
  url: "",
  headers: [{ id: crypto.randomUUID(), key: "", value: "" }],
  body: "",
  activeTab: "headers",
};

function requestPanelReducer(state: RequestPanelState, action: RequestPanelAction): RequestPanelState {
  switch (action.type) {
    case "SET_NAME":
      return { ...state, name: action.payload };
    case "SET_METHOD":
      return { ...state, method: action.payload };
    case "SET_URL":
      return { ...state, url: action.payload };
    case "SET_HEADERS":
      return { ...state, headers: action.payload };
    case "SET_BODY":
      return { ...state, body: action.payload };
    case "SET_ACTIVE_TAB":
      return { ...state, activeTab: action.payload };
    case "ADD_HEADER":
      return {
        ...state,
        headers: [...state.headers, { id: crypto.randomUUID(), key: "", value: "" }],
      };
    case "UPDATE_HEADER": {
      const newHeaders = [...state.headers];
      newHeaders[action.payload.index][action.payload.field] = action.payload.value;
      return { ...state, headers: newHeaders };
    }
    case "REMOVE_HEADER":
      return {
        ...state,
        headers: state.headers.filter((_, i) => i !== action.payload),
      };
    case "RESET_FROM_REQUEST": {
      const request = action.payload;
      if (request) {
        return {
          name: request.name,
          method: request.method,
          url: request.url,
          headers:
            Object.entries(request.headers).length > 0
              ? Object.entries(request.headers).map(([key, value]) => ({
                  id: crypto.randomUUID(),
                  key,
                  value,
                }))
              : [{ id: crypto.randomUUID(), key: "", value: "" }],
          body: request.body,
          activeTab: state.activeTab,
        };
      }
      return {
        ...initialState,
        activeTab: state.activeTab,
      };
    }
    default:
      return state;
  }
}

export function useRequestPanelReducer() {
  const [state, dispatch] = useReducer(requestPanelReducer, initialState);

  const setName = (name: string) => dispatch({ type: "SET_NAME", payload: name });
  const setMethod = (method: string) => dispatch({ type: "SET_METHOD", payload: method });
  const setUrl = (url: string) => dispatch({ type: "SET_URL", payload: url });
  const setHeaders = (headers: Array<{ id: string; key: string; value: string }>) =>
    dispatch({ type: "SET_HEADERS", payload: headers });
  const setBody = (body: string) => dispatch({ type: "SET_BODY", payload: body });
  const setActiveTab = (tab: "headers" | "body") => dispatch({ type: "SET_ACTIVE_TAB", payload: tab });
  const addHeader = () => dispatch({ type: "ADD_HEADER" });
  const updateHeader = (index: number, field: "key" | "value", value: string) =>
    dispatch({ type: "UPDATE_HEADER", payload: { index, field, value } });
  const removeHeader = (index: number) => dispatch({ type: "REMOVE_HEADER", payload: index });
  const resetFromRequest = (request: Request | null) =>
    dispatch({ type: "RESET_FROM_REQUEST", payload: request });

  const getRequestObject = (request: Request | null) => {
    const headersObj = state.headers.reduce(
      (acc, { key, value }) => {
        if (key) acc[key] = value;
        return acc;
      },
      {} as Record<string, string>,
    );

    return {
      id: request?.id,
      name: state.name,
      folder_id: request?.folder_id || null,
      method: state.method,
      url: state.url,
      headers: headersObj,
      body: state.body,
    };
  };

  return {
    state,
    setName,
    setMethod,
    setUrl,
    setHeaders,
    setBody,
    setActiveTab,
    addHeader,
    updateHeader,
    removeHeader,
    resetFromRequest,
    getRequestObject,
  };
}