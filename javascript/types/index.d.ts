declare module 'labnocturne' {
  export interface UploadResponse {
    id: string;
    url: string;
    size: number;
    mime_type: string;
    created_at: string;
  }

  export interface FileInfo {
    id: string;
    url: string;
    size: number;
    mime_type: string;
    created_at: string;
  }

  export interface ListFilesResponse {
    files: FileInfo[];
    pagination: {
      page: number;
      limit: number;
      total: number;
      total_pages: number;
    };
  }

  export interface StatsResponse {
    storage_used_bytes: number;
    storage_used_mb: number;
    file_count: number;
    quota_bytes: number;
    quota_mb: number;
    usage_percent: number;
  }

  export interface ListFilesOptions {
    page?: number;
    limit?: number;
    sort?: string;
  }

  export default class LabNocturneClient {
    constructor(apiKey: string, baseUrl?: string);

    upload(filePath: string): Promise<UploadResponse>;
    listFiles(options?: ListFilesOptions): Promise<ListFilesResponse>;
    getStats(): Promise<StatsResponse>;
    deleteFile(imageId: string): Promise<void>;

    static generateTestKey(baseUrl?: string): Promise<string>;
  }

  export { LabNocturneClient };
}
