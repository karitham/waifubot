package collection

import "errors"

// ErrInsufficientTokens is returned when a user has insufficient tokens.
var ErrInsufficientTokens = errors.New("insufficient tokens")

// ErrInvalidUserID is returned when an operation names user 0.
var ErrInvalidUserID = errors.New("invalid user id")

// ErrInvalidAmount is returned when a token amount is not positive.
var ErrInvalidAmount = errors.New("amount must be positive")

// ErrSameUserTransfer is returned when transferring to yourself.
var ErrSameUserTransfer = errors.New("cannot transfer to yourself")

// ErrNoUnownedCharacters is returned when the user already owns all characters from the media.
var ErrNoUnownedCharacters = errors.New("no unowned characters remaining")

// ErrMediaNotFound is returned when the media has no characters.
var ErrMediaNotFound = errors.New("no characters found for this series")

// ErrCharactersNotSynced is returned when none of a series' unowned
// characters are in the catalog yet.
var ErrCharactersNotSynced = errors.New("characters not in the catalog yet")
