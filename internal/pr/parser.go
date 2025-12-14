package pr

import (
	"errors"
	"net/url"
	"strconv"
	"strings"
)

type PRInfo struct {
	Owner  string
	Repo   string
	Number int
}

func ParsePRURL(raw string) (PRInfo, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return PRInfo{}, err
	}

	parts := strings.Split(u.Path, "/")
	// /org/repo/pull/123
	if len(parts) < 5 {
		return PRInfo{}, errors.New("invalid PR url")
	}

	num, err := strconv.Atoi(parts[4])
	if err != nil {
		return PRInfo{}, err
	}

	return PRInfo{
		Owner:  parts[1],
		Repo:   parts[2],
		Number: num,
	}, nil
}
