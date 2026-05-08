import * as React from "react";
import type { Notification } from "@/shared/types/api";
import { notificationTypeLabel } from "@/shared/lib/enums";
import { formatDateTime } from "@/shared/lib/date";
import { Badge } from "@/shared/ui/badge";
import { Button } from "@/shared/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/shared/ui/card";

export function NotificationsList({
  items,
  onMarkRead,
  renderActions,
}: {
  items: Notification[];
  onMarkRead: (id: string) => void;
  renderActions?: (notification: Notification) => React.ReactNode;
}) {
  return (
    <div className="grid gap-4">
      {items.map((notification) => (
        <Card key={notification.id} className={notification.is_read ? "opacity-80" : ""}>
          <CardHeader className="gap-3">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <div className="space-y-1">
                <div className="flex flex-wrap items-center gap-2">
                  <CardTitle className="text-base">{notification.title || notificationTypeLabel[notification.type]}</CardTitle>
                  {!notification.is_read ? <Badge>Новое</Badge> : null}
                </div>
                <p className="text-xs text-[#90afd4]">{formatDateTime(notification.created_at)}</p>
              </div>
              <div className="flex gap-2">
                {!notification.is_read ? (
                  <Button variant="outline" size="sm" onClick={() => onMarkRead(notification.id)}>
                    Прочитано
                  </Button>
                ) : null}
              </div>
            </div>
          </CardHeader>
          <CardContent className="grid gap-3">
            <p className="text-sm text-[#90afd4]">{notification.message}</p>
            {renderActions ? <div className="flex flex-wrap gap-2">{renderActions(notification)}</div> : null}
          </CardContent>
        </Card>
      ))}
    </div>
  );
}