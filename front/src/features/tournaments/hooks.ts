import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { tournamentsApi } from "@/features/tournaments/api";
import { queryKeys } from "@/shared/lib/query-keys";
import type {
  ManagerFormValues,
  TournamentFormValues,
  TournamentStatusFormValues,
} from "@/features/tournaments/schemas";

export function useTournaments() {
  return useQuery({
    queryKey: queryKeys.tournaments,
    queryFn: tournamentsApi.list,
  });
}

export function useTournament(id?: string) {
  return useQuery({
    queryKey: id ? queryKeys.tournament(id) : ["tournaments", "empty"],
    queryFn: () => tournamentsApi.getById(id!),
    enabled: Boolean(id),
  });
}

export function useCreateTournament() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: TournamentFormValues) => tournamentsApi.create(payload),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.tournaments });
    },
  });
}

export function useUpdateTournament(id: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: TournamentFormValues) => tournamentsApi.update(id, payload),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.tournament(id) });
      await queryClient.invalidateQueries({ queryKey: queryKeys.tournaments });
    },
  });
}

export function useDeleteTournament(id: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => tournamentsApi.remove(id),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.tournaments });
    },
  });
}

export function useChangeTournamentStatus(id: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: TournamentStatusFormValues) =>
      tournamentsApi.changeStatus(id, payload),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.tournament(id) });
      await queryClient.invalidateQueries({ queryKey: queryKeys.tournaments });
    },
  });
}

export function useAddManager(id: string) {
  return useMutation({
    mutationFn: (payload: ManagerFormValues) => tournamentsApi.addManager(id, payload),
  });
}

export function useRemoveManager(id: string) {
  return useMutation({
    mutationFn: (userId: string) => tournamentsApi.removeManager(id, userId),
  });
}

export function useTournamentAudit(id?: string, enabled = true) {
  return useQuery({
    queryKey: id ? queryKeys.tournamentAudit(id) : ["audit", "empty"],
    queryFn: () => tournamentsApi.getAudit(id!),
    enabled: Boolean(id) && enabled,
  });
}