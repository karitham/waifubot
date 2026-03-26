package collection

import "errors"

// ErrInsufficientTokens is returned when a user has insufficient tokens.
var ErrInsufficientTokens = errors.New("insufficient tokens")

// ErrInvalidAmount is returned when a transfer amount is invalid.
var ErrInvalidAmount = errors.New("amount must be positive")

// ErrSameUserTransfer is returned when transferring to yourself.
var ErrSameUserTransfer = errors.New("cannot transfer to yourself")

// ErrNoUnownedCharacters is returned when the user already owns all characters from the media.
var ErrNoUnownedCharacters = errors.New("no unowned characters remaining in this series")

// ErrMediaNotFound is returned when the media has no characters.
var ErrMediaNotFound = errors.New("no characters found for this series")
