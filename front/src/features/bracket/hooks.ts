import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { bracketApi } from "@/features/bracket/api";
import { queryKeys } from "@/shared/lib/query-keys";

export function useTournamentBracket(tournamentId?: string) {
  return useQuery({
    queryKey: tournamentId ? queryKeys.tournamentBracket(tournamentId) : ["bracket", "empty"],
    queryFn: () => bracketApi.getByTournamentId(tournamentId!),
    enabled: Boolean(tournamentId),
  });
}

export function useGenerateBracket(tournamentId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => bracketApi.generate(tournamentId),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.tournamentBracket(tournamentId) });
      await queryClient.invalidateQueries({ queryKey: queryKeys.tournamentMatches(tournamentId) });
    },
  });
}

export function useRegenerateBracket(tournamentId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => bracketApi.regenerate(tournamentId),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.tournamentBracket(tournamentId) });
      await queryClient.invalidateQueries({ queryKey: queryKeys.tournamentMatches(tournamentId) });
    },
  });
}

export function useReseedBracket(tournamentId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (payload: { ordered_team_ids: string[] }) => bracketApi.reseed(tournamentId, payload),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.tournamentBracket(tournamentId) });
    },
  });
}

export function useTournamentPlacements(tournamentId?: string) {
  return useQuery({
    queryKey: tournamentId ? queryKeys.tournamentPlacements(tournamentId) : ["placements", "empty"],
    queryFn: () => bracketApi.getPlacements(tournamentId!),
    enabled: Boolean(tournamentId),
  });
}

export function useAdvanceToPlayoff(tournamentId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => bracketApi.advanceToPlayoff(tournamentId),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.tournamentBracket(tournamentId) });
      await queryClient.invalidateQueries({ queryKey: queryKeys.tournamentMatches(tournamentId) });
    },
  });
}