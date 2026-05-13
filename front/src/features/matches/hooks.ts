import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { matchesApi } from "@/features/matches/api";
import { queryKeys } from "@/shared/lib/query-keys";
import type {
  RequestReasonValues,
  ScheduleMatchValues,
  SubmitResultValues,
} from "@/features/matches/schemas";

export function useTournamentMatches(tournamentId?: string) {
  return useQuery({
    queryKey: tournamentId ? queryKeys.tournamentMatches(tournamentId) : ["matches", "empty"],
    queryFn: () => matchesApi.getPublicByTournamentId(tournamentId!),
    enabled: Boolean(tournamentId),
  });
}

export function useTournamentAdminMatches(tournamentId?: string, enabled = true) {
  return useQuery({
    queryKey: tournamentId ? queryKeys.tournamentAdminMatches(tournamentId) : ["admin-matches", "empty"],
    queryFn: () => matchesApi.getAdminByTournamentId(tournamentId!),
    enabled: Boolean(tournamentId) && enabled,
    retry: false,
  });
}

export function useScheduleMatch(tournamentId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ matchId, payload }: { matchId: string; payload: ScheduleMatchValues }) =>
      matchesApi.schedule(matchId, payload),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.tournamentMatches(tournamentId) });
      await queryClient.invalidateQueries({ queryKey: queryKeys.tournamentAdminMatches(tournamentId) });
    },
  });
}

export function useConfirmReady(tournamentId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (matchId: string) => matchesApi.confirmReady(matchId),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.tournamentMatches(tournamentId) });
      await queryClient.invalidateQueries({ queryKey: queryKeys.tournamentAdminMatches(tournamentId) });
    },
  });
}

export function useRequestReschedule(tournamentId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ matchId, payload }: { matchId: string; payload: RequestReasonValues }) =>
      matchesApi.requestReschedule(matchId, payload),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.tournamentMatches(tournamentId) });
      await queryClient.invalidateQueries({ queryKey: queryKeys.tournamentAdminMatches(tournamentId) });
    },
  });
}

export function useReportIssue(tournamentId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ matchId, payload }: { matchId: string; payload: RequestReasonValues }) =>
      matchesApi.reportIssue(matchId, payload),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.tournamentMatches(tournamentId) });
      await queryClient.invalidateQueries({ queryKey: queryKeys.tournamentAdminMatches(tournamentId) });
    },
  });
}

export function useSubmitResult(tournamentId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ matchId, payload }: { matchId: string; payload: SubmitResultValues }) =>
      matchesApi.submitResult(matchId, payload),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.tournamentMatches(tournamentId) });
      await queryClient.invalidateQueries({ queryKey: queryKeys.tournamentAdminMatches(tournamentId) });
      await queryClient.invalidateQueries({ queryKey: queryKeys.tournamentBracket(tournamentId) });
    },
  });
}

export function useAdminSetResult(tournamentId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ matchId, payload }: { matchId: string; payload: { winner_team_id?: string; winner_participant_id?: string; score_text?: string } }) =>
      matchesApi.adminSetResult(matchId, payload),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.tournamentMatches(tournamentId) });
      await queryClient.invalidateQueries({ queryKey: queryKeys.tournamentAdminMatches(tournamentId) });
      await queryClient.invalidateQueries({ queryKey: queryKeys.tournamentBracket(tournamentId) });
    },
  });
}

export function useApproveResult(tournamentId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (matchId: string) => matchesApi.approveResult(matchId),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.tournamentMatches(tournamentId) });
      await queryClient.invalidateQueries({ queryKey: queryKeys.tournamentAdminMatches(tournamentId) });
      await queryClient.invalidateQueries({ queryKey: queryKeys.tournamentBracket(tournamentId) });
    },
  });
}

export function useRejectResult(tournamentId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ matchId, payload }: { matchId: string; payload?: RequestReasonValues }) =>
      matchesApi.rejectResult(matchId, payload),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.tournamentMatches(tournamentId) });
      await queryClient.invalidateQueries({ queryKey: queryKeys.tournamentAdminMatches(tournamentId) });
    },
  });
}