package ami

import (
	"context"
	"strings"
)

// Queues shows queues list
func Queues(ctx context.Context, client Client, actionID string) ([]Response, error) {
	b, err := command("Queues", actionID, nil)
	if err != nil {
		return nil, err
	}
	if err := client.Send(string(b)); err != nil {
		return nil, err
	}

	data, err := readRaw(ctx, client)
	if err != nil {
		return nil, err
	}

	blocks := strings.Split(data.String(), "\r\n\r\n")
	response := make([]Response, 0)

	for _, block := range blocks {
		lines := strings.Split(block, "\n")
		if len(lines) < 1 {
			continue
		}

		headline := strings.TrimSpace(lines[0])
		if len(headline) < 1 {
			continue
		}

		queue := Response{
			"Header": []string{headline},
			"Queue":  []string{strings.SplitN(headline, " ", 2)[0]},
		}

		section := ""
		for _, line := range lines[1:] {
			line = strings.TrimSpace(line)
			if len(line) < 1 {
				continue
			}

			if line == "Members:" {
				section = "Members"
				continue
			} else if line == "Callers:" {
				section = "Callers"
				continue
			} else if line == "No Members" || line == "No Callers" {
				section = ""
				continue
			}

			if _, ok := queue[section]; !ok {
				queue[section] = make([]string, 0)
			}
			queue[section] = append(queue[section], line)
		}

		response = append(response, queue)
	}

	return response, nil
}

// QueueAdd adds interface to queue.
func QueueAdd(ctx context.Context, client Client, actionID string, queueData QueueData) (Response, error) {
	return send(ctx, client, "QueueAdd", actionID, queueData)
}

// QueueLog adds custom entry in queue_log.
func QueueLog(ctx context.Context, client Client, actionID string, queueData QueueData) (Response, error) {
	return send(ctx, client, "QueueLog", actionID, queueData)
}

// QueuePause makes a queue member temporarily unavailable.
func QueuePause(ctx context.Context, client Client, actionID string, queueData QueueData) (Response, error) {
	return send(ctx, client, "QueuePause", actionID, queueData)
}

// QueuePenalty sets the penalty for a queue member.
func QueuePenalty(ctx context.Context, client Client, actionID string, queueData QueueData) (Response, error) {
	return send(ctx, client, "QueuePenalty", actionID, queueData)
}

// QueueReload reloads a queue, queues, or any sub-section of a queue or queues.
func QueueReload(ctx context.Context, client Client, actionID string, queueData QueueData) (Response, error) {
	return send(ctx, client, "QueueReload", actionID, queueData)
}

// QueueRemove removes interface from queue.
func QueueRemove(ctx context.Context, client Client, actionID string, queueData QueueData) (Response, error) {
	return send(ctx, client, "QueueRemove", actionID, queueData)
}

// QueueReset resets queue statistics.
func QueueReset(ctx context.Context, client Client, actionID, queue string) (Response, error) {
	return send(ctx, client, "QueueReset", actionID, QueueData{Queue: queue})
}

// QueueRule queues Rules.
func QueueRule(ctx context.Context, client Client, actionID, rule string) (Response, error) {
	return send(ctx, client, "QueueRule", actionID, map[string]string{
		"Rule": rule,
	})
}

// QueueStatus show queue status by member.
func QueueStatus(ctx context.Context, client Client, actionID, queue, member string) (Response, error) {
	return send(ctx, client, "QueueStatus", actionID, map[string]string{
		"Queue":  queue,
		"Member": member,
	})
}

// QueueStatuses show status all members in queue.
func QueueStatuses(ctx context.Context, client Client, actionID, queue string) ([]Response, error) {
	return requestMultiEvent(ctx, client, "QueueStatus", actionID, []string{"QueueMember", "QueueEntry"}, "QueueStatusComplete", map[string]string{
		"Queue": queue,
	})
}

// QueueSummary show queue summary.
func QueueSummary(ctx context.Context, client Client, actionID, queue string) ([]Response, error) {
	return requestList(ctx, client, "QueueSummary", actionID, "QueueSummary", "QueueSummaryComplete", map[string]string{
		"Queue": queue,
	})
}

// QueueMemberRingInUse set the ringinuse value for a queue member.
func QueueMemberRingInUse(ctx context.Context, client Client, actionID, iface, ringInUse, queue string) (Response, error) {
	return send(ctx, client, "QueueMemberRingInUse", actionID, map[string]string{
		"Interface": iface,
		"RingInUse": ringInUse,
		"Queue":     queue,
	})
}
