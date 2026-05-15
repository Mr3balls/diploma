import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { teamsApi } from "@/features/teams/api";
import { queryKeys } from "@/shared/lib/query-keys";

export function useTournamentTeams(tournamentId?: string) {
  return useQuery({
    queryKey: tournamentId ? queryKeys.tournamentTeams(tournamentId) : ["teams", "empty"],
    queryFn: () => teamsApi.getTournamentTeams(tournamentId!),
    enabled: Boolean(tournamentId),
  });
}

export function useTournamentAdminTeams(tournamentId?: string, enabled = true) {
  return useQuery({
    queryKey: tournamentId ? queryKeys.tournamentAdminTeams(tournamentId) : ["admin-teams", "empty"],
    queryFn: () => teamsApi.getAdminTournamentTeams(tournamentId!),
    enabled: Boolean(tournamentId) && enabled,
    retry: false,
  });
}

export function useTeam(teamId?: string, enabled = true) {
  return useQuery({
    queryKey: teamId ? queryKeys.team(teamId) : ["team", "empty"],
    queryFn: () => teamsApi.getTeam(teamId!),
    enabled: Boolean(teamId) && enabled,
  });
}

export function useApproveTeam(tournamentId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (teamId: string) => teamsApi.approveTeam(teamId),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.tournamentAdminTeams(tournamentId) });
      await queryClient.invalidateQueries({ queryKey: queryKeys.tournamentTeams(tournamentId) });
    },
  });
}

export function useRejectTeam(tournamentId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ teamId, reason }: { teamId: string; reason: string }) => teamsApi.rejectTeam(teamId, reason),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.tournamentAdminTeams(tournamentId) });
      await queryClient.invalidateQueries({ queryKey: queryKeys.tournamentTeams(tournamentId) });
    },
  });
}

export function useRemoveMember(tournamentId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ teamId, memberId }: { teamId: string; memberId: string }) => teamsApi.removeMember(teamId, memberId),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.tournamentAdminTeams(tournamentId) });
    },
  });
}

export function useAdminDeleteTeam(tournamentId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (teamId: string) => teamsApi.adminDeleteTeam(tournamentId, teamId),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.tournamentAdminTeams(tournamentId) });
      await queryClient.invalidateQueries({ queryKey: queryKeys.tournamentTeams(tournamentId) });
    },
  });
}

export function useAdminCreateTeam(tournamentId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (payload: { team_name: string; members: string[] }) =>
      teamsApi.adminCreateTeam(tournamentId, payload),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.tournamentAdminTeams(tournamentId) });
      await queryClient.invalidateQueries({ queryKey: queryKeys.tournamentTeams(tournamentId) });
    },
  });
}

export function useMyTeam(tournamentId?: string, userId?: string) {
  return useQuery({
    queryKey: ["my-team", tournamentId ?? "empty", userId ?? "none"],
    queryFn: () => teamsApi.getMyTeam(tournamentId!),
    enabled: Boolean(tournamentId) && Boolean(userId),
    retry: false,
  });
}

export function useReplaceMember(tournamentId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ teamId, memberId, email }: { teamId: string; memberId: string; email: string }) =>
      teamsApi.replaceMember(teamId, memberId, { email }),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["my-team", tournamentId] });
      await queryClient.invalidateQueries({ queryKey: queryKeys.tournamentAdminTeams(tournamentId) });
    },
  });
}

export function useAcceptParticipation() {
  return useMutation({
    mutationFn: (id: string) => teamsApi.acceptParticipation(id),
  });
}

export function useDeclineParticipation() {
  return useMutation({
    mutationFn: (id: string) => teamsApi.declineParticipation(id),
  });
}