package repository

import sq "github.com/Masterminds/squirrel"

func Builder() sq.StatementBuilderType {
	return sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
}
