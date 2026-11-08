package lang

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Translator manages multi-language translation dictionaries and string interpolations.
type Translator struct {
	locale         string
	fallbackLocale string
	dictionaries   map[string]map[string]string
	mu             sync.RWMutex
}

var (
	globalTranslator *Translator
	translatorOnce   sync.Once
)

// GetTranslator returns the singleton global translator instance.
func GetTranslator() *Translator {
	translatorOnce.Do(func() {
		globalTranslator = New("id", "en")
	})
	return globalTranslator
}

// New creates a new Translator instance with a primary and fallback locale.
func New(locale, fallback string) *Translator {
	if locale == "" {
		locale = "id"
	}
	if fallback == "" {
		fallback = "en"
	}
	return &Translator{
		locale:         locale,
		fallbackLocale: fallback,
		dictionaries:   make(map[string]map[string]string),
	}
}

// SetLocale sets the active locale.
func (t *Translator) SetLocale(locale string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.locale = strings.ToLower(strings.TrimSpace(locale))
}

// GetLocale returns the active locale.
func (t *Translator) GetLocale() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.locale
}

// SetFallbackLocale sets the fallback locale used when a key is missing in active locale.
func (t *Translator) SetFallbackLocale(fallback string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.fallbackLocale = strings.ToLower(strings.TrimSpace(fallback))
}

// GetFallbackLocale returns the fallback locale.
func (t *Translator) GetFallbackLocale() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.fallbackLocale
}

// AddMessages loads a flat map of messages for a given locale.
func (t *Translator) AddMessages(locale string, msgs map[string]string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	loc := strings.ToLower(strings.TrimSpace(locale))
	if _, ok := t.dictionaries[loc]; !ok {
		t.dictionaries[loc] = make(map[string]string)
	}
	for k, v := range msgs {
		t.dictionaries[loc][k] = v
	}
}

// LoadJSON parses a JSON string or byte slice into the dictionary for the specified locale.
func (t *Translator) LoadJSON(locale string, data []byte) error {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("gagal maca JSON kamus: %w", err)
	}

	flat := make(map[string]string)
	flattenDictionary("", raw, flat)
	t.AddMessages(locale, flat)
	return nil
}

// LoadFile reads a JSON dictionary file (e.g. "id.json", "auth.json").
// If locale is empty, it uses the base filename without extension as the locale.
func (t *Translator) LoadFile(filePath string, locale string) error {
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	if locale == "" {
		base := filepath.Base(filePath)
		locale = strings.TrimSuffix(base, filepath.Ext(base))
	}
	return t.LoadJSON(locale, bytes)
}

// LoadDir scans a directory for JSON dictionary files and registers them.
func (t *Translator) LoadDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			// Subdirectories can be locale names (e.g., kamus/id/messages.json)
			subDir := filepath.Join(dir, entry.Name())
			subEntries, err := os.ReadDir(subDir)
			if err == nil {
				for _, sub := range subEntries {
					if !sub.IsDir() && strings.HasSuffix(strings.ToLower(sub.Name()), ".json") {
						_ = t.LoadFile(filepath.Join(subDir, sub.Name()), entry.Name())
					}
				}
			}
			continue
		}

		if strings.HasSuffix(strings.ToLower(entry.Name()), ".json") {
			_ = t.LoadFile(filepath.Join(dir, entry.Name()), "")
		}
	}
	return nil
}

// Trans translates a message key using the active locale, with fallback to fallbackLocale.
// It replaces ":key" placeholders with values from params.
func (t *Translator) Trans(key string, params ...any) string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	loc := t.locale
	var template string
	var found bool

	// 1. Try active locale
	if dict, ok := t.dictionaries[loc]; ok {
		if val, exists := dict[key]; exists {
			template = val
			found = true
		}
	}

	// 2. Fallback locale
	if !found && t.fallbackLocale != "" && t.fallbackLocale != loc {
		if dict, ok := t.dictionaries[t.fallbackLocale]; ok {
			if val, exists := dict[key]; exists {
				template = val
				found = true
			}
		}
	}

	// 3. Not found: return raw key
	if !found {
		template = key
	}

	if len(params) == 0 {
		return template
	}

	return interpolate(template, params...)
}

// Get is an alias for Trans.
func (t *Translator) Get(key string, params ...any) string {
	return t.Trans(key, params...)
}

// Terjemah is an alias for Trans.
func (t *Translator) Terjemah(key string, params ...any) string {
	return t.Trans(key, params...)
}

// Jupuk is an alias for Trans.
func (t *Translator) Jupuk(key string, params ...any) string {
	return t.Trans(key, params...)
}

// Helper to flatten nested maps into dot notation keys
func flattenDictionary(prefix string, src map[string]any, dest map[string]string) {
	for k, v := range src {
		fullKey := k
		if prefix != "" {
			fullKey = prefix + "." + k
		}
		switch val := v.(type) {
		case string:
			dest[fullKey] = val
		case map[string]any:
			flattenDictionary(fullKey, val, dest)
		default:
			dest[fullKey] = fmt.Sprintf("%v", val)
		}
	}
}

// Helper to interpolate :key or {key} placeholders
func interpolate(msg string, params ...any) string {
	if len(params) == 0 {
		return msg
	}

	// Case 1: First param is map[string]any or map[string]string
	if m, ok := params[0].(map[string]any); ok {
		for k, v := range m {
			msg = replacePlaceholder(msg, k, fmt.Sprintf("%v", v))
		}
		return msg
	}
	if m, ok := params[0].(map[string]string); ok {
		for k, v := range m {
			msg = replacePlaceholder(msg, k, v)
		}
		return msg
	}

	// Case 2: Key-value pair arguments: Trans("key", "name", "Budi", "age", 25)
	if len(params)%2 == 0 {
		for i := 0; i < len(params); i += 2 {
			k := fmt.Sprintf("%v", params[i])
			v := fmt.Sprintf("%v", params[i+1])
			msg = replacePlaceholder(msg, k, v)
		}
		return msg
	}

	// Case 3: Positional replacement: :1, :2, etc.
	for i, p := range params {
		msg = replacePlaceholder(msg, fmt.Sprintf("%d", i+1), fmt.Sprintf("%v", p))
	}

	return msg
}

func replacePlaceholder(msg, key, val string) string {
	msg = strings.ReplaceAll(msg, ":"+key, val)
	msg = strings.ReplaceAll(msg, "{"+key+"}", val)
	return msg
}

// Global convenience wrappers
func Trans(key string, params ...any) string {
	return GetTranslator().Trans(key, params...)
}

func SetLocale(locale string) {
	GetTranslator().SetLocale(locale)
}

func GetLocale() string {
	return GetTranslator().GetLocale()
}
