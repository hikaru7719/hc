import type { Request, Response } from "@/types";
import { DEFAULT_METHOD, DEFAULT_REQUEST_NAME } from "./http";

export const DEFAULT_REQUEST: Request = {
  name: DEFAULT_REQUEST_NAME,
  folder_id: null,
  method: DEFAULT_METHOD,
  url: "",
  headers: {},
  body: "",
};

export const DEFAULT_RESPONSE: Response = {
  status_code: 0,
  headers: {},
  body: "",
  duration: 0,
};
