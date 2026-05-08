import { apiClient } from "@/shared/api/client";
import type {
  GoogleSheetLink,
  ImportBatchDetailsResponse,
  ImportPreviewResponse,
  ListResponse,
  ValidateGoogleSheetResponse,
  ImportBatch,
} from "@/shared/types/api";
import type { GoogleSheetFormValues } from "@/features/import/schemas";

export const importApi = {
  async connect(tournamentId: string, payload: GoogleSheetFormValues) {
    const { data } = await apiClient.post<GoogleSheetLink>(`/tournaments/${tournamentId}/google-sheet/connect`, payload);
    return data;
  },
  async validate(tournamentId: string, payload: GoogleSheetFormValues) {
    const { data } = await apiClient.post<ValidateGoogleSheetResponse>(
      `/tournaments/${tournamentId}/google-sheet/validate`,
      payload,
    );
    return data;
  },
  async preview(tournamentId: string, payload: GoogleSheetFormValues) {
    const { data } = await apiClient.post<ImportPreviewResponse>(`/tournaments/${tournamentId}/imports/preview`, payload);
    return data;
  },
  async confirm(tournamentId: string, payload: { batch_id: string }) {
    const { data } = await apiClient.post<{ batch: ImportBatch }>(`/tournaments/${tournamentId}/imports/confirm`, payload);
    return data;
  },
  async list(tournamentId: string) {
    const { data } = await apiClient.get<ListResponse<ImportBatch>>(`/tournaments/${tournamentId}/imports`);
    return data;
  },
  async getBatch(batchId: string) {
    const { data } = await apiClient.get<ImportBatchDetailsResponse>(`/imports/${batchId}`);
    return data;
  },
};