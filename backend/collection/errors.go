package collection

import "errors"

var ErrUserDoesNotOwnCharacter = errors.New("user does not own character")

var ErrInvalidURL = errors.New("invalid URL")
var ErrInvalidAnilistURL = errors.New("invalid Anilist URL")
var ErrQuoteTooLong = errors.New("quote is too long")

var ErrNotInGuild = errors.New("this command can only be used in servers")
var ErrGuildNotIndexed = errors.New("guild members not indexed yet, please try again later")
var ErrCharacterNotHeld = errors.New("no one in this server has this character")

var ErrUserAlreadyOwnsCharacter = errors.New("user already owns this character")
