package organizations

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

var slugInvalidChars = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(name string) string {
	slug := strings.Trim(slugInvalidChars.ReplaceAllString(strings.ToLower(name), "-"), "-")
	if slug == "" {
		slug = "org"
	}
	return slug
}

// GenerateUniqueSlug slugifies name and appends "-2", "-3", ... until it
// finds one not already taken. repo should be scoped to the same
// transaction as the eventual insert, otherwise two concurrent signups
// could still race between this check and the create — the create's own
// unique constraint is the real guarantee, this just avoids the common case.
func GenerateUniqueSlug(ctx context.Context, repo *Repository, name string) (string, error) {
	base := slugify(name)
	slug := base
	for suffix := 2; ; suffix++ {
		if _, err := repo.GetBySlug(ctx, slug); err != nil {
			return slug, nil
		}
		slug = fmt.Sprintf("%s-%d", base, suffix)
	}
}
