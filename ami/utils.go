package ami

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
)

// GetUUID returns a new UUID based on /dev/urandom (unix).
func GetUUID() (string, error) {
	f, err := os.Open("/dev/urandom")
	if err != nil {
		return "", fmt.Errorf("open /dev/urandom error:[%v]", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("Error closing file: %s\n", err)
		}
	}()
	b := make([]byte, 16)

	_, err = f.Read(b)
	if err != nil {
		return "", err
	}
	uuid := fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	return uuid, nil
}

func command(action string, id string, v ...interface{}) ([]byte, error) {
	if action == "" {
		return nil, errors.New("invalid Action")
	}
	return marshal(&struct {
		Action string `ami:"Action"`
		ID     string `ami:"ActionID, omitempty"`
		V      []interface{}
	}{Action: action, ID: id, V: v})
}

func send(ctx context.Context, client Client, action, id string, v interface{}) (Response, error) {
	b, err := command(action, id, v)
	if err != nil {
		return nil, err
	}
	if err := client.Send(string(b)); err != nil {
		return nil, err
	}
	return read(ctx, client)
}

func readRaw(ctx context.Context, client Client) (*bytes.Buffer, error) {
	var buffer bytes.Buffer
	for {
		input, err := client.Recv(ctx)
		if err != nil {
			return nil, err
		}
		buffer.WriteString(input)
		if buffer.Len() > 4 && strings.HasSuffix(buffer.String(), "\r\n\r\n") {
			break
		}
	}

	return &buffer, nil
}

func read(ctx context.Context, client Client) (Response, error) {
	data, err := readRaw(ctx, client)
	if err != nil {
		return nil, err
	}

	return parseResponse(data.String())
}

func parseResponse(input string) (Response, error) {
	resp := make(Response)
	lines := strings.Split(input, "\r\n")
	for _, line := range lines {
		keys := strings.SplitAfterN(line, ":", 2)
		if len(keys) == 2 {
			key := strings.TrimSpace(strings.Trim(keys[0], ":"))
			value := strings.TrimSpace(keys[1])
			resp[key] = append(resp[key], value)
		} else if strings.Contains(line, "\r\n\r\n") || line == "" {
			return resp, nil
		}
	}
	return resp, nil
}

func requestList(ctx context.Context, client Client, action, id, event, complete string, v ...interface{}) ([]Response, error) {
	b, err := command(action, id, v)
	if err != nil {
		return nil, err
	}
	if err := client.Send(string(b)); err != nil {
		return nil, err
	}

	response := make([]Response, 0)
	for {
		rsp, err := read(ctx, client)
		if err != nil {
			return nil, err
		}
		if rsp.Get("EventList") == "start" {
			continue
		}

		e := rsp.Get("Event")
		r := rsp.Get("Response")
		if e == event {
			response = append(response, rsp)
		} else if e == complete || r != "" && r != "Success" {
			break
		}
	}
	return response, nil
}

// requestMultiEvent allows for a list of events to be specified, used in cases where a command
// returns multiple types of events.
func requestMultiEvent(ctx context.Context, client Client, action, id string, events []string, complete string, v ...interface{}) ([]Response, error) {

	set := make(map[string]struct{}, len(events))
	for _, evt := range events {
		set[evt] = struct{}{}
	}

	b, err := command(action, id, v)
	if err != nil {
		return nil, err
	}
	if err := client.Send(string(b)); err != nil {
		return nil, err
	}

	response := make([]Response, 0)
	for {
		rsp, err := read(ctx, client)
		if err != nil {
			return nil, err
		}

		if rsp.Get("EventList") == "start" {
			continue
		}

		e := rsp.Get("Event")
		r := rsp.Get("Response")

		_, ok := set[e]
		if ok {
			response = append(response, rsp)
		} else if e == complete || r != "" && r != "Success" {
			break
		} else {
			fmt.Printf("Unexpected event: %s %v\n", e, rsp)
		}
	}
	return response, nil
}
