package domain

import "errors"

var (
	ErrInvalidID           = errors.New("invalid id (must be uuid string)")
	ErrPRMerged            = errors.New("pr is merged")
	ErrReviewerNotAssigned = errors.New("reviewer is not assigned")
	ErrNoCandidate         = errors.New("no replacement candidate available")
)
