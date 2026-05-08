import { Card, CardContent } from "@/shared/ui/card";

export function ErrorState({
  title = "Не удалось загрузить данные",
  description = "Попробуйте обновить страницу или повторить действие позже.",
}: {
  title?: string;
  description?: string;
}) {
  return (
    <Card className="border-red-200">
      <CardContent className="py-8">
        <h3 className="text-base font-semibold text-red-700">{title}</h3>
        <p className="mt-2 text-sm text-red-600">{description}</p>
      </CardContent>
    </Card>
  );
}