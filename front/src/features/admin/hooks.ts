import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { adminApi } from "@/features/admin/api";
import { queryKeys } from "@/shared/lib/query-keys";

export function useAdminUsers(enabled = true) {
  return useQuery({
    queryKey: queryKeys.adminUsers,
    queryFn: adminApi.getUsers,
    enabled,
  });
}

export function useAdminTournaments(enabled = true) {
  return useQuery({
    queryKey: queryKeys.adminTournaments,
    queryFn: adminApi.getTournaments,
    enabled,
  });
}

export function useBlockUser() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => adminApi.blockUser(id),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.adminUsers });
    },
  });
}

export function useUnblockUser() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => adminApi.unblockUser(id),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.adminUsers });
    },
  });
}