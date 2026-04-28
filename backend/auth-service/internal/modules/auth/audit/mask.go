package audit

import (
	"net"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/ports"
)

const userAgentMaxRunes = 200

// MaskEvent returns a shallow copy safe for persistence / outbox (PII reduced).
func MaskEvent(e ports.AuditEvent) ports.AuditEvent {
	out := e
	if out.Meta != nil {
		out.Meta = cloneMeta(out.Meta)
		maskMetaMap(out.Meta)
	} else {
		out.Meta = map[string]any{}
	}
	if out.IP != nil && *out.IP != "" {
		hint := maskIPString(*out.IP)
		if hint != "" {
			out.Meta["ip_class"] = hint
		}
		out.IP = nil
	}
	if out.UserAgent != nil && *out.UserAgent != "" {
		u := truncateRunes(*out.UserAgent, userAgentMaxRunes)
		out.UserAgent = &u
	}
	return out
}

func cloneMeta(m map[string]any) map[string]any {
	cp := make(map[string]any, len(m)+2)
	for k, v := range m {
		cp[k] = v
	}
	return cp
}

var sensitiveMetaKeys = map[string]struct{}{
	"email": {}, "mail": {}, "phone": {}, "password": {}, "token": {},
	"access_token": {}, "refresh_token": {}, "authorization": {}, "cookie": {},
	"secret": {}, "hash": {}, "otp": {}, "code": {},
}

func maskMetaMap(m map[string]any) {
	for k, v := range m {
		lk := strings.ToLower(k)
		if _, ok := sensitiveMetaKeys[lk]; ok {
			m[k] = "[redacted]"
			continue
		}
		if nested, ok := v.(map[string]any); ok {
			maskMetaMap(nested)
		}
	}
	// Shorten obvious UUID session references for correlation without full id.
	if sid, ok := m["session_id"].(string); ok && len(sid) == 36 {
		if _, err := uuid.Parse(sid); err == nil {
			m["session_id"] = sid[:8] + "…"
		}
	}
}

func maskIPString(s string) string {
	ip := net.ParseIP(strings.TrimSpace(s))
	if ip == nil {
		return ""
	}
	if v4 := ip.To4(); v4 != nil {
		return strconv.Itoa(int(v4[0])) + "." + strconv.Itoa(int(v4[1])) + ".*.*"
	}
	// IPv6: keep first hextet only
	parts := strings.Split(ip.String(), ":")
	if len(parts) > 0 && parts[0] != "" {
		return parts[0] + ":…"
	}
	return ""
}

func truncateRunes(s string, max int) string {
	if max <= 0 {
		return ""
	}
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	var b strings.Builder
	n := 0
	for _, r := range s {
		if n >= max-1 {
			b.WriteRune('…')
			break
		}
		b.WriteRune(r)
		n++
	}
	return b.String()
}
