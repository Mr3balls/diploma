import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { profileApi } from "@/features/profile/api";
import type { ProfileFormValues } from "@/features/profile/schemas";
import { queryKeys } from "@/shared/lib/query-keys";

export function useMe(enabled = true) {
  return useQuery({
    queryKey: queryKeys.me,
    queryFn: profileApi.getMe,
    enabled,
  });
}

export function useUpdateMe() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: ProfileFormValues) => profileApi.updateMe(payload),
    onSuccess: (user) => {
      queryClient.setQueryData(queryKeys.me, user);
    },
  });
}

export function useDeleteMe() {
  return useMutation({
    mutationFn: profileApi.deleteMe,
  });
}