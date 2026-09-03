package runner

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/murilopereirame/dyndns/pkg/ddns"
	"github.com/murilopereirame/dyndns/pkg/notification"
)

type Runner struct {
	Manager  ddns.DDNS
	Notifier notification.NotificationSender
	Interval int64
}

type UpdatePayload struct {
	Old  ddns.DDNSState
	New  ddns.DDNSState
	Time time.Time
}

func NewRunner(manager ddns.DDNS, notifier notification.NotificationSender, interval int64) *Runner {
	return &Runner{
		Manager:  manager,
		Notifier: notifier,
		Interval: interval,
	}
}

func (r *Runner) Run() {
	for true {
		notificationPayload := &UpdatePayload{
			Old: r.Manager.State,
		}

		slog.Info("Running entries check", "state", r.Manager.State)
		hasChanges, state := r.Manager.CheckEntries()

		if hasChanges && r.Manager.Config.Webhook.Enabled {
			slog.Info("State updated", "state", state)

			notificationPayload.New = state
			notificationPayload.Time = time.Now()

			notificationBody, err := json.Marshal(notificationPayload)

			if err == nil {
				sent, err := r.Notifier.Send(&notification.NotificationPayload{
					Title:    "IP Address updated",
					Body:     string(notificationBody),
					Priority: r.Manager.Config.Webhook.Priority,
				})

				if err != nil || !sent {
					slog.Warn("Update notification not sent", "error", err)
				}
			}
		}

		time.Sleep(time.Second * time.Duration(r.Interval))
	}
}
