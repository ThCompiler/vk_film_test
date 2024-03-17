package types

import (
	"database/sql/driver"
	"github.com/pkg/errors"
)

type Id uint64

type Rating uint8

type Sexes string

func (s *Sexes) Scan(src any) error {
	if str, ok := src.(string); ok {
		*s = Sexes(str)
		return nil
	}

	if str, ok := src.([]byte); ok {
		*s = Sexes(str)
		return nil
	}

	return errors.Errorf("invalid type of data for Sexes %v", src)
}

func (s *Sexes) Value() (driver.Value, error) {
	return driver.Value(string(*s)), nil
}

const (
	MALE   Sexes = "male"
	FEMALE Sexes = "female"
)

type Order string

const (
	ASC  Order = "ASC"
	DESC Order = "DESC"
)

type OrderField string

const (
	RatingField      OrderField = "rating"
	NameField        OrderField = "name"
	DataPublishField OrderField = "publish_date"
)

type SearchField string

const (
	ActorField SearchField = "actor"
	FilmField  SearchField = "film"
)

type Roles string

const (
	ADMIN Roles = "admin"
	USER  Roles = "user"
)

type ContextField string
