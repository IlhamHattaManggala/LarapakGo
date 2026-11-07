package rikatlive

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"reflect"
	"strconv"
	"strings"
)

// RequestPayload represents the incoming AJAX or WebSocket action payload from RikatLive client.
type RequestPayload struct {
	Component string         `json:"component"`
	ID        string         `json:"id"`
	Snapshot  map[string]any `json:"snapshot"`
	Checksum  string         `json:"checksum"`
	Action    string         `json:"action,omitempty"`
	Args      []any          `json:"args,omitempty"`
	Updates   map[string]any `json:"updates,omitempty"`
}

// ResponsePayload represents the rendered response returned to RikatLive client for DOM morphing.
type ResponsePayload struct {
	ID       string         `json:"id"`
	HTML     string         `json:"html"`
	Snapshot map[string]any `json:"snapshot"`
	Checksum string         `json:"checksum"`
	Error    string         `json:"error,omitempty"`
}

func generateID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return "r-" + hex.EncodeToString(b)
}

// ExtractState scans the component's struct fields or dynamic state into a map.
func ExtractState(comp Component) map[string]any {
	if dyn, ok := comp.(DynamicComponent); ok {
		return dyn.GetState()
	}

	val := reflect.ValueOf(comp)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}

	state := make(map[string]any)
	if val.Kind() != reflect.Struct {
		return state
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		// Only consider exported fields
		if !field.IsExported() {
			continue
		}
		if tag := field.Tag.Get("rikat"); tag == "-" {
			continue
		}
		state[field.Name] = val.Field(i).Interface()
	}

	return state
}

// RestoreState populates exported struct fields or dynamic state from a state map.
func RestoreState(comp Component, state map[string]any) error {
	if dyn, ok := comp.(DynamicComponent); ok {
		return dyn.SetState(state)
	}

	val := reflect.ValueOf(comp)
	if val.Kind() != reflect.Pointer || val.Elem().Kind() != reflect.Struct {
		return errors.New("rikatlive: component must be a pointer to a struct")
	}
	val = val.Elem()

	for k, v := range state {
		fieldVal := val.FieldByName(k)
		if !fieldVal.IsValid() || !fieldVal.CanSet() {
			// Try case-insensitive lookup
			for i := 0; i < val.NumField(); i++ {
				f := val.Type().Field(i)
				if strings.EqualFold(f.Name, k) && f.IsExported() {
					fieldVal = val.Field(i)
					break
				}
			}
		}

		if !fieldVal.IsValid() || !fieldVal.CanSet() {
			continue
		}

		setFieldValue(fieldVal, v)
	}

	return nil
}

func setFieldValue(field reflect.Value, rawVal any) {
	if rawVal == nil {
		return
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(fmt.Sprintf("%v", rawVal))

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch v := rawVal.(type) {
		case float64:
			field.SetInt(int64(v))
		case string:
			if n, err := strconv.ParseInt(v, 10, 64); err == nil {
				field.SetInt(n)
			}
		case int:
			field.SetInt(int64(v))
		case int64:
			field.SetInt(v)
		}

	case reflect.Float32, reflect.Float64:
		switch v := rawVal.(type) {
		case float64:
			field.SetFloat(v)
		case string:
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				field.SetFloat(f)
			}
		}

	case reflect.Bool:
		switch v := rawVal.(type) {
		case bool:
			field.SetBool(v)
		case string:
			field.SetBool(strings.ToLower(v) == "true" || v == "1" || v == "bener")
		}
	}
}

// RenderInitial renders a component on the server for the first page load.
func RenderInitial(name string, initialProps ...map[string]any) (string, error) {
	comp, err := Create(name)
	if err != nil {
		return "", err
	}

	id := generateID()

	if len(initialProps) > 0 && initialProps[0] != nil {
		if mounter, ok := comp.(Mounter); ok {
			_ = mounter.Mount(initialProps[0])
		} else {
			_ = RestoreState(comp, initialProps[0])
		}
	}

	state := ExtractState(comp)
	snapshotJSON, _ := json.Marshal(state)
	checksum := SignSnapshot(snapshotJSON)

	innerHTML := comp.Render()

	// Safe escaping for snapshot JSON inside HTML attribute
	escapedSnapshot := html.EscapeString(string(snapshotJSON))

	root := fmt.Sprintf(`<div r-component="%s" r-id="%s" r-snapshot="%s" r-checksum="%s">
%s
</div>`, name, id, escapedSnapshot, checksum, innerHTML)

	return root, nil
}

// ProcessRequest handles an action or input update request from the RikatLive client.
func ProcessRequest(req RequestPayload) (*ResponsePayload, error) {
	comp, err := Create(req.Component)
	if err != nil {
		return nil, err
	}

	// 1. Verify snapshot integrity (HMAC anti-tampering)
	snapshotJSON, err := json.Marshal(req.Snapshot)
	if err != nil {
		return nil, fmt.Errorf("rikatlive: invalid snapshot json: %w", err)
	}

	if !VerifySnapshot(snapshotJSON, req.Checksum) {
		return nil, errors.New("rikatlive: snapshot verification failed (tampered state)")
	}

	// 2. Restore state into component struct
	if err := RestoreState(comp, req.Snapshot); err != nil {
		return nil, err
	}

	// 3. Apply model input updates
	for field, val := range req.Updates {
		stateMap := map[string]any{field: val}
		_ = RestoreState(comp, stateMap)
		if updater, ok := comp.(Updater); ok {
			_ = updater.Updated(field, val)
		}
	}

	// 4. Call action method if requested
	if req.Action != "" {
		if dyn, ok := comp.(DynamicComponent); ok {
			if err := dyn.CallAction(req.Action, req.Args); err != nil {
				return nil, err
			}
		} else {
			compVal := reflect.ValueOf(comp)
			method := compVal.MethodByName(req.Action)
			if !method.IsValid() {
				// Try title-cased method name
				titleAction := strings.ToUpper(req.Action[:1]) + req.Action[1:]
				method = compVal.MethodByName(titleAction)
			}

			if !method.IsValid() {
				return nil, fmt.Errorf("rikatlive: method '%s' not found on component '%s'", req.Action, req.Component)
			}

			// Prepare method arguments
			methodType := method.Type()
			numArgs := methodType.NumIn()
			callArgs := make([]reflect.Value, 0, numArgs)

			for i := 0; i < numArgs; i++ {
				expectedType := methodType.In(i)
				if i < len(req.Args) {
					argVal := reflect.ValueOf(req.Args[i])
					if argVal.Type().ConvertibleTo(expectedType) {
						callArgs = append(callArgs, argVal.Convert(expectedType))
					} else {
						callArgs = append(callArgs, reflect.Zero(expectedType))
					}
				} else {
					callArgs = append(callArgs, reflect.Zero(expectedType))
				}
			}

			results := method.Call(callArgs)
			// If method returned error as last return value, check it
			if len(results) > 0 {
				lastRes := results[len(results)-1]
				if lastRes.Type().Implements(reflect.TypeOf((*error)(nil)).Elem()) && !lastRes.IsNil() {
					return nil, lastRes.Interface().(error)
				}
			}
		}
	}

	// 5. Extract fresh updated state
	newState := ExtractState(comp)
	newSnapshotJSON, _ := json.Marshal(newState)
	newChecksum := SignSnapshot(newSnapshotJSON)

	// 6. Render updated component HTML
	innerHTML := comp.Render()
	escapedSnapshot := html.EscapeString(string(newSnapshotJSON))

	wrappedHTML := fmt.Sprintf(`<div r-component="%s" r-id="%s" r-snapshot="%s" r-checksum="%s">
%s
</div>`, req.Component, req.ID, escapedSnapshot, newChecksum, innerHTML)

	return &ResponsePayload{
		ID:       req.ID,
		HTML:     wrappedHTML,
		Snapshot: newState,
		Checksum: newChecksum,
	}, nil
}

