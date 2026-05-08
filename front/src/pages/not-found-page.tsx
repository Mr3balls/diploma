import { Link } from "react-router-dom";
import { Button } from "@/shared/ui/button";
import { Card, CardContent } from "@/shared/ui/card";

export function NotFoundPage() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-slate-100 p-6">
      <Card className="w-full max-w-lg">
        <CardContent className="grid gap-4 py-8">
          <h1 className="text-2xl font-semibold">Страница не найдена</h1>
          <p className="text-sm text-slate-500">Проверьте адрес или вернитесь к списку турниров.</p>
          <Link to="/">
            <Button>На главную</Button>
          </Link>
        </CardContent>
      </Card>
    </div>
  );
}