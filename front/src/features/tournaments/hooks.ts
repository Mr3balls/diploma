import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { tournamentsApi } from "@/features/tournaments/api";
import { queryKeys } from "@/shared/lib/query-keys";
import type {
  ManagerFormValues,
  TournamentFormValues,
  TournamentStatusFormValues,
} from "@/features/tournaments/schemas";

export type TournamentListParams = {
  limit?: number;
  offset?: number;
  status?: string;
  format?: string;
  discipline?: string;
  q?: string;
};

export function useTournaments(params?: TournamentListParams) {
  return useQuery({
    queryKey: [...queryKeys.tournaments, params],
    queryFn: () => tournamentsApi.list(params),
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

export function useTournamentParticipants(id?: string) {
  return useQuery({
    queryKey: id ? ["tournaments", id, "participants"] : ["tournaments", "empty", "participants"],
    queryFn: () => tournamentsApi.getParticipants(id!),
    enabled: Boolean(id),
    refetchInterval: 5000,
  });
}

export function useAddTournamentParticipant(id: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (name: string) => tournamentsApi.addParticipant(id, name),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["tournaments", id, "participants"] });
      await queryClient.invalidateQueries({ queryKey: queryKeys.tournament(id) });
    },
  });
}

export function useBulkAddTournamentParticipants(id: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (names: string[]) => tournamentsApi.bulkAddParticipants(id, names),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["tournaments", id, "participants"] });
    },
  });
}

export function useRemoveTournamentParticipant(id: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (participantId: string) => tournamentsApi.removeParticipant(id, participantId),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["tournaments", id, "participants"] });
    },
  });
}

export function useShuffleTournamentParticipants(id: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => tournamentsApi.shuffleParticipants(id),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["tournaments", id, "participants"] });
    },
  });
}

export function useStartTournamentBracket(id: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => tournamentsApi.startBracket(id),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.tournament(id) });
      await queryClient.invalidateQueries({ queryKey: ["tournaments", id, "participants"] });
    },
  });
}

export function useJoinIndividualTournament(id: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => tournamentsApi.joinIndividual(id),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["tournaments", id, "participants"] });
    },
  });
}

export function useRegisterTeam(id: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (payload: { team_name: string; members: string[] }) =>
      tournamentsApi.registerTeam(id, payload),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.tournament(id) });
      await queryClient.invalidateQueries({ queryKey: ["tournaments", id, "teams"] });
    },
  });
}