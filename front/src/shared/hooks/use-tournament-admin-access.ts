import { useMemo } from "react";
import { useAuth } from "@/app/providers/auth-provider";
import { useTournamentAdminTeams } from "@/features/teams/hooks";
import type { Tournament } from "@/shared/types/api";
import { isPlatformAdmin } from "@/shared/lib/roles";

export function useTournamentAdminAccess(tournamentId?: string, tournament?: Tournament | null) {
  const { user } = useAuth();
  const isOwner = Boolean(user?.id && tournament?.owner_user_id && user.id === tournament.owner_user_id);
  const isAdmin = isPlatformAdmin(user);

  const probe = useTournamentAdminTeams(tournamentId, Boolean(tournamentId) && !isOwner && !isAdmin && Boolean(user));
  const canAccessAdmin = useMemo(() => {
    if (!user) return false;
    if (isOwner || isAdmin) return true;
    if (probe.isSuccess) return true;
    return false;
  }, [isAdmin, isOwner, probe.isSuccess, user]);

  return {
    isOwner,
    isAdmin,
    canAccessAdmin,
    isLoading: probe.isLoading,
    probeError: probe.error,
  };
}