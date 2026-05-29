import { Link } from "react-router-dom";
import { useLang } from "@/app/providers/lang-provider";
import { Button } from "@/shared/ui/button";
import { Card, CardContent } from "@/shared/ui/card";

export function NotFoundPage() {
  const { t } = useLang();

  return (
    <div className="flex min-h-screen items-center justify-center bg-[#1a1a1a] p-6">
      <Card className="w-full max-w-lg">
        <CardContent className="grid gap-4 py-8">
          <h1 className="text-2xl font-semibold">{t("notFound.title")}</h1>
          <p className="text-sm text-[#9e9e9e]">{t("notFound.desc")}</p>
          <Link to="/">
            <Button>{t("notFound.home")}</Button>
          </Link>
        </CardContent>
      </Card>
    </div>
  );
}
