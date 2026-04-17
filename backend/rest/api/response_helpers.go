package api

// Error marker methods for response interfaces.
// These allow api.Error to be used as error responses for all operations.

func (*Error) updateProfileRes()            {}
func (*Error) updateFavoriteRes()           {}
func (*Error) addWishlistCharactersRes()    {}
func (*Error) removeWishlistCharactersRes() {}
func (*Error) clearWishlistRes()            {}
func (*Error) addWishlistMediaRes()         {}
