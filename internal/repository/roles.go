package repository

import (
	"context"
	"database/sql"
	"fmt"
)

type Role struct {
	Id          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Level       int    `json:"level"`
}

type RoleRepo struct {
	db *sql.DB
}

func (s *RoleRepo) GetByName(ctx context.Context, slug string) (*Role, error) {
	query := `SELECT id,name,description,level FROM roles WHERE name =$1`

	role := &Role{}
	err := s.db.QueryRowContext(ctx, query, slug).Scan(&role.Id, &role.Name, &role.Description, &role.Level)

	if err != nil {
		return nil, err
	}

	fmt.Printf(role.Name)

	return role, nil
}
