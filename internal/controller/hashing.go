package controller

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	pbv1 "github.com/nicjohnson145/plantr/gen/plantr/controller/v1"
)

func baseHash(parts []string) string {
	h := md5.Sum([]byte(strings.Join(parts, ""))) //nolint: gosec // its a hash, it doesnt have to be cryptographically secure
	return hex.EncodeToString(h[:])
}

func ComputeHash(s *pbv1.Seed) (string, error) {
	digest := sha256.New()
	// hash the object, ignoring the metadata portion
	s.HashPB(
		digest,
		map[string]struct{}{
			"plantr.controller.v1.Seed.metadata": {},
		},
	)

	// add our type to the mix
	_, err := fmt.Fprintf(digest, "%T", s.Element)
	if err != nil {
		return "", fmt.Errorf("error adding to hash: %w", err)
	}

	// encode as hex and return
	return fmt.Sprintf("%x\n", digest.Sum(nil)), nil
}
