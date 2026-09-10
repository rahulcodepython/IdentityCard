package utils

import (
    "github.com/google/uuid"
    "github.com/gosimple/slug"
)

// Slugify generates a URL-safe slug from a string using Unicode-aware
// transliteration — a non-Latin-script title still produces a real slug
// instead of collapsing to nothing — with a short random suffix for
// uniqueness instead of a 19-digit nanosecond timestamp.
func Slugify(title string) string {
    s := slug.Make(title)
    if s == "" {
        s = "item"
    }
    return s + "-" + uuid.NewString()[:8]
}
