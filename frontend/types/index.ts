export interface Folder {
  id: number;
  name: string;
  parent_id: number | null;
  created_at: string;
  updated_at: string;
}

export interface Request {
  id?: number;
  name: string;
  folder_id: number | null;
  method: string;
  url: string;
  headers: Record<string, string>;
  body: string;
  created_at?: string;
  updated_at?: string;
}

export interface Response {
  status_code: number;
  headers: Record<string, string>;
  body: string;
  duration: number;
}
