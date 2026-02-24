package gitlab

import "strings"

type SearchQuery struct {
	Text      string
	Author    string
	Assignee  string
	Labels    []string
	Milestone string
}

func ParseSearchQuery(raw string) SearchQuery {
	tokens := splitSearchTokens(raw)
	query := SearchQuery{}
	textTerms := make([]string, 0, len(tokens))

	for i := 0; i < len(tokens); i++ {
		token := tokens[i]
		trimmed := strings.TrimSpace(token)
		if trimmed == "" {
			continue
		}
		key, value, allowExpand, ok := parseQualifierToken(trimmed)
		if !ok {
			textTerms = append(textTerms, trimmed)
			continue
		}
		v := value
		for allowExpand && i+1 < len(tokens) {
			next := strings.TrimSpace(tokens[i+1])
			if next == "" {
				i++
				continue
			}
			if _, _, _, isQualifier := parseQualifierToken(next); isQualifier {
				break
			}
			if strings.Contains(next, ":") {
				break
			}
			v = strings.TrimSpace(v + " " + next)
			i++
		}

		switch key {
		case "author":
			query.Author = v
		case "assignee":
			query.Assignee = v
		case "label":
			query.Labels = append(query.Labels, v)
		case "milestone":
			query.Milestone = v
		default:
			textTerms = append(textTerms, trimmed)
		}
	}

	query.Text = strings.TrimSpace(strings.Join(textTerms, " "))
	query.Labels = uniqueNonEmpty(query.Labels)
	return query
}

func parseQualifierToken(token string) (string, string, bool, bool) {
	key, value, ok := strings.Cut(token, ":")
	if !ok {
		return "", "", false, false
	}
	k := strings.ToLower(strings.TrimSpace(key))
	if k != "author" && k != "assignee" && k != "label" && k != "milestone" {
		return "", "", false, false
	}
	v := strings.TrimSpace(value)
	if v == "" {
		return "", "", false, false
	}
	if strings.HasPrefix(v, "\"") && strings.HasSuffix(v, "\"") && len(v) >= 2 {
		unquoted := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(v, "\""), "\""))
		if unquoted == "" {
			return "", "", false, false
		}
		return k, unquoted, false, true
	}
	return k, v, true, true
}

func splitSearchTokens(raw string) []string {
	input := strings.TrimSpace(raw)
	if input == "" {
		return nil
	}

	runes := []rune(input)
	tokens := make([]string, 0, 8)
	var current strings.Builder
	inQuote := false

	flush := func() {
		if current.Len() == 0 {
			return
		}
		tokens = append(tokens, current.String())
		current.Reset()
	}

	for _, r := range runes {
		switch {
		case r == '"':
			inQuote = !inQuote
			current.WriteRune(r)
		case !inQuote && (r == ' ' || r == '\t' || r == '\n'):
			flush()
		default:
			current.WriteRune(r)
		}
	}

	flush()
	return tokens
}

func uniqueNonEmpty(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(values))
	unique := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		unique = append(unique, trimmed)
	}
	if len(unique) == 0 {
		return nil
	}
	return unique
}
