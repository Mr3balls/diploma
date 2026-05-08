import * as React from "react";
import { Card, CardContent } from "@/shared/ui/card";

export function EmptyState({
  title,
  description,
  action,
}: {
  title: string;
  description: string;
  action?: React.ReactNode;
}) {
  return (
    <Card>
      <CardContent className="flex flex-col items-start gap-3 py-8">
        <h3 className="text-base font-semibold">{title}</h3>
        <p className="max-w-2xl text-sm text-slate-500">{description}</p>
        {action}
      </CardContent>
    </Card>
  );
}