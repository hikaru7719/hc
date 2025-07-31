export const HTTP_METHODS = {
  GET: "GET",
  POST: "POST",
  PUT: "PUT",
  DELETE: "DELETE",
  PATCH: "PATCH",
  HEAD: "HEAD",
  OPTIONS: "OPTIONS",
} as const;

export const HTTP_METHOD_LIST = Object.values(HTTP_METHODS);

export const DEFAULT_METHOD = HTTP_METHODS.GET;

export const METHODS_WITH_BODY = [HTTP_METHODS.POST, HTTP_METHODS.PUT, HTTP_METHODS.PATCH] as const;

export const DEFAULT_REQUEST_NAME = "New Request";

export const COPY_FEEDBACK_DURATION = 2000;
