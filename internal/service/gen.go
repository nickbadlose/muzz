package service

// generates mock repositories.

// nolint:lll // these are commands.
//go:generate go run github.com/vektra/mockery/v2 --with-expecter --name=UserRepository --packageprefix=mock
//go:generate go run github.com/vektra/mockery/v2 --with-expecter --name=MatchRepository --packageprefix=mock
//go:generate go run github.com/vektra/mockery/v2 --with-expecter --name=Authenticator --packageprefix=mock
