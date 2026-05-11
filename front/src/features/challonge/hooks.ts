import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { challongeApi, type CreateChallongePayload } from "./api";

const keys = {
  bracket: (slug: string) => ["challonge", slug] as const,
  standings: (slug: string) => ["challonge", slug, "standings"] as const,
};

export function useChallongeBracket(slug?: string) {
  return useQuery({
    queryKey: slug ? keys.bracket(slug) : ["challonge", "empty"],
    queryFn: () => challongeApi.getBracket(slug!),
    enabled: Boolean(slug),
    refetchInterval: 5000,
  });
}

export function useChallongeStandings(slug?: string) {
  return useQuery({
    queryKey: slug ? keys.standings(slug) : ["challonge", "standings", "empty"],
    queryFn: () => challongeApi.getStandings(slug!),
    enabled: Boolean(slug),
  });
}

export function useCreateChallonge() {
  return useMutation({
    mutationFn: (payload: CreateChallongePayload) => challongeApi.create(payload),
  });
}

export function useAddParticipant(slug: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (name: string) => challongeApi.addParticipant(slug, name),
    onSuccess: () => qc.invalidateQueries({ queryKey: keys.bracket(slug) }),
  });
}

export function useBulkAdd(slug: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (names: string[]) => challongeApi.bulkAdd(slug, names),
    onSuccess: () => qc.invalidateQueries({ queryKey: keys.bracket(slug) }),
  });
}

export function useRemoveParticipant(slug: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (participantID: string) => challongeApi.removeParticipant(slug, participantID),
    onSuccess: () => qc.invalidateQueries({ queryKey: keys.bracket(slug) }),
  });
}

export function useShuffle(slug: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => challongeApi.shuffle(slug),
    onSuccess: () => qc.invalidateQueries({ queryKey: keys.bracket(slug) }),
  });
}

export function useStartTournament(slug: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => challongeApi.start(slug),
    onSuccess: () => qc.invalidateQueries({ queryKey: keys.bracket(slug) }),
  });
}

export function useResetTournament(slug: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => challongeApi.reset(slug),
    onSuccess: () => qc.invalidateQueries({ queryKey: keys.bracket(slug) }),
  });
}

export function useSubmitResult(slug: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      matchID,
      winnerID,
      score1,
      score2,
    }: {
      matchID: string;
      winnerID: string;
      score1?: number;
      score2?: number;
    }) => challongeApi.submitResult(slug, matchID, winnerID, score1, score2),
    onSuccess: () => qc.invalidateQueries({ queryKey: keys.bracket(slug) }),
  });
}

export function useResetMatch(slug: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (matchID: string) => challongeApi.resetMatch(slug, matchID),
    onSuccess: () => qc.invalidateQueries({ queryKey: keys.bracket(slug) }),
  });
}
