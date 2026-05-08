import type { AxiosError } from "axios";
import type { ApiErrorResponse } from "@/shared/types/api";

export function getErrorMessage(error: unknown) {
  const axiosError = error as AxiosError<ApiErrorResponse>;
  return (
    axiosError.response?.data?.error?.message ||
    axiosError.message ||
    "Произошла ошибка"
  );
}