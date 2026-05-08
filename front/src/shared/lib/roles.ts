import type { User } from "@/shared/types/api";

export function isPlatformAdmin(user: User | null | undefined) {
  if (!user) return false;
  return user.role === "platform_admin" || user.is_platform_admin === true;
}