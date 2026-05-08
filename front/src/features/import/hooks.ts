import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { importApi } from "@/features/import/api";
import { queryKeys } from "@/shared/lib/query-keys";
import type { GoogleSheetFormValues } from "@/features/import/schemas";

export function useTournamentImports(tournamentId?: string, enabled = true) {
  return useQuery({
    queryKey: tournamentId ? queryKeys.tournamentImports(tournamentId) : ["imports", "empty"],
    queryFn: () => importApi.list(tournamentId!),
    enabled: Boolean(tournamentId) && enabled,
  });
}

export function useImportBatch(batchId?: string, enabled = true) {
  return useQuery({
    queryKey: batchId ? queryKeys.importBatch(batchId) : ["import-batch", "empty"],
    queryFn: () => importApi.getBatch(batchId!),
    enabled: Boolean(batchId) && enabled,
  });
}

export function useConnectGoogleSheet(tournamentId: string) {
  return useMutation({
    mutationFn: (payload: GoogleSheetFormValues) => importApi.connect(tournamentId, payload),
  });
}

export function useValidateGoogleSheet(tournamentId: string) {
  return useMutation({
    mutationFn: (payload: GoogleSheetFormValues) => importApi.validate(tournamentId, payload),
  });
}

export function usePreviewImport(tournamentId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (payload: GoogleSheetFormValues) => importApi.preview(tournamentId, payload),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.tournamentImports(tournamentId) });
    },
  });
}

export function useConfirmImport(tournamentId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (payload: { batch_id: string }) => importApi.confirm(tournamentId, payload),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.tournamentImports(tournamentId) });
      await queryClient.invalidateQueries({ queryKey: queryKeys.tournamentAdminTeams(tournamentId) });
      await queryClient.invalidateQueries({ queryKey: queryKeys.tournamentTeams(tournamentId) });
    },
  });
}