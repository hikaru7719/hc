import { useReducer } from "react";
import type { Request, Response } from "@/types";

interface AppState {
  selectedRequest: Request | null;
  response: Response | null;
  loading: boolean;
}

type AppAction =
  | { type: "SET_SELECTED_REQUEST"; payload: Request | null }
  | { type: "SET_RESPONSE"; payload: Response | null }
  | { type: "SET_LOADING"; payload: boolean }
  | { type: "CLEAR_REQUEST_AND_RESPONSE" }
  | { type: "REQUEST_START" }
  | { type: "REQUEST_SUCCESS"; payload: Response }
  | { type: "REQUEST_FAILURE" };

const initialState: AppState = {
  selectedRequest: null,
  response: null,
  loading: false,
};

function appReducer(state: AppState, action: AppAction): AppState {
  switch (action.type) {
    case "SET_SELECTED_REQUEST":
      return { ...state, selectedRequest: action.payload };
    case "SET_RESPONSE":
      return { ...state, response: action.payload };
    case "SET_LOADING":
      return { ...state, loading: action.payload };
    case "CLEAR_REQUEST_AND_RESPONSE":
      return { ...state, selectedRequest: null, response: null };
    case "REQUEST_START":
      return { ...state, loading: true };
    case "REQUEST_SUCCESS":
      return { ...state, response: action.payload, loading: false };
    case "REQUEST_FAILURE":
      return { ...state, loading: false };
    default:
      return state;
  }
}

export function useAppReducer() {
  const [state, dispatch] = useReducer(appReducer, initialState);

  const setSelectedRequest = (request: Request | null) => {
    dispatch({ type: "SET_SELECTED_REQUEST", payload: request });
  };

  const setResponse = (response: Response | null) => {
    dispatch({ type: "SET_RESPONSE", payload: response });
  };

  const setLoading = (loading: boolean) => {
    dispatch({ type: "SET_LOADING", payload: loading });
  };

  const clearRequestAndResponse = () => {
    dispatch({ type: "CLEAR_REQUEST_AND_RESPONSE" });
  };

  const startRequest = () => {
    dispatch({ type: "REQUEST_START" });
  };

  const requestSuccess = (response: Response) => {
    dispatch({ type: "REQUEST_SUCCESS", payload: response });
  };

  const requestFailure = () => {
    dispatch({ type: "REQUEST_FAILURE" });
  };

  return {
    state,
    setSelectedRequest,
    setResponse,
    setLoading,
    clearRequestAndResponse,
    startRequest,
    requestSuccess,
    requestFailure,
  };
}