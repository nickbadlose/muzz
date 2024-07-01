package store

// TODO docs for generating mocks

// generates mocks for both our internal database interfaces and also the required upper.io interfaces.
// requires mockgen to be installed for this to work
//
//  go get -u github.com/golang/mock/mockgen@latest
//
// see: https://github.com/golang/mock for more information.

// nolint:lll // these are commands.
//go:generate go run github.com/vektra/mockery/v2 --with-expecter --name=Store --packageprefix=mock
