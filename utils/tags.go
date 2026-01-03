package utils

import (
	"strings"
	"encoding/json"
)

func ParseTags(tagStr string) []string {
    if tagStr == "" {
        return []string{}
    }

    if strings.HasPrefix(tagStr, "[") && strings.HasSuffix(tagStr, "]") {
        var tags []string
        if err := json.Unmarshal([]byte(tagStr), &tags); err == nil {
            var cleanedTags []string
            for _, tag := range tags {
                trimmedTag := strings.TrimSpace(tag)
                if trimmedTag != "" {
                    cleanedTags = append(cleanedTags, trimmedTag)
                }
            }
            return cleanedTags
        }
    }

    tagStr = strings.ReplaceAll(tagStr, "，", ",")

    rawTags := strings.Split(tagStr, ",")
    var cleanedTags []string

    for _, tag := range rawTags {
        trimmedTag := strings.TrimSpace(tag)
        if trimmedTag != "" {
            cleanedTags = append(cleanedTags, trimmedTag)
        }
    }

    return cleanedTags
}

